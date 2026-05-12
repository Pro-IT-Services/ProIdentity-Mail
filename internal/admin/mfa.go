package admin

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"

	"proidentity-mail/internal/domain"
	"proidentity-mail/internal/session"
)

const defaultProIdentityTimeoutSeconds = 90
const proIdentityAuthServiceURL = "https://verify.proidentity.cloud"
const adminStepUpTTL = 5 * time.Minute

type adminMFAVerificationRequest struct {
	MFAToken string `json:"mfa_token"`
	Code     string `json:"code"`
}

func (h handler) beginAdminMFA(ctx context.Context, r *http.Request, username string, settings domain.AdminMFASettings) (map[string]any, error) {
	provider := effectiveMFAProvider(settings)
	if provider == "" {
		return nil, nil
	}
	if provider == "webauthn" {
		return h.beginSessionWebAuthn(ctx, r, username, settings)
	}
	token, err := randomHexToken(32)
	if err != nil {
		return nil, err
	}
	challenge := domain.AdminMFAChallenge{
		Token:     token,
		Username:  username,
		Provider:  provider,
		ExpiresAt: time.Now().UTC().Add(mfaTimeout(settings)),
	}
	response := map[string]any{
		"mfa_required": true,
		"provider":     provider,
		"providers":    mfaProviders(settings),
		"mfa_token":    token,
		"expires_at":   challenge.ExpiresAt.Unix(),
	}
	if provider == "proidentity" {
		created, err := createProIdentityAuthRequest(ctx, settings, proIdentityAuthRequestPayload{
			UserEmail:     settings.ProIdentityUserEmail,
			DisplayName:   "ProIdentity Mail Admin",
			ContextTitle:  "Sign in to ProIdentity Mail Admin",
			ContextDetail: "Admin panel browser session approval",
			ClientIP:      requestClientIP(r),
		})
		if err != nil {
			return nil, err
		}
		challenge.RequestID = created.RequestID
		if !created.ExpiresAt.IsZero() && created.ExpiresAt.Before(challenge.ExpiresAt) {
			challenge.ExpiresAt = created.ExpiresAt
			response["expires_at"] = challenge.ExpiresAt.Unix()
		}
		response["request_id"] = created.RequestID
	}
	if h.store == nil {
		return nil, errors.New("mfa store unavailable")
	}
	if err := h.store.CreateAdminMFAChallenge(ctx, challenge); err != nil {
		return nil, err
	}
	h.recordAudit(ctx, domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.challenge_created", TargetType: "admin", TargetID: username, MetadataJSON: fmt.Sprintf(`{"provider":%q}`, provider)})
	return response, nil
}

func (h handler) verifySessionMFA(w http.ResponseWriter, r *http.Request) {
	if h.auth.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	var req adminMFAVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	challenge, ok := h.verifyAdminMFAChallenge(w, r, req)
	if !ok {
		return
	}
	session.SuccessAll(h.auth.Limiter, loginKeys("admin", challenge.Username, r))
	created, err := h.auth.Sessions.Create(r, challenge.Username, "admin")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create session failed")
		return
	}
	_ = h.store.DeleteAdminMFAChallenge(r.Context(), challenge.Token)
	http.SetCookie(w, created.Cookie)
	h.recordAdminLoginSuccess(r.Context(), r, challenge.Username, challenge.Provider)
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": created.CSRFToken})
}

