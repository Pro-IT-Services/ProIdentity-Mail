package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/security"
)

type Store interface {
	CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error)
	CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error)
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetDomainDNS(ctx context.Context, domainID uint64) (domain.DomainDNS, error)
}

type handler struct {
	store Store
}

func NewRouter(store Store) http.Handler {
	h := handler{store: store}
	r := chi.NewRouter()
	r.Get("/healthz", health)
	r.Post("/api/v1/tenants", h.createTenant)
	r.Post("/api/v1/domains", h.createDomain)
	r.Get("/api/v1/domains/{domainID}/dns", h.getDomainDNS)
	r.Post("/api/v1/users", h.createUser)
	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h handler) createTenant(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	tenant, err := h.store.CreateTenant(r.Context(), domain.Tenant{Name: req.Name, Slug: req.Slug})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create tenant failed")
		return
	}
	writeJSON(w, http.StatusCreated, tenant)
}

func (h handler) createDomain(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		TenantID uint64 `json:"tenant_id"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TenantID == 0 || req.Name == "" {
		writeError(w, http.StatusBadRequest, "tenant_id and name are required")
		return
	}
	mailDomain, err := h.store.CreateDomain(r.Context(), domain.Domain{TenantID: req.TenantID, Name: req.Name})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create domain failed")
		return
	}
	writeJSON(w, http.StatusCreated, mailDomain)
}

func (h handler) createUser(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	var req struct {
		TenantID        uint64 `json:"tenant_id"`
		PrimaryDomainID uint64 `json:"primary_domain_id"`
		LocalPart       string `json:"local_part"`
		DisplayName     string `json:"display_name"`
		Password        string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TenantID == 0 || req.PrimaryDomainID == 0 || req.LocalPart == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "tenant_id, primary_domain_id, local_part, and password are required")
		return
	}
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid password")
		return
	}
	user, err := h.store.CreateUser(r.Context(), domain.User{
		TenantID:        req.TenantID,
		PrimaryDomainID: req.PrimaryDomainID,
		LocalPart:       req.LocalPart,
		DisplayName:     req.DisplayName,
		PasswordHash:    hash,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create user failed")
		return
	}
	user.PasswordHash = ""
	writeJSON(w, http.StatusCreated, user)
}

func (h handler) getDomainDNS(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	domainID, err := strconv.ParseUint(chi.URLParam(r, "domainID"), 10, 64)
	if err != nil || domainID == 0 {
		writeError(w, http.StatusBadRequest, "valid domain id is required")
		return
	}
	dns, err := h.store.GetDomainDNS(r.Context(), domainID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get domain dns failed")
		return
	}
	writeJSON(w, http.StatusOK, dns)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
