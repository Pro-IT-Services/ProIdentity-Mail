package webmail

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image/png"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/i18n"
	"proidentity-mail/internal/security"
	"proidentity-mail/internal/session"
)

type Store interface {
	VerifyUserPassword(ctx context.Context, email, password string) (bool, error)
	ListMailboxes(ctx context.Context, email string) ([]MailboxAccount, error)
	ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error)
	ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error)
	GetMessage(ctx context.Context, email, id string) (MessageDetail, error)
	SendMessage(ctx context.Context, message OutboundMessage) error
	SaveSentMessage(ctx context.Context, message OutboundMessage) error
	SaveDraftMessage(ctx context.Context, message OutboundMessage) (string, error)
	ReportMessage(ctx context.Context, email, id, verdict string) error
	MoveMessage(ctx context.Context, email, id, folder string) error
	DeleteMessage(ctx context.Context, email, id string) error
	ListFolders(ctx context.Context, email string) ([]MailFolder, error)
	CreateFolder(ctx context.Context, email, name string) (MailFolder, error)
	DeleteFolder(ctx context.Context, email, name string) error
	ListFilters(ctx context.Context, email string) ([]MailFilter, error)
	CreateFilter(ctx context.Context, email string, filter MailFilter) (MailFilter, error)
	UpdateFilter(ctx context.Context, email, id string, filter MailFilter) (MailFilter, error)
	DeleteFilter(ctx context.Context, email, id string) error
	ListContacts(ctx context.Context, email string) ([]Contact, error)
	CreateContact(ctx context.Context, email string, contact Contact) (Contact, error)
	UpdateContact(ctx context.Context, email, id string, contact Contact) (Contact, error)
	DeleteContact(ctx context.Context, email, id string) error
	ListCalendarEvents(ctx context.Context, email string) ([]CalendarEvent, error)
	CreateCalendarEvent(ctx context.Context, email string, event CalendarEvent) (CalendarEvent, error)
	UpdateCalendarEvent(ctx context.Context, email, id string, event CalendarEvent) (CalendarEvent, error)
	DeleteCalendarEvent(ctx context.Context, email, id string) error
	GetProfile(ctx context.Context, email string) (UserProfile, error)
	UpdateProfile(ctx context.Context, email string, profile UserProfile) (UserProfile, error)
	ChangePassword(ctx context.Context, email, newPassword string) error
	ListContentTrust(ctx context.Context, email string) ([]ContentTrustEntry, error)
	AddContentTrust(ctx context.Context, email string, entry ContentTrustEntry) (ContentTrustEntry, error)
}

type MailboxMFAStore interface {
	GetMailboxSecurity(ctx context.Context, email string) (MailboxSecurity, error)
	CreateMailboxMFAChallenge(ctx context.Context, challenge MailboxMFAChallenge) error
	GetMailboxMFAChallenge(ctx context.Context, token string) (MailboxMFAChallenge, error)
	DeleteMailboxMFAChallenge(ctx context.Context, token string) error
	BeginMailboxTOTPEnrollment(ctx context.Context, email string) (MailboxTOTPEnrollment, error)
	VerifyMailboxTOTPEnrollment(ctx context.Context, email, code string) (MailboxSecurity, error)
	VerifyMailboxTOTPCode(ctx context.Context, email, code string) (bool, error)
}

type ProIdentitySettingsStore interface {
	GetAdminMFASettings(ctx context.Context) (domain.AdminMFASettings, error)
}

type AppPasswordStore interface {
	ListAppPasswords(ctx context.Context, email string) ([]AppPassword, error)
	CreateAppPassword(ctx context.Context, email string, req AppPassword) (AppPassword, error)
	RevokeAppPassword(ctx context.Context, email, id string) error
}

type ProtocolPasswordVerifier interface {
	VerifyProtocolPassword(ctx context.Context, email, password, protocol string) (bool, error)
}

type MessageReadMarker interface {
	MarkMessageRead(ctx context.Context, email, id string) (MessageDetail, error)
}

type UserAuditRecorder interface {
	RecordUserAudit(ctx context.Context, email, action, targetType, targetID string, metadata map[string]any) error
}

type OutboundMessage struct {
	From        string               `json:"from"`
	To          []string             `json:"to"`
	Subject     string               `json:"subject"`
	Body        string               `json:"body"`
	BodyHTML    string               `json:"body_html,omitempty"`
	Attachments []OutboundAttachment `json:"attachments,omitempty"`
	DraftID     string               `json:"draft_id,omitempty"`
}

type OutboundAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"-"`
	Size        int64  `json:"size"`
}

type UserProfile struct {
	Email            string `json:"email"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	DisplayName      string `json:"display_name"`
	SignatureHTML    string `json:"signature_html"`
	SignatureAutoAdd bool   `json:"signature_auto_add"`
	Language         string `json:"language"`
}

type MailboxSecurity struct {
	Email        string `json:"email,omitempty"`
	MFAAvailable bool   `json:"mfa_available"`
	ForceMFA     bool   `json:"force_mfa"`
	MFAEnabled   bool   `json:"mfa_enabled"`
	TOTPEnabled  bool   `json:"totp_enabled"`
	SetupNeeded  bool   `json:"setup_needed"`
}

type MailboxMFAChallenge struct {
	Token     string
	Email     string
	Purpose   string
	Provider  string
	RequestID string
	ExpiresAt time.Time
}

type MailboxTOTPEnrollment struct {
	Email      string `json:"email"`
	OTPAuthURL string `json:"otpauth_url"`
	QRDataURL  string `json:"qr_data_url"`
}

type AppPassword struct {
	ID               string     `json:"id,omitempty"`
	Name             string     `json:"name"`
	Protocols        []string   `json:"protocols"`
	Secret           string     `json:"secret,omitempty"`
	Status           string     `json:"status,omitempty"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	LastUsedProtocol string     `json:"last_used_protocol,omitempty"`
	CreatedAt        time.Time  `json:"created_at,omitempty"`
}

type ContentTrustEntry struct {
	ID        string `json:"id,omitempty"`
	Scope     string `json:"scope"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at,omitempty"`
}

type Contact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type MailFolder struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	System bool   `json:"system"`
	Unread int    `json:"unread"`
	Total  int    `json:"total"`
}

type MailFilter struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Field     string `json:"field"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
	Action    string `json:"action"`
	Folder    string `json:"folder,omitempty"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type CalendarEvent struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

type MailboxAccount struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Address         string `json:"address"`
	Kind            string `json:"kind"`
	CanRead         bool   `json:"can_read"`
	CanSendAs       bool   `json:"can_send_as"`
	CanSendOnBehalf bool   `json:"can_send_on_behalf"`
	CanManage       bool   `json:"can_manage"`
}

type handler struct {
	store    Store
	sessions *session.Manager
	limiter  session.Limiter
}

const bulkSendAlertRecipientThreshold = 20

func NewRouter(store Store, managers ...*session.Manager) http.Handler {
	var manager *session.Manager
	if len(managers) > 0 {
		manager = managers[0]
	}
	return NewRouterWithLimiter(store, manager, session.NewLoginLimiter(session.Options{}))
}

