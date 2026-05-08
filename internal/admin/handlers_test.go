package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/session"
)

func TestHealthEndpoint(t *testing.T) {
	handler := NewRouter(nil)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestAdminIndexServesWebUI(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want text/html", contentType)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("ProIdentity Mail Admin")) {
		t.Fatalf("index missing product title: %s", rec.Body.String())
	}
}

func TestAdminAPIRequiresAuthWhenConfigured(t *testing.T) {
	handler := NewRouter(&fakeStore{}, AuthConfig{Username: "admin", Password: "secret"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got == "" {
		t.Fatal("WWW-Authenticate header is empty")
	}
}

func TestAdminAPIAcceptsConfiguredAuth(t *testing.T) {
	handler := NewRouter(&fakeStore{}, AuthConfig{Username: "admin", Password: "secret"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	req.SetBasicAuth("admin", "secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestAdminSessionLoginAndCSRF(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	handler := NewRouter(&fakeStore{}, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(`{"name":"Example Org","slug":"example"}`))
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("Accept-Language", "en-US")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("missing csrf status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBufferString(`{"name":"Example Org","slug":"example"}`))
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("X-CSRF-Token", loginResponse.CSRFToken)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("session request status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
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
		Username  string `json:"username"`
	}
	if err := json.NewDecoder(currentRec.Body).Decode(&currentResponse); err != nil {
		t.Fatalf("decode current session: %v", err)
	}
	if currentResponse.CSRFToken != loginResponse.CSRFToken || currentResponse.Username != "admin" {
		t.Fatalf("unexpected current session: %+v", currentResponse)
	}
}

func TestAdminSessionProtectedAPIWithoutCookieDoesNotTriggerBasicPopup(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	handler := NewRouter(&fakeStore{}, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for browser session auth", got)
	}
}

func TestDiscoveryStaysPublicWhenAdminAuthConfigured(t *testing.T) {
	handler := NewRouter(&fakeStore{}, AuthConfig{Username: "admin", Password: "secret"})
	req := httptest.NewRequest(http.MethodGet, "/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=marko@example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestListEndpoints(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	tests := []struct {
		path string
		want string
	}{
		{path: "/api/v1/tenants", want: "Example Org"},
		{path: "/api/v1/domains", want: "example.com"},
		{path: "/api/v1/users", want: "marko"},
		{path: "/api/v1/aliases", want: "sales"},
		{path: "/api/v1/catch-all", want: "catchall@example.com"},
		{path: "/api/v1/shared-permissions", want: "can_send_as"},
		{path: "/api/v1/quarantine", want: "EICAR"},
		{path: "/api/v1/audit", want: "message.report_spam"},
		{path: "/api/v1/policies", want: "quarantine"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d, body %s", tt.path, rec.Code, http.StatusOK, rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(tt.want)) {
			t.Fatalf("%s response missing %q: %s", tt.path, tt.want, rec.Body.String())
		}
	}
}

func TestMailAutoconfigEndpoint(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=marko@example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/xml; charset=utf-8" {
		t.Fatalf("content-type = %q, want xml", got)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"<emailProvider id=\"example.com\">",
		"<incomingServer type=\"imap\">",
		"<hostname>mail.example.com</hostname>",
		"<outgoingServer type=\"smtp\">",
		"<port>587</port>",
	} {
		if !bytes.Contains([]byte(body), []byte(want)) {
			t.Fatalf("autoconfig missing %q: %s", want, body)
		}
	}
}

func TestServiceDiscoveryEndpoint(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/.well-known/proidentity-mail/config.json?emailaddress=marko@example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"\"imap\"", "\"smtp\"", "\"caldav\"", "\"carddav\"", "https://mail.example.com/dav/calendars/marko@example.com/"} {
		if !bytes.Contains([]byte(body), []byte(want)) {
			t.Fatalf("service discovery missing %q: %s", want, body)
		}
	}
}

func TestWellKnownGroupwareRedirects(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	for _, tt := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/.well-known/caldav"},
		{method: http.MethodGet, path: "/.well-known/carddav"},
		{method: http.MethodHead, path: "/.well-known/caldav"},
		{method: http.MethodHead, path: "/.well-known/carddav"},
	} {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		response := rec.Result()
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, response.Body)

		if response.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, response.StatusCode, http.StatusTemporaryRedirect)
		}
		if location := response.Header.Get("Location"); location != "/dav/" {
			t.Fatalf("%s %s location = %q, want /dav/", tt.method, tt.path, location)
		}
	}
}

