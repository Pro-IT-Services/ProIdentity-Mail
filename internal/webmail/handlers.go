package webmail

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Store interface {
	VerifyUserPassword(ctx context.Context, email, password string) (bool, error)
	ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error)
	GetMessage(ctx context.Context, email, id string) (MessageDetail, error)
}

type handler struct {
	store Store
}

func NewRouter(store Store) http.Handler {
	h := handler{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/api/v1/messages", h.messages)
	mux.HandleFunc("/api/v1/messages/", h.message)
	mux.HandleFunc("/", index)
	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(webmailIndexHTML))
}

func (h handler) messages(w http.ResponseWriter, r *http.Request) {
	email, ok := h.authorized(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	messages, err := h.store.ListRecentMessages(r.Context(), email, limit)
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
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/messages/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "message id is required")
		return
	}
	message, err := h.store.GetMessage(r.Context(), email, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	writeJSON(w, http.StatusOK, message)
}

func (h handler) authorized(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.store == nil {
		writeUnauthorized(w)
		return "", false
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

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="ProIdentity Webmail", charset="UTF-8"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