func NewRouterWithLimiter(store Store, manager *session.Manager, limiter session.Limiter) http.Handler {
	if limiter == nil {
		limiter = session.NewLoginLimiter(session.Options{})
	}
	h := handler{store: store, sessions: manager, limiter: limiter}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/api/v1/session", h.session)
	mux.HandleFunc("/api/v1/session/mfa", h.sessionMFA)
	mux.HandleFunc("/api/v1/mfa", h.mailboxMFA)
	mux.HandleFunc("/api/v1/mfa/totp/enroll", h.mailboxTOTPEnroll)
	mux.HandleFunc("/api/v1/mfa/totp/verify", h.mailboxTOTPVerify)
	mux.HandleFunc("/api/v1/app-passwords", h.appPasswords)
	mux.HandleFunc("/api/v1/app-passwords/", h.appPassword)
	mux.HandleFunc("/api/v1/mailboxes", h.mailboxes)
	mux.HandleFunc("/api/v1/messages", h.messages)
	mux.HandleFunc("/api/v1/messages/batch/move", h.batchMoveMessages)
	mux.HandleFunc("/api/v1/messages/batch/delete", h.batchDeleteMessages)
	mux.HandleFunc("/api/v1/messages/", h.message)
	mux.HandleFunc("/api/v1/send", h.send)
	mux.HandleFunc("/api/v1/drafts", h.draft)
	mux.HandleFunc("/api/v1/folders", h.folders)
	mux.HandleFunc("/api/v1/folders/", h.folder)
	mux.HandleFunc("/api/v1/filters", h.filters)
	mux.HandleFunc("/api/v1/filters/", h.filter)
	mux.HandleFunc("/api/v1/contacts", h.contacts)
	mux.HandleFunc("/api/v1/contacts/", h.contact)
	mux.HandleFunc("/api/v1/calendar", h.calendar)
	mux.HandleFunc("/api/v1/calendar/", h.calendarEvent)
	mux.HandleFunc("/api/v1/profile", h.profile)
	mux.HandleFunc("/api/v1/password", h.changePassword)
	mux.HandleFunc("/api/v1/content-trust", h.contentTrust)
	mux.HandleFunc("/", index)
	return security.BrowserHeaders(security.LimitRequestBody(35 << 20)(mux))
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
}

func index(w http.ResponseWriter, r *http.Request) {
	nonce, err := security.NewCSPNonce()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create security nonce failed")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", security.BrowserCSP(nonce))
	html := strings.Replace(webmailIndexHTML, "__PROIDENTITY_I18N_CATALOG__", i18n.CatalogJSON(), 1)
	html = strings.ReplaceAll(html, "__PROIDENTITY_CSP_NONCE__", nonce)
	_, _ = w.Write([]byte(html))
}

func (h handler) messages(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	targetEmail, ok := h.authorizedMailbox(w, r, email, false)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	folder := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("folder")))
	if folder == "" {
		folder = "inbox"
	}
	messages, err := h.store.ListMessages(r.Context(), targetEmail, folder, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list messages failed")
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

func (h handler) message(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	targetEmail, ok := h.authorizedMailbox(w, r, email, false)
	if !ok {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/messages/")
	if strings.HasSuffix(id, "/report") {
		h.reportMessage(w, r, email, targetEmail, strings.TrimSuffix(id, "/report"))
		return
	}
	if strings.HasSuffix(id, "/move") {
		h.moveMessage(w, r, email, targetEmail, strings.TrimSuffix(id, "/move"))
		return
	}
	if strings.HasSuffix(id, "/delete") {
		h.deleteMessage(w, r, email, targetEmail, strings.TrimSuffix(id, "/delete"))
		return
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "message id is required")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var message MessageDetail
	var err error
	if marker, ok := h.store.(MessageReadMarker); ok {
		message, err = marker.MarkMessageRead(r.Context(), targetEmail, id)
	} else {
		message, err = h.store.GetMessage(r.Context(), targetEmail, id)
	}
	if err != nil {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	writeJSON(w, http.StatusOK, message)
}

func (h handler) mailboxes(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mailboxes, err := h.store.ListMailboxes(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list mailboxes failed")
		return
	}
	writeJSON(w, http.StatusOK, mailboxes)
}

func (h handler) reportMessage(w http.ResponseWriter, r *http.Request, actorEmail, mailboxEmail, id string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "message id is required")
		return
	}
	var req struct {
		Verdict string `json:"verdict"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Verdict = strings.ToLower(strings.TrimSpace(req.Verdict))
	if req.Verdict != "spam" && req.Verdict != "ham" {
		writeError(w, http.StatusBadRequest, "verdict must be spam or ham")
		return
	}
	if err := h.store.ReportMessage(r.Context(), mailboxEmail, id, req.Verdict); err != nil {
		log.Printf("webmail report failed email=%q mailbox=%q id=%q verdict=%q: %v", actorEmail, mailboxEmail, id, req.Verdict, err)
		writeError(w, http.StatusInternalServerError, "report message failed")
		return
	}
	h.recordUserAudit(r.Context(), actorEmail, "message.report", "message", id, map[string]any{"mailbox": mailboxEmail, "verdict": req.Verdict})
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "recorded"})
}

func (h handler) moveMessage(w http.ResponseWriter, r *http.Request, actorEmail, mailboxEmail, id string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "message id is required")
		return
	}
	var req struct {
		Folder string `json:"folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Folder = strings.TrimSpace(req.Folder)
	if req.Folder == "" || strings.ContainsAny(req.Folder, `/\:`) || strings.Contains(req.Folder, "..") {
		writeError(w, http.StatusBadRequest, "valid folder is required")
		return
	}
	if err := h.validateMove(r.Context(), mailboxEmail, id, req.Folder); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.MoveMessage(r.Context(), mailboxEmail, id, req.Folder); err != nil {
		log.Printf("webmail move failed email=%q mailbox=%q id=%q folder=%q: %v", actorEmail, mailboxEmail, id, req.Folder, err)
		writeError(w, http.StatusInternalServerError, "move message failed")
		return
	}
	h.recordUserAudit(r.Context(), actorEmail, "message.move", "message", id, map[string]any{"mailbox": mailboxEmail, "folder": req.Folder, "count": 1})
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "moved"})
}

