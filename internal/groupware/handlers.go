package groupware

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrNotFound = errors.New("not found")

type DAVObject struct {
	Href string
	ETag string
	Body []byte
}

type Store interface {
	VerifyUserPassword(ctx context.Context, email, password string) (bool, error)
	PutContact(ctx context.Context, email, href string, body []byte) (DAVObject, error)
	GetContact(ctx context.Context, email, href string) (DAVObject, error)
	PutCalendarObject(ctx context.Context, email, href string, body []byte) (DAVObject, error)
	GetCalendarObject(ctx context.Context, email, href string) (DAVObject, error)
}

type handler struct {
	store Store
}

func NewRouter(store Store) http.Handler {
	h := handler{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/dav/", h.dav)
	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
}

func (h handler) dav(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("DAV", "1, 3, calendar-access, addressbook")
	w.Header().Set("MS-Author-Via", "DAV")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case "PROPFIND":
		if _, ok := h.authorized(w, r); !ok {
			return
		}
		writeMultiStatus(w, r.URL.Path)
	case http.MethodPut:
		email, ok := h.authorized(w, r)
		if !ok {
			return
		}
		h.putObject(w, r, email)
	case http.MethodGet:
		email, ok := h.authorized(w, r)
		if !ok {
			return
		}
		h.getObject(w, r, email)
	default:
		w.Header().Set("Allow", "OPTIONS, PROPFIND, GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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
	w.Header().Set("WWW-Authenticate", `Basic realm="ProIdentity DAV", charset="UTF-8"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func (h handler) putObject(w http.ResponseWriter, r *http.Request, authEmail string) {
	kind, pathEmail, href, ok := objectPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if !sameEmail(authEmail, pathEmail) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 10<<20))
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	var object DAVObject
	switch kind {
	case "addressbooks":
		object, err = h.store.PutContact(r.Context(), authEmail, href, body)
	case "calendars":
		object, err = h.store.PutCalendarObject(r.Context(), authEmail, href, body)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "store object failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("ETag", object.ETag)
	w.WriteHeader(http.StatusCreated)
}

func (h handler) getObject(w http.ResponseWriter, r *http.Request, authEmail string) {
	kind, pathEmail, href, ok := objectPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if !sameEmail(authEmail, pathEmail) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var object DAVObject
	var err error
	switch kind {
	case "addressbooks":
		object, err = h.store.GetContact(r.Context(), authEmail, href)
		w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
	case "calendars":
		object, err = h.store.GetCalendarObject(r.Context(), authEmail, href)
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	default:
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, ErrNotFound) || len(object.Body) == 0 {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "get object failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("ETag", object.ETag)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(object.Body)
}

func objectPath(path string) (kind, email, href string, ok bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 5 || parts[0] != "dav" || parts[3] != "default" || parts[4] == "" {
		return "", "", "", false
	}
	if parts[1] != "addressbooks" && parts[1] != "calendars" {
		return "", "", "", false
	}
	if !strings.Contains(parts[2], "@") {
		return "", "", "", false
	}
	return parts[1], strings.ToLower(parts[2]), parts[4], true
}

func sameEmail(left, right string) bool {
	return strings.EqualFold(left, right)
}

func writeMultiStatus(w http.ResponseWriter, path string) {
	href := cleanDAVPath(path)
	email := emailFromDAVPath(href)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)
	_, _ = fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:Card="urn:ietf:params:xml:ns:carddav">
  <D:response>
    <D:href>%s</D:href>
    <D:propstat>
      <D:prop>
        <D:displayname>%s</D:displayname>
        <D:current-user-principal><D:href>/dav/principals/%s/</D:href></D:current-user-principal>
        <C:calendar-home-set><D:href>/dav/calendars/%s/</D:href></C:calendar-home-set>
        <Card:addressbook-home-set><D:href>/dav/addressbooks/%s/</D:href></Card:addressbook-home-set>
        <D:resourcetype>%s</D:resourcetype>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>
`, xmlText(href), xmlText(displayNameForPath(href)), xmlText(email), xmlText(email), xmlText(email), resourceTypeForPath(href))
}

func cleanDAVPath(path string) string {
	if path == "" || path == "/dav" {
		return "/dav/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func emailFromDAVPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, part := range parts {
		if strings.Contains(part, "@") {
			return part
		}
	}
	return "anonymous"
}

func displayNameForPath(path string) string {
	if strings.Contains(path, "/calendars/") {
		return "Default Calendar"
	}
	if strings.Contains(path, "/addressbooks/") {
		return "Default Address Book"
	}
	if email := emailFromDAVPath(path); email != "anonymous" {
		return email
	}
	return "ProIdentity DAV"
}

func resourceTypeForPath(path string) string {
	if strings.Contains(path, "/calendars/") {
		return "<D:collection/><C:calendar/>"
	}
	if strings.Contains(path, "/addressbooks/") {
		return "<D:collection/><Card:addressbook/>"
	}
	return "<D:collection/><D:principal/>"
}

func xmlText(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}