func TestCreateTenantEndpoint(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := bytes.NewBufferString(`{"name":"Example Org","slug":"example"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.tenant.Name != "Example Org" || store.tenant.Slug != "example" {
		t.Fatalf("tenant not passed to store: %+v", store.tenant)
	}
	responseBody := rec.Body.String()
	var response domain.Tenant
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.ID != 11 {
		t.Fatalf("tenant ID = %d, want 11", response.ID)
	}
	if bytes.Contains([]byte(responseBody), []byte(`"ID"`)) {
		t.Fatalf("response uses Go field names instead of JSON field names: %s", responseBody)
	}
}

func TestCreateDomainEndpoint(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := bytes.NewBufferString(`{"tenant_id":11,"name":"example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/domains", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.mailDomain.TenantID != 11 || store.mailDomain.Name != "example.com" {
		t.Fatalf("domain not passed to store: %+v", store.mailDomain)
	}
}

func TestCreateUserEndpointHashesPassword(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := bytes.NewBufferString(`{"tenant_id":11,"primary_domain_id":22,"local_part":"marko","display_name":"Marko","password":"secret123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.user.PasswordHash == "" {
		t.Fatal("password hash is empty")
	}
	if store.user.PasswordHash == "secret123456" {
		t.Fatal("plaintext password stored")
	}
	if store.user.LocalPart != "marko" {
		t.Fatalf("local part = %q", store.user.LocalPart)
	}
	if store.user.MailboxType != "user" || store.user.QuotaBytes != 0 {
		t.Fatalf("unexpected user metadata: %+v", store.user)
	}
	var response domain.User
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.PasswordHash != "" {
		t.Fatal("response exposed password hash")
	}
}

func TestCreateSharedMailboxEndpointAllowsEmptyPasswordAndQuota(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := bytes.NewBufferString(`{"tenant_id":11,"primary_domain_id":22,"local_part":"support","display_name":"Support","mailbox_type":"shared","quota_bytes":21474836480}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.user.MailboxType != "shared" || store.user.PasswordHash != "" || store.user.QuotaBytes != 21474836480 {
		t.Fatalf("unexpected shared mailbox: %+v", store.user)
	}
}

func TestCreateAliasCatchAllAndSharedPermissionEndpoints(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	tests := []struct {
		path string
		body string
	}{
		{"/api/v1/aliases", `{"tenant_id":11,"domain_id":22,"source_local_part":"sales","destination":"marko@example.com"}`},
		{"/api/v1/catch-all", `{"tenant_id":11,"domain_id":22,"destination":"catchall@example.com"}`},
		{"/api/v1/shared-permissions", `{"tenant_id":11,"shared_mailbox_id":44,"user_id":33,"can_read":true,"can_send_as":true}`},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("%s status = %d, want %d, body %s", tt.path, rec.Code, http.StatusCreated, rec.Body.String())
		}
	}
	if store.alias.SourceLocalPart != "sales" || store.catchAll.Destination != "catchall@example.com" || !store.sharedPermission.CanSendAs {
		t.Fatalf("unexpected stored values: alias=%+v catch=%+v permission=%+v", store.alias, store.catchAll, store.sharedPermission)
	}
}

func TestDomainDNSEndpoint(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/domains/22/dns", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var response domain.DomainDNS
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.DomainID != 22 {
		t.Fatalf("domain id = %d, want 22", response.DomainID)
	}
	if len(response.Records) < 4 {
		t.Fatalf("expected MX/SPF/DMARC/DKIM records, got %+v", response.Records)
	}
}

func TestUpdateTenantPolicyEndpoint(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := bytes.NewBufferString(`{"tenant_id":11,"spam_action":"quarantine","malware_action":"reject","require_tls_for_auth":true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/policies/11", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.policy.TenantID != 11 || store.policy.SpamAction != "quarantine" || store.policy.MalwareAction != "reject" || !store.policy.RequireTLSForAuth {
		t.Fatalf("unexpected policy passed to store: %+v", store.policy)
	}
}

func TestQuarantineEndpointRequiresConfiguredAuth(t *testing.T) {
	handler := NewRouter(&fakeStore{}, AuthConfig{Username: "admin", Password: "secret"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/quarantine", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestResolveQuarantineEndpointReleasesEvent(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := bytes.NewBufferString(`{"resolution_note":"false positive"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/quarantine/44/release", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.resolvedQuarantineID != 44 || store.resolvedQuarantineStatus != "released" || store.resolvedQuarantineNote != "false positive" {
		t.Fatalf("unexpected resolution: id=%d status=%q note=%q", store.resolvedQuarantineID, store.resolvedQuarantineStatus, store.resolvedQuarantineNote)
	}
}

func TestNormalizeDKIMTXTExtractsTXTValue(t *testing.T) {
	raw := "mail._domainkey IN TXT ( \"v=DKIM1; k=rsa; \"\n\t\"p=abc123\"\n) ;"
	got := normalizeDKIMTXT(raw)
	want := "v=DKIM1; k=rsa; p=abc123"
	if got != want {
		t.Fatalf("normalizeDKIMTXT() = %q, want %q", got, want)
	}
}

type fakeStore struct {
	tenant                   domain.Tenant
	mailDomain               domain.Domain
	user                     domain.User
	alias                    domain.Alias
	catchAll                 domain.CatchAllRoute
	sharedPermission         domain.SharedMailboxPermission
	policy                   domain.TenantPolicy
	resolvedQuarantineID     uint64
	resolvedQuarantineStatus string
	resolvedQuarantineNote   string
	auditActions             []string
}

func (s *fakeStore) CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	s.tenant = tenant
	tenant.ID = 11
	tenant.Status = "active"
	return tenant, nil
}

func (s *fakeStore) ListTenants(ctx context.Context) ([]domain.Tenant, error) {
	return []domain.Tenant{{ID: 11, Name: "Example Org", Slug: "example", Status: "active"}}, nil
}

func (s *fakeStore) CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error) {
	s.mailDomain = mailDomain
	mailDomain.ID = 22
	mailDomain.Status = "pending"
	return mailDomain, nil
}

func (s *fakeStore) ListDomains(ctx context.Context) ([]domain.Domain, error) {
	return []domain.Domain{{ID: 22, TenantID: 11, Name: "example.com", Status: "pending", DKIMSelector: "mail"}}, nil
}

func (s *fakeStore) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	s.user = user
	user.ID = 33
	user.Status = "active"
	if user.MailboxType == "" {
		user.MailboxType = "user"
	}
	return user, nil
}