func (h handler) batchMoveMessages(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	targetEmail, ok := h.authorizedMailbox(w, r, email, false)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		IDs    []string `json:"ids"`
		Folder string   `json:"folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Folder = strings.TrimSpace(req.Folder)
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "message ids are required")
		return
	}
	for _, id := range req.IDs {
		if err := h.validateMove(r.Context(), targetEmail, id, req.Folder); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	for _, id := range req.IDs {
		if err := h.store.MoveMessage(r.Context(), targetEmail, id, req.Folder); err != nil {
			log.Printf("webmail batch move failed email=%q mailbox=%q id=%q folder=%q: %v", email, targetEmail, id, req.Folder, err)
			writeError(w, http.StatusInternalServerError, "move messages failed")
			return
		}
	}
	h.recordUserAudit(r.Context(), email, "message.move", "message", "batch", map[string]any{"mailbox": targetEmail, "folder": req.Folder, "count": len(req.IDs)})
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "moved", "count": len(req.IDs)})
}

func (h handler) deleteMessage(w http.ResponseWriter, r *http.Request, actorEmail, mailboxEmail, id string) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		w.Header().Set("Allow", "DELETE, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "message id is required")
		return
	}
	if err := h.validateDelete(r.Context(), mailboxEmail, id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.DeleteMessage(r.Context(), mailboxEmail, id); err != nil {
		log.Printf("webmail delete failed email=%q mailbox=%q id=%q: %v", actorEmail, mailboxEmail, id, err)
		writeError(w, http.StatusInternalServerError, "delete message failed")
		return
	}
	h.recordUserAudit(r.Context(), actorEmail, "message.delete", "message", id, map[string]any{"mailbox": mailboxEmail, "count": 1})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) batchDeleteMessages(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	targetEmail, ok := h.authorizedMailbox(w, r, email, false)
	if !ok {
		return
	}
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		w.Header().Set("Allow", "DELETE, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "message ids are required")
		return
	}
	for _, id := range req.IDs {
		if err := h.validateDelete(r.Context(), targetEmail, id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	for _, id := range req.IDs {
		if err := h.store.DeleteMessage(r.Context(), targetEmail, id); err != nil {
			log.Printf("webmail batch delete failed email=%q mailbox=%q id=%q: %v", email, targetEmail, id, err)
			writeError(w, http.StatusInternalServerError, "delete messages failed")
			return
		}
	}
	h.recordUserAudit(r.Context(), email, "message.delete", "message", "batch", map[string]any{"mailbox": targetEmail, "count": len(req.IDs)})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) validateMove(ctx context.Context, email, id, folder string) error {
	id = strings.TrimSpace(id)
	folder = strings.TrimSpace(folder)
	if id == "" {
		return errors.New("message id is required")
	}
	if folder == "" || strings.ContainsAny(folder, `/\:`) || strings.Contains(folder, "..") {
		return errors.New("valid folder is required")
	}
	message, err := h.store.GetMessage(ctx, email, id)
	if err != nil {
		return errors.New("message not found")
	}
	source := messageFolderID(message.Mailbox)
	target := strings.ToLower(strings.TrimPrefix(folder, "."))
	switch source {
	case "spam":
		if target == "trash" {
			return nil
		}
		return errors.New("spam messages can only be moved to trash")
	case "sent":
		if target == "trash" {
			return nil
		}
	case "inbox":
		if target == "trash" || target == "archive" || isCustomFolderTarget(folder) {
			return nil
		}
	case "trash":
		origin := strings.ToLower(strings.TrimPrefix(message.TrashOrigin, "."))
		if origin == "sent" {
			if target == "sent" {
				return nil
			}
			return errors.New("sent trash messages can only be restored to sent")
		}
		if target == "inbox" || isCustomFolderTarget(folder) {
			return nil
		}
	}
	return errors.New("this move is not allowed")
}

func (h handler) validateDelete(ctx context.Context, email, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("message id is required")
	}
	message, err := h.store.GetMessage(ctx, email, id)
	if err != nil {
		return errors.New("message not found")
	}
	if messageFolderID(message.Mailbox) != "trash" {
		return errors.New("messages can only be permanently deleted from trash")
	}
	return nil
}

func messageFolderID(mailbox string) string {
	name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(mailbox), "."))
	switch name {
	case "", "new", "cur":
		return "inbox"
	case "trash":
		return "trash"
	case "spam":
		return "spam"
	case "sent":
		return "sent"
	case "drafts", "draft":
		return "drafts"
	case "archive":
		return "archive"
	default:
		return name
	}
}

func isCustomFolderTarget(folder string) bool {
	name := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(folder), "."))
	switch name {
	case "", "inbox", "new", "cur", "sent", "drafts", "draft", "archive", "spam", "trash":
		return false
	default:
		return true
	}
}

func (h handler) send(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req, err := parseOutboundRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.To) == 0 || strings.TrimSpace(req.Subject) == "" {
		writeError(w, http.StatusBadRequest, "recipient and subject are required")
		return
	}
	from, ok := h.authorizedSender(w, r, email, req.From)
	if !ok {
		return
	}
	message := OutboundMessage{From: from, To: req.To, Subject: req.Subject, Body: req.Body, BodyHTML: req.BodyHTML, Attachments: req.Attachments, DraftID: req.DraftID}
	if err := h.store.SendMessage(r.Context(), message); err != nil {
		log.Printf("webmail send failed from=%q to=%q: %v", message.From, message.To, err)
		writeError(w, http.StatusInternalServerError, "send message failed")
		return
	}
	if message.DraftID != "" {
		if err := h.store.DeleteMessage(r.Context(), message.From, message.DraftID); err != nil {
			log.Printf("webmail draft cleanup failed from=%q draft=%q: %v", message.From, message.DraftID, err)
		}
	}
	if len(message.To) >= bulkSendAlertRecipientThreshold {
		h.recordUserAudit(r.Context(), email, "security.alert.bulk_send", "message", "outbound", map[string]any{"from": message.From, "recipient_count": len(message.To), "subject": message.Subject})
	}
	h.recordUserAudit(r.Context(), email, "message.send", "message", "outbound", map[string]any{"from": message.From, "recipient_count": len(message.To), "subject": message.Subject})
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
}

func (h handler) draft(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req, err := parseOutboundRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	from, ok := h.authorizedSender(w, r, email, req.From)
	if !ok {
		return
	}
	message := OutboundMessage{From: from, To: req.To, Subject: req.Subject, Body: req.Body, BodyHTML: req.BodyHTML, Attachments: req.Attachments, DraftID: req.DraftID}
	id, err := h.store.SaveDraftMessage(r.Context(), message)
	if err != nil {
		log.Printf("webmail draft save failed from=%q: %v", message.From, err)
		writeError(w, http.StatusInternalServerError, "save draft failed")
		return
	}
	h.recordUserAudit(r.Context(), email, "message.draft_save", "message", id, map[string]any{"from": message.From, "subject": message.Subject})
	writeJSON(w, http.StatusCreated, map[string]string{"status": "saved", "id": id})
}

const maxSendRequestBytes = 30 << 20

func parseOutboundRequest(w http.ResponseWriter, r *http.Request) (OutboundMessage, error) {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxSendRequestBytes)
		if err := r.ParseMultipartForm(maxSendRequestBytes); err != nil {
			return OutboundMessage{}, errors.New("invalid multipart form")
		}
		message := OutboundMessage{
			From:     strings.TrimSpace(r.FormValue("from")),
			Subject:  strings.TrimSpace(r.FormValue("subject")),
			Body:     strings.TrimSpace(r.FormValue("body")),
			BodyHTML: strings.TrimSpace(r.FormValue("body_html")),
			DraftID:  strings.TrimSpace(r.FormValue("draft_id")),
		}
		recipients, err := parseRecipients(r.MultipartForm.Value["to"])
		if err != nil {
			return OutboundMessage{}, err
		}
		message.To = recipients
		attachments, err := readOutboundAttachments(r)
		if err != nil {
			return OutboundMessage{}, err
		}
		message.Attachments = attachments
		return message, nil
	}
	var req OutboundMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return OutboundMessage{}, errors.New("invalid json")
	}
	req.From = strings.TrimSpace(req.From)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Body = strings.TrimSpace(req.Body)
	req.BodyHTML = strings.TrimSpace(req.BodyHTML)
	req.DraftID = strings.TrimSpace(req.DraftID)
	recipients, err := normalizeRecipients(req.To)
	if err != nil {
		return OutboundMessage{}, err
	}
	req.To = recipients
	return req, nil
}

func parseRecipients(values []string) ([]string, error) {
	recipients := make([]string, 0, len(values))
	for _, value := range values {
		recipients = append(recipients, strings.Split(value, ",")...)
	}
	return normalizeRecipients(recipients)
}

func normalizeRecipients(values []string) ([]string, error) {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		address, ok := normalizeRecipient(value)
		if !ok {
			if strings.TrimSpace(value) == "" {
				continue
			}
			return nil, errors.New("valid recipient email is required")
		}
		seenKey := strings.ToLower(address)
		if seen[seenKey] {
			continue
		}
		seen[seenKey] = true
		out = append(out, address)
	}
	return out, nil
}

func normalizeRecipient(value string) (string, bool) {
	parsed, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil || parsed.Address == "" {
		return "", false
	}
	if strings.ContainsAny(parsed.Address, "\r\n\x00") {
		return "", false
	}
	parts := strings.Split(parsed.Address, "@")
	if len(parts) != 2 {
		return "", false
	}
	local := strings.TrimSpace(parts[0])
	domain := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(parts[1]), "."))
	if local == "" || domain == "" || strings.Contains(local, "..") || strings.Contains(domain, "..") {
		return "", false
	}
	if len(local) > 64 || len(domain) > 253 || strings.ContainsAny(local, `()<>,;:\"[] `) || !validRecipientDomain(domain) {
		return "", false
	}
	return local + "@" + domain, true
}

