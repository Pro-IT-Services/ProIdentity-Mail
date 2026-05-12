package webmail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"proidentity-mail/internal/domain"
)

const proIdentityAuthServiceURL = "https://verify.proidentity.cloud"

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

func writeProIdentityMailboxAPIError(w http.ResponseWriter, err error) bool {
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
		if apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 && apiErr.HTTPStatus != http.StatusUnauthorized && apiErr.HTTPStatus != http.StatusForbidden {
			status = http.StatusConflict
		}
	}
	writeJSON(w, status, map[string]string{
		"error": apiErr.Error(),
		"code":  apiErr.Code,
	})
	return true
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

type proIdentityAuthRequestStatus struct {
	RequestID    string `json:"request_id"`
	Status       string `json:"status"`
	TOTPVerified bool   `json:"totp_verified"`
	RespondedAt  int64  `json:"responded_at"`
}

func createMailboxProIdentityAuthRequest(ctx context.Context, settings domain.AdminMFASettings, payload proIdentityAuthRequestPayload) (proIdentityAuthRequestCreated, error) {
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

func readMailboxProIdentityAuthRequestStatus(ctx context.Context, settings domain.AdminMFASettings, requestID string) (proIdentityAuthRequestStatus, error) {
	var out proIdentityAuthRequestStatus
	if strings.TrimSpace(requestID) == "" {
		return out, errors.New("missing ProIdentity Auth request id")
	}
	if err := proIdentityJSON(ctx, http.MethodGet, settings, "/api/v1/sp/auth-requests/"+url.PathEscape(requestID), nil, &out); err != nil {
		return out, err
	}
	return out, nil
}

func verifyMailboxProIdentityTOTP(ctx context.Context, settings domain.AdminMFASettings, email, code string) (bool, error) {
	var out struct {
		Verified  bool   `json:"verified"`
		UserEmail string `json:"user_email"`
	}
	payload := map[string]string{
		"user_email": strings.ToLower(strings.TrimSpace(email)),
		"code":       strings.TrimSpace(code),
	}
	if err := proIdentityJSON(ctx, http.MethodPost, settings, "/api/v1/sp/verify-totp", payload, &out); err != nil {
		return false, err
	}
	return out.Verified && strings.EqualFold(strings.TrimSpace(out.UserEmail), strings.TrimSpace(email)), nil
}

func proIdentityJSON(ctx context.Context, method string, settings domain.AdminMFASettings, path string, payload any, out any) error {
	if err := validateMailboxProIdentitySettings(settings); err != nil {
		return err
	}
	baseURL, _ := url.Parse(mailboxProIdentityRequestBaseURL(settings))
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

func validateMailboxProIdentitySettings(settings domain.AdminMFASettings) error {
	baseURL, err := url.Parse(mailboxProIdentityRequestBaseURL(settings))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" || baseURL.User != nil {
		return errors.New("valid ProIdentity Auth URL is required")
	}
	if baseURL.Scheme != "https" && !isLoopbackHost(baseURL.Hostname()) {
		return errors.New("ProIdentity Auth URL must use https unless it is loopback")
	}
	if strings.TrimSpace(settings.ProIdentityAPIKey) == "" {
		return errors.New("ProIdentity Auth API key is required")
	}
	return nil
}

func mailboxProIdentityRequestBaseURL(settings domain.AdminMFASettings) string {
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

func isLoopbackHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
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