func (h handler) verifyAdminMFAChallenge(w http.ResponseWriter, r *http.Request, req adminMFAVerificationRequest) (domain.AdminMFAChallenge, bool) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return domain.AdminMFAChallenge{}, false
	}
	challenge, err := h.store.GetAdminMFAChallenge(r.Context(), strings.TrimSpace(req.MFAToken))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "mfa challenge not found")
		return domain.AdminMFAChallenge{}, false
	}
	if time.Now().UTC().After(challenge.ExpiresAt) {
		_ = h.store.DeleteAdminMFAChallenge(r.Context(), challenge.Token)
		writeError(w, http.StatusUnauthorized, "mfa challenge expired")
		return domain.AdminMFAChallenge{}, false
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return domain.AdminMFAChallenge{}, false
	}
	switch challenge.Provider {
	case "totp":
		if !settings.LocalTOTPEnabled || !verifyTOTPCode(settings.LocalTOTPSecret, req.Code) {
			h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa_failed", TargetType: "admin", TargetID: challenge.Username, MetadataJSON: `{"provider":"totp"}`})
			writeError(w, http.StatusUnauthorized, "invalid mfa code")
			return domain.AdminMFAChallenge{}, false
		}
	case "proidentity":
		if strings.TrimSpace(req.Code) != "" {
			verified, err := verifyProIdentityTOTP(r.Context(), settings, req.Code)
			if err != nil {
				if writeProIdentityAPIError(w, err) {
					return domain.AdminMFAChallenge{}, false
				}
				writeError(w, http.StatusBadGateway, err.Error())
				return domain.AdminMFAChallenge{}, false
			}
			if !verified {
				h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa_failed", TargetType: "admin", TargetID: challenge.Username, MetadataJSON: `{"provider":"proidentity_totp"}`})
				writeError(w, http.StatusUnauthorized, "invalid proidentity totp code")
				return domain.AdminMFAChallenge{}, false
			}
			break
		}
		status, err := readProIdentityAuthRequestStatus(r.Context(), settings, challenge.RequestID)
		if err != nil {
			if writeProIdentityAPIError(w, err) {
				return domain.AdminMFAChallenge{}, false
			}
			writeError(w, http.StatusBadGateway, err.Error())
			return domain.AdminMFAChallenge{}, false
		}
		switch status.Status {
		case "approved":
		case "pending":
			writeJSON(w, http.StatusAccepted, map[string]any{"mfa_required": true, "provider": "proidentity", "status": "pending", "request_id": challenge.RequestID, "expires_at": challenge.ExpiresAt.Unix()})
			return domain.AdminMFAChallenge{}, false
		case "denied", "expired":
			_ = h.store.DeleteAdminMFAChallenge(r.Context(), challenge.Token)
			h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa_failed", TargetType: "admin", TargetID: challenge.Username, MetadataJSON: fmt.Sprintf(`{"provider":"proidentity","status":%q}`, status.Status)})
			writeError(w, http.StatusUnauthorized, "proidentity auth "+status.Status)
			return domain.AdminMFAChallenge{}, false
		default:
			writeError(w, http.StatusBadGateway, "unexpected proidentity auth status")
			return domain.AdminMFAChallenge{}, false
		}
	default:
		writeError(w, http.StatusUnauthorized, "unsupported mfa provider")
		return domain.AdminMFAChallenge{}, false
	}
	return challenge, true
}

func (h handler) beginAdminStepUp(w http.ResponseWriter, r *http.Request) {
	if h.auth.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	principal := currentAdminPrincipal(r.Context())
	if strings.TrimSpace(principal.Subject) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	if effectiveMFAProvider(settings) == "" {
		writeError(w, http.StatusForbidden, "admin mfa must be enabled before dangerous operations can be approved")
		return
	}
	response, err := h.beginAdminMFA(r.Context(), r, principal.Subject, settings)
	if err != nil {
		if writeProIdentityAPIError(w, err) {
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if response == nil {
		writeError(w, http.StatusForbidden, "admin mfa must be enabled before dangerous operations can be approved")
		return
	}
	response["step_up_required"] = true
	writeJSON(w, http.StatusOK, response)
}

func (h handler) verifyAdminStepUp(w http.ResponseWriter, r *http.Request) {
	if h.auth.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "sessions unavailable")
		return
	}
	principal := currentAdminPrincipal(r.Context())
	if strings.TrimSpace(principal.Subject) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req adminMFAVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	challenge, verified := h.verifyAdminMFAChallenge(w, r, req)
	if !verified {
		return
	}
	if !strings.EqualFold(challenge.Username, principal.Subject) {
		writeError(w, http.StatusForbidden, "mfa challenge does not match current admin")
		return
	}
	if !h.auth.Sessions.MarkStepUp(r, adminStepUpTTL) {
		writeError(w, http.StatusUnauthorized, "session step-up failed")
		return
	}
	_ = h.store.DeleteAdminMFAChallenge(r.Context(), challenge.Token)
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.step_up", TargetType: "admin", TargetID: challenge.Username, MetadataJSON: fmt.Sprintf(`{"provider":%q}`, challenge.Provider)})
	writeJSON(w, http.StatusOK, map[string]any{"step_up_verified": true, "valid_for_seconds": int(adminStepUpTTL.Seconds())})
}

