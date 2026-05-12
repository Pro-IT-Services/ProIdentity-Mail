package groupware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestOptionsAdvertisesCalendarAndAddressBookDAV(t *testing.T) {
	handler := NewRouter(nil)
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
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
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
	req.SetBasicAuth("marko@example.com", "secret123456")
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

func TestPropfindRequiresValidBasicAuth(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest("PROPFIND", "/dav/principals/marko@example.com/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got == "" {
		t.Fatal("WWW-Authenticate header is empty")
	}
}

func TestPropfindRejectsInvalidBasicAuth(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest("PROPFIND", "/dav/principals/marko@example.com/", nil)
	req.SetBasicAuth("marko@example.com", "wrong")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestDAVAuthRecordsLimiterFailuresWithRealClientIP(t *testing.T) {
	limiter := &recordingLimiter{}
	handler := NewRouterWithLimiter(&fakeStore{}, limiter)
	req := httptest.NewRequest("PROPFIND", "/dav/principals/marko@example.com/", nil)
	req.RemoteAddr = "127.0.0.1:38123"
	req.Header.Set("X-Real-IP", "203.0.113.45")
	req.SetBasicAuth("Marko@Example.com", "wrong")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	for _, want := range []string{
		"dav|ip|203.0.113.45",
		"dav|account|marko@example.com",
		"dav|pair|marko@example.com|203.0.113.45",
	} {
		if !slices.Contains(limiter.failed, want) {
			t.Fatalf("limiter failures missing %q: %+v", want, limiter.failed)
		}
	}
}

func TestDAVAuthStopsWhenLimiterLocked(t *testing.T) {
	limiter := &recordingLimiter{locked: map[string]bool{"dav|ip|203.0.113.45": true}}
	store := &fakeStore{valid: true}
	handler := NewRouterWithLimiter(store, limiter)
	req := httptest.NewRequest("PROPFIND", "/dav/principals/marko@example.com/", nil)
	req.RemoteAddr = "127.0.0.1:38123"
	req.Header.Set("X-Real-IP", "203.0.113.45")
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if store.authCalls != 0 {
		t.Fatalf("password verifier should not run while locked, got %d calls", store.authCalls)
	}
}

func TestDefaultCollectionsPropfind(t *testing.T) {
	handler := NewRouter(&fakeStore{valid: true})
	for _, path := range []string{
		"/dav/calendars/marko@example.com/default/",
		"/dav/addressbooks/marko@example.com/default/",
	} {
		req := httptest.NewRequest("PROPFIND", path, nil)
		req.Header.Set("Depth", "0")
		req.SetBasicAuth("marko@example.com", "secret123456")
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

func TestPutAndGetContactObject(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	vcard := "BEGIN:VCARD\r\nVERSION:4.0\r\nUID:contact-1\r\nFN:Marko Test\r\nEMAIL:marko@example.com\r\nEND:VCARD\r\n"

	put := httptest.NewRequest(http.MethodPut, "/dav/addressbooks/marko@example.com/default/contact-1.vcf", bytes.NewBufferString(vcard))
	put.SetBasicAuth("marko@example.com", "secret123456")
	putRec := httptest.NewRecorder()
	handler.ServeHTTP(putRec, put)

	if putRec.Code != http.StatusCreated {
		t.Fatalf("PUT status = %d, want %d, body %s", putRec.Code, http.StatusCreated, putRec.Body.String())
	}
	if putRec.Header().Get("ETag") == "" {
		t.Fatal("PUT response missing ETag")
	}

	get := httptest.NewRequest(http.MethodGet, "/dav/addressbooks/marko@example.com/default/contact-1.vcf", nil)
	get.SetBasicAuth("marko@example.com", "secret123456")
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, get)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d, body %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}
	if got := getRec.Header().Get("Content-Type"); got != "text/vcard; charset=utf-8" {
		t.Fatalf("content-type = %q, want vcard", got)
	}
	if getRec.Body.String() != vcard {
		t.Fatalf("GET body = %q, want vcard", getRec.Body.String())
	}
}

func TestPutRejectsDifferentAuthenticatedUser(t *testing.T) {
	handler := NewRouter(&fakeStore{valid: true})
	req := httptest.NewRequest(http.MethodPut, "/dav/addressbooks/other@example.com/default/contact-1.vcf", bytes.NewBufferString("BEGIN:VCARD\r\nEND:VCARD\r\n"))
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestPutAndGetCalendarObject(t *testing.T) {
	store := &fakeStore{valid: true}
	handler := NewRouter(store)
	ics := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nUID:event-1\r\nSUMMARY:Test\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

	put := httptest.NewRequest(http.MethodPut, "/dav/calendars/marko@example.com/default/event-1.ics", bytes.NewBufferString(ics))
	put.SetBasicAuth("marko@example.com", "secret123456")
	putRec := httptest.NewRecorder()
	handler.ServeHTTP(putRec, put)

	if putRec.Code != http.StatusCreated {
		t.Fatalf("PUT status = %d, want %d, body %s", putRec.Code, http.StatusCreated, putRec.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/dav/calendars/marko@example.com/default/event-1.ics", nil)
	get.SetBasicAuth("marko@example.com", "secret123456")
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, get)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d, body %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}
	if got := getRec.Header().Get("Content-Type"); got != "text/calendar; charset=utf-8" {
		t.Fatalf("content-type = %q, want calendar", got)
	}
	if getRec.Body.String() != ics {
		t.Fatalf("GET body = %q, want ics", getRec.Body.String())
	}
}

func TestReportListsAddressBookObjects(t *testing.T) {
	store := &fakeStore{valid: true, contacts: map[string]DAVObject{
		"marko@example.com/contact-1.vcf": {Href: "contact-1.vcf", ETag: `"contact-etag"`, Body: []byte("BEGIN:VCARD\r\nEND:VCARD\r\n")},
	}}
	handler := NewRouter(store)
	req := httptest.NewRequest("REPORT", "/dav/addressbooks/marko@example.com/default/", nil)
	req.SetBasicAuth("marko@example.com", "secret123456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusMultiStatus, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("/dav/addressbooks/marko@example.com/default/contact-1.vcf")) {
		t.Fatalf("REPORT missing contact href: %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("contact-etag")) {
		t.Fatalf("REPORT missing etag: %s", rec.Body.String())
	}
}

type fakeStore struct {
	valid     bool
	authCalls int
	contacts  map[string]DAVObject
	events    map[string]DAVObject
}

func (s *fakeStore) VerifyUserPassword(ctx context.Context, email, password string) (bool, error) {
	s.authCalls++
	return s.valid && email == "marko@example.com" && password == "secret123456", nil
}

type recordingLimiter struct {
	locked    map[string]bool
	failed    []string
	succeeded []string
}

func (l *recordingLimiter) Locked(key string) bool {
	return l.locked[key]
}

func (l *recordingLimiter) Fail(key string) {
	l.failed = append(l.failed, key)
}

func (l *recordingLimiter) Success(key string) {
	l.succeeded = append(l.succeeded, key)
}

func (s *fakeStore) PutContact(ctx context.Context, email, href string, body []byte) (DAVObject, error) {
	if s.contacts == nil {
		s.contacts = map[string]DAVObject{}
	}
	object := DAVObject{Href: href, ETag: `"contact-etag"`, Body: append([]byte(nil), body...)}
	s.contacts[email+"/"+href] = object
	return object, nil
}

func (s *fakeStore) GetContact(ctx context.Context, email, href string) (DAVObject, error) {
	return s.contacts[email+"/"+href], nil
}

func (s *fakeStore) PutCalendarObject(ctx context.Context, email, href string, body []byte) (DAVObject, error) {
	if s.events == nil {
		s.events = map[string]DAVObject{}
	}
	object := DAVObject{Href: href, ETag: `"event-etag"`, Body: append([]byte(nil), body...)}
	s.events[email+"/"+href] = object
	return object, nil
}

func (s *fakeStore) GetCalendarObject(ctx context.Context, email, href string) (DAVObject, error) {
	return s.events[email+"/"+href], nil
}

func (s *fakeStore) ListContacts(ctx context.Context, email string) ([]DAVObject, error) {
	var objects []DAVObject
	for key, object := range s.contacts {
		if strings.HasPrefix(key, email+"/") {
			objects = append(objects, object)
		}
	}
	return objects, nil
}

func (s *fakeStore) ListCalendarObjects(ctx context.Context, email string) ([]DAVObject, error) {
	var objects []DAVObject
	for key, object := range s.events {
		if strings.HasPrefix(key, email+"/") {
			objects = append(objects, object)
		}
	}
	return objects, nil
}
