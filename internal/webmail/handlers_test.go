package webmail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/session"
)

func TestMessagesEndpointRequiresAuth(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestIndexDisablesBrowserCaching(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestMessagesEndpointReturnsRecentMessages(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?limit=1", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var messages []MessageSummary
	if err := json.NewDecoder(rec.Body).Decode(&messages); err != nil {
		t.Fatalf("decode messages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(messages))
	}
	if messages[0].Subject != "Welcome" {
		t.Fatalf("subject = %q, want Welcome", messages[0].Subject)
	}
	if store.folder != "inbox" {
		t.Fatalf("folder = %q, want inbox", store.folder)
	}
}

func TestMessagesEndpointSupportsSpamFolder(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?folder=spam", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.folder != "spam" {
		t.Fatalf("folder = %q, want spam", store.folder)
	}
}

func TestMailboxesEndpointReturnsSharedMailboxes(t *testing.T) {
	store := &fakeStore{valid: true, mailboxes: []MailboxAccount{
		{ID: "marko@example.com", Name: "Marko", Address: "marko@example.com", Kind: "personal", CanRead: true, CanManage: true},
		{ID: "support@example.com", Name: "Support", Address: "support@example.com", Kind: "shared", CanRead: true, CanSendAs: true},
	}}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/mailboxes", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var mailboxes []MailboxAccount
	if err := json.NewDecoder(rec.Body).Decode(&mailboxes); err != nil {
		t.Fatalf("decode mailboxes: %v", err)
	}
	if len(mailboxes) != 2 || mailboxes[1].Address != "support@example.com" || mailboxes[1].Kind != "shared" {
		t.Fatalf("mailboxes = %+v, want personal and shared", mailboxes)
	}
}

func TestMessagesEndpointUsesSelectedSharedMailbox(t *testing.T) {
	store := &fakeStore{valid: true, mailboxes: []MailboxAccount{
		{ID: "marko@example.com", Name: "Marko", Address: "marko@example.com", Kind: "personal", CanRead: true, CanManage: true},
		{ID: "support@example.com", Name: "Support", Address: "support@example.com", Kind: "shared", CanRead: true},
	}}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?mailbox=support@example.com", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.listEmail != "support@example.com" {
		t.Fatalf("listEmail = %q, want shared mailbox", store.listEmail)
	}
}

func TestMessagesEndpointRejectsUnsharedMailbox(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages?mailbox=other@example.com", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestMessageEndpointReturnsFullMessage(t *testing.T) {
	handler := NewRouter(&fakeStore{valid: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages/1", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var message MessageDetail
	if err := json.NewDecoder(rec.Body).Decode(&message); err != nil {
		t.Fatalf("decode message: %v", err)
	}
	if message.ID != "1" || message.Body != "Full body" {
		t.Fatalf("unexpected message: %+v", message)
	}
}

func TestContentTrustEndpointStoresTrustPerAuthenticatedUser(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/content-trust", strings.NewReader(`{"scope":"sender","value":"sender@example.net"}`))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.trustEmail != "marko@example.com" {
		t.Fatalf("trustEmail = %q, want authenticated user", store.trustEmail)
	}
	if store.trustEntry.Scope != "sender" || store.trustEntry.Value != "sender@example.net" {
		t.Fatalf("trust entry = %+v, want sender trust", store.trustEntry)
	}
	if !containsString(store.auditActions, "webmail.content_trust_add") {
		t.Fatalf("content trust audit event missing: %+v", store.auditActions)
	}
}

func TestContentTrustEndpointListsTrustForAuthenticatedUser(t *testing.T) {
	store := &fakeStore{
		valid: true,
		trustEntries: []ContentTrustEntry{
			{Scope: "sender", Value: "sender@example.net"},
			{Scope: "domain", Value: "example.net"},
		},
	}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/content-trust", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.listTrustEmail != "marko@example.com" {
		t.Fatalf("listTrustEmail = %q, want authenticated user", store.listTrustEmail)
	}
	var entries []ContentTrustEntry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode trust entries: %v", err)
	}
	if len(entries) != 2 || entries[1].Value != "example.net" {
		t.Fatalf("entries = %+v, want per-user content trust", entries)
	}
}

func TestContentTrustEndpointRejectsPublicProviderDomainTrust(t *testing.T) {
	for _, domain := range []string{"gmail.com", "outlook.com", "outlook.xyz", "azet.sk", "mail.google.com"} {
		t.Run(domain, func(t *testing.T) {
			store := &fakeStore{valid: true}
			handler := NewRouter(store)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/content-trust", strings.NewReader(`{"scope":"domain","value":"`+domain+`"}`))
			req.SetBasicAuth("marko@example.com", "secret123456")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if store.trustEmail != "" {
				t.Fatalf("public provider domain was stored for %q", store.trustEmail)
			}
			if !strings.Contains(strings.ToLower(rec.Body.String()), "public") {
				t.Fatalf("body = %q, want public provider explanation", rec.Body.String())
			}
		})
	}
}

func TestSendEndpointUsesAuthenticatedSender(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"to":["marko@example.com"],"subject":"Hello","body":"Sent from webmail"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/send", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.sent.From != "marko@example.com" {
		t.Fatalf("from = %q, want authenticated user", store.sent.From)
	}
	if store.sent.Subject != "Hello" || store.sent.Body != "Sent from webmail" {
		t.Fatalf("unexpected sent message: %+v", store.sent)
	}
}