func (h handler) getAdminMFASettings(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	settings = h.addAdminWebAuthnCount(r.Context(), settings)
	writeJSON(w, http.StatusOK, publicAdminMFASettings(settings))
}

func (h handler) updateProIdentityAuthSettings(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	current, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	var req struct {
		Enabled        bool   `json:"enabled"`
		BaseURL        string `json:"base_url"`
		APIKey         string `json:"api_key"`
		UserEmail      string `json:"user_email"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	oldBaseURL := strings.TrimSpace(current.ProIdentityBaseURL)
	oldUserEmail := strings.ToLower(strings.TrimSpace(current.ProIdentityUserEmail))
	apiKeyChanged := strings.TrimSpace(req.APIKey) != ""
	current.ProIdentityEnabled = req.Enabled
	current.ProIdentityBaseURL = proIdentityAuthServiceURL
	current.ProIdentityUserEmail = strings.ToLower(strings.TrimSpace(req.UserEmail))
	current.ProIdentityTimeoutSeconds = req.TimeoutSeconds
	if apiKeyChanged {
		current.ProIdentityAPIKey = strings.TrimSpace(req.APIKey)
	}
	if !current.ProIdentityEnabled || oldBaseURL != current.ProIdentityBaseURL || oldUserEmail != current.ProIdentityUserEmail || apiKeyChanged {
		current.ProIdentityTOTPEnabled = false
	}
	if current.ProIdentityTimeoutSeconds <= 0 {
		current.ProIdentityTimeoutSeconds = defaultProIdentityTimeoutSeconds
	}
	if current.ProIdentityEnabled {
		if err := validateProIdentitySettings(current); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	saved, err := h.store.SaveAdminMFASettings(r.Context(), current)
	if err != nil {
		writeStoreError(w, err, "save proidentity auth settings failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.proidentity.update", TargetType: "admin_mfa", TargetID: "proidentity", MetadataJSON: fmt.Sprintf(`{"enabled":%t,"base_url":%q}`, saved.ProIdentityEnabled, saved.ProIdentityBaseURL)})
	saved = h.addAdminWebAuthnCount(r.Context(), saved)
	writeJSON(w, http.StatusOK, publicAdminMFASettings(saved))
}

func (h handler) createAdminTOTPEnrollment(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	if effectiveMFAProvider(settings) == "proidentity" {
		writeError(w, http.StatusConflict, "local totp is disabled while ProIdentity Auth is enabled")
		return
	}
	secret, otpURL, qrDataURL, err := createTOTPEnrollment(h.auth.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create totp enrollment failed")
		return
	}
	settings.LocalTOTPPendingSecret = secret
	saved, err := h.store.SaveAdminMFASettings(r.Context(), settings)
	if err != nil {
		writeStoreError(w, err, "save totp enrollment failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.totp_enrollment_created", TargetType: "admin_mfa", TargetID: "totp", MetadataJSON: `{}`})
	writeJSON(w, http.StatusOK, map[string]any{
		"secret":       saved.LocalTOTPPendingSecret,
		"otpauth_url":  otpURL,
		"qr_data_url":  qrDataURL,
		"mfa_settings": publicAdminMFASettings(saved),
	})
}

func (h handler) verifyAdminTOTPEnrollment(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	if settings.LocalTOTPPendingSecret == "" || !verifyTOTPCode(settings.LocalTOTPPendingSecret, req.Code) {
		writeError(w, http.StatusUnauthorized, "invalid totp code")
		return
	}
	settings.LocalTOTPSecret = settings.LocalTOTPPendingSecret
	settings.LocalTOTPPendingSecret = ""
	settings.LocalTOTPEnabled = true
	saved, err := h.store.SaveAdminMFASettings(r.Context(), settings)
	if err != nil {
		writeStoreError(w, err, "enable totp failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.totp_enabled", TargetType: "admin_mfa", TargetID: "totp", MetadataJSON: `{}`})
	saved = h.addAdminWebAuthnCount(r.Context(), saved)
	writeJSON(w, http.StatusOK, publicAdminMFASettings(saved))
}

func (h handler) createProIdentityTOTPEnrollment(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	if effectiveMFAProvider(settings) != "proidentity" {
		writeError(w, http.StatusConflict, "enable ProIdentity Auth before creating provider TOTP")
		return
	}
	enrollment, err := createProIdentityTOTPEnrollment(r.Context(), settings, "ProIdentity Mail Admin", settings.ProIdentityUserEmail)
	if err != nil {
		if writeProIdentityAPIError(w, err) {
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.proidentity_totp_enrollment_created", TargetType: "admin_mfa", TargetID: "proidentity-totp", MetadataJSON: `{}`})
	writeJSON(w, http.StatusOK, enrollment)
}

func (h handler) verifyProIdentityTOTPEnrollment(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	if effectiveMFAProvider(settings) != "proidentity" {
		writeError(w, http.StatusConflict, "enable ProIdentity Auth before verifying provider TOTP")
		return
	}
	verified, err := verifyProIdentityTOTP(r.Context(), settings, req.Code)
	if err != nil {
		if writeProIdentityAPIError(w, err) {
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if !verified {
		h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa_failed", TargetType: "admin_mfa", TargetID: "proidentity-totp", MetadataJSON: `{"provider":"proidentity_totp_enrollment"}`})
		writeError(w, http.StatusUnauthorized, "invalid proidentity totp code")
		return
	}
	created, err := createProIdentityAuthRequest(r.Context(), settings, proIdentityAuthRequestPayload{
		UserEmail:     settings.ProIdentityUserEmail,
		DisplayName:   "ProIdentity Mail Admin",
		ContextTitle:  "Confirm ProIdentity hosted TOTP setup",
		ContextDetail: "Approve this request to finish enabling hosted TOTP for admin login",
		ClientIP:      requestClientIP(r),
	})
	if err != nil {
		if writeProIdentityAPIError(w, err) {
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	token, err := randomHexToken(32)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create hosted totp confirmation failed")
		return
	}
	expiresAt := time.Now().UTC().Add(mfaTimeout(settings))
	if !created.ExpiresAt.IsZero() && created.ExpiresAt.Before(expiresAt) {
		expiresAt = created.ExpiresAt
	}
	challenge := domain.AdminMFAChallenge{
		Token:     token,
		Username:  h.auth.Username,
		Provider:  "proidentity_totp_enrollment",
		RequestID: created.RequestID,
		ExpiresAt: expiresAt,
	}
	if err := h.store.CreateAdminMFAChallenge(r.Context(), challenge); err != nil {
		writeStoreError(w, err, "save hosted totp confirmation failed")
		return
	}
	h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.proidentity_totp_code_verified", TargetType: "admin_mfa", TargetID: "proidentity-totp", MetadataJSON: fmt.Sprintf(`{"request_id":%q}`, created.RequestID)})
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "pending", "mfa_token": token, "request_id": created.RequestID, "expires_at": expiresAt.Unix()})
}

func (h handler) confirmProIdentityTOTPEnrollment(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable")
		return
	}
	if !requireSuperAdmin(w, r) {
		return
	}
	var req struct {
		MFAToken string `json:"mfa_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	challenge, err := h.store.GetAdminMFAChallenge(r.Context(), strings.TrimSpace(req.MFAToken))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "hosted totp confirmation not found")
		return
	}
	if challenge.Provider != "proidentity_totp_enrollment" || time.Now().UTC().After(challenge.ExpiresAt) {
		_ = h.store.DeleteAdminMFAChallenge(r.Context(), challenge.Token)
		writeError(w, http.StatusUnauthorized, "hosted totp confirmation expired")
		return
	}
	settings, err := h.store.GetAdminMFASettings(r.Context())
	if err != nil {
		writeStoreError(w, err, "mfa settings unavailable")
		return
	}
	status, err := readProIdentityAuthRequestStatus(r.Context(), settings, challenge.RequestID)
	if err != nil {
		if writeProIdentityAPIError(w, err) {
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	switch status.Status {
	case "approved":
		settings.ProIdentityTOTPEnabled = true
		saved, err := h.store.SaveAdminMFASettings(r.Context(), settings)
		if err != nil {
			writeStoreError(w, err, "save hosted totp settings failed")
			return
		}
		_ = h.store.DeleteAdminMFAChallenge(r.Context(), challenge.Token)
		h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa.proidentity_totp_enabled", TargetType: "admin_mfa", TargetID: "proidentity-totp", MetadataJSON: fmt.Sprintf(`{"request_id":%q}`, challenge.RequestID)})
		saved = h.addAdminWebAuthnCount(r.Context(), saved)
		writeJSON(w, http.StatusOK, publicAdminMFASettings(saved))
	case "pending":
		writeJSON(w, http.StatusAccepted, map[string]any{"status": "pending", "mfa_token": challenge.Token, "request_id": challenge.RequestID, "expires_at": challenge.ExpiresAt.Unix()})
	case "denied", "expired":
		_ = h.store.DeleteAdminMFAChallenge(r.Context(), challenge.Token)
		h.recordAudit(r.Context(), domain.AuditEvent{ActorType: "admin", Action: "admin.mfa_failed", TargetType: "admin_mfa", TargetID: "proidentity-totp", MetadataJSON: fmt.Sprintf(`{"provider":"proidentity_totp_enrollment","status":%q}`, status.Status)})
		writeError(w, http.StatusUnauthorized, "proidentity auth "+status.Status)
	default:
		writeError(w, http.StatusBadGateway, "unexpected proidentity auth status")
	}
}

func effectiveMFAProvider(settings domain.AdminMFASettings) string {
	if settings.ProIdentityEnabled && strings.TrimSpace(proIdentityRequestBaseURL(settings)) != "" && strings.TrimSpace(settings.ProIdentityAPIKey) != "" && strings.TrimSpace(settings.ProIdentityUserEmail) != "" {
		return "proidentity"
	}
	if settings.NativeWebAuthnEnabled {
		return "webauthn"
	}
	if settings.LocalTOTPEnabled && strings.TrimSpace(settings.LocalTOTPSecret) != "" {
		return "totp"
	}
	return ""
}

func (h handler) addAdminWebAuthnCount(ctx context.Context, settings domain.AdminMFASettings) domain.AdminMFASettings {
	if h.store == nil {
		return settings
	}
	credentials, err := h.store.ListAdminWebAuthnCredentials(ctx)
	if err == nil {
		settings.NativeWebAuthnCredentialCount = len(credentials)
	}
	return settings
}

func mfaProviders(settings domain.AdminMFASettings) []string {
	if effectiveMFAProvider(settings) == "proidentity" {
		return []string{"proidentity_push", "proidentity_totp"}
	}
	var providers []string
	if settings.NativeWebAuthnEnabled {
		providers = append(providers, "webauthn")
	}
	if settings.LocalTOTPEnabled && strings.TrimSpace(settings.LocalTOTPSecret) != "" {
		providers = append(providers, "totp")
	}
	return providers
}

func publicAdminMFASettings(settings domain.AdminMFASettings) domain.AdminMFASettings {
	settings.ProIdentityBaseURL = proIdentityAuthServiceURL
	settings.ProIdentityAPIKeyConfigured = strings.TrimSpace(settings.ProIdentityAPIKey) != ""
	settings.EffectiveProvider = effectiveMFAProvider(settings)
	settings.NativeWebAuthnEnabled = settings.NativeWebAuthnEnabled && settings.NativeWebAuthnCredentialCount > 0
	settings.LocalTOTPSecret = ""
	settings.LocalTOTPPendingSecret = ""
	settings.ProIdentityAPIKey = ""
	if settings.ProIdentityTimeoutSeconds <= 0 {
		settings.ProIdentityTimeoutSeconds = defaultProIdentityTimeoutSeconds
	}
	return settings
}

func mfaTimeout(settings domain.AdminMFASettings) time.Duration {
	seconds := settings.ProIdentityTimeoutSeconds
	if seconds <= 0 {
		seconds = defaultProIdentityTimeoutSeconds
	}
	if seconds < 30 {
		seconds = 30
	}
	if seconds > 300 {
		seconds = 300
	}
	return time.Duration(seconds) * time.Second
}

func createTOTPEnrollment(username string) (string, string, string, error) {
	account := strings.TrimSpace(username)
	if account == "" {
		account = "admin"
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ProIdentity Mail Admin",
		AccountName: account,
		Period:      30,
		Digits:      6,
		SecretSize:  20,
	})
	if err != nil {
		return "", "", "", err
	}
	img, err := key.Image(220, 220)
	if err != nil {
		return "", "", "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", "", err
	}
	return key.Secret(), key.URL(), "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func verifyTOTPCode(secret, code string) bool {
	secret = strings.TrimSpace(secret)
	code = strings.TrimSpace(code)
	if secret == "" || code == "" {
		return false
	}
	return totp.Validate(code, secret)
}

func randomHexToken(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func requestClientIP(r *http.Request) string {
	remote := requestRemoteIP(r.RemoteAddr)
	if isLoopbackIP(remote) {
		if ip := headerClientIP(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
		if ip := forwardedForClientIP(r.Header.Get("X-Forwarded-For")); ip != "" {
			return ip
		}
	}
	return remote
}

func requestRemoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return strings.TrimSpace(host)
}

func isLoopbackIP(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	return ip != nil && ip.IsLoopback()
}

func headerClientIP(value string) string {
	value = strings.Trim(strings.TrimSpace(value), `"`)
	if value == "" {
		return ""
	}
	if ip := net.ParseIP(value); ip != nil {
		return ip.String()
	}
	host, _, err := net.SplitHostPort(value)
	if err == nil {
		if ip := net.ParseIP(strings.Trim(host, "[]")); ip != nil {
			return ip.String()
		}
	}
	return ""
}

func forwardedForClientIP(value string) string {
	parts := strings.Split(value, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		if ip := headerClientIP(parts[i]); ip != "" {
			return ip
		}
	}
	return ""
}

func validateProIdentitySettings(settings domain.AdminMFASettings) error {
	baseURL, err := url.Parse(proIdentityRequestBaseURL(settings))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" || baseURL.User != nil {
		return errors.New("valid ProIdentity Auth URL is required")
	}
	if baseURL.Scheme != "https" && !isLoopbackHost(baseURL.Hostname()) {
		return errors.New("ProIdentity Auth URL must use https unless it is loopback")
	}
	if strings.TrimSpace(settings.ProIdentityAPIKey) == "" {
		return errors.New("ProIdentity Auth API key is required")
	}
	if _, ok := normalizeEmailAddress(settings.ProIdentityUserEmail); !ok {
		return errors.New("valid ProIdentity Auth admin email is required")
	}
	return nil
}

func proIdentityRequestBaseURL(settings domain.AdminMFASettings) string {
	baseURL := strings.TrimRight(strings.TrimSpace(settings.ProIdentityBaseURL), "/")
	if baseURL == "" {
		return proIdentityAuthServiceURL
	}
	parsed, err := url.Parse(baseURL)
	if err == nil && parsed.Host != "" && parsed.User == nil && isLoopbackHost(parsed.Hostname()) {
		return baseURL
	}
	return proIdentityAuthServiceURL
}

func isLoopbackHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

type proIdentityAuthRequestPayload struct {
	UserEmail     string `json:"user_email"`
	DisplayName   string `json:"display_name"`
	ContextTitle  string `json:"context_title"`
	ContextDetail string `json:"context_detail"`
	ClientIP      string `json:"client_ip"`
}

type proIdentityAuthRequestCreated struct {
	RequestID string
	ExpiresAt time.Time
}

type proIdentityAPIError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e proIdentityAPIError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	switch e.Code {
	case "user_created_needs_setup":
		return "ProIdentity user was created, but must finish mobile app setup before push approval can be sent."
	case "email_domain_not_trusted":
		return "Email domain is not trusted in ProIdentity Auth."
	default:
		if strings.TrimSpace(e.Code) != "" {
			return "ProIdentity Auth returned " + e.Code
		}
		if e.HTTPStatus > 0 {
			return fmt.Sprintf("ProIdentity Auth returned HTTP %d", e.HTTPStatus)
		}
		return "ProIdentity Auth request failed"
	}
}

