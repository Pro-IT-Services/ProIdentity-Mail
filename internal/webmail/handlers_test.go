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
	valid           bool
	sent            OutboundMessage
	reportedEmail   string
	reportedID      string
	reportedVerdict string
	folder          string
	movedEmail      string
	movedID         string
	movedFolder     string
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
