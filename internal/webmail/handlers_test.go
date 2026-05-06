package webmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	handler := NewRouter(&fakeStore{valid: true})
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

type fakeStore struct {
	valid bool
}

func (s *fakeStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	return s.valid && email == "marko@example.com" && password == "secret123456", nil
}

func (s *fakeStore) ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error) {
	return []MessageSummary{{ID: "1", From: "sender@example.net", To: email, Subject: "Welcome", Preview: "Hello"}}, nil
}

func (s *fakeStore) GetMessage(ctx context.Context, email, id string) (MessageDetail, error) {
	return MessageDetail{ID: id, From: "sender@example.net", To: email, Subject: "Welcome", Body: "Full body"}, nil
}