func writeProIdentityAPIError(w http.ResponseWriter, err error) bool {
	var apiErr proIdentityAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	status := http.StatusBadGateway
	switch apiErr.Code {
	case "user_created_needs_setup":
		status = http.StatusPreconditionRequired
	case "email_domain_not_trusted":
		status = http.StatusConflict
	default:
		if apiErr.HTTPStatus == http.StatusUnauthorized || apiErr.HTTPStatus == http.StatusForbidden {
			status = http.StatusBadGateway
		} else if apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
			status = http.StatusConflict
		}
	}
	writeJSON(w, status, map[string]string{
		"error": apiErr.Error(),
		"code":  apiErr.Code,
	})
	return true
}

type proIdentityAuthRequestStatus struct {
	RequestID    string `json:"request_id"`
	Status       string `json:"status"`
	TOTPVerified bool   `json:"totp_verified"`
	RespondedAt  int64  `json:"responded_at"`
}

func createProIdentityAuthRequest(ctx context.Context, settings domain.AdminMFASettings, payload proIdentityAuthRequestPayload) (proIdentityAuthRequestCreated, error) {
	var out struct {
		RequestID string `json:"request_id"`
		ExpiresAt int64  `json:"expires_at"`
		Status    string `json:"status"`
		Code      string `json:"code"`
		Message   string `json:"message"`
	}
	if err := proIdentityJSON(ctx, http.MethodPost, settings, "/api/v1/sp/auth-requests", payload, &out); err != nil {
		return proIdentityAuthRequestCreated{}, err
	}
	if out.Code != "" {
		return proIdentityAuthRequestCreated{}, proIdentityAPIError{Code: out.Code, Message: out.Message}
	}
	if strings.TrimSpace(out.RequestID) == "" {
		return proIdentityAuthRequestCreated{}, errors.New("ProIdentity Auth did not return request id")
	}
	var expiresAt time.Time
	if out.ExpiresAt > 0 {
		expiresAt = time.Unix(out.ExpiresAt, 0).UTC()
	}
	return proIdentityAuthRequestCreated{RequestID: out.RequestID, ExpiresAt: expiresAt}, nil
}