func TestSendEndpointRecordsBulkSendSecurityAlert(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	recipients := make([]string, 0, bulkSendAlertRecipientThreshold)
	for i := 0; i < bulkSendAlertRecipientThreshold; i++ {
		recipients = append(recipients, fmt.Sprintf("person%d@example.net", i))
	}
	bodyBytes, err := json.Marshal(map[string]any{
		"to":      recipients,
		"subject": "Team update",
		"body":    "Hello all",
	})
	if err != nil {
		t.Fatalf("marshal send body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/send", bytes.NewReader(bodyBytes))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if !containsString(store.auditActions, "security.alert.bulk_send") {
		t.Fatalf("bulk send security alert missing from audit actions: %+v", store.auditActions)
	}
	if !containsString(store.auditActions, "message.send") {
		t.Fatalf("normal send audit missing from audit actions: %+v", store.auditActions)
	}
}

func TestSendEndpointAllowsSharedMailboxSenderWithPermission(t *testing.T) {
	store := &fakeStore{
		valid: true,
		mailboxes: []MailboxAccount{
			{ID: "marko@example.com", Name: "Marko", Address: "marko@example.com", Kind: "personal", CanRead: true, CanManage: true, CanSendAs: true},
			{ID: "support@example.com", Name: "Support", Address: "support@example.com", Kind: "shared", CanRead: true, CanSendAs: true},
		},
	}
	handler := NewRouter(store)
	body := `{"from":"support@example.com","to":["customer@example.net"],"subject":"Hello","body":"Sent from support"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/send", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.sent.From != "support@example.com" {
		t.Fatalf("from = %q, want shared mailbox sender", store.sent.From)
	}
}

func TestSendEndpointRejectsSharedMailboxSenderWithoutPermission(t *testing.T) {
	store := &fakeStore{
		valid: true,
		mailboxes: []MailboxAccount{
			{ID: "marko@example.com", Name: "Marko", Address: "marko@example.com", Kind: "personal", CanRead: true, CanManage: true, CanSendAs: true},
			{ID: "support@example.com", Name: "Support", Address: "support@example.com", Kind: "shared", CanRead: true, CanSendAs: false},
		},
	}
	handler := NewRouter(store)
	body := `{"from":"support@example.com","to":["customer@example.net"],"subject":"Hello","body":"Sent from support"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/send", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestSendEndpointAcceptsMultipartAttachments(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("from", "marko@example.com")
	_ = writer.WriteField("to", "customer@example.net")
	_ = writer.WriteField("subject", "Contract")
	_ = writer.WriteField("body", "See attached")
	_ = writer.WriteField("body_html", "<p>See <strong>attached</strong></p>")
	part, err := writer.CreateFormFile("attachments", "contract.txt")
	if err != nil {
		t.Fatalf("create attachment: %v", err)
	}
	if _, err := part.Write([]byte("signed")); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	part, err = writer.CreateFormFile("attachments", "invoice.pdf")
	if err != nil {
		t.Fatalf("create second attachment: %v", err)
	}
	if _, err := part.Write([]byte("%PDF invoice")); err != nil {
		t.Fatalf("write second attachment: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/send", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.sent.BodyHTML != "<p>See <strong>attached</strong></p>" {
		t.Fatalf("html body = %q", store.sent.BodyHTML)
	}
	if len(store.sent.Attachments) != 2 {
		t.Fatalf("attachments = %+v, want two", store.sent.Attachments)
	}
	attachment := store.sent.Attachments[0]
	if attachment.Filename != "contract.txt" || string(attachment.Data) != "signed" || attachment.ContentType == "" {
		t.Fatalf("unexpected attachment: %+v", attachment)
	}
	second := store.sent.Attachments[1]
	if second.Filename != "invoice.pdf" || string(second.Data) != "%PDF invoice" || second.ContentType == "" {
		t.Fatalf("unexpected second attachment: %+v", second)
	}
}

func TestAttachmentMetadataIsSanitized(t *testing.T) {
	if got := sanitizeAttachmentFilename(`..\evil` + string(rune(0x202e)) + `.gpj.exe`); got != "evil.gpj.exe" {
		t.Fatalf("filename = %q, want sanitized base name without bidi control", got)
	}
	if got := sanitizeAttachmentContentType("text/plain\r\nX-Evil: yes", []byte("%PDF-1.7")); got != "application/pdf" {
		t.Fatalf("content type = %q, want detected safe type", got)
	}
	if got := sanitizeAttachmentContentType("application/x-custom+json; charset=utf-8", nil); got != "application/x-custom+json" {
		t.Fatalf("content type = %q, want normalized media type", got)
	}
}

func TestDraftEndpointSavesAuthenticatedDraft(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"to":["ada@example.net"],"subject":"Draft subject","body":"Draft body","body_html":"<p>Draft body</p>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.savedDraft.From != "marko@example.com" || store.savedDraft.Subject != "Draft subject" || store.savedDraft.BodyHTML != "<p>Draft body</p>" {
		t.Fatalf("saved draft = %+v", store.savedDraft)
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != "draft-1" {
		t.Fatalf("draft id = %q, want draft-1", response.ID)
	}
}

func TestSendEndpointDeletesServerDraftAfterSending(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"to":["ada@example.net"],"subject":"Ready","body":"Send it","draft_id":"draft-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/send", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.deletedMessageEmail != "marko@example.com" || store.deletedMessageID != "draft-1" {
		t.Fatalf("deleted draft = %q/%q, want marko@example.com/draft-1", store.deletedMessageEmail, store.deletedMessageID)
	}
}

func TestWebmailSessionLoginAndCSRF(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{valid: true}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	login.Header.Set("User-Agent", "Browser A")
	login.Header.Set("Accept-Language", "en-US")
	loginRec := httptest.NewRecorder()

	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	var loginResponse struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResponse); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResponse.CSRFToken == "" {
		t.Fatal("csrf token is empty")
	}
	cookie := loginRec.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodPost, "/api/v1/send", strings.NewReader(`{"to":["marko@example.com"],"subject":"Hello","body":"Body"}`))
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("Accept-Language", "en-US")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("missing csrf status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/send", strings.NewReader(`{"to":["marko@example.com"],"subject":"Hello","body":"Body"}`))
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("X-CSRF-Token", loginResponse.CSRFToken)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("session send status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.sent.From != "marko@example.com" {
		t.Fatalf("from = %q, want session subject", store.sent.From)
	}

	current := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	current.Header.Set("User-Agent", "Browser A")
	current.Header.Set("Accept-Language", "en-US")
	current.AddCookie(cookie)
	currentRec := httptest.NewRecorder()
	handler.ServeHTTP(currentRec, current)
	if currentRec.Code != http.StatusOK {
		t.Fatalf("current session status = %d, want %d, body %s", currentRec.Code, http.StatusOK, currentRec.Body.String())
	}
	var currentResponse struct {
		CSRFToken string `json:"csrf_token"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(currentRec.Body).Decode(&currentResponse); err != nil {
		t.Fatalf("decode current session: %v", err)
	}
	if currentResponse.CSRFToken != loginResponse.CSRFToken || currentResponse.Email != "marko@example.com" {
		t.Fatalf("unexpected current session: %+v", currentResponse)
	}
}

func TestProfileEndpointReadsAndUpdatesSettings(t *testing.T) {
	store := &fakeStore{
		valid: true,
		profile: UserProfile{
			Email:            "marko@example.com",
			FirstName:        "Marko",
			LastName:         "Tester",
			DisplayName:      "Marko Tester",
			SignatureHTML:    "<p>Regards</p>",
			SignatureAutoAdd: true,
			Language:         "en",
		},
	}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Marko Tester") || !strings.Contains(rec.Body.String(), "Regards") || !strings.Contains(rec.Body.String(), `"language":"en"`) {
		t.Fatalf("profile response missing values: %s", rec.Body.String())
	}

	update := `{"first_name":"Ada","last_name":"Lovelace","display_name":"Ada Lovelace","signature_html":"<p>Ada</p>","signature_auto_add":false,"language":"sk"}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/profile", strings.NewReader(update))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.updatedProfile.FirstName != "" || store.updatedProfile.LastName != "" || store.updatedProfile.DisplayName != "" {
		t.Fatalf("profile update should not accept user-owned identity fields: %+v", store.updatedProfile)
	}
	if store.updatedProfile.SignatureHTML != "<p>Ada</p>" || store.updatedProfile.SignatureAutoAdd || store.updatedProfile.Language != "sk" {
		t.Fatalf("unexpected updated profile: %+v", store.updatedProfile)
	}
	if store.profile.DisplayName != "Marko Tester" {
		t.Fatalf("admin-owned display name changed: %+v", store.profile)
	}
	if !containsString(store.auditActions, "webmail.profile_update") {
		t.Fatalf("profile update audit event missing: %+v", store.auditActions)
	}
}

func TestProfileEndpointRejectsUnsupportedLanguage(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/profile", strings.NewReader(`{"signature_html":"","signature_auto_add":false,"language":"xx"}`))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestSessionLoginAndLogoutRecordAuditEvents(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{valid: true}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	login.Header.Set("User-Agent", "Browser A")
	login.Header.Set("Accept-Language", "en-US")
	loginRec := httptest.NewRecorder()

	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	if !containsString(store.auditActions, "webmail.login") {
		t.Fatalf("webmail login audit event missing: %+v", store.auditActions)
	}

	logout := httptest.NewRequest(http.MethodDelete, "/api/v1/session", nil)
	logout.Header.Set("User-Agent", "Browser A")
	logout.Header.Set("Accept-Language", "en-US")
	for _, cookie := range loginRec.Result().Cookies() {
		logout.AddCookie(cookie)
	}
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logout)

	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d, body %s", logoutRec.Code, http.StatusNoContent, logoutRec.Body.String())
	}
	if !containsString(store.auditActions, "webmail.logout") {
		t.Fatalf("webmail logout audit event missing: %+v", store.auditActions)
	}
}