func (s *fakeStore) ListUsers(ctx context.Context) ([]domain.User, error) {
	return []domain.User{{ID: 33, TenantID: 11, PrimaryDomainID: 22, LocalPart: "marko", DisplayName: "Marko", MailboxType: "user", Status: "active"}}, nil
}

func (s *fakeStore) CreateAlias(ctx context.Context, alias domain.Alias) (domain.Alias, error) {
	s.alias = alias
	alias.ID = 66
	return alias, nil
}

func (s *fakeStore) ListAliases(ctx context.Context) ([]domain.Alias, error) {
	return []domain.Alias{{ID: 66, TenantID: 11, DomainID: 22, SourceLocalPart: "sales", Destination: "marko@example.com"}}, nil
}

func (s *fakeStore) CreateCatchAllRoute(ctx context.Context, route domain.CatchAllRoute) (domain.CatchAllRoute, error) {
	s.catchAll = route
	route.ID = 77
	route.Status = "active"
	return route, nil
}

func (s *fakeStore) ListCatchAllRoutes(ctx context.Context) ([]domain.CatchAllRoute, error) {
	return []domain.CatchAllRoute{{ID: 77, TenantID: 11, DomainID: 22, Destination: "catchall@example.com", Status: "active"}}, nil
}

func (s *fakeStore) CreateSharedMailboxPermission(ctx context.Context, permission domain.SharedMailboxPermission) (domain.SharedMailboxPermission, error) {
	s.sharedPermission = permission
	permission.ID = 88
	return permission, nil
}

