package webmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

type fakeStore struct {
	valid                bool
	sent                 OutboundMessage
	reportedEmail        string
	reportedID           string
	reportedVerdict      string
	folder               string
	movedEmail           string
	movedID              string
	movedFolder          string
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
}

func (s *fakeStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	return s.valid && email == "marko@example.com" && password == "secret123456", nil
}

func (s *fakeStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	s.folder = "inbox"
	return []MessageSummary{{ID: "1", From: "sender@example.net", To: email, Subject: "Welcome", Preview: "Hello"}}, nil
}

func (s *fakeStore) ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error) {
	s.folder = folder
	return []MessageSummary{{ID: "1", From: "sender@example.net", To: email, Subject: "Welcome", Preview: "Hello"}}, nil
}

func (s *fakeStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	return MessageDetail{ID: id, From: "sender@example.net", To: email, Subject: "Welcome", Body: "Full body"}, nil
}

func (s *fakeStore) SendMessage(ctx context.Context, message OutboundMessage) error {
	s.sent = message
	return nil
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
	s.movedFolder = folder
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