func TestFailedSessionLoginRecordsAuditEvent(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{valid: false}
	handler := NewRouter(store, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if !containsString(store.auditActions, "webmail.login_failed") {
		t.Fatalf("webmail failed login audit event missing: %+v", store.auditActions)
	}
}

func TestSessionLoginRequiresMailboxTOTPWhenEnabled(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{valid: true, mailboxSecurity: MailboxSecurity{MFAEnabled: true, TOTPEnabled: true}}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, login)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if cookies := rec.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("password step must not create session before MFA, got %d cookie(s)", len(cookies))
	}
	var first struct {
		MFARequired bool   `json:"mfa_required"`
		Provider    string `json:"provider"`
		MFAToken    string `json:"mfa_token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&first); err != nil {
		t.Fatalf("decode password step: %v", err)
	}
	if !first.MFARequired || first.Provider != "totp" || first.MFAToken == "" {
		t.Fatalf("unexpected MFA step: %+v", first)
	}

	verify := httptest.NewRequest(http.MethodPost, "/api/v1/session/mfa", strings.NewReader(`{"mfa_token":"`+first.MFAToken+`","code":"123456"}`))
	verify.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("mfa verify status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if cookies := verifyRec.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("MFA verification did not create webmail session cookie")
	}
}

func TestWebmailLoginUsesProIdentityPushWhenConfigured(t *testing.T) {
	var requestCreated bool
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/auth-requests":
			requestCreated = true
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode proidentity request: %v", err)
			}
			if body["user_email"] != "marko@example.com" || body["context_title"] == "" || body["client_ip"] != "203.0.113.123" {
				t.Fatalf("unexpected proidentity request body: %+v", body)
			}
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "mail-push-1", "expires_at": time.Now().Add(time.Minute).Unix(), "status": "pending"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/sp/auth-requests/mail-push-1":
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "mail-push-1", "status": "approved", "totp_verified": true, "responded_at": time.Now().Unix()})
		default:
			http.NotFound(w, r)
		}
	}))
	defer authServer.Close()

	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{
		valid:           true,
		mailboxSecurity: MailboxSecurity{MFAAvailable: true},
		proIdentitySettings: domain.AdminMFASettings{
			ProIdentityEnabled:        true,
			ProIdentityBaseURL:        authServer.URL,
			ProIdentityAPIKey:         "sp-secret",
			ProIdentityTimeoutSeconds: 60,
		},
	}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	login.RemoteAddr = "127.0.0.1:38123"
	login.Header.Set("X-Real-IP", "203.0.113.123")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	if !requestCreated {
		t.Fatal("proidentity auth request was not created")
	}
	if cookies := loginRec.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("password step must not create session before push approval, got %d cookie(s)", len(cookies))
	}
	var first struct {
		MFARequired bool   `json:"mfa_required"`
		Provider    string `json:"provider"`
		MFAToken    string `json:"mfa_token"`
		RequestID   string `json:"request_id"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&first); err != nil {
		t.Fatalf("decode password step: %v", err)
	}
	if !first.MFARequired || first.Provider != "proidentity" || first.MFAToken == "" || first.RequestID != "mail-push-1" {
		t.Fatalf("unexpected password step response: %+v", first)
	}

	verify := httptest.NewRequest(http.MethodPost, "/api/v1/session/mfa", strings.NewReader(`{"mfa_token":"`+first.MFAToken+`"}`))
	verify.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("mfa status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if cookies := verifyRec.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("proidentity approval did not create webmail session cookie")
	}
}

func TestWebmailLoginLimiterKeysUseTrustedProxyClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", nil)
	req.RemoteAddr = "127.0.0.1:38123"
	req.Header.Set("X-Real-IP", "203.0.113.123")

	keys := loginKeys("webmail", "Tester@Example.COM", req)
	if !slices.Contains(keys, "webmail|ip|203.0.113.123") {
		t.Fatalf("login limiter keys should use real client IP, got %+v", keys)
	}
	if slices.Contains(keys, "webmail|ip|127.0.0.1") {
		t.Fatalf("login limiter keys still use loopback proxy IP: %+v", keys)
	}
}