func validRecipientDomain(domain string) bool {
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if !validRecipientDomainLabel(label) {
			return false
		}
	}
	return true
}

func validRecipientDomainLabel(label string) bool {
	if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for _, r := range label {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}

func readOutboundAttachments(r *http.Request) ([]OutboundAttachment, error) {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, nil
	}
	files := append([]*multipart.FileHeader{}, r.MultipartForm.File["attachments"]...)
	files = append(files, r.MultipartForm.File["file"]...)
	if len(files) > 10 {
		return nil, errors.New("maximum 10 attachments are allowed")
	}
	attachments := make([]OutboundAttachment, 0, len(files))
	var total int64
	for _, header := range files {
		if header == nil || header.Size == 0 {
			continue
		}
		total += header.Size
		if total > 25<<20 {
			return nil, errors.New("attachments exceed 25 MB")
		}
		file, err := header.Open()
		if err != nil {
			return nil, errors.New("read attachment failed")
		}
		data, readErr := io.ReadAll(io.LimitReader(file, header.Size+1))
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			return nil, errors.New("read attachment failed")
		}
		contentType := header.Header.Get("Content-Type")
		attachments = append(attachments, OutboundAttachment{
			Filename:    sanitizeAttachmentFilename(header.Filename),
			ContentType: sanitizeAttachmentContentType(contentType, data),
			Data:        data,
			Size:        int64(len(data)),
		})
	}
	return attachments, nil
}

func sanitizeAttachmentFilename(filename string) string {
	filename = filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	filename = strings.TrimSpace(sanitizeHeader(filename))
	if filename == "." || filename == "/" || filename == "" {
		return "attachment"
	}
	var builder strings.Builder
	for _, r := range filename {
		if r < 32 || r == 127 || r == 0x202a || r == 0x202b || r == 0x202d || r == 0x202e || r == 0x2066 || r == 0x2067 || r == 0x2068 || r == 0x2069 {
			continue
		}
		builder.WriteRune(r)
	}
	cleaned := strings.TrimSpace(builder.String())
	if cleaned == "" {
		return "attachment"
	}
	if runes := []rune(cleaned); len(runes) > 180 {
		cleaned = string(runes[:180])
	}
	return cleaned
}

func sanitizeAttachmentContentType(value string, data []byte) string {
	if strings.ContainsAny(value, "\r\n\x00") {
		value = ""
	} else {
		value = sanitizeHeader(value)
	}
	if mediaType, _, err := mime.ParseMediaType(value); err == nil && safeAttachmentMediaType(mediaType) {
		return strings.ToLower(mediaType)
	}
	if detected := http.DetectContentType(data); detected != "" {
		if mediaType, _, err := mime.ParseMediaType(detected); err == nil && safeAttachmentMediaType(mediaType) {
			return strings.ToLower(mediaType)
		}
	}
	return "application/octet-stream"
}

func safeAttachmentMediaType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	major, sub, ok := strings.Cut(value, "/")
	if !ok || major == "" || sub == "" || len(value) > 127 {
		return false
	}
	return safeAttachmentMediaToken(major) && safeAttachmentMediaToken(sub)
}

func safeAttachmentMediaToken(value string) bool {
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '!' || r == '#' || r == '$' || r == '&' || r == '-' || r == '^' || r == '_' || r == '.' || r == '+' {
			continue
		}
		return false
	}
	return true
}

