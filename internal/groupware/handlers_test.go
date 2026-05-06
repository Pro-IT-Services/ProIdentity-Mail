package groupware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOptionsAdvertisesCalendarAndAddressBookDAV(t *testing.T) {
	handler := NewRouter()
	req := httptest.NewRequest(http.MethodOptions, "/dav/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	dav := rec.Header().Get("DAV")
	for _, want := range []string{"1", "3", "calendar-access", "addressbook"} {
		if !bytes.Contains([]byte(dav), []byte(want)) {
			t.Fatalf("DAV header missing %q: %q", want, dav)
		}
	}
}

func TestPrincipalPropfindReturnsCalendarAndAddressBookHomes(t *testing.T) {
	handler := NewRouter()
	body := bytes.NewBufferString(`<?xml version="1.0" encoding="utf-8" ?>
<propfind xmlns="DAV:">
  <prop>
    <current-user-principal/>
    <calendar-home-set xmlns="urn:ietf:params:xml:ns:caldav"/>
    <addressbook-home-set xmlns="urn:ietf:params:xml:ns:carddav"/>
  </prop>
</propfind>`)
	req := httptest.NewRequest("PROPFIND", "/dav/principals/marko@example.com/", body)
	req.Header.Set("Depth", "0")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusMultiStatus, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/xml; charset=utf-8" {
		t.Fatalf("content-type = %q, want xml", got)
	}
	response := rec.Body.String()
	for _, want := range []string{
		"<D:multistatus",
		"/dav/principals/marko@example.com/",
		"/dav/calendars/marko@example.com/",
		"/dav/addressbooks/marko@example.com/",
	} {
		if !bytes.Contains([]byte(response), []byte(want)) {
			t.Fatalf("response missing %q: %s", want, response)
		}
	}
}

func TestDefaultCollectionsPropfind(t *testing.T) {
	handler := NewRouter()
	for _, path := range []string{
		"/dav/calendars/marko@example.com/default/",
		"/dav/addressbooks/marko@example.com/default/",
	} {
		req := httptest.NewRequest("PROPFIND", path, nil)
		req.Header.Set("Depth", "0")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMultiStatus {
			t.Fatalf("%s status = %d, want %d, body %s", path, rec.Code, http.StatusMultiStatus, rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(path)) {
			t.Fatalf("%s response missing href: %s", path, rec.Body.String())
		}
	}
}