func TestWebmailLoginCanVerifyProIdentityHostedTOTP(t *testing.T) {
	var statusRead bool
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/auth-requests":
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "mail-push-2", "expires_at": time.Now().Add(time.Minute).Unix(), "status": "pending"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/verify-totp":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode verify totp request: %v", err)
			}
			if body["user_email"] != "marko@example.com" || body["code"] != "123456" {
				t.Fatalf("unexpected verify totp body: %+v", body)
			}
			writeJSON(w, http.StatusOK, map[string]any{"verified": true, "user_email": "marko@example.com"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/sp/auth-requests/mail-push-2":
			statusRead = true
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "mail-push-2", "status": "pending"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer authServer.Close()

	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{
		valid:           true,
		mailboxSecurity: MailboxSecurity{MFAAvailable: true},
		proIdentitySettings: domain.AdminMFASettings{
			ProIdentityEnabled:        true,
			ProIdentityBaseURL:        authServer.URL,
			ProIdentityAPIKey:         "sp-secret",
			ProIdentityTimeoutSeconds: 60,
		},
	}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)
	var first struct {
		MFAToken string `json:"mfa_token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&first); err != nil {
		t.Fatalf("decode password step: %v", err)
	}

	verify := httptest.NewRequest(http.MethodPost, "/api/v1/session/mfa", strings.NewReader(`{"mfa_token":"`+first.MFAToken+`","code":"123456"}`))
	verify.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("mfa status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if statusRead {
		t.Fatal("hosted TOTP verification should not poll push status")
	}
	if cookies := verifyRec.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("proidentity hosted totp did not create webmail session cookie")
	}
}

func TestWebmailLoginReportsProIdentityAutocreateSetupNeeded(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/sp/auth-requests" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"code":       "user_created_needs_setup",
			"user_email": "marko@example.com",
		})
	}))
	defer authServer.Close()

	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{
		valid:           true,
		mailboxSecurity: MailboxSecurity{MFAAvailable: true},
		proIdentitySettings: domain.AdminMFASettings{
			ProIdentityEnabled:        true,
			ProIdentityBaseURL:        authServer.URL,
			ProIdentityAPIKey:         "sp-secret",
			ProIdentityTimeoutSeconds: 60,
		},
	}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusPreconditionRequired {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusPreconditionRequired, loginRec.Body.String())
	}
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != "user_created_needs_setup" || !strings.Contains(body.Error, "mobile app setup") {
		t.Fatalf("unexpected setup-needed response: %+v", body)
	}
	if cookies := loginRec.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("setup-needed response should not create a session, got %d cookie(s)", len(cookies))
	}
}

func TestSessionLoginRequiresMailboxMFASetupWhenForced(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{valid: true, mailboxSecurity: MailboxSecurity{MFAAvailable: true, ForceMFA: true}}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, login)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var first struct {
		SetupRequired bool   `json:"mfa_setup_required"`
		MFAToken      string `json:"mfa_token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&first); err != nil {
		t.Fatalf("decode setup step: %v", err)
	}
	if !first.SetupRequired || first.MFAToken == "" {
		t.Fatalf("unexpected setup step: %+v", first)
	}

	enroll := httptest.NewRequest(http.MethodPost, "/api/v1/mfa/totp/enroll", strings.NewReader(`{"mfa_token":"`+first.MFAToken+`"}`))
	enroll.Header.Set("Content-Type", "application/json")
	enrollRec := httptest.NewRecorder()
	handler.ServeHTTP(enrollRec, enroll)
	if enrollRec.Code != http.StatusOK {
		t.Fatalf("enroll status = %d, want %d, body %s", enrollRec.Code, http.StatusOK, enrollRec.Body.String())
	}

	verify := httptest.NewRequest(http.MethodPost, "/api/v1/mfa/totp/verify", strings.NewReader(`{"mfa_token":"`+first.MFAToken+`","code":"123456"}`))
	verify.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("setup verify status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if !store.mailboxSecurity.MFAEnabled || !store.mailboxSecurity.TOTPEnabled {
		t.Fatalf("MFA was not enabled after setup: %+v", store.mailboxSecurity)
	}
	if cookies := verifyRec.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("setup verification did not create webmail session cookie")
	}
}

func TestAppPasswordLifecycle(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	store := &fakeStore{valid: true}
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)

	create := httptest.NewRequest(http.MethodPost, "/api/v1/app-passwords", strings.NewReader(`{"name":"Thunderbird laptop","protocols":["imap","smtp","dav"]}`))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("X-CSRF-Token", csrfFromBody(t, loginRec.Body.Bytes()))
	for _, cookie := range loginRec.Result().Cookies() {
		create.AddCookie(cookie)
	}
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, create)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d, body %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created AppPassword
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode created app password: %v", err)
	}
	if created.Secret == "" || created.Name != "Thunderbird laptop" || !containsString(created.Protocols, "smtp") {
		t.Fatalf("unexpected app password response: %+v", created)
	}
	if ok, _ := store.VerifyProtocolPassword(context.Background(), "marko@example.com", created.Secret, "smtp"); !ok {
		t.Fatal("created app password did not verify for allowed protocol")
	}
	if ok, _ := store.VerifyProtocolPassword(context.Background(), "marko@example.com", created.Secret, "webmail"); ok {
		t.Fatal("app password should not verify for webmail")
	}

	revoke := httptest.NewRequest(http.MethodDelete, "/api/v1/app-passwords/"+created.ID, nil)
	revoke.Header.Set("X-CSRF-Token", csrfFromBody(t, loginRec.Body.Bytes()))
	for _, cookie := range loginRec.Result().Cookies() {
		revoke.AddCookie(cookie)
	}
	revokeRec := httptest.NewRecorder()
	handler.ServeHTTP(revokeRec, revoke)
	if revokeRec.Code != http.StatusNoContent {
		t.Fatalf("revoke status = %d, want %d, body %s", revokeRec.Code, http.StatusNoContent, revokeRec.Body.String())
	}
	if ok, _ := store.VerifyProtocolPassword(context.Background(), "marko@example.com", created.Secret, "smtp"); ok {
		t.Fatal("revoked app password still verifies")
	}
}

func TestWebmailSessionAPIWithoutCookieDoesNotTriggerBasicPopup(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	handler := NewRouter(&fakeStore{}, manager)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for browser session auth", got)
	}
}