func (h handler) folders(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	requireManage := r.Method != http.MethodGet
	targetEmail, ok := h.authorizedMailbox(w, r, email, requireManage)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		folders, err := h.store.ListFolders(r.Context(), targetEmail)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list folders failed")
			return
		}
		writeJSON(w, http.StatusOK, folders)
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		folder, err := h.store.CreateFolder(r.Context(), targetEmail, req.Name)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.folder_create", "folder", folder.ID, map[string]any{"mailbox": targetEmail, "name": folder.Name})
		writeJSON(w, http.StatusCreated, folder)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) folder(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	targetEmail, ok := h.authorizedMailbox(w, r, email, true)
	if !ok {
		return
	}
	name := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/folders/"), "/")
	if decoded, err := url.PathUnescape(name); err == nil {
		name = decoded
	}
	if name == "" {
		writeError(w, http.StatusBadRequest, "folder name is required")
		return
	}
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.store.DeleteFolder(r.Context(), targetEmail, name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.recordUserAudit(r.Context(), email, "webmail.folder_delete", "folder", name, map[string]any{"mailbox": targetEmail, "name": name})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) filters(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		filters, err := h.store.ListFilters(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list filters failed")
			return
		}
		writeJSON(w, http.StatusOK, filters)
	case http.MethodPost:
		var req MailFilter
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		filter, err := h.store.CreateFilter(r.Context(), email, normalizeFilter(req))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.filter_create", "filter", filter.ID, map[string]any{"name": filter.Name, "field": filter.Field, "action": filter.Action, "folder": filter.Folder})
		writeJSON(w, http.StatusCreated, filter)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) filter(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/filters/"), "/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "filter id is required")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req MailFilter
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		filter, err := h.store.UpdateFilter(r.Context(), email, id, normalizeFilter(req))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.filter_update", "filter", filter.ID, map[string]any{"name": filter.Name, "field": filter.Field, "action": filter.Action, "folder": filter.Folder, "enabled": filter.Enabled})
		writeJSON(w, http.StatusOK, filter)
	case http.MethodDelete:
		if err := h.store.DeleteFilter(r.Context(), email, id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.filter_delete", "filter", id, map[string]any{"filter_id": id})
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "PUT, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func normalizeFilter(filter MailFilter) MailFilter {
	filter.Name = strings.TrimSpace(filter.Name)
	filter.Field = strings.ToLower(strings.TrimSpace(filter.Field))
	filter.Operator = strings.ToLower(strings.TrimSpace(filter.Operator))
	filter.Value = strings.TrimSpace(filter.Value)
	filter.Action = strings.ToLower(strings.TrimSpace(filter.Action))
	filter.Folder = strings.TrimSpace(filter.Folder)
	if filter.Field == "" {
		filter.Field = "subject"
	}
	if filter.Operator == "" {
		filter.Operator = "contains"
	}
	if filter.Action == "" {
		filter.Action = "move"
	}
	return filter
}

func (h handler) contacts(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		contacts, err := h.store.ListContacts(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list contacts failed")
			return
		}
		writeJSON(w, http.StatusOK, contacts)
	case http.MethodPost:
		var req Contact
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" {
			writeError(w, http.StatusBadRequest, "name and email are required")
			return
		}
		contact, err := h.store.CreateContact(r.Context(), email, Contact{Name: strings.TrimSpace(req.Name), Email: strings.TrimSpace(req.Email)})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create contact failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.contact_create", "contact", contact.ID, map[string]any{"name": contact.Name, "contact_email": contact.Email})
		writeJSON(w, http.StatusCreated, contact)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) contact(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/contacts/"), "/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "contact id is required")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req Contact
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" {
			writeError(w, http.StatusBadRequest, "name and email are required")
			return
		}
		contact, err := h.store.UpdateContact(r.Context(), email, id, Contact{Name: strings.TrimSpace(req.Name), Email: strings.TrimSpace(req.Email)})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "update contact failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.contact_update", "contact", contact.ID, map[string]any{"name": contact.Name, "contact_email": contact.Email})
		writeJSON(w, http.StatusOK, contact)
	case http.MethodDelete:
		if err := h.store.DeleteContact(r.Context(), email, id); err != nil {
			writeError(w, http.StatusInternalServerError, "delete contact failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.contact_delete", "contact", id, map[string]any{"contact_id": id})
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "PUT, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) calendar(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		events, err := h.store.ListCalendarEvents(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list calendar failed")
			return
		}
		writeJSON(w, http.StatusOK, events)
	case http.MethodPost:
		var req CalendarEvent
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.Title) == "" || req.StartsAt.IsZero() || req.EndsAt.IsZero() {
			writeError(w, http.StatusBadRequest, "title, starts_at, and ends_at are required")
			return
		}
		event, err := h.store.CreateCalendarEvent(r.Context(), email, CalendarEvent{Title: strings.TrimSpace(req.Title), StartsAt: req.StartsAt, EndsAt: req.EndsAt})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create calendar event failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.calendar_create", "calendar_event", event.ID, map[string]any{"title": event.Title})
		writeJSON(w, http.StatusCreated, event)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) calendarEvent(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/calendar/"), "/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "event id is required")
		return
	}
	switch r.Method {
	case http.MethodPut:
		var req CalendarEvent
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if strings.TrimSpace(req.Title) == "" || req.StartsAt.IsZero() || req.EndsAt.IsZero() {
			writeError(w, http.StatusBadRequest, "title, starts_at, and ends_at are required")
			return
		}
		event, err := h.store.UpdateCalendarEvent(r.Context(), email, id, CalendarEvent{Title: strings.TrimSpace(req.Title), StartsAt: req.StartsAt, EndsAt: req.EndsAt})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "update calendar event failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.calendar_update", "calendar_event", event.ID, map[string]any{"title": event.Title})
		writeJSON(w, http.StatusOK, event)
	case http.MethodDelete:
		if err := h.store.DeleteCalendarEvent(r.Context(), email, id); err != nil {
			writeError(w, http.StatusInternalServerError, "delete calendar event failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.calendar_delete", "calendar_event", id, map[string]any{"event_id": id})
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "PUT, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) profile(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		profile, err := h.store.GetProfile(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "load profile failed")
			return
		}
		writeJSON(w, http.StatusOK, profile)
	case http.MethodPut:
		var req struct {
			SignatureHTML    string `json:"signature_html"`
			SignatureAutoAdd bool   `json:"signature_auto_add"`
			Language         string `json:"language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		req.SignatureHTML = strings.TrimSpace(req.SignatureHTML)
		if len(req.SignatureHTML) > 20000 {
			writeError(w, http.StatusBadRequest, "signature is too long")
			return
		}
		req.Language = i18n.NormalizeLanguage(req.Language)
		if req.Language == "" {
			writeError(w, http.StatusBadRequest, "language is not supported")
			return
		}
		profile, err := h.store.UpdateProfile(r.Context(), email, UserProfile{
			Email:            email,
			SignatureHTML:    req.SignatureHTML,
			SignatureAutoAdd: req.SignatureAutoAdd,
			Language:         req.Language,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "update profile failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.profile_update", "user", email, map[string]any{"changed": "signature_settings,language"})
		writeJSON(w, http.StatusOK, profile)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) changePassword(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	valid, err := h.store.VerifyUserPassword(r.Context(), email, req.CurrentPassword)
	if err != nil || !valid {
		h.recordUserAudit(r.Context(), email, "webmail.password_change_failed", "user", email, map[string]any{"email": email})
		writeError(w, http.StatusUnauthorized, "current password is invalid")
		return
	}
	if len(req.NewPassword) < 12 {
		writeError(w, http.StatusBadRequest, "new password must be at least 12 characters")
		return
	}
	if err := h.store.ChangePassword(r.Context(), email, req.NewPassword); err != nil {
		writeError(w, http.StatusInternalServerError, "change password failed")
		return
	}
	if h.sessions != nil {
		h.sessions.InvalidateSubject(email, "webmail")
		h.sessions.Clear(w, r)
	}
	h.recordUserAudit(r.Context(), email, "webmail.password_change", "user", email, map[string]any{"email": email})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) contentTrust(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		entries, err := h.store.ListContentTrust(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "load content trust failed")
			return
		}
		writeJSON(w, http.StatusOK, entries)
	case http.MethodPost:
		var req ContentTrustEntry
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		entry, err := normalizeContentTrustEntry(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		created, err := h.store.AddContentTrust(r.Context(), email, entry)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "add content trust failed")
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.content_trust_add", "content_trust", created.Scope+":"+created.Value, map[string]any{"scope": created.Scope, "value": created.Value})
		writeJSON(w, http.StatusCreated, created)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

var publicMailHostingDomains = map[string]struct{}{
	"aol.com":        {},
	"azet.sk":        {},
	"centrum.sk":     {},
	"fastmail.com":   {},
	"gmail.com":      {},
	"gmx.com":        {},
	"gmx.net":        {},
	"google.com":     {},
	"googlemail.com": {},
	"hotmail.com":    {},
	"icloud.com":     {},
	"laposte.net":    {},
	"libero.it":      {},
	"live.com":       {},
	"mail.com":       {},
	"mail.ru":        {},
	"me.com":         {},
	"msn.com":        {},
	"outlook.com":    {},
	"outlook.xyz":    {},
	"pm.me":          {},
	"post.cz":        {},
	"proton.me":      {},
	"protonmail.com": {},
	"seznam.cz":      {},
	"t-online.de":    {},
	"tutanota.com":   {},
	"web.de":         {},
	"yahoo.com":      {},
	"yandex.com":     {},
	"yandex.ru":      {},
	"ymail.com":      {},
	"zoho.com":       {},
}

func normalizeContentTrustEntry(entry ContentTrustEntry) (ContentTrustEntry, error) {
	scope := strings.ToLower(strings.TrimSpace(entry.Scope))
	switch scope {
	case "sender":
		value := normalizeTrustSender(entry.Value)
		if value == "" {
			return ContentTrustEntry{}, errors.New("valid sender email is required")
		}
		return ContentTrustEntry{Scope: scope, Value: value}, nil
	case "domain":
		value := normalizeTrustDomain(entry.Value)
		if value == "" {
			return ContentTrustEntry{}, errors.New("valid sender domain is required")
		}
		if isPublicHostingTrustDomain(value) {
			return ContentTrustEntry{}, errors.New("domain trust is not allowed for public mail hosting services; trust this specific sender instead")
		}
		return ContentTrustEntry{Scope: scope, Value: value}, nil
	default:
		return ContentTrustEntry{}, errors.New("scope must be sender or domain")
	}
}

func normalizeTrustSender(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return ""
	}
	if address, err := mail.ParseAddress(value); err == nil {
		value = address.Address
	}
	value = strings.ToLower(strings.Trim(value, `"' `))
	if strings.ContainsAny(value, " <>") {
		return ""
	}
	at := strings.LastIndex(value, "@")
	if at <= 0 || at == len(value)-1 {
		return ""
	}
	local := value[:at]
	domain := normalizeTrustDomain(value[at+1:])
	if local == "" || domain == "" || len(value) > 254 {
		return ""
	}
	return local + "@" + domain
}

