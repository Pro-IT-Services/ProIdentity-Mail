package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"proidentity-mail/internal/domain"
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
	var response domain.User
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.PasswordHash != "" {
		t.Fatal("response exposed password hash")
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

func TestNormalizeDKIMTXTExtractsTXTValue(t *testing.T) {
	raw := "mail._domainkey IN TXT ( \"v=DKIM1; k=rsa; \"\n\t\"p=abc123\"\n) ;"
	got := normalizeDKIMTXT(raw)
	want := "v=DKIM1; k=rsa; p=abc123"
	if got != want {
		t.Fatalf("normalizeDKIMTXT() = %q, want %q", got, want)
	}
}

type fakeStore struct {
	tenant     domain.Tenant
	mailDomain domain.Domain
	user       domain.User
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
	return user, nil
}

func (s *fakeStore) ListUsers(ctx context.Context) ([]domain.User, error) {
	return []domain.User{{ID: 33, TenantID: 11, PrimaryDomainID: 22, LocalPart: "marko", DisplayName: "Marko", Status: "active"}}, nil
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