func TestWebmailSessionAPIRejectsBasicFallback(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	handler := NewRouter(&fakeStore{valid: true}, manager)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for browser session auth", got)
	}
}

func TestReportMessageEndpointRecordsSpamTraining(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"verdict":"spam"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/1/report", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.reportedEmail != "marko@example.com" || store.reportedID != "1" || store.reportedVerdict != "spam" {
		t.Fatalf("unexpected report: email=%q id=%q verdict=%q", store.reportedEmail, store.reportedID, store.reportedVerdict)
	}
}

func TestMoveMessageEndpointMovesSelectedMessage(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"folder":"trash"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/1/move", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.movedEmail != "marko@example.com" || store.movedID != "1" || store.movedFolder != "trash" {
		t.Fatalf("unexpected move: email=%q id=%q folder=%q", store.movedEmail, store.movedID, store.movedFolder)
	}
}

func TestDeleteMessageEndpointDeletesSelectedMessage(t *testing.T) {
	store := &fakeStore{valid: true, messageMailbox: ".Trash"}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/messages/1/delete", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if store.deletedMessageEmail != "marko@example.com" || store.deletedMessageID != "1" {
		t.Fatalf("unexpected delete: email=%q id=%q", store.deletedMessageEmail, store.deletedMessageID)
	}
}

func TestMoveMessageEndpointMovesSpamToTrash(t *testing.T) {
	store := &fakeStore{valid: true, messageMailbox: ".Spam"}
	handler := NewRouter(store)
	body := `{"folder":"trash"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/1/move", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.movedID != "1" || store.movedFolder != "trash" {
		t.Fatalf("unexpected spam move: id=%q folder=%q", store.movedID, store.movedFolder)
	}
}

func TestMoveMessageEndpointRejectsSpamMoveOutsideTrash(t *testing.T) {
	store := &fakeStore{valid: true, messageMailbox: ".Spam"}
	handler := NewRouter(store)
	body := `{"folder":"inbox"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/1/move", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if store.movedID != "" {
		t.Fatalf("spam message was moved unexpectedly: id=%q", store.movedID)
	}
}

func TestDeleteMessageEndpointRejectsNonTrashMessage(t *testing.T) {
	store := &fakeStore{valid: true, messageMailbox: "new"}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/messages/1/delete", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if store.deletedMessageID != "" {
		t.Fatalf("non-trash message was deleted unexpectedly: id=%q", store.deletedMessageID)
	}
}

func TestBatchMoveMessagesEndpointMovesSelectedMessages(t *testing.T) {
	store := &fakeStore{valid: true, messageMailboxes: map[string]string{"1": "new", "2": "new"}}
	handler := NewRouter(store)
	body := `{"ids":["1","2"],"folder":"trash"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/batch/move", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if strings.Join(store.movedIDs, ",") != "1,2" || store.movedFolder != "trash" {
		t.Fatalf("unexpected batch move: ids=%v folder=%q", store.movedIDs, store.movedFolder)
	}
}

func TestBatchDeleteMessagesEndpointDeletesTrashMessages(t *testing.T) {
	store := &fakeStore{valid: true, messageMailboxes: map[string]string{"1": ".Trash", "2": ".Trash"}}
	handler := NewRouter(store)
	body := `{"ids":["1","2"]}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/messages/batch/delete", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if strings.Join(store.deletedMessageIDs, ",") != "1,2" {
		t.Fatalf("unexpected batch delete ids=%v", store.deletedMessageIDs)
	}
}

func TestMoveMessageEndpointRestoresSentTrashOnlyToSent(t *testing.T) {
	store := &fakeStore{valid: true, messageMailbox: ".Trash", messageTrashOrigin: "sent"}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/1/move", strings.NewReader(`{"folder":"sent"}`))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	if store.movedFolder != "sent" {
		t.Fatalf("moved folder = %q, want sent", store.movedFolder)
	}
}

func TestMoveMessageEndpointRejectsSentTrashRestoreToInbox(t *testing.T) {
	store := &fakeStore{valid: true, messageMailbox: ".Trash", messageTrashOrigin: "sent"}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/messages/1/move", strings.NewReader(`{"folder":"inbox"}`))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if store.movedID != "" {
		t.Fatalf("message was moved unexpectedly: id=%q", store.movedID)
	}
}

func TestFoldersEndpointCreatesFolder(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/folders", strings.NewReader(`{"name":"Projects"}`))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.createdFolder != "Projects" {
		t.Fatalf("created folder = %q, want Projects", store.createdFolder)
	}
}

func TestFiltersEndpointCreatesFilter(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"name":"Boss","field":"from","operator":"contains","value":"boss@example.com","action":"move","folder":"Projects","enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/filters", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.createdFilter.Name != "Boss" || store.createdFilter.Folder != "Projects" {
		t.Fatalf("unexpected filter: %+v", store.createdFilter)
	}
}

func TestContactsEndpointReturnsContacts(t *testing.T) {
	handler := NewRouter(&fakeStore{valid: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Ada Lovelace") {
		t.Fatalf("contacts missing expected person: %s", rec.Body.String())
	}
}

func TestCreateContactEndpointUsesAuthenticatedUser(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"name":"Ada Lovelace","email":"ada@example.net"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.createdContact.Email != "ada@example.net" || store.createdContact.Name != "Ada Lovelace" {
		t.Fatalf("unexpected created contact: %+v", store.createdContact)
	}
}

func TestUpdateContactEndpointUsesAuthenticatedUser(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"name":"Ada Byron","email":"ada@lovelace.example"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/contacts/ada", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.updatedContactID != "ada" || store.updatedContact.Email != "ada@lovelace.example" {
		t.Fatalf("unexpected updated contact id=%q contact=%+v", store.updatedContactID, store.updatedContact)
	}
}

func TestDeleteContactEndpointUsesAuthenticatedUser(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/contacts/ada", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if store.deletedContactID != "ada" {
		t.Fatalf("deleted contact id = %q, want ada", store.deletedContactID)
	}
}