func normalizeTrustDomain(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "@")
	value = strings.TrimSuffix(value, ".")
	if at := strings.LastIndex(value, "@"); at >= 0 {
		value = value[at+1:]
	}
	if value == "" || len(value) > 253 || !strings.Contains(value, ".") || strings.ContainsAny(value, " /\\:\r\n\t") {
		return ""
	}
	labels := strings.Split(value, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return ""
		}
		for _, char := range label {
			if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
				continue
			}
			return ""
		}
	}
	return value
}

func isPublicHostingTrustDomain(domain string) bool {
	domain = normalizeTrustDomain(domain)
	if domain == "" {
		return false
	}
	for provider := range publicMailHostingDomains {
		if domain == provider || strings.HasSuffix(domain, "."+provider) {
			return true
		}
	}
	return false
}

func (h handler) authorized(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.store == nil {
		writeUnauthorized(w)
		return "", false
	}
	if h.sessions != nil {
		if safeMethod(r.Method) {
			if session, ok := h.sessions.Validate(r); ok && session.Kind == "webmail" {
				return session.Subject, true
			}
			writeSessionUnauthorized(w)
			return "", false
		} else {
			session, ok := h.sessions.ValidateUnsafe(r)
			if !ok || session.Kind != "webmail" {
				http.Error(w, "csrf required", http.StatusForbidden)
				return "", false
			}
			return session.Subject, true
		}
	}
	email, password, ok := r.BasicAuth()
	if !ok || email == "" || password == "" {
		writeUnauthorized(w)
		return "", false
	}
	email = strings.ToLower(email)
	valid, err := h.store.VerifyUserPassword(r.Context(), email, password)
	if err != nil || !valid {
		writeUnauthorized(w)
		return "", false
	}
	return email, true
}

func (h handler) authorizedMailbox(w http.ResponseWriter, r *http.Request, actorEmail string, requireManage bool) (string, bool) {
	requested := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mailbox")))
	if requested == "" || requested == strings.ToLower(actorEmail) {
		return actorEmail, true
	}
	mailboxes, err := h.store.ListMailboxes(r.Context(), actorEmail)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list mailboxes failed")
		return "", false
	}
	for _, mailbox := range mailboxes {
		address := strings.ToLower(strings.TrimSpace(mailbox.Address))
		id := strings.ToLower(strings.TrimSpace(mailbox.ID))
		if requested != address && requested != id {
			continue
		}
		if !mailbox.CanRead || (requireManage && !mailbox.CanManage) {
			writeError(w, http.StatusForbidden, "mailbox permission denied")
			return "", false
		}
		return address, true
	}
	writeError(w, http.StatusForbidden, "mailbox permission denied")
	return "", false
}

func (h handler) authorizedSender(w http.ResponseWriter, r *http.Request, actorEmail, requestedSender string) (string, bool) {
	requested := strings.ToLower(strings.TrimSpace(requestedSender))
	if requested == "" || requested == strings.ToLower(actorEmail) {
		return actorEmail, true
	}
	mailboxes, err := h.store.ListMailboxes(r.Context(), actorEmail)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list mailboxes failed")
		return "", false
	}
	for _, mailbox := range mailboxes {
		address := strings.ToLower(strings.TrimSpace(mailbox.Address))
		id := strings.ToLower(strings.TrimSpace(mailbox.ID))
		if requested != address && requested != id {
			continue
		}
		if !mailbox.CanSendAs {
			writeError(w, http.StatusForbidden, "sender permission denied")
			return "", false
		}
		if address != "" {
			return address, true
		}
		return id, true
	}
	writeError(w, http.StatusForbidden, "sender permission denied")
	return "", false
}