func readProIdentityAuthRequestStatus(ctx context.Context, settings domain.AdminMFASettings, requestID string) (proIdentityAuthRequestStatus, error) {
	var out proIdentityAuthRequestStatus
	if strings.TrimSpace(requestID) == "" {
		return out, errors.New("missing ProIdentity Auth request id")
	}
	if err := proIdentityJSON(ctx, http.MethodGet, settings, "/api/v1/sp/auth-requests/"+url.PathEscape(requestID), nil, &out); err != nil {
		return out, err
	}
	return out, nil
}

func createProIdentityTOTPEnrollment(ctx context.Context, settings domain.AdminMFASettings, issuer, accountName string) (map[string]any, error) {
	payload := map[string]any{
		"user_email":   settings.ProIdentityUserEmail,
		"issuer":       issuer,
		"account_name": accountName,
		"algorithm":    "SHA1",
		"digits":       6,
		"period":       30,
	}
	var out map[string]any
	if err := proIdentityJSON(ctx, http.MethodPost, settings, "/api/v1/sp/totp/enrollments", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func verifyProIdentityTOTP(ctx context.Context, settings domain.AdminMFASettings, code string) (bool, error) {
	var out struct {
		Verified  bool   `json:"verified"`
		UserEmail string `json:"user_email"`
	}
	payload := map[string]string{
		"user_email": settings.ProIdentityUserEmail,
		"code":       strings.TrimSpace(code),
	}
	if err := proIdentityJSON(ctx, http.MethodPost, settings, "/api/v1/sp/verify-totp", payload, &out); err != nil {
		return false, err
	}
	return out.Verified && strings.EqualFold(strings.TrimSpace(out.UserEmail), strings.TrimSpace(settings.ProIdentityUserEmail)), nil
}

func proIdentityJSON(ctx context.Context, method string, settings domain.AdminMFASettings, path string, payload any, out any) error {
	if err := validateProIdentitySettings(settings); err != nil {
		return err
	}
	baseURL, _ := url.Parse(proIdentityRequestBaseURL(settings))
	endpoint, _ := url.Parse(path)
	target := baseURL.ResolveReference(endpoint).String()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, target, &body)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", settings.ProIdentityAPIKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return proIdentityAPIError{Code: errBody.Code, Message: errBody.Message, HTTPStatus: resp.StatusCode}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