func TestCalendarEndpointReturnsEvents(t *testing.T) {
	handler := NewRouter(&fakeStore{valid: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/calendar", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Planning") {
		t.Fatalf("calendar missing expected event: %s", rec.Body.String())
	}
}

func TestCreateCalendarEndpointUsesAuthenticatedUser(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"title":"Planning","starts_at":"2026-05-07T10:00:00Z","ends_at":"2026-05-07T11:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calendar", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.createdEvent.Title != "Planning" {
		t.Fatalf("unexpected created event: %+v", store.createdEvent)
	}
}

func TestUpdateCalendarEndpointUsesAuthenticatedUser(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"title":"Planning updated","starts_at":"2026-05-07T12:00:00Z","ends_at":"2026-05-07T13:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/calendar/planning", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.updatedEventID != "planning" || store.updatedEvent.Title != "Planning updated" {
		t.Fatalf("unexpected updated event id=%q event=%+v", store.updatedEventID, store.updatedEvent)
	}
}

func TestDeleteCalendarEndpointUsesAuthenticatedUser(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/calendar/planning", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if store.deletedEventID != "planning" {
		t.Fatalf("deleted event id = %q, want planning", store.deletedEventID)
	}
}

func TestPasswordChangeEndpointRequiresCurrentPassword(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	body := `{"current_password":"secret123456","new_password":"newsecret123456"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/password", strings.NewReader(body))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if store.changedPasswordEmail != "marko@example.com" || store.changedPassword != "newsecret123456" {
		t.Fatalf("unexpected password change: email=%q password=%q", store.changedPasswordEmail, store.changedPassword)
	}
	if !containsString(store.auditActions, "webmail.password_change") {
		t.Fatalf("password change audit event missing: %+v", store.auditActions)
	}
}

func TestPasswordChangeInvalidatesWebmailSessions(t *testing.T) {
	store := &fakeStore{valid: true}
	manager := session.NewManager(session.Options{CookieName: "webmail_sid"})
	handler := NewRouter(store, manager)
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"email":"marko@example.com","password":"secret123456"}`))
	login.Header.Set("Content-Type", "application/json")
	login.Header.Set("User-Agent", "Browser A")
	login.Header.Set("Accept-Language", "en-US")
	login.RemoteAddr = "203.0.113.10:48123"
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	csrf := csrfFromBody(t, loginRec.Body.Bytes())
	cookie := loginRec.Result().Cookies()[0]

	change := httptest.NewRequest(http.MethodPost, "/api/v1/password", strings.NewReader(`{"current_password":"secret123456","new_password":"newsecret123456"}`))
	change.Header.Set("Content-Type", "application/json")
	change.Header.Set("User-Agent", "Browser A")
	change.Header.Set("Accept-Language", "en-US")
	change.Header.Set("X-CSRF-Token", csrf)
	change.RemoteAddr = "203.0.113.10:48123"
	change.AddCookie(cookie)
	changeRec := httptest.NewRecorder()
	handler.ServeHTTP(changeRec, change)
	if changeRec.Code != http.StatusNoContent {
		t.Fatalf("change status = %d, want %d, body %s", changeRec.Code, http.StatusNoContent, changeRec.Body.String())
	}

	current := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	current.Header.Set("User-Agent", "Browser A")
	current.Header.Set("Accept-Language", "en-US")
	current.RemoteAddr = "203.0.113.10:48123"
	current.AddCookie(cookie)
	currentRec := httptest.NewRecorder()
	handler.ServeHTTP(currentRec, current)
	if currentRec.Code != http.StatusUnauthorized {
		t.Fatalf("session status after password change = %d, want %d", currentRec.Code, http.StatusUnauthorized)
	}
}

func TestCompositeStoreReportsSpamByLearningAndMovingMessage(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir maildir: %v", err)
	}
	messageID := "message-1"
	messagePath := filepath.Join(messageDir, messageID)
	if err := os.WriteFile(messagePath, []byte("From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Bad\r\n\r\nbody"), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}
	auth := &reportRecorder{}
	learner := &fakeLearner{}
	store := CompositeStore{Auth: auth, Mailbox: MaildirStore{Root: root}, Learner: learner}

	if err := store.ReportMessage(context.Background(), "marko@example.com", messageID, "spam"); err != nil {
		t.Fatalf("ReportMessage returned error: %v", err)
	}
	if learner.verdict != "spam" || learner.path != messagePath {
		t.Fatalf("unexpected learner call: verdict=%q path=%q", learner.verdict, learner.path)
	}
	if auth.verdict != "spam" || auth.messageID != messageID {
		t.Fatalf("unexpected audit call: verdict=%q id=%q", auth.verdict, auth.messageID)
	}
	if _, err := os.Stat(filepath.Join(root, "example.com", "marko", "Maildir", ".Spam", "new", messageID)); err != nil {
		t.Fatalf("message was not moved to spam: %v", err)
	}
}

func TestCompositeStoreReportsHamByLearningAndMovingMessageToInbox(t *testing.T) {
	root := t.TempDir()
	messageDir := filepath.Join(root, "example.com", "marko", "Maildir", ".Spam", "new")
	if err := os.MkdirAll(messageDir, 0750); err != nil {
		t.Fatalf("mkdir spam: %v", err)
	}
	messageID := "message-1"
	messagePath := filepath.Join(messageDir, messageID)
	if err := os.WriteFile(messagePath, []byte("From: sender@example.net\r\nTo: marko@example.com\r\nSubject: Good\r\n\r\nbody"), 0640); err != nil {
		t.Fatalf("write message: %v", err)
	}
	auth := &reportRecorder{}
	learner := &fakeLearner{}
	store := CompositeStore{Auth: auth, Mailbox: MaildirStore{Root: root}, Learner: learner}

	if err := store.ReportMessage(context.Background(), "marko@example.com", messageID, "ham"); err != nil {
		t.Fatalf("ReportMessage returned error: %v", err)
	}
	if learner.verdict != "ham" || learner.path != messagePath {
		t.Fatalf("unexpected learner call: verdict=%q path=%q", learner.verdict, learner.path)
	}
	if auth.verdict != "ham" || auth.messageID != messageID {
		t.Fatalf("unexpected audit call: verdict=%q id=%q", auth.verdict, auth.messageID)
	}
	if _, err := os.Stat(filepath.Join(root, "example.com", "marko", "Maildir", "new", messageID)); err != nil {
		t.Fatalf("message was not moved to inbox: %v", err)
	}
}