func (h handler) session(w http.ResponseWriter, r *http.Request) {
	if h.sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	switch r.Method {
	case http.MethodGet:
		current, ok := h.sessions.Validate(r)
		if !ok || current.Kind != "webmail" {
			writeSessionUnauthorized(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"csrf_token": current.CSRFToken, "email": current.Subject})
	case http.MethodPost:
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		email := strings.ToLower(strings.TrimSpace(req.Email))
		keys := loginKeys("webmail", email, r)
		if session.AnyLocked(h.limiter, keys) {
			h.recordUserAudit(r.Context(), email, "webmail.login_locked", "user", email, map[string]any{"email": email})
			writeError(w, http.StatusTooManyRequests, "login temporarily locked")
			return
		}
		valid, err := h.store.VerifyUserPassword(r.Context(), email, req.Password)
		if err != nil || !valid {
			session.FailAll(h.limiter, keys)
			h.recordUserAudit(r.Context(), email, "webmail.login_failed", "user", email, map[string]any{"email": email})
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		session.SuccessAll(h.limiter, keys)
		if mfaStore, ok := h.store.(MailboxMFAStore); ok {
			securityState, err := mfaStore.GetMailboxSecurity(r.Context(), email)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "mailbox security unavailable")
				return
			}
			if securityState.MFAAvailable {
				proIdentitySettings, ok, err := h.proIdentitySettings(r.Context())
				if err != nil {
					writeError(w, http.StatusInternalServerError, "proidentity auth settings unavailable")
					return
				}
				if ok {
					response, err := h.beginMailboxProIdentityMFA(r.Context(), r, email, proIdentitySettings)
					if err != nil {
						if writeProIdentityMailboxAPIError(w, err) {
							return
						}
						writeError(w, http.StatusBadGateway, err.Error())
						return
					}
					h.recordUserAudit(r.Context(), email, "webmail.mfa_required", "user", email, map[string]any{"email": email, "provider": "proidentity"})
					writeJSON(w, http.StatusOK, response)
					return
				}
			}
			switch {
			case securityState.MFAAvailable && securityState.ForceMFA && !securityState.MFAEnabled:
				challenge, err := h.createMailboxMFAChallenge(r.Context(), email, "setup")
				if err != nil {
					writeError(w, http.StatusInternalServerError, "create mfa challenge failed")
					return
				}
				h.recordUserAudit(r.Context(), email, "webmail.mfa_setup_required", "user", email, map[string]any{"email": email})
				writeJSON(w, http.StatusOK, map[string]any{
					"mfa_setup_required": true,
					"provider":           "totp",
					"mfa_token":          challenge.Token,
					"email":              email,
				})
				return
			case securityState.MFAEnabled && securityState.TOTPEnabled:
				challenge, err := h.createMailboxMFAChallenge(r.Context(), email, "login")
				if err != nil {
					writeError(w, http.StatusInternalServerError, "create mfa challenge failed")
					return
				}
				h.recordUserAudit(r.Context(), email, "webmail.mfa_required", "user", email, map[string]any{"email": email})
				writeJSON(w, http.StatusOK, map[string]any{
					"mfa_required": true,
					"provider":     "totp",
					"mfa_token":    challenge.Token,
					"email":        email,
				})
				return
			}
		}
		created, err := h.sessions.Create(r, email, "webmail")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create session failed")
			return
		}
		http.SetCookie(w, created.Cookie)
		h.recordUserAudit(r.Context(), email, "webmail.login", "user", email, map[string]any{"email": email})
		writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken, "email": email})
	case http.MethodDelete:
		if current, ok := h.sessions.Validate(r); ok && current.Kind == "webmail" {
			h.recordUserAudit(r.Context(), current.Subject, "webmail.logout", "user", current.Subject, map[string]any{"email": current.Subject})
		}
		h.sessions.Clear(w, r)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) sessionMFA(w http.ResponseWriter, r *http.Request) {
	if h.sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mfaStore, ok := h.store.(MailboxMFAStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "mailbox mfa unavailable")
		return
	}
	var req struct {
		Token string `json:"mfa_token"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	challenge, err := mfaStore.GetMailboxMFAChallenge(r.Context(), req.Token)
	if err != nil || challenge.Purpose != "login" || time.Now().After(challenge.ExpiresAt) {
		writeError(w, http.StatusUnauthorized, "mfa challenge expired")
		return
	}
	switch challenge.Provider {
	case "proidentity":
		settings, ok, err := h.proIdentitySettings(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "proidentity auth settings unavailable")
			return
		}
		if !ok {
			writeError(w, http.StatusServiceUnavailable, "proidentity auth unavailable")
			return
		}
		if strings.TrimSpace(req.Code) != "" {
			verified, err := verifyMailboxProIdentityTOTP(r.Context(), settings, challenge.Email, req.Code)
			if err != nil {
				if writeProIdentityMailboxAPIError(w, err) {
					return
				}
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			if !verified {
				h.recordUserAudit(r.Context(), challenge.Email, "webmail.mfa_failed", "user", challenge.Email, map[string]any{"email": challenge.Email, "provider": "proidentity_totp"})
				writeError(w, http.StatusUnauthorized, "invalid proidentity totp code")
				return
			}
			break
		}
		status, err := readMailboxProIdentityAuthRequestStatus(r.Context(), settings, challenge.RequestID)
		if err != nil {
			if writeProIdentityMailboxAPIError(w, err) {
				return
			}
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		switch status.Status {
		case "approved":
		case "pending":
			writeJSON(w, http.StatusAccepted, map[string]any{"mfa_required": true, "provider": "proidentity", "status": "pending", "request_id": challenge.RequestID, "expires_at": challenge.ExpiresAt.Unix(), "email": challenge.Email})
			return
		case "denied", "expired":
			_ = mfaStore.DeleteMailboxMFAChallenge(r.Context(), req.Token)
			h.recordUserAudit(r.Context(), challenge.Email, "webmail.mfa_failed", "user", challenge.Email, map[string]any{"email": challenge.Email, "provider": "proidentity", "status": status.Status})
			writeError(w, http.StatusUnauthorized, "proidentity auth "+status.Status)
			return
		default:
			writeError(w, http.StatusBadGateway, "unexpected proidentity auth status")
			return
		}
	default:
		valid, err := mfaStore.VerifyMailboxTOTPCode(r.Context(), challenge.Email, req.Code)
		if err != nil || !valid {
			h.recordUserAudit(r.Context(), challenge.Email, "webmail.mfa_failed", "user", challenge.Email, map[string]any{"email": challenge.Email})
			writeError(w, http.StatusUnauthorized, "invalid authenticator code")
			return
		}
	}
	_ = mfaStore.DeleteMailboxMFAChallenge(r.Context(), req.Token)
	created, err := h.sessions.Create(r, challenge.Email, "webmail")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}
	http.SetCookie(w, created.Cookie)
	h.recordUserAudit(r.Context(), challenge.Email, "webmail.mfa_login", "user", challenge.Email, map[string]any{"email": challenge.Email})
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken, "email": challenge.Email})
}

func (h handler) mailboxMFA(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	mfaStore, ok := h.store.(MailboxMFAStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "mailbox mfa unavailable")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	state, err := mfaStore.GetMailboxSecurity(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mailbox security unavailable")
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h handler) mailboxTOTPEnroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mfaStore, ok := h.store.(MailboxMFAStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "mailbox mfa unavailable")
		return
	}
	var req struct {
		Token string `json:"mfa_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var email string
	if strings.TrimSpace(req.Token) != "" {
		challenge, err := mfaStore.GetMailboxMFAChallenge(r.Context(), req.Token)
		if err != nil || challenge.Purpose != "setup" || time.Now().After(challenge.ExpiresAt) {
			writeError(w, http.StatusUnauthorized, "mfa challenge expired")
			return
		}
		email = challenge.Email
	} else {
		var authorized bool
		email, authorized = h.authorized(w, r)
		if !authorized {
			return
		}
	}
	enrollment, err := mfaStore.BeginMailboxTOTPEnrollment(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create totp enrollment failed")
		return
	}
	h.recordUserAudit(r.Context(), email, "webmail.mfa_totp_enroll", "user", email, map[string]any{"email": email})
	writeJSON(w, http.StatusOK, enrollment)
}

func (h handler) mailboxTOTPVerify(w http.ResponseWriter, r *http.Request) {
	if h.sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mfaStore, ok := h.store.(MailboxMFAStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "mailbox mfa unavailable")
		return
	}
	var req struct {
		Token string `json:"mfa_token"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var email string
	var loginChallenge bool
	if strings.TrimSpace(req.Token) != "" {
		challenge, err := mfaStore.GetMailboxMFAChallenge(r.Context(), req.Token)
		if err != nil || challenge.Purpose != "setup" || time.Now().After(challenge.ExpiresAt) {
			writeError(w, http.StatusUnauthorized, "mfa challenge expired")
			return
		}
		email = challenge.Email
		loginChallenge = true
	} else {
		var authorized bool
		email, authorized = h.authorized(w, r)
		if !authorized {
			return
		}
	}
	securityState, err := mfaStore.VerifyMailboxTOTPEnrollment(r.Context(), email, req.Code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid authenticator code")
		return
	}
	if !loginChallenge {
		h.recordUserAudit(r.Context(), email, "webmail.mfa_totp_enabled", "user", email, map[string]any{"email": email})
		writeJSON(w, http.StatusOK, securityState)
		return
	}
	_ = mfaStore.DeleteMailboxMFAChallenge(r.Context(), req.Token)
	created, err := h.sessions.Create(r, email, "webmail")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}
	http.SetCookie(w, created.Cookie)
	h.recordUserAudit(r.Context(), email, "webmail.mfa_totp_enabled", "user", email, map[string]any{"email": email})
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken, "email": email})
}

func (h handler) appPasswords(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	store, ok := h.store.(AppPasswordStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "app passwords unavailable")
		return
	}
	switch r.Method {
	case http.MethodGet:
		passwords, err := store.ListAppPasswords(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list app passwords failed")
			return
		}
		writeJSON(w, http.StatusOK, passwords)
	case http.MethodPost:
		var req AppPassword
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		created, err := store.CreateAppPassword(r.Context(), email, req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.recordUserAudit(r.Context(), email, "webmail.app_password_create", "user", email, map[string]any{"name": created.Name, "protocols": created.Protocols})
		writeJSON(w, http.StatusCreated, created)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h handler) appPassword(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	store, ok := h.store.(AppPasswordStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "app passwords unavailable")
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/app-passwords/"), "/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := store.RevokeAppPassword(r.Context(), email, id); err != nil {
		writeError(w, http.StatusNotFound, "app password not found")
		return
	}
	h.recordUserAudit(r.Context(), email, "webmail.app_password_revoke", "user", email, map[string]any{"app_password_id": id})
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) emailForMFASetup(w http.ResponseWriter, r *http.Request, purpose string) (string, bool) {
	var req struct {
		Token string `json:"mfa_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return "", false
	}
	mfaStore, ok := h.store.(MailboxMFAStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "mailbox mfa unavailable")
		return "", false
	}
	challenge, err := mfaStore.GetMailboxMFAChallenge(r.Context(), req.Token)
	if err != nil || challenge.Purpose != purpose || time.Now().After(challenge.ExpiresAt) {
		writeError(w, http.StatusUnauthorized, "mfa challenge expired")
		return "", false
	}
	return challenge.Email, true
}

