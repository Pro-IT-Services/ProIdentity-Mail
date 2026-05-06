package groupware

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/dav/", dav)
	return mux
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
}

func dav(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("DAV", "1, 3, calendar-access, addressbook")
	w.Header().Set("MS-Author-Via", "DAV")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	case "PROPFIND":
		writeMultiStatus(w, r.URL.Path)
	default:
		w.Header().Set("Allow", "OPTIONS, PROPFIND")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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