type fakeStore struct {
	valid                bool
	mailboxes            []MailboxAccount
	listEmail            string
	sent                 OutboundMessage
	reportedEmail        string
	reportedID           string
	reportedVerdict      string
	folder               string
	createdFolder        string
	deletedFolder        string
	createdFilter        MailFilter
	updatedFilterID      string
	updatedFilter        MailFilter
	deletedFilterID      string
	movedEmail           string
	movedID              string
	movedIDs             []string
	movedFolder          string
	deletedMessageEmail  string
	deletedMessageID     string
	deletedMessageIDs    []string
	messageMailbox       string
	messageTrashOrigin   string
	messageMailboxes     map[string]string
	createdContact       Contact
	updatedContactID     string
	updatedContact       Contact
	deletedContactID     string
	createdEvent         CalendarEvent
	updatedEventID       string
	updatedEvent         CalendarEvent
	deletedEventID       string
	changedPasswordEmail string
	changedPassword      string
	profile              UserProfile
	updatedProfile       UserProfile
	savedDraft           OutboundMessage
	listTrustEmail       string
	trustEmail           string
	trustEntry           ContentTrustEntry
	trustEntries         []ContentTrustEntry
	mailboxSecurity      MailboxSecurity
	proIdentitySettings  domain.AdminMFASettings
	mfaChallenges        map[string]MailboxMFAChallenge
	appPasswords         map[string]AppPassword
	appPasswordHashes    map[string]string
	auditActions         []string
}

func (s *fakeStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	return s.valid && email == "marko@example.com" && password == "secret123456", nil
}

func (s *fakeStore) ListMailboxes(ctx context.Context, email string) ([]MailboxAccount, error) {
	if len(s.mailboxes) > 0 {
		return s.mailboxes, nil
	}
	return []MailboxAccount{{ID: email, Name: strings.Split(email, "@")[0], Address: email, Kind: "personal", CanRead: true, CanManage: true, CanSendAs: true, CanSendOnBehalf: true}}, nil
}

func (s *fakeStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	s.listEmail = email
	s.folder = "inbox"
	return []MessageSummary{{ID: "1", From: "sender@example.net", To: email, Subject: "Welcome", Preview: "Hello"}}, nil
}

func (s *fakeStore) ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error) {
	s.listEmail = email
	s.folder = folder
	return []MessageSummary{{ID: "1", From: "sender@example.net", To: email, Subject: "Welcome", Preview: "Hello"}}, nil
}

func (s *fakeStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	mailbox := s.messageMailbox
	if s.messageMailboxes != nil && s.messageMailboxes[id] != "" {
		mailbox = s.messageMailboxes[id]
	}
	return MessageDetail{ID: id, From: "sender@example.net", To: email, Subject: "Welcome", Body: "Full body", Mailbox: mailbox, TrashOrigin: s.messageTrashOrigin}, nil
}

func (s *fakeStore) SendMessage(ctx context.Context, message OutboundMessage) error {
	s.sent = message
	return nil
}

func (s *fakeStore) SaveSentMessage(ctx context.Context, message OutboundMessage) error {
	return nil
}

func (s *fakeStore) SaveDraftMessage(ctx context.Context, message OutboundMessage) (string, error) {
	s.savedDraft = message
	return "draft-1", nil
}

func (s *fakeStore) ReportMessage(ctx context.Context, email, id, verdict string) error {
	s.reportedEmail = email
	s.reportedID = id
	s.reportedVerdict = verdict
	return nil
}

func (s *fakeStore) MoveMessage(ctx context.Context, email, id, folder string) error {
	s.movedEmail = email
	s.movedID = id
	s.movedIDs = append(s.movedIDs, id)
	s.movedFolder = folder
	return nil
}

func (s *fakeStore) DeleteMessage(ctx context.Context, email, id string) error {
	s.deletedMessageEmail = email
	s.deletedMessageID = id
	s.deletedMessageIDs = append(s.deletedMessageIDs, id)
	return nil
}

func (s *fakeStore) ListFolders(ctx context.Context, email string) ([]MailFolder, error) {
	return []MailFolder{{ID: "inbox", Name: "Inbox", System: true, Total: 1}}, nil
}

func (s *fakeStore) CreateFolder(ctx context.Context, email, name string) (MailFolder, error) {
	s.createdFolder = name
	return MailFolder{ID: name, Name: name}, nil
}

func (s *fakeStore) DeleteFolder(ctx context.Context, email, name string) error {
	s.deletedFolder = name
	return nil
}

func (s *fakeStore) ListFilters(ctx context.Context, email string) ([]MailFilter, error) {
	return []MailFilter{{ID: "1", Name: "Boss", Field: "from", Operator: "contains", Value: "boss@example.com", Action: "move", Folder: "Projects", Enabled: true}}, nil
}

func (s *fakeStore) CreateFilter(ctx context.Context, email string, filter MailFilter) (MailFilter, error) {
	s.createdFilter = filter
	filter.ID = "1"
	return filter, nil
}

func (s *fakeStore) UpdateFilter(ctx context.Context, email, id string, filter MailFilter) (MailFilter, error) {
	s.updatedFilterID = id
	s.updatedFilter = filter
	filter.ID = id
	return filter, nil
}

func (s *fakeStore) DeleteFilter(ctx context.Context, email, id string) error {
	s.deletedFilterID = id
	return nil
}

func (s *fakeStore) ListContacts(ctx context.Context, email string) ([]Contact, error) {
	return []Contact{{ID: "ada", Name: "Ada Lovelace", Email: "ada@example.net"}}, nil
}

func (s *fakeStore) CreateContact(ctx context.Context, email string, contact Contact) (Contact, error) {
	s.createdContact = contact
	contact.ID = "ada"
	return contact, nil
}

func (s *fakeStore) UpdateContact(ctx context.Context, email, id string, contact Contact) (Contact, error) {
	s.updatedContactID = id
	s.updatedContact = contact
	contact.ID = id
	return contact, nil
}

func (s *fakeStore) DeleteContact(ctx context.Context, email, id string) error {
	s.deletedContactID = id
	return nil
}

func (s *fakeStore) ListCalendarEvents(ctx context.Context, email string) ([]CalendarEvent, error) {
	return []CalendarEvent{{ID: "planning", Title: "Planning"}}, nil
}

func (s *fakeStore) CreateCalendarEvent(ctx context.Context, email string, event CalendarEvent) (CalendarEvent, error) {
	s.createdEvent = event
	event.ID = "planning"
	return event, nil
}

func (s *fakeStore) UpdateCalendarEvent(ctx context.Context, email, id string, event CalendarEvent) (CalendarEvent, error) {
	s.updatedEventID = id
	s.updatedEvent = event
	event.ID = id
	return event, nil
}

func (s *fakeStore) DeleteCalendarEvent(ctx context.Context, email, id string) error {
	s.deletedEventID = id
	return nil
}

func (s *fakeStore) ChangePassword(ctx context.Context, email, newPassword string) error {
	s.changedPasswordEmail = email
	s.changedPassword = newPassword
	return nil
}