func (s *fakeStore) ListSharedMailboxPermissions(ctx context.Context) ([]domain.SharedMailboxPermission, error) {
	return []domain.SharedMailboxPermission{{ID: 88, TenantID: 11, SharedMailboxID: 44, UserID: 33, CanRead: true, CanSendAs: true}}, nil
}

func (s *fakeStore) ListQuarantineEvents(ctx context.Context) ([]domain.QuarantineEvent, error) {
	return []domain.QuarantineEvent{{
		ID:          44,
		TenantID:    11,
		Recipient:   "marko@example.com",
		Verdict:     "malware",
		Action:      "quarantine",
		Scanner:     "ClamAV",
		SymbolsJSON: `{"signature":"EICAR-Test-Signature"}`,
		Status:      "held",
	}}, nil
}

func (s *fakeStore) ResolveQuarantineEvent(ctx context.Context, eventID uint64, status, note string) (domain.QuarantineEvent, error) {
	s.resolvedQuarantineID = eventID
	s.resolvedQuarantineStatus = status
	s.resolvedQuarantineNote = note
	return domain.QuarantineEvent{ID: eventID, TenantID: 11, Recipient: "marko@example.com", Verdict: "malware", Action: "quarantine", Scanner: "ClamAV", SymbolsJSON: `{}`, Status: status, ResolutionNote: note}, nil
}

func (s *fakeStore) ListAuditEvents(ctx context.Context) ([]domain.AuditEvent, error) {
	return []domain.AuditEvent{{
		ID:           55,
		ActorType:    "user",
		Action:       "message.report_spam",
		TargetType:   "message",
		TargetID:     "1",
		MetadataJSON: `{"verdict":"spam"}`,
	}}, nil
}

func (s *fakeStore) RecordAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	s.auditActions = append(s.auditActions, event.Action)
	return nil
}

func (s *fakeStore) ListTenantPolicies(ctx context.Context) ([]domain.TenantPolicy, error) {
	return []domain.TenantPolicy{{TenantID: 11, SpamAction: "quarantine", MalwareAction: "quarantine", RequireTLSForAuth: true}}, nil
}

func (s *fakeStore) UpdateTenantPolicy(ctx context.Context, policy domain.TenantPolicy) (domain.TenantPolicy, error) {
	s.policy = policy
	return policy, nil
}

func (s *fakeStore) GetDomainDNS(ctx context.Context, domainID uint64) (domain.DomainDNS, error) {
	priority := 10
	return domain.DomainDNS{
		DomainID: domainID,
		Domain:   "example.com",
		Records: []domain.DNSRecord{
			{Type: "MX", Name: "example.com", Value: "mail.example.com", Priority: &priority},
			{Type: "TXT", Name: "example.com", Value: "v=spf1 mx -all"},
			{Type: "TXT", Name: "_dmarc.example.com", Value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"},
			{Type: "TXT", Name: "mail._domainkey.example.com", Value: "v=DKIM1; k=rsa; p=test"},
		},
	}, nil
}
