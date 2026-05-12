package admin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"proidentity-mail/internal/domain"
)

type adminWebAuthnUser struct {
	username    string
	credentials []webauthn.Credential
}

func (u adminWebAuthnUser) WebAuthnID() []byte {
	sum := sha256.Sum256([]byte("proidentity-admin:" + strings.ToLower(strings.TrimSpace(u.username))))
	return sum[:]
}

func (u adminWebAuthnUser) WebAuthnName() string {
	return u.username
}

func (u adminWebAuthnUser) WebAuthnDisplayName() string {
	return "ProIdentity Mail Admin"
}

func (u adminWebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (u adminWebAuthnUser) WebAuthnIcon() string {
	return ""
}

func (h handler) beginAdminWebAuthnRegistration(w http.ResponseWriter, r *http.Request) {
	if !requireSuperAdmin(w, r) {
		return
	}
	wa, err := newRequestWebAuthn(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	credentials, _, err := h.adminWebAuthnCredentials(r.Context())
	if err != nil {
		writeStoreError(w, err, "load hardware keys failed")
		return
	}
	user := adminWebAuthnUser{username: h.auth.Username, credentials: credentials}
	creation, sessionData, err := wa.BeginRegistration(user, webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
		UserVerification: protocol.VerificationRequired,
		ResidentKey:      protocol.ResidentKeyRequirementPreferred,
	}))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := randomHexToken(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create hardware-key challenge failed")
		return
	}
	if err := h.saveAdminWebAuthnSession(r.Context(), token, "registration", *sessionData); err != nil {
		writeStoreError(w, err, "save hardware-key challenge failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "publicKey": creation.Response})
}

func (h handler) finishAdminWebAuthnRegistration(w http.ResponseWriter, r *http.Request) {
	if !requireSuperAdmin(w, r) {
		return
	}
	var req struct {
		Token      string          `json:"token"`
		Name       string          `json:"name"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	sessionData, err := h.loadAdminWebAuthnSession(r.Context(), req.Token, "registration")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "hardware-key challenge not found")
		return
	}
	wa, err := newRequestWebAuthn(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	credentials, _, err := h.adminWebAuthnCredentials(r.Context())
	if err != nil {
		writeStoreError(w, err, "load hardware keys failed")
		return
	}
	user := adminWebAuthnUser{username: h.auth.Username, credentials: credentials}
	finishReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.String(), bytes.NewReader(req.Credential))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid credential")
		return
	}
	finishReq.Header.Set("Content-Type", "application/json")
	credential, err := wa.FinishRegistration(user, sessionData, finishReq)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "serialize credential failed")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "Hardware key"
	}
	saved, err := h.store.CreateAdminWebAuthnCredential(r.Context(), domain.AdminWebAuthnCredential{Name: name, CredentialID: credential.ID, CredentialJSON: credentialJSON})
	if err != nil {
		writeStoreError(w, err, "save hardware key failed")
		return
	}
	_ = h.store.DeleteAdminWebAuthnSession(r.Context(), req.Token)
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err == nil {
		settings.NativeWebAuthnEnabled = true
		_, _ = h.store.SaveAdminMFASettings(r.Context(), settings)
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.webauthn_registered", TargetType: "admin_mfa", TargetID: fmt.Sprint(saved.ID), MetadataJSON: fmt.Sprintf(`{"name":%q}`, saved.Name)})
	writeJSON(w, http.StatusOK, publicAdminMFASettingsWithCount(settings, len(credentials)+1))
}

func (h handler) beginSessionWebAuthn(ctx context.Context, r *http.Request, username string, settings domain.AdminMFASettings) (map[string]any, error) {
	wa, err := newRequestWebAuthn(r)
	if err != nil {
		return nil, err
	}
	credentials, _, err := h.adminWebAuthnCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if len(credentials) == 0 {
		return nil, errors.New("no hardware keys registered")
	}
	user := adminWebAuthnUser{username: username, credentials: credentials}
	assertion, sessionData, err := wa.BeginLogin(user, webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, err
	}
	token, err := randomHexToken(32)
	if err != nil {
		return nil, err
	}
	if err := h.saveAdminWebAuthnSession(ctx, token, "login", *sessionData); err != nil {
		return nil, err
	}
	return map[string]any{
		"mfa_required": true,
		"provider":     "webauthn",
		"providers":    mfaProviders(settings),
		"mfa_token":    token,
		"publicKey":    assertion.Response,
		"expires_at":   sessionData.Expires.Unix(),
	}, nil
}

func (h handler) finishSessionWebAuthn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MFAToken   string          `json:"mfa_token"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if !h.finishAdminWebAuthnAssertion(w, r, req.MFAToken, req.Credential, h.auth.Username) {
		return
	}
	created, err := h.auth.Sessions.Create(r, h.auth.Username, "admin")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}
	http.SetCookie(w, created.Cookie)
	h.recordAdminLoginSuccess(r.Context(), r, h.auth.Username, "webauthn")
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken})
}