func (s *fakeStore) GetProfile(ctx context.Context, email string) (UserProfile, error) {
	if s.profile.Email == "" {
		s.profile = UserProfile{Email: email, DisplayName: strings.Split(email, "@")[0], Language: "en"}
	}
	if s.profile.Language == "" {
		s.profile.Language = "en"
	}
	return s.profile, nil
}

func (s *fakeStore) UpdateProfile(ctx context.Context, email string, profile UserProfile) (UserProfile, error) {
	profile.Email = email
	s.updatedProfile = profile
	s.profile.SignatureHTML = profile.SignatureHTML
	s.profile.SignatureAutoAdd = profile.SignatureAutoAdd
	s.profile.Language = profile.Language
	if s.profile.Email == "" {
		s.profile.Email = email
	}
	return s.profile, nil
}

func (s *fakeStore) ListContentTrust(ctx context.Context, email string) ([]ContentTrustEntry, error) {
	s.listTrustEmail = email
	return s.trustEntries, nil
}

func (s *fakeStore) AddContentTrust(ctx context.Context, email string, entry ContentTrustEntry) (ContentTrustEntry, error) {
	s.trustEmail = email
	s.trustEntry = entry
	if entry.ID == "" {
		entry.ID = entry.Scope + ":" + entry.Value
	}
	s.trustEntries = append(s.trustEntries, entry)
	return entry, nil
}

func (s *fakeStore) GetMailboxSecurity(ctx context.Context, email string) (MailboxSecurity, error) {
	state := s.mailboxSecurity
	state.Email = email
	if state.MFAEnabled || state.TOTPEnabled || state.ForceMFA {
		state.MFAAvailable = true
	}
	state.SetupNeeded = state.MFAAvailable && state.ForceMFA && !state.MFAEnabled
	return state, nil
}

func (s *fakeStore) GetAdminMFASettings(ctx context.Context) (domain.AdminMFASettings, error) {
	return s.proIdentitySettings, nil
}

func (s *fakeStore) CreateMailboxMFAChallenge(ctx context.Context, challenge MailboxMFAChallenge) error {
	if s.mfaChallenges == nil {
		s.mfaChallenges = map[string]MailboxMFAChallenge{}
	}
	s.mfaChallenges[challenge.Token] = challenge
	return nil
}

func (s *fakeStore) GetMailboxMFAChallenge(ctx context.Context, token string) (MailboxMFAChallenge, error) {
	if challenge, ok := s.mfaChallenges[token]; ok {
		return challenge, nil
	}
	return MailboxMFAChallenge{}, os.ErrNotExist
}

func (s *fakeStore) DeleteMailboxMFAChallenge(ctx context.Context, token string) error {
	delete(s.mfaChallenges, token)
	return nil
}

func (s *fakeStore) BeginMailboxTOTPEnrollment(ctx context.Context, email string) (MailboxTOTPEnrollment, error) {
	return MailboxTOTPEnrollment{Email: email, OTPAuthURL: "otpauth://totp/ProIdentity%20Mail:" + email, QRDataURL: "data:image/png;base64,fake"}, nil
}

func (s *fakeStore) VerifyMailboxTOTPEnrollment(ctx context.Context, email, code string) (MailboxSecurity, error) {
	if code != "123456" {
		return MailboxSecurity{}, os.ErrPermission
	}
	s.mailboxSecurity = MailboxSecurity{Email: email, MFAAvailable: true, MFAEnabled: true, TOTPEnabled: true, ForceMFA: s.mailboxSecurity.ForceMFA}
	return s.mailboxSecurity, nil
}

func (s *fakeStore) VerifyMailboxTOTPCode(ctx context.Context, email, code string) (bool, error) {
	return code == "123456", nil
}

func (s *fakeStore) ListAppPasswords(ctx context.Context, email string) ([]AppPassword, error) {
	out := make([]AppPassword, 0, len(s.appPasswords))
	for _, password := range s.appPasswords {
		password.Secret = ""
		out = append(out, password)
	}
	return out, nil
}

func (s *fakeStore) CreateAppPassword(ctx context.Context, email string, req AppPassword) (AppPassword, error) {
	if s.appPasswords == nil {
		s.appPasswords = map[string]AppPassword{}
	}
	if s.appPasswordHashes == nil {
		s.appPasswordHashes = map[string]string{}
	}
	secret, err := newAppPasswordSecret()
	if err != nil {
		return AppPassword{}, err
	}
	created := AppPassword{
		ID:        "app-1",
		Name:      strings.TrimSpace(req.Name),
		Protocols: normalizeAppPasswordProtocols(req.Protocols),
		Secret:    secret,
		Status:    "active",
		CreatedAt: time.Now().UTC(),
	}
	s.appPasswords[created.ID] = created
	s.appPasswordHashes[appPasswordFingerprint(secret)] = created.ID
	return created, nil
}

func (s *fakeStore) RevokeAppPassword(ctx context.Context, email, id string) error {
	password, ok := s.appPasswords[id]
	if !ok {
		return os.ErrNotExist
	}
	password.Status = "revoked"
	s.appPasswords[id] = password
	return nil
}

func (s *fakeStore) VerifyProtocolPassword(ctx context.Context, email, password, protocol string) (bool, error) {
	if normalizeProtocol(protocol) == "webmail" {
		return false, nil
	}
	id := s.appPasswordHashes[appPasswordFingerprint(password)]
	if id == "" {
		return false, nil
	}
	appPassword := s.appPasswords[id]
	return appPassword.Status == "active" && protocolAllowed(appPassword.Protocols, protocol), nil
}

func (s *fakeStore) RecordUserAudit(ctx context.Context, email, action, targetType, targetID string, metadata map[string]any) error {
	s.auditActions = append(s.auditActions, action)
	return nil
}

func csrfFromBody(t *testing.T, body []byte) string {
	t.Helper()
	var response struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode csrf token: %v", err)
	}
	if response.CSRFToken == "" {
		t.Fatalf("csrf token missing in body %s", string(body))
	}
	return response.CSRFToken
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type reportRecorder struct {
	messageID string
	verdict   string
}

func (r *reportRecorder) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	return true, nil
}

func (r *reportRecorder) ReportMessage(ctx context.Context, email, id, verdict string) error {
	r.messageID = id
	r.verdict = verdict
	return nil
}

type fakeLearner struct {
	path    string
	verdict string
}

func (l *fakeLearner) Learn(ctx context.Context, path, verdict string) error {
	l.path = path
	l.verdict = verdict
	return nil
}