func (h handler) createMailboxMFAChallenge(ctx context.Context, email, purpose string) (MailboxMFAChallenge, error) {
	return h.createMailboxMFAChallengeWithProvider(ctx, email, purpose, "totp", "", time.Now().Add(10*time.Minute))
}

func (h handler) createMailboxMFAChallengeWithProvider(ctx context.Context, email, purpose, provider, requestID string, expiresAt time.Time) (MailboxMFAChallenge, error) {
	mfaStore, ok := h.store.(MailboxMFAStore)
	if !ok {
		return MailboxMFAChallenge{}, errors.New("mailbox mfa unavailable")
	}
	token, err := randomURLToken(32)
	if err != nil {
		return MailboxMFAChallenge{}, err
	}
	challenge := MailboxMFAChallenge{
		Token:     token,
		Email:     strings.ToLower(strings.TrimSpace(email)),
		Purpose:   purpose,
		Provider:  normalizeMFAProvider(provider),
		RequestID: strings.TrimSpace(requestID),
		ExpiresAt: expiresAt.UTC(),
	}
	if challenge.ExpiresAt.IsZero() {
		challenge.ExpiresAt = time.Now().Add(10 * time.Minute).UTC()
	}
	if err := mfaStore.CreateMailboxMFAChallenge(ctx, challenge); err != nil {
		return MailboxMFAChallenge{}, err
	}
	return challenge, nil
}

func normalizeMFAProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "proidentity":
		return "proidentity"
	default:
		return "totp"
	}
}

func (h handler) proIdentitySettings(ctx context.Context) (domain.AdminMFASettings, bool, error) {
	store, ok := h.store.(ProIdentitySettingsStore)
	if !ok {
		return domain.AdminMFASettings{}, false, nil
	}
	settings, err := store.GetAdminMFASettings(ctx)
	if err != nil {
		return domain.AdminMFASettings{}, false, err
	}
	if !mailboxProIdentityConfigured(settings) {
		return settings, false, nil
	}
	return settings, true, nil
}

func mailboxProIdentityConfigured(settings domain.AdminMFASettings) bool {
	return settings.ProIdentityEnabled &&
		strings.TrimSpace(mailboxProIdentityRequestBaseURL(settings)) != "" &&
		strings.TrimSpace(settings.ProIdentityAPIKey) != ""
}

func (h handler) beginMailboxProIdentityMFA(ctx context.Context, r *http.Request, email string, settings domain.AdminMFASettings) (map[string]any, error) {
	created, err := createMailboxProIdentityAuthRequest(ctx, settings, proIdentityAuthRequestPayload{
		UserEmail:     email,
		DisplayName:   email,
		ContextTitle:  "Sign in to ProIdentity Mail",
		ContextDetail: "Webmail browser session approval",
		ClientIP:      requestClientIP(r),
	})
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(mailboxProIdentityTimeout(settings))
	if !created.ExpiresAt.IsZero() && created.ExpiresAt.Before(expiresAt) {
		expiresAt = created.ExpiresAt
	}
	challenge, err := h.createMailboxMFAChallengeWithProvider(ctx, email, "login", "proidentity", created.RequestID, expiresAt)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"mfa_required": true,
		"provider":     "proidentity",
		"providers":    []string{"proidentity_push", "proidentity_totp"},
		"mfa_token":    challenge.Token,
		"request_id":   created.RequestID,
		"status":       "pending",
		"expires_at":   expiresAt.Unix(),
		"email":        strings.ToLower(strings.TrimSpace(email)),
	}, nil
}

func mailboxProIdentityTimeout(settings domain.AdminMFASettings) time.Duration {
	seconds := settings.ProIdentityTimeoutSeconds
	if seconds <= 0 {
		seconds = 90
	}
	if seconds < 30 {
		seconds = 30
	}
	if seconds > 300 {
		seconds = 300
	}
	return time.Duration(seconds) * time.Second
}

func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func loginKeys(service, subject string, r *http.Request) []string {
	subject = strings.ToLower(strings.TrimSpace(subject))
	host := strings.ToLower(strings.TrimSpace(requestClientIP(r)))
	return []string{
		service + "|ip|" + host,
		service + "|account|" + subject,
		service + "|pair|" + subject + "|" + host,
	}
}

func randomURLToken(bytesLen int) (string, error) {
	if bytesLen <= 0 {
		bytesLen = 32
	}
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func newAppPasswordSecret() (string, error) {
	token, err := randomURLToken(24)
	if err != nil {
		return "", err
	}
	return "pim-" + token, nil
}

func appPasswordFingerprint(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func normalizeAppPasswordProtocols(protocols []string) []string {
	allowed := map[string]bool{"imap": true, "smtp": true, "pop3": true, "dav": true}
	seen := map[string]bool{}
	var out []string
	for _, protocol := range protocols {
		protocol = normalizeProtocol(protocol)
		if !allowed[protocol] || seen[protocol] {
			continue
		}
		seen[protocol] = true
		out = append(out, protocol)
	}
	if len(out) == 0 {
		out = []string{"imap", "smtp", "pop3", "dav"}
	}
	return out
}

func normalizeProtocol(protocol string) string {
	protocol = strings.ToLower(strings.TrimSpace(protocol))
	switch protocol {
	case "submission", "smtps":
		return "smtp"
	case "carddav", "caldav", "groupware":
		return "dav"
	default:
		return protocol
	}
}

func protocolAllowed(protocols []string, protocol string) bool {
	protocol = normalizeProtocol(protocol)
	if protocol == "webmail" || protocol == "" {
		return false
	}
	for _, candidate := range normalizeAppPasswordProtocols(protocols) {
		if candidate == protocol {
			return true
		}
	}
	return false
}

func totpQRCodeDataURL(otpURL string) (string, error) {
	key, err := otp.NewKeyFromURL(otpURL)
	if err != nil {
		return "", err
	}
	image, err := key.Image(220, 220)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, image); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (h handler) recordUserAudit(ctx context.Context, email, action, targetType, targetID string, metadata map[string]any) {
	recorder, ok := h.store.(UserAuditRecorder)
	if !ok {
		return
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["email"] = strings.ToLower(strings.TrimSpace(email))
	_ = recorder.RecordUserAudit(ctx, email, action, targetType, targetID, metadata)
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="ProIdentity Webmail", charset="UTF-8"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func writeSessionUnauthorized(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, "unauthorized")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
