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

func (s *fakeStore) CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error) {
	s.mailDomain = mailDomain
	mailDomain.ID = 22
	mailDomain.Status = "pending"
	return mailDomain, nil
}

func (s *fakeStore) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	s.user = user
	user.ID = 33
	user.Status = "active"
	return user, nil
}
