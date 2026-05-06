package webmail

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"proidentity-mail/internal/session"
)

type Store interface {
	VerifyUserPassword(ctx context.Context, email, password string) (bool, error)
	ListRecentMessages(ctx context.Context, email string, limit int) ([]MessageSummary, error)
	ListMessages(ctx context.Context, email, folder string, limit int) ([]MessageSummary, error)
	GetMessage(ctx context.Context, email, id string) (MessageDetail, error)
	SendMessage(ctx context.Context, message OutboundMessage) error
	ReportMessage(ctx context.Context, email, id, verdict string) error
	MoveMessage(ctx context.Context, email, id, folder string) error
	ListContacts(ctx context.Context, email string) ([]Contact, error)
	CreateContact(ctx context.Context, email string, contact Contact) (Contact, error)
	UpdateContact(ctx context.Context, email, id string, contact Contact) (Contact, error)
	DeleteContact(ctx context.Context, email, id string) error
	ListCalendarEvents(ctx context.Context, email string) ([]CalendarEvent, error)
	CreateCalendarEvent(ctx context.Context, email string, event CalendarEvent) (CalendarEvent, error)
	UpdateCalendarEvent(ctx context.Context, email, id string, event CalendarEvent) (CalendarEvent, error)
	DeleteCalendarEvent(ctx context.Context, email, id string) error
	ChangePassword(ctx context.Context, email, newPassword string) error
}

type OutboundMessage struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

type Contact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CalendarEvent struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

type handler struct {
	store    Store
	sessions *session.Manager
	limiter  *session.LoginLimiter
}

func NewRouter(store Store, managers ...*session.Manager) http.Handler {
	var manager *session.Manager
	if len(managers) > 0 {
		manager = managers[0]
	}
	h := handler{store: store, sessions: manager, limiter: session.NewLoginLimiter(session.Options{})}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/api/v1/session", h.session)
	mux.HandleFunc("/api/v1/messages", h.messages)
	mux.HandleFunc("/api/v1/messages/", h.message)
	mux.HandleFunc("/api/v1/send", h.send)
	mux.HandleFunc("/api/v1/contacts", h.contacts)
	mux.HandleFunc("/api/v1/contacts/", h.contact)
	mux.HandleFunc("/api/v1/calendar", h.calendar)
	mux.HandleFunc("/api/v1/calendar/", h.calendarEvent)
	mux.HandleFunc("/api/v1/password", h.changePassword)
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
	folder := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("folder")))
	if folder == "" {
		folder = "inbox"
	}
	messages, err := h.store.ListMessages(r.Context(), email, folder, limit)
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
	if strings.HasSuffix(id, "/report") {
		h.reportMessage(w, r, email, strings.TrimSuffix(id, "/report"))
		return
	}
	if strings.HasSuffix(id, "/move") {
		h.moveMessage(w, r, email, strings.TrimSuffix(id, "/move"))
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
	message, err := h.store.GetMessage(r.Context(), email, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	writeJSON(w, http.StatusOK, message)
}

func (h handler) reportMessage(w http.ResponseWriter, r *http.Request, email, id string) {
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
	if err := h.store.ReportMessage(r.Context(), email, id, req.Verdict); err != nil {
		log.Printf("webmail report failed email=%q id=%q verdict=%q: %v", email, id, req.Verdict, err)
		writeError(w, http.StatusInternalServerError, "report message failed")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "recorded"})
}

func (h handler) moveMessage(w http.ResponseWriter, r *http.Request, email, id string) {
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
	req.Folder = strings.ToLower(strings.TrimSpace(req.Folder))
	if req.Folder != "inbox" && req.Folder != "spam" && req.Folder != "trash" && req.Folder != "archive" {
		writeError(w, http.StatusBadRequest, "folder must be inbox, spam, trash, or archive")
		return
	}
	if err := h.store.MoveMessage(r.Context(), email, id, req.Folder); err != nil {
		log.Printf("webmail move failed email=%q id=%q folder=%q: %v", email, id, req.Folder, err)
		writeError(w, http.StatusInternalServerError, "move message failed")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "moved"})
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
	var req struct {
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Body    string   `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.To) == 0 || strings.TrimSpace(req.Subject) == "" {
		writeError(w, http.StatusBadRequest, "recipient and subject are required")
		return
	}
	message := OutboundMessage{From: email, To: req.To, Subject: req.Subject, Body: req.Body}
	if err := h.store.SendMessage(r.Context(), message); err != nil {
		log.Printf("webmail send failed from=%q to=%q: %v", message.From, message.To, err)
		writeError(w, http.StatusInternalServerError, "send message failed")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "queued"})
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
		writeJSON(w, http.StatusOK, contact)
	case http.MethodDelete:
		if err := h.store.DeleteContact(r.Context(), email, id); err != nil {
			writeError(w, http.StatusInternalServerError, "delete contact failed")
			return
		}
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
		writeJSON(w, http.StatusOK, event)
	case http.MethodDelete:
		if err := h.store.DeleteCalendarEvent(r.Context(), email, id); err != nil {
			writeError(w, http.StatusInternalServerError, "delete calendar event failed")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "PUT, DELETE")
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
	w.WriteHeader(http.StatusNoContent)
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
		} else if _, _, hasBasic := r.BasicAuth(); !hasBasic {
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

func (h handler) session(w http.ResponseWriter, r *http.Request) {
	if h.sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	switch r.Method {
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
		key := loginKey(email, r)
		if h.limiter != nil && h.limiter.Locked(key) {
			writeError(w, http.StatusTooManyRequests, "login temporarily locked")
			return
		}
		valid, err := h.store.VerifyUserPassword(r.Context(), email, req.Password)
		if err != nil || !valid {
			if h.limiter != nil {
				h.limiter.Fail(key)
			}
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		if h.limiter != nil {
			h.limiter.Success(key)
		}
		created, err := h.sessions.Create(r, email, "webmail")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create session failed")
			return
		}
		http.SetCookie(w, created.Cookie)
		writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken, "email": email})
	case http.MethodDelete:
		h.sessions.Clear(w, r)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "POST, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func loginKey(subject string, r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return strings.ToLower(strings.TrimSpace(subject)) + "|" + host
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