func (h handler) finishAdminStepUpWebAuthn(w http.ResponseWriter, r *http.Request) {
	if h.auth.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	principal := currentAdminPrincipal(r.Context())
	if strings.TrimSpace(principal.Subject) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		MFAToken   string          `json:"mfa_token"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if !h.finishAdminWebAuthnAssertion(w, r, req.MFAToken, req.Credential, principal.Subject) {
		return
	}
	if !h.auth.Sessions.MarkStepUp(r, adminStepUpTTL) {
		writeError(w, http.StatusUnauthorized, "session step-up failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.step_up", TargetType: "admin", TargetID: principal.Subject, MetadataJSON: `{"provider":"webauthn"}`})
	writeJSON(w, http.StatusOK, map[string]any{"step_up_verified": true, "valid_for_seconds": int(adminStepUpTTL.Seconds())})
}

func (h handler) finishAdminWebAuthnAssertion(w http.ResponseWriter, r *http.Request, token string, credentialBody json.RawMessage, username string) bool {
	sessionData, err := h.loadAdminWebAuthnSession(r.Context(), token, "login")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "hardware-key challenge not found")
		return false
	}
	wa, err := newRequestWebAuthn(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return false
	}
	credentials, stored, err := h.adminWebAuthnCredentials(r.Context())
	if err != nil {
		writeStoreError(w, err, "load hardware keys failed")
		return false
	}
	user := adminWebAuthnUser{username: username, credentials: credentials}
	finishReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, r.URL.String(), bytes.NewReader(credentialBody))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid credential")
		return false
	}
	finishReq.Header.Set("Content-Type", "application/json")
	credential, err := wa.FinishLogin(user, sessionData, finishReq)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return false
	}
	credentialJSON, _ := json.Marshal(credential)
	for _, item := range stored {
		if bytes.Equal(item.CredentialID, credential.ID) {
			item.CredentialJSON = credentialJSON
			_ = h.store.UpdateAdminWebAuthnCredential(r.Context(), item)
			break
		}
	}
	_ = h.store.DeleteAdminWebAuthnSession(r.Context(), token)
	return true
}

func (h handler) adminWebAuthnCredentials(ctx context.Context) ([]webauthn.Credential, []domain.AdminWebAuthnCredential, error) {
	if h.store == nil {
		return nil, nil, errors.New("store unavailable")
	}
	rows, err := h.store.ListAdminWebAuthnCredentials(ctx)
	if err != nil {
		return nil, nil, err
	}
	credentials := make([]webauthn.Credential, 0, len(rows))
	for _, row := range rows {
		var credential webauthn.Credential
		if len(row.CredentialJSON) > 0 {
			if err := json.Unmarshal(row.CredentialJSON, &credential); err != nil {
				return nil, nil, err
			}
		}
		if len(credential.ID) == 0 {
			credential.ID = row.CredentialID
		}
		credentials = append(credentials, credential)
	}
	return credentials, rows, nil
}

func (h handler) saveAdminWebAuthnSession(ctx context.Context, token, ceremony string, sessionData webauthn.SessionData) error {
	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}
	return h.store.CreateAdminWebAuthnSession(ctx, domain.AdminWebAuthnSession{Token: token, Ceremony: ceremony, SessionJSON: sessionJSON, ExpiresAt: sessionData.Expires.UTC()})
}

func (h handler) loadAdminWebAuthnSession(ctx context.Context, token, ceremony string) (webauthn.SessionData, error) {
	row, err := h.store.GetAdminWebAuthnSession(ctx, strings.TrimSpace(token))
	if err != nil {
		return webauthn.SessionData{}, err
	}
	if row.Ceremony != ceremony || time.Now().UTC().After(row.ExpiresAt) {
		return webauthn.SessionData{}, errors.New("expired hardware-key challenge")
	}
	var sessionData webauthn.SessionData
	if err := json.Unmarshal(row.SessionJSON, &sessionData); err != nil {
		return webauthn.SessionData{}, err
	}
	return sessionData, nil
}

func newRequestWebAuthn(r *http.Request) (*webauthn.WebAuthn, error) {
	host := r.Host
	if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
		host = strings.Split(forwardedHost, ",")[0]
	}
	host = strings.TrimSpace(host)
	rpID := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		rpID = h
	}
	if rpID == "" {
		return nil, errors.New("request host is required for hardware keys")
	}
	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else if isLoopbackHost(rpID) {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}
	origin := scheme + "://" + host
	return webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: "ProIdentity Mail Admin",
		RPOrigins:     []string{origin},
	})
}

func publicAdminMFASettingsWithCount(settings domain.AdminMFASettings, count int) domain.AdminMFASettings {
	settings.NativeWebAuthnCredentialCount = count
	settings.NativeWebAuthnEnabled = settings.NativeWebAuthnEnabled && count > 0
	return publicAdminMFASettings(settings)
}
