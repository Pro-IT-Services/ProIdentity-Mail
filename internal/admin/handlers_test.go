package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"proidentity-mail/internal/configdrift"
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

func TestAdminIndexIncludesGlobalShowAllScopeOptions(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{[]byte("All tenants"), []byte("All domains")} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing global scope option %q", want)
		}
	}
}

func TestAdminIndexIncludesMailboxQuotaUnitSelector(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte("Mailbox storage quota"),
		[]byte("Storage quota"),
		[]byte(`name=\"quota_value\"`),
		[]byte(`name=\"quota_unit\"`),
		[]byte(">MB<"),
		[]byte(">GB<"),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing mailbox quota UI %q", want)
		}
	}
}

func TestAdminIndexIncludesResourceEditAndRemoveControls(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte(`id="modal"`),
		[]byte(`data-edit=`),
		[]byte(`data-delete=`),
		[]byte("Edit tenant"),
		[]byte("Remove"),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing admin edit/remove UI %q", want)
		}
	}
}

func TestAdminIndexUsesModalsForCreateAndDNSFlows(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte(`data-create=`),
		[]byte(`function openCreate`),
		[]byte(`function openDNSModal`),
		[]byte(`function openCloudflareSettings`),
		[]byte(`function openCloudflareProvisionModal`),
		[]byte(`function cloudflareBusyStep`),
		[]byte(`Applying Cloudflare DNS`),
		[]byte(`Applying DNS...`),
		[]byte(`data-cloudflare-settings`),
		[]byte(`data-cloudflare-provision`),
		[]byte(`function renderClientSetup`),
		[]byte(`createAction("domain"`),
		[]byte(`data-create=\"user\"`),
		[]byte(`createAction("shared-permission"`),
		[]byte(`setView("users")`),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing modal create/DNS UI %q", want)
		}
	}
}

func TestAdminIndexUsesMailServerBehaviorSettings(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte("Mail server behavior"),
		[]byte("/api/v1/mail-server-settings"),
		[]byte(`data-save-mail-settings`),
		[]byte("Default interface language"),
		[]byte("supportedLanguages"),
		[]byte("Slovenčina"),
		[]byte("i18nCatalog"),
		[]byte("translateUI"),
		[]byte("Enable SNI certificate maps"),
		[]byte("HTTPS and SSL"),
		[]byte("Existing certificate"),
		[]byte("Force HTTPS redirect"),
		[]byte("/api/v1/tls/certificates"),
		[]byte("Mailbox two-factor authentication"),
		[]byte("Force mailbox 2FA setup"),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing mail server behavior UI %q", want)
		}
	}
	if bytes.Contains(body, []byte("Client endpoints")) {
		t.Fatalf("system settings still show old client endpoints panel")
	}
}

func TestAdminIndexIncludesConfigDriftUI(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte("Config drift"),
		[]byte("/api/v1/system/config-drift"),
		[]byte("/api/v1/system/config-apply"),
		[]byte("data-check-config-drift"),
		[]byte("data-apply-config-drift"),
		[]byte("configApplyInProgress"),
		[]byte("Configuration reload completed"),
		[]byte("Live system differs from database"),
		[]byte("diff-box"),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing config drift UI %q", want)
		}
	}
}

func TestConfigDriftEndpointReportsLiveDifferences(t *testing.T) {
	root := t.TempDir()
	liveRoot := filepath.Join(root, "live")
	rendered := map[string]string{}
	runner := func(ctx context.Context, command string, args ...string) ([]byte, error) {
		if command != "mailctl" {
			return nil, fmt.Errorf("unexpected command %q", command)
		}
		if len(args) == 0 {
			return nil, errors.New("missing subcommand")
		}
		targetDir := commandTargetDir(args)
		if targetDir == "" {
			return nil, fmt.Errorf("missing target dir for %v", args)
		}
		rendered[args[0]] = targetDir
		if rendered["render"] != "" && rendered["render-proxy"] != "" {
			writeConfigDriftFixture(t, rendered["render"], rendered["render-proxy"], liveRoot)
		}
		return []byte("ok"), nil
	}
	handler := NewRouter(&fakeStore{}, AuthConfig{System: SystemConfig{MailctlPath: "mailctl", LiveRoot: liveRoot, CommandRunner: runner}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/config-drift", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var report configdrift.Report
	if err := json.NewDecoder(rec.Body).Decode(&report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if report.Status != "drift" || report.Summary.Drifted != 1 {
		t.Fatalf("expected one drifted item, got %+v", report.Summary)
	}
	found := false
	for _, item := range report.Items {
		if item.ID == "postfix-main" {
			found = item.Status == "drift" && strings.Contains(item.Diff, "-myhostname = old.example") && strings.Contains(item.Diff, "+myhostname = mail.example")
		}
	}
	if !found {
		t.Fatalf("postfix-main drift not reported correctly: %+v", report.Items)
	}
}

func TestConfigApplyEndpointQueuesRootApplyRequest(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "apply-request")
	handler := NewRouter(&fakeStore{}, AuthConfig{System: SystemConfig{ConfigApplyRequestPath: marker}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/config-apply", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("apply marker was not written: %v", err)
	}
	if !strings.Contains(string(data), "requested_at=") {
		t.Fatalf("apply marker missing request timestamp: %q", string(data))
	}
}

func TestAdminIndexIncludesDomainTLSManagement(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte("Certificates"),
		[]byte("/api/v1/domains/"),
		[]byte("/tls/settings"),
		[]byte(`data-tls-request`),
		[]byte(`data-tls-save`),
		[]byte(`data-tls-queue`),
		[]byte("Publish optional webmail."),
		[]byte("Request progress"),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing TLS management UI %q", want)
		}
	}
}

func TestAdminIndexUsesReadableSplitAuditView(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte("Admin & access"),
		[]byte("Mail security"),
		[]byte("User activity"),
		[]byte("System & DNS"),
		[]byte(`data-audit-tab`),
		[]byte(`data-audit-search`),
		[]byte(`data-audit-action`),
		[]byte(`function auditCard`),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing split readable audit UI %q", want)
		}
	}
	if bytes.Contains(body, []byte(`table(["Action", "Actor", "Target", "Tenant", "Metadata", "Date"]`)) {
		t.Fatalf("audit view still renders raw metadata table")
	}
}

func TestAdminIndexIncludesNativeLoginProtectionUI(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte("Login protection"),
		[]byte("Locked accounts"),
		[]byte("/api/v1/security/login-rate-limits"),
		[]byte(`data-unlock-user`),
		[]byte(`data-clear-rate-limit`),
		[]byte(`function unlockUser`),
		[]byte(`function clearRateLimit`),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing native login protection UI %q", want)
		}
	}
}

func TestAdminIndexIncludesMFASettingsAndProIdentityAuthTab(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.Bytes()
	for _, want := range [][]byte{
		[]byte("Admin MFA"),
		[]byte("ProIdentity Auth"),
		[]byte("/api/v1/admin-mfa/settings"),
		[]byte("/api/v1/admin-mfa/totp/enroll"),
		[]byte("/api/v1/admin-mfa/totp/verify"),
		[]byte("/api/v1/admin-mfa/proidentity"),
		[]byte("/api/v1/admin-mfa/proidentity/totp/verify"),
		[]byte("/api/v1/admin-mfa/proidentity/totp/confirm"),
		[]byte("/api/v1/admin-mfa/webauthn/register/begin"),
		[]byte("/api/v1/admin-mfa/webauthn/register/finish"),
		[]byte("/api/v1/session/step-up"),
		[]byte("/api/v1/session/step-up/verify"),
		[]byte("/api/v1/session/step-up/webauthn"),
		[]byte("Confirm admin action"),
		[]byte("Native hardware keys"),
		[]byte(`id="mfa-panel"`),
		[]byte(`id="push-mfa-view"`),
		[]byte("Push Verification"),
		[]byte("Enter code manually"),
		[]byte("showProIdentityPushView"),
		[]byte(`id="login-status"`),
		[]byte("Checking credentials..."),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("index missing MFA settings UI %q", want)
		}
	}
	if bytes.Contains(body, []byte(`prompt("Enter your 6-digit authenticator code"`)) {
		t.Fatalf("login flow still uses browser prompt for MFA")
	}
	if bytes.Contains(body, []byte("You can also enter a hosted TOTP code")) {
		t.Fatalf("proidentity push flow shows TOTP code entry by default")
	}
	if bytes.Contains(body, []byte(`.login-cover.push-mode { background: #0f1218`)) {
		t.Fatalf("proidentity push flow should use the light admin login design")
	}
	if !bytes.Contains(body, []byte("https://verify.proidentity.cloud")) {
		t.Fatalf("proidentity auth tab should show the fixed service URL")
	}
	if bytes.Contains(body, []byte(`name="base_url"`)) {
		t.Fatalf("proidentity auth service URL should not be editable")
	}
	if !bytes.Contains(body, []byte("toast.textContent = t(message)")) {
		t.Fatalf("admin status messages should pass through i18n")
	}
	for _, want := range [][]byte{
		[]byte(`confirm(t("Unlock this mailbox user and clear its failed-login limiter entries?"))`),
		[]byte(`confirm(t("Reset this user's mailbox 2FA? They will need to set it up again when force 2FA is enabled."))`),
		[]byte(`confirm(t("Clear this login protection entry?"))`),
		[]byte(`setPushMFAStatus(t("Waiting for approval..."))`),
	} {
		if !bytes.Contains(body, want) {
			t.Fatalf("admin dynamic i18n hook missing %q", want)
		}
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

func TestAdminLoginFromNewClientIPRecordsSecurityAlert(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{auditEvents: []domain.AuditEvent{}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	login.RemoteAddr = "203.0.113.7:48123"
	loginRec := httptest.NewRecorder()

	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	if !containsString(store.auditActions, "security.alert.admin_new_ip") {
		t.Fatalf("new-ip security alert missing from audit actions: %+v", store.auditActions)
	}
	if !containsString(store.auditActions, "admin.login") {
		t.Fatalf("admin login audit missing from audit actions: %+v", store.auditActions)
	}
}

func TestAdminLoginFromKnownClientIPDoesNotRecordSecurityAlert(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{auditEvents: []domain.AuditEvent{{
		Action:       "admin.login",
		TargetID:     "admin",
		MetadataJSON: `{"client_ip":"203.0.113.7"}`,
	}}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	login.RemoteAddr = "203.0.113.7:48123"
	loginRec := httptest.NewRecorder()

	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	if containsString(store.auditActions, "security.alert.admin_new_ip") {
		t.Fatalf("known-ip login should not record new-ip alert: %+v", store.auditActions)
	}
}

func TestDangerousAdminOperationRequiresFreshStepUp(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	base := httptest.NewRequest(http.MethodGet, "/", nil)
	base.Header.Set("User-Agent", "Browser A")
	created, err := manager.Create(base, "admin", "admin")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/domains/22/cloudflare/apply", strings.NewReader(`{"replace":false}`))
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("X-CSRF-Token", created.CSRFToken)
	req.AddCookie(created.Cookie)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusPreconditionRequired {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusPreconditionRequired, rec.Body.String())
	}
}

func TestAdminStepUpFlowAllowsDangerousOperation(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{adminMFASettings: domain.AdminMFASettings{LocalTOTPEnabled: true, LocalTOTPSecret: firstTestTOTPSecret()}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	base := httptest.NewRequest(http.MethodGet, "/", nil)
	base.Header.Set("User-Agent", "Browser A")
	created, err := manager.Create(base, "admin", "admin")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	begin := httptest.NewRequest(http.MethodPost, "/api/v1/session/step-up", nil)
	begin.Header.Set("User-Agent", "Browser A")
	begin.Header.Set("X-CSRF-Token", created.CSRFToken)
	begin.AddCookie(created.Cookie)
	beginRec := httptest.NewRecorder()
	handler.ServeHTTP(beginRec, begin)
	if beginRec.Code != http.StatusOK {
		t.Fatalf("begin status = %d, want %d, body %s", beginRec.Code, http.StatusOK, beginRec.Body.String())
	}
	var beginResponse struct {
		StepUpRequired bool   `json:"step_up_required"`
		MFAToken       string `json:"mfa_token"`
	}
	if err := json.NewDecoder(beginRec.Body).Decode(&beginResponse); err != nil {
		t.Fatalf("decode begin response: %v", err)
	}
	if !beginResponse.StepUpRequired || beginResponse.MFAToken == "" {
		t.Fatalf("unexpected step-up begin response: %+v", beginResponse)
	}
	code, err := generateTOTPCode(firstTestTOTPSecret())
	if err != nil {
		t.Fatalf("generate totp: %v", err)
	}
	verify := httptest.NewRequest(http.MethodPost, "/api/v1/session/step-up/verify", strings.NewReader(`{"mfa_token":"`+beginResponse.MFAToken+`","code":"`+code+`"}`))
	verify.Header.Set("User-Agent", "Browser A")
	verify.Header.Set("X-CSRF-Token", created.CSRFToken)
	verify.AddCookie(created.Cookie)
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/domains/22/cloudflare/apply", strings.NewReader(`{"replace":true}`))
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("X-CSRF-Token", created.CSRFToken)
	req.AddCookie(created.Cookie)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dangerous operation status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !store.appliedCloudflareReplace {
		t.Fatal("cloudflare apply did not run after step-up")
	}
	if !containsString(store.auditActions, "admin.mfa.step_up") {
		t.Fatalf("step-up audit action missing: %+v", store.auditActions)
	}
}

func TestAdminLoginRequiresLocalTOTPWhenEnabled(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{adminMFASettings: domain.AdminMFASettings{LocalTOTPEnabled: true, LocalTOTPSecret: "JBSWY3DPEHPK3PXP"}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})

	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	login.Header.Set("Content-Type", "application/json")
	login.RemoteAddr = "203.0.113.8:48121"
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	if cookies := loginRec.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("password step should not create session cookie before MFA, got %d cookie(s)", len(cookies))
	}
	var first struct {
		MFARequired bool   `json:"mfa_required"`
		Provider    string `json:"provider"`
		MFAToken    string `json:"mfa_token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&first); err != nil {
		t.Fatalf("decode password step: %v", err)
	}
	if !first.MFARequired || first.Provider != "totp" || first.MFAToken == "" {
		t.Fatalf("unexpected password step response: %+v", first)
	}
	code, err := generateTOTPCode(firstTestTOTPSecret())
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	verifyBody := fmt.Sprintf(`{"mfa_token":%q,"code":%q}`, first.MFAToken, code)
	verify := httptest.NewRequest(http.MethodPost, "/api/v1/session/mfa", strings.NewReader(verifyBody))
	verify.Header.Set("Content-Type", "application/json")
	verify.RemoteAddr = "203.0.113.8:48121"
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)

	if verifyRec.Code != http.StatusOK {
		t.Fatalf("mfa status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if cookies := verifyRec.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("mfa verification did not create admin session cookie")
	}
	var second struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.NewDecoder(verifyRec.Body).Decode(&second); err != nil {
		t.Fatalf("decode mfa response: %v", err)
	}
	if second.CSRFToken == "" {
		t.Fatal("csrf token is empty after mfa verification")
	}
}

func TestAdminLoginUsesProIdentityAuthWhenConfigured(t *testing.T) {
	var requestCreated bool
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/auth-requests":
			requestCreated = true
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode proidentity request: %v", err)
			}
			if body["user_email"] != "admin@example.com" || body["context_title"] == "" || body["client_ip"] == "" {
				t.Fatalf("unexpected proidentity request body: %+v", body)
			}
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "push-1", "expires_at": time.Now().Add(time.Minute).Unix(), "status": "pending"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/sp/auth-requests/push-1":
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "push-1", "status": "approved", "totp_verified": true, "responded_at": time.Now().Unix()})
		default:
			http.NotFound(w, r)
		}
	}))
	defer authServer.Close()

	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{adminMFASettings: domain.AdminMFASettings{
		ProIdentityEnabled:        true,
		ProIdentityBaseURL:        authServer.URL,
		ProIdentityAPIKey:         "sp-secret",
		ProIdentityUserEmail:      "admin@example.com",
		ProIdentityTimeoutSeconds: 60,
	}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})

	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	login.Header.Set("Content-Type", "application/json")
	login.RemoteAddr = "198.51.100.9:38123"
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	if !requestCreated {
		t.Fatal("proidentity auth request was not created")
	}
	var first struct {
		MFARequired bool   `json:"mfa_required"`
		Provider    string `json:"provider"`
		MFAToken    string `json:"mfa_token"`
		RequestID   string `json:"request_id"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&first); err != nil {
		t.Fatalf("decode password step: %v", err)
	}
	if !first.MFARequired || first.Provider != "proidentity" || first.MFAToken == "" || first.RequestID != "push-1" {
		t.Fatalf("unexpected password step response: %+v", first)
	}

	verifyBody := fmt.Sprintf(`{"mfa_token":%q}`, first.MFAToken)
	verify := httptest.NewRequest(http.MethodPost, "/api/v1/session/mfa", strings.NewReader(verifyBody))
	verify.Header.Set("Content-Type", "application/json")
	verify.RemoteAddr = "198.51.100.9:38123"
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("mfa status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if cookies := verifyRec.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("proidentity approval did not create admin session cookie")
	}
}

func TestAdminLoginReportsProIdentityAutocreateSetupNeeded(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/sp/auth-requests" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"code":       "user_created_needs_setup",
			"message":    "",
			"user_email": "newadmin@example.com",
		})
	}))
	defer authServer.Close()

	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{adminMFASettings: domain.AdminMFASettings{
		ProIdentityEnabled:        true,
		ProIdentityBaseURL:        authServer.URL,
		ProIdentityAPIKey:         "sp-secret",
		ProIdentityUserEmail:      "newadmin@example.com",
		ProIdentityTimeoutSeconds: 60,
	}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})

	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	login.Header.Set("Content-Type", "application/json")
	login.RemoteAddr = "127.0.0.1:38123"
	login.Header.Set("X-Real-IP", "203.0.113.99")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusPreconditionRequired {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusPreconditionRequired, loginRec.Body.String())
	}
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != "user_created_needs_setup" || !strings.Contains(body.Error, "mobile app setup") {
		t.Fatalf("unexpected setup-needed response: %+v", body)
	}
	if cookies := loginRec.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("setup-needed response should not create a session, got %d cookie(s)", len(cookies))
	}
}

func TestAdminLoginReportsProIdentityUntrustedDomain(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/sp/auth-requests" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"code":    "email_domain_not_trusted",
			"message": "Email domain is not trusted for this company.",
		})
	}))
	defer authServer.Close()

	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{adminMFASettings: domain.AdminMFASettings{
		ProIdentityEnabled:        true,
		ProIdentityBaseURL:        authServer.URL,
		ProIdentityAPIKey:         "sp-secret",
		ProIdentityUserEmail:      "admin@outside.example",
		ProIdentityTimeoutSeconds: 60,
	}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})

	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	login.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)

	if loginRec.Code != http.StatusConflict {
		t.Fatalf("login status = %d, want %d, body %s", loginRec.Code, http.StatusConflict, loginRec.Body.String())
	}
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != "email_domain_not_trusted" || !strings.Contains(body.Error, "not trusted") {
		t.Fatalf("unexpected untrusted-domain response: %+v", body)
	}
}

func TestAdminLoginCanVerifyProIdentityHostedTOTP(t *testing.T) {
	var statusRead bool
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/auth-requests":
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "push-2", "expires_at": time.Now().Add(time.Minute).Unix(), "status": "pending"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/verify-totp":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode verify totp request: %v", err)
			}
			if body["user_email"] != "admin@example.com" || body["code"] != "123456" {
				t.Fatalf("unexpected verify totp body: %+v", body)
			}
			writeJSON(w, http.StatusOK, map[string]any{"verified": true, "user_email": "admin@example.com"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/sp/auth-requests/push-2":
			statusRead = true
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "push-2", "status": "pending"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer authServer.Close()

	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{adminMFASettings: domain.AdminMFASettings{
		ProIdentityEnabled:        true,
		ProIdentityBaseURL:        authServer.URL,
		ProIdentityAPIKey:         "sp-secret",
		ProIdentityUserEmail:      "admin@example.com",
		ProIdentityTimeoutSeconds: 60,
	}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	login.Header.Set("Content-Type", "application/json")
	login.RemoteAddr = "198.51.100.9:38123"
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, login)
	var first struct {
		MFAToken string `json:"mfa_token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&first); err != nil {
		t.Fatalf("decode password step: %v", err)
	}
	verify := httptest.NewRequest(http.MethodPost, "/api/v1/session/mfa", strings.NewReader(fmt.Sprintf(`{"mfa_token":%q,"code":"123456"}`, first.MFAToken)))
	verify.Header.Set("Content-Type", "application/json")
	verify.RemoteAddr = "198.51.100.9:38123"
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("mfa status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if statusRead {
		t.Fatal("hosted TOTP verification should not poll push status")
	}
	if cookies := verifyRec.Result().Cookies(); len(cookies) == 0 {
		t.Fatal("proidentity hosted totp did not create admin session cookie")
	}
}

func TestProIdentityHostedTOTPEnrollmentRequiresCodeThenPushBeforeSaving(t *testing.T) {
	var verifyCalled bool
	var pushCreated bool
	var statusRead bool
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sp-secret" {
			http.Error(w, "bad api key", http.StatusForbidden)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/totp/enrollments":
			writeJSON(w, http.StatusOK, map[string]any{"otpauth_url": "otpauth://totp/ProIdentity:admin@example.com", "qr_data_url": "data:image/png;base64,abc"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/verify-totp":
			verifyCalled = true
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode verify body: %v", err)
			}
			if body["user_email"] != "admin@example.com" || body["code"] != "123456" {
				t.Fatalf("unexpected verify body: %+v", body)
			}
			writeJSON(w, http.StatusOK, map[string]any{"verified": true, "user_email": "admin@example.com"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sp/auth-requests":
			pushCreated = true
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode push body: %v", err)
			}
			if body["context_title"] != "Confirm ProIdentity hosted TOTP setup" {
				t.Fatalf("unexpected push context: %+v", body)
			}
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "enroll-push-1", "expires_at": time.Now().Add(time.Minute).Unix(), "status": "pending"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/sp/auth-requests/enroll-push-1":
			statusRead = true
			writeJSON(w, http.StatusOK, map[string]any{"request_id": "enroll-push-1", "status": "approved", "responded_at": time.Now().Unix()})
		default:
			http.NotFound(w, r)
		}
	}))
	defer authServer.Close()

	store := &fakeStore{adminMFASettings: domain.AdminMFASettings{
		ProIdentityEnabled:        true,
		ProIdentityBaseURL:        authServer.URL,
		ProIdentityAPIKey:         "sp-secret",
		ProIdentityUserEmail:      "admin@example.com",
		ProIdentityTimeoutSeconds: 60,
	}}
	handler := NewRouter(store)

	enroll := httptest.NewRequest(http.MethodPost, "/api/v1/admin-mfa/proidentity/totp/enroll", nil)
	enrollRec := httptest.NewRecorder()
	handler.ServeHTTP(enrollRec, enroll)
	if enrollRec.Code != http.StatusOK {
		t.Fatalf("enroll status = %d, want %d, body %s", enrollRec.Code, http.StatusOK, enrollRec.Body.String())
	}
	if store.adminMFASettings.ProIdentityTOTPEnabled {
		t.Fatal("hosted TOTP should not be saved before code and push approval")
	}

	verify := httptest.NewRequest(http.MethodPost, "/api/v1/admin-mfa/proidentity/totp/verify", strings.NewReader(`{"code":"123456"}`))
	verify.Header.Set("Content-Type", "application/json")
	verify.RemoteAddr = "203.0.113.44:49101"
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusAccepted {
		t.Fatalf("verify status = %d, want %d, body %s", verifyRec.Code, http.StatusAccepted, verifyRec.Body.String())
	}
	if !verifyCalled || !pushCreated {
		t.Fatalf("verifyCalled=%t pushCreated=%t, want both true", verifyCalled, pushCreated)
	}
	if store.adminMFASettings.ProIdentityTOTPEnabled {
		t.Fatal("hosted TOTP should not be saved until push approval is confirmed")
	}
	var pending struct {
		MFAToken  string `json:"mfa_token"`
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(verifyRec.Body).Decode(&pending); err != nil {
		t.Fatalf("decode verify response: %v", err)
	}
	if pending.MFAToken == "" || pending.RequestID != "enroll-push-1" || pending.Status != "pending" {
		t.Fatalf("unexpected verify response: %+v", pending)
	}

	confirm := httptest.NewRequest(http.MethodPost, "/api/v1/admin-mfa/proidentity/totp/confirm", strings.NewReader(fmt.Sprintf(`{"mfa_token":%q}`, pending.MFAToken)))
	confirm.Header.Set("Content-Type", "application/json")
	confirmRec := httptest.NewRecorder()
	handler.ServeHTTP(confirmRec, confirm)
	if confirmRec.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want %d, body %s", confirmRec.Code, http.StatusOK, confirmRec.Body.String())
	}
	if !statusRead {
		t.Fatal("push status was not checked before saving hosted TOTP")
	}
	if !store.adminMFASettings.ProIdentityTOTPEnabled {
		t.Fatal("hosted TOTP was not saved after code and push approval")
	}
	var settings domain.AdminMFASettings
	if err := json.NewDecoder(confirmRec.Body).Decode(&settings); err != nil {
		t.Fatalf("decode confirm response: %v", err)
	}
	if !settings.ProIdentityTOTPEnabled {
		t.Fatalf("public settings did not show hosted TOTP enabled: %+v", settings)
	}
}

func TestAdminTOTPEnrollmentReturnsQRAndVerifyEnablesMFA(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin-mfa/totp/enroll", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("enroll status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var enrollment struct {
		Secret     string `json:"secret"`
		OTPAuthURL string `json:"otpauth_url"`
		QRDataURL  string `json:"qr_data_url"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&enrollment); err != nil {
		t.Fatalf("decode enrollment: %v", err)
	}
	if enrollment.Secret == "" || !strings.HasPrefix(enrollment.OTPAuthURL, "otpauth://totp/") || !strings.HasPrefix(enrollment.QRDataURL, "data:image/png;base64,") {
		t.Fatalf("unexpected enrollment: %+v", enrollment)
	}
	code, err := generateTOTPCode(enrollment.Secret)
	if err != nil {
		t.Fatalf("generate enrollment code: %v", err)
	}
	verify := httptest.NewRequest(http.MethodPost, "/api/v1/admin-mfa/totp/verify", strings.NewReader(fmt.Sprintf(`{"code":%q}`, code)))
	verify.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verify)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify status = %d, want %d, body %s", verifyRec.Code, http.StatusOK, verifyRec.Body.String())
	}
	if !store.adminMFASettings.LocalTOTPEnabled || store.adminMFASettings.LocalTOTPSecret == "" || store.adminMFASettings.LocalTOTPPendingSecret != "" {
		t.Fatalf("totp settings not enabled correctly: %+v", store.adminMFASettings)
	}
}

func TestProIdentityAuthSettingsDoNotExposeAPIKey(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin-mfa/proidentity", strings.NewReader(`{"enabled":true,"base_url":"https://auth.example.com","api_key":"sp-secret","user_email":"admin@example.com","timeout_seconds":120}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "sp-secret") {
		t.Fatalf("api key leaked in response: %s", body)
	}
	var settings domain.AdminMFASettings
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&settings); err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	if !settings.ProIdentityEnabled || !settings.ProIdentityAPIKeyConfigured || settings.EffectiveProvider != "proidentity" || settings.NativeWebAuthnEnabled {
		t.Fatalf("unexpected public settings: %+v", settings)
	}
	if store.adminMFASettings.ProIdentityAPIKey != "sp-secret" {
		t.Fatalf("api key was not stored server-side")
	}
	if store.adminMFASettings.ProIdentityBaseURL != "https://verify.proidentity.cloud" || settings.ProIdentityBaseURL != "https://verify.proidentity.cloud" {
		t.Fatalf("proidentity base URL should be fixed, stored=%q response=%q", store.adminMFASettings.ProIdentityBaseURL, settings.ProIdentityBaseURL)
	}
}

func firstTestTOTPSecret() string {
	return "JBSWY3DPEHPK3PXP"
}

func generateTOTPCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

func TestRequestClientIPUsesProxyHeadersOnlyFromLoopback(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", nil)
	req.RemoteAddr = "127.0.0.1:38123"
	req.Header.Set("X-Real-IP", "203.0.113.77")
	req.Header.Set("X-Forwarded-For", "198.51.100.10, 203.0.113.77")
	if got := requestClientIP(req); got != "203.0.113.77" {
		t.Fatalf("proxied client IP = %q, want 203.0.113.77", got)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/session", nil)
	req.RemoteAddr = "127.0.0.1:38123"
	req.Header.Set("X-Forwarded-For", "198.51.100.10, 203.0.113.88")
	if got := requestClientIP(req); got != "203.0.113.88" {
		t.Fatalf("proxied fallback client IP = %q, want closest forwarded peer", got)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/session", nil)
	req.RemoteAddr = "198.51.100.20:38123"
	req.Header.Set("X-Real-IP", "203.0.113.77")
	req.Header.Set("X-Forwarded-For", "203.0.113.77")
	if got := requestClientIP(req); got != "198.51.100.20" {
		t.Fatalf("direct client spoofed IP = %q, want remote address", got)
	}
}

func TestAdminLoginLimiterKeysUseTrustedProxyClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session", nil)
	req.RemoteAddr = "127.0.0.1:38123"
	req.Header.Set("X-Real-IP", "203.0.113.77")

	keys := loginKeys("admin", "Admin", req)
	if !slices.Contains(keys, "admin|ip|203.0.113.77") {
		t.Fatalf("login limiter keys should use real client IP, got %+v", keys)
	}
	if slices.Contains(keys, "admin|ip|127.0.0.1") {
		t.Fatalf("login limiter keys still use loopback proxy IP: %+v", keys)
	}
}

func TestTenantAdminSessionIsScopedToAssignedTenant(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{
		tenants: []domain.Tenant{
			{ID: 11, Name: "Allowed", Slug: "allowed", Status: "active"},
			{ID: 22, Name: "Blocked", Slug: "blocked", Status: "active"},
		},
		tenantAdmins: []domain.TenantAdmin{{ID: 99, TenantID: 11, UserID: 33, Role: "tenant_admin", Status: "active"}},
	}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})

	list := adminSessionRequest(t, manager, http.MethodGet, "/api/v1/tenants", nil, "tenant-admin@example.com")
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, list)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list tenants status = %d, want %d, body %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	if strings.Contains(listRec.Body.String(), "Blocked") || !strings.Contains(listRec.Body.String(), "Allowed") {
		t.Fatalf("tenant admin should only see assigned tenant, body %s", listRec.Body.String())
	}

	createBlocked := adminSessionRequest(t, manager, http.MethodPost, "/api/v1/users", strings.NewReader(`{"tenant_id":22,"primary_domain_id":44,"local_part":"blocked","password":"secret123456","mailbox_type":"user"}`), "tenant-admin@example.com")
	blockedRec := httptest.NewRecorder()
	handler.ServeHTTP(blockedRec, createBlocked)
	if blockedRec.Code != http.StatusForbidden {
		t.Fatalf("cross-tenant create status = %d, want %d, body %s", blockedRec.Code, http.StatusForbidden, blockedRec.Body.String())
	}
}

func TestReadOnlyTenantAdminCannotWrite(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{tenantAdmins: []domain.TenantAdmin{{ID: 99, TenantID: 11, UserID: 33, Role: "read_only", Status: "active"}}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})

	req := adminSessionRequest(t, manager, http.MethodPost, "/api/v1/users", strings.NewReader(`{"tenant_id":11,"primary_domain_id":22,"local_part":"marko","password":"secret123456","mailbox_type":"user"}`), "readonly@example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("read-only create status = %d, want %d, body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestTenantAdminCannotAccessSystemOperations(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	store := &fakeStore{tenantAdmins: []domain.TenantAdmin{{ID: 99, TenantID: 11, UserID: 33, Role: "tenant_admin", Status: "active"}}}
	handler := NewRouter(store, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})

	req := adminSessionRequest(t, manager, http.MethodPost, "/api/v1/system/config-apply", nil, "tenant-admin@example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("system operation status = %d, want %d, body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func adminSessionRequest(t *testing.T, manager *session.Manager, method, target string, body io.Reader, subject string) *http.Request {
	t.Helper()
	seed := httptest.NewRequest(http.MethodGet, "/", nil)
	seed.RemoteAddr = "203.0.113.10:48123"
	seed.Header.Set("User-Agent", "Browser A")
	seed.Header.Set("Accept-Language", "en-US")
	created, err := manager.Create(seed, subject, "admin")
	if err != nil {
		t.Fatalf("create admin session: %v", err)
	}
	req := httptest.NewRequest(method, target, body)
	req.RemoteAddr = seed.RemoteAddr
	req.Header.Set("User-Agent", "Browser A")
	req.Header.Set("Accept-Language", "en-US")
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		req.Header.Set("X-CSRF-Token", created.CSRFToken)
	}
	req.AddCookie(created.Cookie)
	return req
}

func TestDovecotAuthPolicyRequiresLoopbackAndToken(t *testing.T) {
	limiter := &recordingLimiter{}
	handler := NewRouter(&fakeStore{}, AuthConfig{
		AuthPolicyLimiter: limiter,
		AuthPolicyToken:   "policy-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/dovecot/auth-policy?command=allow", strings.NewReader(`{"login":"tester@example.com","remote":"198.51.100.10","protocol":"imap"}`))
	req.RemoteAddr = "198.51.100.20:44444"
	req.Header.Set("X-ProIdentity-Auth-Policy", "policy-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-loopback status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	req = httptest.NewRequest(http.MethodPost, "/internal/dovecot/auth-policy?command=allow", strings.NewReader(`{"login":"tester@example.com","remote":"198.51.100.10","protocol":"imap"}`))
	req.RemoteAddr = "127.0.0.1:44444"
	req.Header.Set("X-ProIdentity-Auth-Policy", "wrong-secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("bad token status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	req = httptest.NewRequest(http.MethodPost, "/internal/dovecot/auth-policy?command=allow", strings.NewReader(`{"login":"tester@example.com","remote":"198.51.100.10","protocol":"imap"}`))
	req.RemoteAddr = "127.0.0.1:44444"
	req.Header.Set("X-Forwarded-For", "203.0.113.55")
	req.Header.Set("X-ProIdentity-Auth-Policy", "policy-secret")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("proxied external status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestDovecotAuthPolicyReportsFailuresToLimiter(t *testing.T) {
	limiter := &recordingLimiter{}
	handler := NewRouter(&fakeStore{}, AuthConfig{
		AuthPolicyLimiter: limiter,
		AuthPolicyToken:   "policy-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/dovecot/auth-policy?command=report", strings.NewReader(`{"login":"Tester@Example.COM","remote":"198.51.100.10","protocol":"imap","success":false,"fail_type":"credentials"}`))
	req.RemoteAddr = "127.0.0.1:44444"
	req.Header.Set("X-ProIdentity-Auth-Policy", "policy-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Status int `json:"status"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != 0 {
		t.Fatalf("policy status = %d, want 0", body.Status)
	}
	want := []string{
		"imap|ip|198.51.100.10",
		"imap|account|tester@example.com",
		"imap|pair|tester@example.com|198.51.100.10",
		"dovecot|ip|198.51.100.10",
		"dovecot|account|tester@example.com",
		"dovecot|pair|tester@example.com|198.51.100.10",
	}
	if !reflect.DeepEqual(limiter.failed, want) {
		t.Fatalf("failed keys = %#v, want %#v", limiter.failed, want)
	}
}

func TestDovecotAuthPolicyBlocksLockedLogin(t *testing.T) {
	limiter := &recordingLimiter{locked: map[string]bool{
		"imap|pair|tester@example.com|198.51.100.10": true,
	}}
	handler := NewRouter(&fakeStore{}, AuthConfig{
		AuthPolicyLimiter: limiter,
		AuthPolicyToken:   "policy-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/dovecot/auth-policy?command=allow", strings.NewReader(`{"login":"tester@example.com","remote":"198.51.100.10","protocol":"imap"}`))
	req.RemoteAddr = "127.0.0.1:44444"
	req.Header.Set("X-ProIdentity-Auth-Policy", "policy-secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != -1 {
		t.Fatalf("policy status = %d, want -1", body.Status)
	}
	if !strings.Contains(body.Msg, "temporarily locked") {
		t.Fatalf("policy msg = %q, want lockout message", body.Msg)
	}
}

func TestDovecotAuthPolicyKeysDoNotTreatPolicyServerLoopbackAsClient(t *testing.T) {
	keys := dovecotAuthPolicyKeys("imap", "tester@example.com", "", "127.0.0.1:44444")
	for _, key := range keys {
		if strings.Contains(key, "127.0.0.1") {
			t.Fatalf("keys should not use policy callback loopback peer as client IP: %#v", keys)
		}
	}
}

func TestListLoginRateLimitsReturnsFriendlyRows(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{loginRateLimits: []domain.LoginRateLimit{{
		ID:           7,
		Service:      "webmail",
		LimiterKey:   "webmail|account|tester@example.com",
		Scope:        "account",
		Subject:      "tester@example.com",
		FailureCount: 4,
		LockedUntil:  &now,
		Locked:       true,
		UpdatedAt:    now,
	}}}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/security/login-rate-limits", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var rows []domain.LoginRateLimit
	if err := json.NewDecoder(rec.Body).Decode(&rows); err != nil {
		t.Fatalf("decode rows: %v", err)
	}
	if len(rows) != 1 || rows[0].Scope != "account" || !rows[0].Locked {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestListLoginRateLimitsReturnsEmptyArray(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/security/login-rate-limits", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Fatalf("body = %q, want []", body)
	}
}

func TestUnlockUserActivatesUserAndClearsLimiterRows(t *testing.T) {
	store := &fakeStore{user: domain.User{ID: 42, TenantID: 11, PrimaryDomainID: 22, LocalPart: "tester", Status: "locked"}}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/42/unlock", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.unlockedUserID != 42 {
		t.Fatalf("unlocked user id = %d, want 42", store.unlockedUserID)
	}
	var user domain.User
	if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	if user.Status != "active" {
		t.Fatalf("unlocked user status = %q, want active", user.Status)
	}
}

func TestResetUserMFADisablesMailboxSecondFactor(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/42/mfa/reset", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if store.resetMFAUserID != 42 {
		t.Fatalf("reset mfa user id = %d, want 42", store.resetMFAUserID)
	}
	if !containsString(store.auditActions, "user_mfa.reset") {
		t.Fatalf("mfa reset audit event missing: %+v", store.auditActions)
	}
}

func TestTenantAdminEndpointsManageTenantPermissions(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	create := httptest.NewRequest(http.MethodPost, "/api/v1/tenant-admins", strings.NewReader(`{"tenant_id":11,"user_id":33,"role":"tenant_admin"}`))
	createRec := httptest.NewRecorder()

	handler.ServeHTTP(createRec, create)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	if store.tenantAdmin.TenantID != 11 || store.tenantAdmin.UserID != 33 || store.tenantAdmin.Role != "tenant_admin" {
		t.Fatalf("stored tenant admin = %+v", store.tenantAdmin)
	}
	if !containsString(store.auditActions, "tenant_admin.create") {
		t.Fatalf("tenant admin create audit event missing: %+v", store.auditActions)
	}

	list := httptest.NewRequest(http.MethodGet, "/api/v1/tenant-admins", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, list)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d body=%s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	if !strings.Contains(listRec.Body.String(), `"tenant_id":11`) {
		t.Fatalf("list body missing tenant admin: %s", listRec.Body.String())
	}

	remove := httptest.NewRequest(http.MethodDelete, "/api/v1/tenant-admins/99", nil)
	removeRec := httptest.NewRecorder()
	handler.ServeHTTP(removeRec, remove)
	if removeRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d body=%s", removeRec.Code, http.StatusNoContent, removeRec.Body.String())
	}
	if store.deletedTenantAdminID != 99 {
		t.Fatalf("deleted tenant admin id = %d, want 99", store.deletedTenantAdminID)
	}
}

func TestMailServerSettingsPersistsMailboxMFAPolicy(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := `{"hostname_mode":"shared","mail_hostname":"mail.example.com","default_language":"en","mailbox_mfa_enabled":true,"force_mailbox_mfa":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/mail-server-settings", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !store.mailServerSettings.MailboxMFAEnabled || !store.mailServerSettings.ForceMailboxMFA {
		t.Fatalf("mailbox mfa settings not stored: %+v", store.mailServerSettings)
	}
}

func TestMailServerSettingsPersistsCloudflareRealIP(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	body := `{"hostname_mode":"shared","mail_hostname":"mail.example.com","default_language":"en","cloudflare_real_ip_enabled":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/mail-server-settings", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !store.mailServerSettings.CloudflareRealIPEnabled {
		t.Fatalf("cloudflare real IP setting not stored: %+v", store.mailServerSettings)
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

func TestAdminSessionProtectedAPIRejectsBasicFallback(t *testing.T) {
	manager := session.NewManager(session.Options{CookieName: "admin_sid"})
	handler := NewRouter(&fakeStore{}, AuthConfig{Username: "admin", Password: "secret", Sessions: manager})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
	req.SetBasicAuth("admin", "secret")
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

func TestDiscoveryEndpointsUseDedicatedRateLimiterWithTrustedClientIP(t *testing.T) {
	limiter := &recordingLimiter{}
	handler := NewRouter(&fakeStore{}, AuthConfig{DiscoveryLimiter: limiter})
	req := httptest.NewRequest(http.MethodGet, "/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=marko@example.com", nil)
	req.RemoteAddr = "127.0.0.1:48123"
	req.Header.Set("X-Forwarded-For", "203.0.113.44")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	for _, want := range []string{"autoconfig|ip|203.0.113.44", "autoconfig|domain|example.com"} {
		if !containsString(limiter.failed, want) {
			t.Fatalf("rate limiter key %q missing from %+v", want, limiter.failed)
		}
	}
}

func TestDiscoveryEndpointReturns429WhenRateLimited(t *testing.T) {
	limiter := &recordingLimiter{locked: map[string]bool{"autodiscover|ip|203.0.113.55": true}}
	handler := NewRouter(&fakeStore{}, AuthConfig{DiscoveryLimiter: limiter})
	req := httptest.NewRequest(http.MethodPost, "/autodiscover/autodiscover.xml", bytes.NewBufferString(`<?xml version="1.0"?><Autodiscover><Request><EMailAddress>marko@example.com</EMailAddress></Request></Autodiscover>`))
	req.RemoteAddr = "127.0.0.1:48123"
	req.Header.Set("X-Real-IP", "203.0.113.55")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
	if len(limiter.failed) != 0 {
		t.Fatalf("locked discovery request should not record more failures: %+v", limiter.failed)
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

func TestAuditEndpointReturnsReadableCategorizedEvents(t *testing.T) {
	store := &fakeStore{auditEvents: []domain.AuditEvent{
		{
			ID:           1,
			TenantID:     uintPtr(11),
			ActorType:    "user",
			Action:       "message.report_spam",
			TargetType:   "message",
			TargetID:     "msg-1",
			MetadataJSON: `{"email":"marko@example.com","verdict":"spam"}`,
		},
		{
			ID:           2,
			ActorType:    "admin",
			Action:       "admin.login_failed",
			TargetType:   "admin",
			TargetID:     "root",
			MetadataJSON: `{}`,
		},
	}}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var events []domain.AuditEvent
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if events[0].Category != "mail_security" || events[0].Title != "User marked message as spam" {
		t.Fatalf("message event not enriched: %+v", events[0])
	}
	if !strings.Contains(events[0].Summary, "marko@example.com") || events[0].Severity != "warning" {
		t.Fatalf("message event summary/severity not friendly: %+v", events[0])
	}
	if len(events[0].Details) == 0 || events[0].Details[0].Label == "" {
		t.Fatalf("message event missing parsed details: %+v", events[0])
	}
	if events[1].Category != "auth" || events[1].Severity != "danger" || !strings.Contains(events[1].Title, "failed") {
		t.Fatalf("auth event not enriched: %+v", events[1])
	}
}

func TestAuditEndpointReturnsReadableSecurityAlerts(t *testing.T) {
	store := &fakeStore{auditEvents: []domain.AuditEvent{
		{
			ActorType:    "system",
			Action:       "security.alert.admin_new_ip",
			TargetType:   "admin",
			TargetID:     "admin",
			MetadataJSON: `{"username":"admin","client_ip":"203.0.113.7"}`,
		},
		{
			ActorType:    "system",
			Action:       "security.alert.auth_spray",
			TargetType:   "client_ip",
			TargetID:     "198.51.100.77",
			MetadataJSON: `{"service":"webmail","client_ip":"198.51.100.77","distinct_accounts":5,"window_seconds":60}`,
		},
		{
			ActorType:    "system",
			Action:       "security.alert.backup_manual",
			TargetType:   "backup",
			TargetID:     "proidentity-mail-20260512-020000.tar.gz.enc",
			MetadataJSON: `{"archive_name":"proidentity-mail-20260512-020000.tar.gz.enc","scheduled":false}`,
		},
	}}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var events []domain.AuditEvent
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if events[0].Category != "security" || events[0].Severity != "warning" || !strings.Contains(events[0].Summary, "203.0.113.7") {
		t.Fatalf("security alert not enriched: %+v", events[0])
	}
	if events[1].Category != "security" || events[1].Severity != "warning" || !strings.Contains(events[1].Summary, "5") || !strings.Contains(events[1].Summary, "198.51.100.77") {
		t.Fatalf("auth spray alert not enriched: %+v", events[1])
	}
	if events[2].Category != "security" || events[2].Severity != "warning" || !strings.Contains(events[2].Summary, "outside the scheduled") {
		t.Fatalf("manual backup alert not enriched: %+v", events[2])
	}
}

func TestCreateAdminResourcesRecordAuditEvents(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	tests := []struct {
		path   string
		body   string
		action string
	}{
		{"/api/v1/tenants", `{"name":"Example Org","slug":"example"}`, "tenant.create"},
		{"/api/v1/domains", `{"tenant_id":11,"name":"example.com"}`, "domain.create"},
		{"/api/v1/users", `{"tenant_id":11,"primary_domain_id":22,"local_part":"marko","display_name":"Marko","password":"secret123456"}`, "user.create"},
		{"/api/v1/aliases", `{"tenant_id":11,"domain_id":22,"source_local_part":"sales","destination":"marko@example.com"}`, "alias.create"},
		{"/api/v1/catch-all", `{"tenant_id":11,"domain_id":22,"destination":"catchall@example.com"}`, "catch_all.create"},
		{"/api/v1/shared-permissions", `{"tenant_id":11,"shared_mailbox_id":44,"user_id":33,"can_read":true,"can_send_as":true}`, "shared_permission.create"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("%s status = %d, want %d, body %s", tt.path, rec.Code, http.StatusCreated, rec.Body.String())
		}
		if !containsString(store.auditActions, tt.action) {
			t.Fatalf("audit action %q not recorded in %+v", tt.action, store.auditActions)
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

func TestOutlookAutodiscoverEndpoint(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	req := httptest.NewRequest(http.MethodPost, "/autodiscover/autodiscover.xml", bytes.NewBufferString(`<?xml version="1.0"?><Autodiscover><Request><EMailAddress>marko@example.com</EMailAddress></Request></Autodiscover>`))
	req.Header.Set("Content-Type", "text/xml")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"<Type>IMAP</Type>",
		"<Type>SMTP</Type>",
		"<Type>POP3</Type>",
		"<Server>mail.example.com</Server>",
		"<LoginName>marko@example.com</LoginName>",
	} {
		if !bytes.Contains([]byte(body), []byte(want)) {
			t.Fatalf("autodiscover missing %q: %s", want, body)
		}
	}
}

func TestAutoconfigRejectsInvalidEmailAndDomainInput(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	tests := []struct {
		name string
		req  *http.Request
	}{
		{"mozilla bad domain", httptest.NewRequest(http.MethodGet, "/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=marko@bad_domain.test", nil)},
		{"service discovery bad email", httptest.NewRequest(http.MethodGet, "/.well-known/proidentity-mail/config.json?emailaddress=not-an-email", nil)},
		{"outlook xxe style body", httptest.NewRequest(http.MethodPost, "/autodiscover/autodiscover.xml", strings.NewReader(`<?xml version="1.0"?><!DOCTYPE x [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><Autodiscover><Request><EMailAddress>&xxe;</EMailAddress></Request></Autodiscover>`))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, tt.req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func commandTargetDir(args []string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--target-dir" {
			return args[i+1]
		}
	}
	return ""
}

func writeConfigDriftFixture(t *testing.T, mailDir, proxyDir, liveRoot string) {
	t.Helper()
	for _, mapping := range configdrift.DefaultMappings(mailDir, proxyDir, liveRoot) {
		desired := []byte("managed = " + mapping.ID + "\n")
		live := desired
		if mapping.ID == "postfix-main" {
			desired = []byte("myhostname = mail.example\n")
			live = []byte("myhostname = old.example\n")
		}
		if err := os.MkdirAll(filepath.Dir(mapping.DesiredPath), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(mapping.DesiredPath, desired, 0640); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(mapping.LivePath), 0750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(mapping.LivePath, live, 0640); err != nil {
			t.Fatal(err)
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
	if store.mailDomain.Status != "" {
		t.Fatalf("create domain request should let store choose default status, got %+v", store.mailDomain)
	}
	wantLocalParts := map[string]string{
		"postmaster": "Postmaster",
		"abuse":      "Abuse Desk",
		"dmarc":      "DMARC Reports",
		"tlsrpt":     "TLS Reports",
	}
	if len(store.createdUsers) != len(wantLocalParts) {
		t.Fatalf("created system mailboxes = %d, want %d: %+v", len(store.createdUsers), len(wantLocalParts), store.createdUsers)
	}
	for _, user := range store.createdUsers {
		wantDisplay, ok := wantLocalParts[user.LocalPart]
		if !ok {
			t.Fatalf("unexpected system mailbox local part: %+v", user)
		}
		if user.TenantID != 11 || user.PrimaryDomainID != 22 || user.MailboxType != "shared" || user.PasswordHash != "" {
			t.Fatalf("unexpected system shared mailbox metadata: %+v", user)
		}
		if user.DisplayName != wantDisplay {
			t.Fatalf("display name for %s = %q, want %q", user.LocalPart, user.DisplayName, wantDisplay)
		}
	}
}

func TestAdminRejectsInvalidMailResourceNames(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{"domain traversal", "/api/v1/domains", `{"tenant_id":11,"name":"../../example.com"}`},
		{"domain underscore", "/api/v1/domains", `{"tenant_id":11,"name":"bad_domain.test"}`},
		{"user local part traversal", "/api/v1/users", `{"tenant_id":11,"primary_domain_id":22,"local_part":"../marko","password":"secret123456"}`},
		{"user local part at sign", "/api/v1/users", `{"tenant_id":11,"primary_domain_id":22,"local_part":"marko@example.com","password":"secret123456"}`},
		{"alias bad destination", "/api/v1/aliases", `{"tenant_id":11,"domain_id":22,"source_local_part":"sales","destination":"not an address"}`},
		{"catch all bad destination", "/api/v1/catch-all", `{"tenant_id":11,"domain_id":22,"destination":"../../postmaster"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRouter(&fakeStore{})
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestAdminNormalizesMailResourceNames(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewBufferString(`{"tenant_id":11,"name":" Example.COM. "}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("domain status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.mailDomain.Name != "example.com" {
		t.Fatalf("domain name = %q, want normalized example.com", store.mailDomain.Name)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBufferString(`{"tenant_id":11,"primary_domain_id":22,"local_part":" Marko.Sales ","password":"secret123456"}`))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("user status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if store.user.LocalPart != "marko.sales" {
		t.Fatalf("local part = %q, want normalized marko.sales", store.user.LocalPart)
	}
}

func TestAdminRejectsUnsafeCustomTLSPaths(t *testing.T) {
	handler := NewRouter(&fakeStore{})
	body := `{"tls_mode":"custom","challenge_type":"custom-import","custom_cert_path":"/etc/passwd","custom_key_path":"/etc/proidentity-mail/certs/example.key"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/domains/22/tls/settings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
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

func TestUpdateAndDeleteAdminResourceEndpoints(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	updates := []struct {
		path string
		body string
	}{
		{"/api/v1/tenants/11", `{"name":"Example Updated","slug":"example-updated","status":"suspended"}`},
		{"/api/v1/domains/22", `{"tenant_id":11,"name":"example.org","status":"active","dkim_selector":"mail2"}`},
		{"/api/v1/users/33", `{"tenant_id":11,"primary_domain_id":22,"local_part":"marko2","display_name":"Marko Two","mailbox_type":"user","status":"locked","quota_bytes":1048576,"password":"newsecret123456"}`},
		{"/api/v1/aliases/66", `{"tenant_id":11,"domain_id":22,"source_local_part":"info","destination":"marko@example.com"}`},
		{"/api/v1/catch-all/77", `{"tenant_id":11,"domain_id":22,"destination":"postmaster@example.com","status":"disabled"}`},
		{"/api/v1/shared-permissions/88", `{"tenant_id":11,"shared_mailbox_id":44,"user_id":33,"can_read":true,"can_send_as":true,"can_send_on_behalf":true,"can_manage":true}`},
	}

	for _, tt := range updates {
		req := httptest.NewRequest(http.MethodPut, tt.path, bytes.NewBufferString(tt.body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d, body %s", tt.path, rec.Code, http.StatusOK, rec.Body.String())
		}
	}
	if store.tenant.Name != "Example Updated" || store.tenant.Status != "suspended" {
		t.Fatalf("tenant update not passed to store: %+v", store.tenant)
	}
	if store.mailDomain.Name != "example.org" || store.mailDomain.DKIMSelector != "mail2" {
		t.Fatalf("domain update not passed to store: %+v", store.mailDomain)
	}
	if store.user.LocalPart != "marko2" || store.user.PasswordHash == "" || store.user.PasswordHash == "newsecret123456" || store.user.QuotaBytes != 1048576 {
		t.Fatalf("user update not passed to store with hashed password: %+v", store.user)
	}
	if store.alias.SourceLocalPart != "info" || store.catchAll.Status != "disabled" || !store.sharedPermission.CanManage {
		t.Fatalf("update values not passed to store: alias=%+v catch=%+v permission=%+v", store.alias, store.catchAll, store.sharedPermission)
	}

	deletes := []struct {
		path   string
		action string
	}{
		{"/api/v1/tenants/11", "tenant.delete"},
		{"/api/v1/domains/22", "domain.delete"},
		{"/api/v1/users/33", "user.delete"},
		{"/api/v1/aliases/66", "alias.delete"},
		{"/api/v1/catch-all/77", "catch_all.delete"},
		{"/api/v1/shared-permissions/88", "shared_permission.delete"},
	}
	for _, tt := range deletes {
		req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("%s status = %d, want %d, body %s", tt.path, rec.Code, http.StatusNoContent, rec.Body.String())
		}
		found := false
		for _, action := range store.auditActions {
			if action == tt.action {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("audit action %q not recorded in %+v", tt.action, store.auditActions)
		}
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
	if len(response.Records) < 8 {
		t.Fatalf("expected MX/SPF/DMARC/DKIM records, got %+v", response.Records)
	}
	if response.MailHost != "mail.example.com" {
		t.Fatalf("mail host = %q, want mail.example.com", response.MailHost)
	}
	foundAutodiscover := false
	for _, record := range response.Records {
		if record.Type == "CNAME" && record.Name == "autodiscover.example.com" && record.Value == "mail.example.com" {
			foundAutodiscover = true
			break
		}
	}
	if !foundAutodiscover {
		t.Fatalf("DNS records missing autodiscover CNAME: %+v", response.Records)
	}
	if len(response.ClientSetup) < 3 {
		t.Fatalf("expected client setup profiles, got %+v", response.ClientSetup)
	}
}

func TestDomainTLSEndpoints(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/domains/22/tls", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d, body %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}
	var tlsState domain.DomainTLS
	if err := json.NewDecoder(getRec.Body).Decode(&tlsState); err != nil {
		t.Fatalf("decode tls response: %v", err)
	}
	if tlsState.Settings.TLSMode != "inherit" || len(tlsState.Settings.DesiredHostnames) == 0 || len(tlsState.Jobs) == 0 {
		t.Fatalf("unexpected tls response: %+v", tlsState)
	}

	saveReq := httptest.NewRequest(http.MethodPut, "/api/v1/domains/22/tls/settings", bytes.NewBufferString(`{"dns_webmail_alias_enabled":false,"dns_admin_alias_enabled":true,"tls_mode":"letsencrypt-dns-cloudflare","challenge_type":"dns-cloudflare","use_for_https":true,"use_for_mail_sni":true,"include_mail_hostname":true,"include_webmail_hostname":false,"include_admin_hostname":true}`))
	saveRec := httptest.NewRecorder()
	handler.ServeHTTP(saveRec, saveReq)
	if saveRec.Code != http.StatusOK {
		t.Fatalf("settings status = %d, want %d, body %s", saveRec.Code, http.StatusOK, saveRec.Body.String())
	}
	if store.tlsSettings.DomainID != 22 || store.tlsSettings.DNSWebmailAliasEnabled {
		t.Fatalf("settings were not saved: %+v", store.tlsSettings)
	}

	jobReq := httptest.NewRequest(http.MethodPost, "/api/v1/domains/22/tls/jobs", bytes.NewBufferString(`{"job_type":"issue","challenge_type":"dns-cloudflare","hostnames":["mail.example.com"]}`))
	jobRec := httptest.NewRecorder()
	handler.ServeHTTP(jobRec, jobReq)
	if jobRec.Code != http.StatusAccepted {
		t.Fatalf("job status = %d, want %d, body %s", jobRec.Code, http.StatusAccepted, jobRec.Body.String())
	}
	if store.tlsJob.JobType != "issue" || store.tlsJob.ChallengeType != "dns-cloudflare" || len(store.tlsJob.Hostnames) != 1 {
		t.Fatalf("unexpected tls job: %+v", store.tlsJob)
	}
}

func TestCloudflareConfigAndProvisionEndpoints(t *testing.T) {
	store := &fakeStore{
		cloudflarePlan: domain.DNSProvisionPlan{
			DomainID: 22,
			Domain:   "example.com",
			Provider: "cloudflare",
			ZoneID:   "zone-123",
			ZoneName: "example.com",
			Status:   "changes",
			Summary:  "1 records can be created safely.",
			Actions: []domain.DNSProvisionAction{{
				Action: "create",
				Type:   "MX",
				Name:   "example.com",
				Value:  "mail.example.com",
				Reason: "record is missing",
			}},
		},
		cloudflareResult: domain.DNSProvisionResult{BackupID: 91, Applied: true, Changed: 1},
	}
	handler := NewRouter(store)

	saveReq := httptest.NewRequest(http.MethodPut, "/api/v1/domains/22/cloudflare", bytes.NewBufferString(`{"zone_id":"zone-123","api_token":"secret-token"}`))
	saveRec := httptest.NewRecorder()
	handler.ServeHTTP(saveRec, saveReq)
	if saveRec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want %d, body %s", saveRec.Code, http.StatusOK, saveRec.Body.String())
	}
	if store.savedCloudflareDomainID != 22 || store.savedCloudflareZoneID != "zone-123" || store.savedCloudflareToken != "secret-token" {
		t.Fatalf("unexpected cloudflare config saved: domain=%d zone=%q token=%q", store.savedCloudflareDomainID, store.savedCloudflareZoneID, store.savedCloudflareToken)
	}
	if bytes.Contains(saveRec.Body.Bytes(), []byte("secret-token")) {
		t.Fatalf("save response leaked API token: %s", saveRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/domains/22/cloudflare", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d, body %s", getRec.Code, http.StatusOK, getRec.Body.String())
	}
	var config domain.CloudflareConfig
	if err := json.NewDecoder(getRec.Body).Decode(&config); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if !config.TokenConfigured || config.ZoneID != "zone-123" {
		t.Fatalf("unexpected cloudflare config response: %+v", config)
	}

	checkReq := httptest.NewRequest(http.MethodPost, "/api/v1/domains/22/cloudflare/check", nil)
	checkRec := httptest.NewRecorder()
	handler.ServeHTTP(checkRec, checkReq)
	if checkRec.Code != http.StatusOK {
		t.Fatalf("check status = %d, want %d, body %s", checkRec.Code, http.StatusOK, checkRec.Body.String())
	}
	var plan domain.DNSProvisionPlan
	if err := json.NewDecoder(checkRec.Body).Decode(&plan); err != nil {
		t.Fatalf("decode plan: %v", err)
	}
	if plan.Status != "changes" || len(plan.Actions) != 1 {
		t.Fatalf("unexpected plan response: %+v", plan)
	}

	applyReq := httptest.NewRequest(http.MethodPost, "/api/v1/domains/22/cloudflare/apply", bytes.NewBufferString(`{"replace":true}`))
	applyRec := httptest.NewRecorder()
	handler.ServeHTTP(applyRec, applyReq)
	if applyRec.Code != http.StatusOK {
		t.Fatalf("apply status = %d, want %d, body %s", applyRec.Code, http.StatusOK, applyRec.Body.String())
	}
	if !store.appliedCloudflareReplace {
		t.Fatal("apply did not pass replace=true to store")
	}
	var result domain.DNSProvisionResult
	if err := json.NewDecoder(applyRec.Body).Decode(&result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if !result.Applied || result.BackupID != 91 {
		t.Fatalf("unexpected apply response: %+v", result)
	}
}

func TestCloudflarePlannerDetectsConflictsBeforeReplacement(t *testing.T) {
	priority := 10
	desired := []cloudflareDNSRecord{
		{Type: "MX", Name: "example.com", Content: "mail.example.com", Priority: &priority},
		{Type: "TXT", Name: "example.com", Content: "v=spf1 mx -all"},
		{Type: "CNAME", Name: "autoconfig.example.com", Content: "mail.example.com"},
	}
	existing := []cloudflareDNSRecord{
		{ID: "old-mx", Type: "MX", Name: "example.com", Content: "oldmail.example.net", Priority: &priority},
		{ID: "other-txt", Type: "TXT", Name: "example.com", Content: "google-site-verification=keep-me"},
		{ID: "old-spf", Type: "TXT", Name: "example.com", Content: "v=spf1 include:old.example -all"},
		{ID: "old-a", Type: "A", Name: "autoconfig.example.com", Content: "192.0.2.10"},
	}

	plan := planCloudflareActions(desired, existing)

	byName := map[string]domain.DNSProvisionAction{}
	for _, item := range plan {
		byName[item.action.Name+"|"+item.action.Type] = item.action
	}
	if byName["example.com|MX"].Action != "conflict" {
		t.Fatalf("MX action = %+v, want conflict", byName["example.com|MX"])
	}
	if byName["example.com|TXT"].Action != "conflict" || len(byName["example.com|TXT"].Existing) != 1 {
		t.Fatalf("SPF action should conflict only with existing SPF record, got %+v", byName["example.com|TXT"])
	}
	if byName["autoconfig.example.com|CNAME"].Action != "blocked" {
		t.Fatalf("CNAME action = %+v, want blocked", byName["autoconfig.example.com|CNAME"])
	}
}

func TestCloudflarePlannerTreatsStructuredSRVAsMatching(t *testing.T) {
	desired := desiredCloudflareRecords([]domain.DNSRecord{
		{Type: "SRV", Name: "_submission._tcp.example.com", Value: "0 1 587 mail.example.com"},
	})
	priority := 0
	existing := []cloudflareDNSRecord{{
		ID:       "srv-submission",
		Type:     "SRV",
		Name:     "_submission._tcp.example.com",
		Content:  "1 587 mail.example.com",
		Priority: &priority,
		Proxied:  boolPtr(false),
		Data: map[string]any{
			"priority": float64(0),
			"weight":   float64(1),
			"port":     float64(587),
			"target":   "mail.example.com",
		},
	}}

	plan := planCloudflareActions(desired, existing)
	if len(plan) != 1 {
		t.Fatalf("plan length = %d, want 1", len(plan))
	}
	if plan[0].action.Action != "ok" {
		t.Fatalf("structured SRV action = %+v, want ok", plan[0].action)
	}
}

func TestCloudflarePlannerTreatsExistingCNAMEAsBlockerForAddressRecord(t *testing.T) {
	desired := desiredCloudflareRecords([]domain.DNSRecord{
		{Type: "A", Name: "webmail.example.com", Value: "203.0.113.10", Proxied: boolPtr(true)},
	})
	existing := []cloudflareDNSRecord{{
		ID:      "old-webmail-cname",
		Type:    "CNAME",
		Name:    "webmail.example.com",
		Content: "mail.example.com",
		Proxied: boolPtr(false),
	}}

	plan := planCloudflareActions(desired, existing)
	if len(plan) != 1 {
		t.Fatalf("plan length = %d, want 1", len(plan))
	}
	if plan[0].action.Action != "blocked" || !strings.Contains(plan[0].action.Reason, "CNAME") {
		t.Fatalf("address over CNAME action = %+v, want blocked CNAME conflict", plan[0].action)
	}
	if len(plan[0].touch) != 1 || plan[0].touch[0].ID != "old-webmail-cname" {
		t.Fatalf("planner did not include blocking CNAME for replacement: %+v", plan[0].touch)
	}
}

func TestCloudflareClientListsAllDNSRecordPages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("authorization header = %q, want bearer token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = io.WriteString(w, `{"success":true,"result":[{"id":"1","type":"MX","name":"example.com","content":"old.example.com"}],"result_info":{"page":1,"total_pages":2}}`)
		case "2":
			_, _ = io.WriteString(w, `{"success":true,"result":[{"id":"2","type":"TXT","name":"example.com","content":"v=spf1 mx -all"}],"result_info":{"page":2,"total_pages":2}}`)
		default:
			t.Fatalf("unexpected page query %q", r.URL.Query().Get("page"))
		}
	}))
	defer server.Close()
	client := cloudflareClient{token: "test-token", baseURL: server.URL, httpClient: server.Client()}

	records, err := client.listRecords(context.Background(), "zone-123")
	if err != nil {
		t.Fatalf("list records: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records length = %d, want 2: %+v", len(records), records)
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

func TestBuildDomainDNSUsesSharedMailHostAndDomainAlias(t *testing.T) {
	priority := 10
	dns := buildDomainDNS(22, "customer.test", "", "", DNSSettings{
		MailHostname:    "mail.example.com",
		AdminHostname:   "madmin.example.com",
		WebmailHostname: "webmail.example.com",
	})
	if dns.MailHost != "mail.example.com" {
		t.Fatalf("mail host = %q, want shared mail host", dns.MailHost)
	}
	if !dns.Provisionable || len(dns.Warnings) != 0 {
		t.Fatalf("dns should be provisionable without local address records: %+v", dns)
	}
	for _, record := range dns.Records {
		if record.Type == "A" || record.Type == "AAAA" {
			t.Fatalf("shared mail host should not create address record in customer zone: %+v", record)
		}
	}
	foundMX := false
	for _, record := range dns.Records {
		if record.Type == "MX" && record.Name == "customer.test" && record.Value == "mail.example.com" && record.Priority != nil && *record.Priority == priority {
			foundMX = true
			break
		}
	}
	if !foundMX {
		t.Fatalf("missing MX to shared mail host: %+v", dns.Records)
	}
	if !hasRecord(dns.Records, "CNAME", "mail.customer.test", "mail.example.com") {
		t.Fatalf("missing mail.customer.test alias to shared mail host: %+v", dns.Records)
	}
	if !hasProxiedRecord(dns.Records, "CNAME", "webmail.customer.test", "webmail.example.com", true) {
		t.Fatalf("missing proxied webmail.customer.test alias: %+v", dns.Records)
	}
	if !hasProxiedRecord(dns.Records, "CNAME", "madmin.customer.test", "madmin.example.com", true) {
		t.Fatalf("missing proxied madmin.customer.test alias: %+v", dns.Records)
	}
}

func TestBuildDomainDNSUsesProxiedAddressRecordsForOwnWebAliases(t *testing.T) {
	dns := buildDomainDNS(22, "example.com", "", "", DNSSettings{
		MailHostname:    "mail.example.com",
		AdminHostname:   "madmin.example.com",
		WebmailHostname: "webmail.example.com",
		PublicIPv4:      "203.0.113.10",
	})

	if !hasProxiedRecord(dns.Records, "A", "webmail.example.com", "203.0.113.10", true) {
		t.Fatalf("missing proxied webmail address record: %+v", dns.Records)
	}
	if !hasProxiedRecord(dns.Records, "A", "madmin.example.com", "203.0.113.10", true) {
		t.Fatalf("missing proxied admin address record: %+v", dns.Records)
	}
}

func TestBuildDomainDNSCanDisableOptionalWebAliases(t *testing.T) {
	dns := buildDomainDNS(22, "customer.test", "", "", DNSSettings{
		MailHostname:        "mail.example.com",
		AdminHostname:       "madmin.example.com",
		WebmailHostname:     "webmail.example.com",
		PublicIPv4:          "203.0.113.10",
		DisableWebmailAlias: true,
		DisableAdminAlias:   true,
	})
	for _, record := range dns.Records {
		if record.Name == "webmail.customer.test" || record.Name == "madmin.customer.test" {
			t.Fatalf("optional web alias was generated while disabled: %+v", dns.Records)
		}
	}
}

func TestBuildDomainDNSRequiresPublicIPWhenMailHostIsInsideDomain(t *testing.T) {
	dns := buildDomainDNS(22, "customer.test", "", "", DNSSettings{MailHostname: "mail.customer.test"})
	if dns.Provisionable {
		t.Fatalf("dns should not be provisionable without public IP: %+v", dns)
	}
	if len(dns.Warnings) == 0 {
		t.Fatalf("expected warning for missing public IP")
	}
	for _, record := range dns.Records {
		if record.Type == "A" || record.Type == "AAAA" {
			t.Fatalf("should not create address record without configured public IP: %+v", record)
		}
	}

	ready := buildDomainDNS(22, "customer.test", "", "", DNSSettings{MailHostname: "mail.customer.test", PublicIPv4: "203.0.113.10", PublicIPv6: "2001:db8::10"})
	if !ready.Provisionable || len(ready.Warnings) != 0 {
		t.Fatalf("dns should be provisionable with public IPs: %+v", ready)
	}
	foundA := false
	foundAAAA := false
	for _, record := range ready.Records {
		if record.Type == "A" && record.Name == "mail.customer.test" && record.Value == "203.0.113.10" {
			foundA = true
		}
		if record.Type == "AAAA" && record.Name == "mail.customer.test" && record.Value == "2001:db8::10" {
			foundAAAA = true
		}
	}
	if !foundA || !foundAAAA {
		t.Fatalf("missing A/AAAA records: %+v", ready.Records)
	}
	if hasRecord(ready.Records, "CNAME", "mail.customer.test", "mail.customer.test") {
		t.Fatalf("mail host must not alias to itself: %+v", ready.Records)
	}
}

func TestBuildDomainDNSPerDomainModeUsesDomainMailAddress(t *testing.T) {
	dns := buildDomainDNS(22, "customer.test", "", "", DNSSettings{
		HostnameMode: "per-domain",
		MailHostname: "mail.example.com",
		PublicIPv4:   "203.0.113.10",
		PublicIPv6:   "2001:db8::10",
	})

	if dns.MailHost != "mail.customer.test" {
		t.Fatalf("mail host = %q, want per-domain host", dns.MailHost)
	}
	if !hasRecord(dns.Records, "A", "mail.customer.test", "203.0.113.10") {
		t.Fatalf("missing per-domain A record: %+v", dns.Records)
	}
	if !hasRecord(dns.Records, "AAAA", "mail.customer.test", "2001:db8::10") {
		t.Fatalf("missing per-domain AAAA record: %+v", dns.Records)
	}
	if hasRecord(dns.Records, "CNAME", "mail.customer.test", "mail.example.com") {
		t.Fatalf("per-domain mode must not alias the mail host to shared host: %+v", dns.Records)
	}
}

func TestBuildDomainDNSHeadDomainModeAliasesOtherDomains(t *testing.T) {
	head := buildDomainDNS(22, "platform.test", "", "", DNSSettings{
		HostnameMode: "head-domain",
		MailHostname: "mail.platform.test",
		PublicIPv4:   "203.0.113.10",
	})
	if head.MailHost != "mail.platform.test" || !hasRecord(head.Records, "A", "mail.platform.test", "203.0.113.10") {
		t.Fatalf("head domain should own the mail address, got %+v", head)
	}

	customer := buildDomainDNS(23, "customer.test", "", "", DNSSettings{
		HostnameMode: "head-domain",
		MailHostname: "mail.platform.test",
		PublicIPv4:   "203.0.113.10",
	})
	if customer.MailHost != "mail.platform.test" {
		t.Fatalf("customer mail host = %q, want head host", customer.MailHost)
	}
	if !hasRecord(customer.Records, "CNAME", "mail.customer.test", "mail.platform.test") {
		t.Fatalf("customer domain should alias mail.customer.test to head host: %+v", customer.Records)
	}
	if hasRecord(customer.Records, "A", "mail.customer.test", "203.0.113.10") {
		t.Fatalf("customer domain should not publish A for head-domain mode: %+v", customer.Records)
	}
}

func TestMailServerSettingsEndpointPersistsBehavior(t *testing.T) {
	store := &fakeStore{}
	marker := filepath.Join(t.TempDir(), "apply-request")
	handler := NewRouter(store, AuthConfig{System: SystemConfig{ConfigApplyRequestPath: marker}})
	payload := `{"hostname_mode":"head-domain","mail_hostname":"mail.platform.test","head_tenant_id":11,"head_domain_id":22,"public_ipv4":"203.0.113.10","public_ipv6":"","sni_enabled":true,"tls_mode":"custom-cert","force_https":true,"https_certificate_id":77,"default_language":"sk"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/mail-server-settings", strings.NewReader(payload))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if store.mailServerSettings.HostnameMode != "head-domain" || !store.mailServerSettings.SNIEnabled {
		t.Fatalf("settings not persisted: %+v", store.mailServerSettings)
	}
	if store.mailServerSettings.TLSMode != "custom-cert" || !store.mailServerSettings.ForceHTTPS {
		t.Fatalf("tls settings not persisted: %+v", store.mailServerSettings)
	}
	if store.mailServerSettings.HTTPSCertificateID == nil || *store.mailServerSettings.HTTPSCertificateID != 77 {
		t.Fatalf("https certificate id not persisted: %+v", store.mailServerSettings)
	}
	if store.mailServerSettings.DefaultLanguage != "sk" {
		t.Fatalf("default language = %q, want sk", store.mailServerSettings.DefaultLanguage)
	}
	var saveResponse domain.MailServerSettings
	if err := json.NewDecoder(rec.Body).Decode(&saveResponse); err != nil {
		t.Fatalf("decode save response: %v", err)
	}
	if !saveResponse.ConfigApplyQueued || saveResponse.ConfigApplyError != "" {
		t.Fatalf("config apply queue status = queued %t error %q", saveResponse.ConfigApplyQueued, saveResponse.ConfigApplyError)
	}
	if data, err := os.ReadFile(marker); err != nil || !strings.Contains(string(data), "requested_at=") {
		t.Fatalf("config apply marker not written: data=%q err=%v", string(data), err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/mail-server-settings", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, body %s", getRec.Code, getRec.Body.String())
	}
	var response domain.MailServerSettings
	if err := json.NewDecoder(getRec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.EffectiveHostname != "mail.platform.test" {
		t.Fatalf("effective hostname = %q", response.EffectiveHostname)
	}
	if response.DefaultLanguage != "sk" {
		t.Fatalf("response default language = %q, want sk", response.DefaultLanguage)
	}
}

func TestMailServerSettingsEndpointRejectsUnsupportedDefaultLanguage(t *testing.T) {
	store := &fakeStore{}
	handler := NewRouter(store)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/mail-server-settings", strings.NewReader(`{"hostname_mode":"shared","default_language":"xx"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCloudflarePlannerTreatsProxiedMailRecordsAsConflicts(t *testing.T) {
	desired := []cloudflareDNSRecord{
		{Type: "CNAME", Name: "mail.example.com", Content: "mail.example.com", TTL: 1, Proxied: boolPtr(false)},
	}
	existing := []cloudflareDNSRecord{
		{ID: "proxied-mail", Type: "CNAME", Name: "mail.example.com", Content: "mail.example.com", TTL: 1, Proxied: boolPtr(true)},
	}

	plan := planCloudflareActions(desired, existing)
	if len(plan) != 1 {
		t.Fatalf("plan length = %d, want 1", len(plan))
	}
	if plan[0].action.Action != "conflict" {
		t.Fatalf("proxied mail record action = %+v, want conflict", plan[0].action)
	}
	if !strings.Contains(plan[0].action.Reason, "DNS-only") {
		t.Fatalf("proxied conflict reason should mention DNS-only, got %+v", plan[0].action)
	}
}

func hasRecord(records []domain.DNSRecord, recordType, name, value string) bool {
	for _, record := range records {
		if record.Type == recordType && record.Name == name && record.Value == value {
			return true
		}
	}
	return false
}

func hasProxiedRecord(records []domain.DNSRecord, recordType, name, value string, proxied bool) bool {
	for _, record := range records {
		if record.Type == recordType && record.Name == name && record.Value == value && record.Proxied != nil && *record.Proxied == proxied {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func uintPtr(value uint64) *uint64 {
	return &value
}

type recordingLimiter struct {
	locked    map[string]bool
	failed    []string
	succeeded []string
}

func (l *recordingLimiter) Locked(key string) bool {
	return l.locked != nil && l.locked[key]
}

func (l *recordingLimiter) Fail(key string) {
	l.failed = append(l.failed, key)
}

func (l *recordingLimiter) Success(key string) {
	l.succeeded = append(l.succeeded, key)
}

type fakeStore struct {
	tenant                   domain.Tenant
	tenants                  []domain.Tenant
	mailDomain               domain.Domain
	domains                  []domain.Domain
	user                     domain.User
	users                    []domain.User
	createdUsers             []domain.User
	tenantAdmin              domain.TenantAdmin
	tenantAdmins             []domain.TenantAdmin
	deletedTenantAdminID     uint64
	resetMFAUserID           uint64
	alias                    domain.Alias
	catchAll                 domain.CatchAllRoute
	sharedPermission         domain.SharedMailboxPermission
	policy                   domain.TenantPolicy
	cloudflareConfig         domain.CloudflareConfig
	cloudflarePlan           domain.DNSProvisionPlan
	cloudflareResult         domain.DNSProvisionResult
	mailServerSettings       domain.MailServerSettings
	tlsState                 domain.DomainTLS
	tlsSettings              domain.DomainTLSSettings
	tlsJob                   domain.TLSCertificateJob
	savedCloudflareDomainID  uint64
	savedCloudflareZoneID    string
	savedCloudflareToken     string
	appliedCloudflareReplace bool
	resolvedQuarantineID     uint64
	resolvedQuarantineStatus string
	resolvedQuarantineNote   string
	auditEvents              []domain.AuditEvent
	auditActions             []string
	loginRateLimits          []domain.LoginRateLimit
	unlockedUserID           uint64
	adminMFASettings         domain.AdminMFASettings
	adminMFAChallenges       map[string]domain.AdminMFAChallenge
	adminWebAuthnCredentials []domain.AdminWebAuthnCredential
	adminWebAuthnSessions    map[string]domain.AdminWebAuthnSession
}

func (s *fakeStore) CreateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	s.tenant = tenant
	tenant.ID = 11
	tenant.Status = "active"
	return tenant, nil
}

func (s *fakeStore) ListTenants(ctx context.Context) ([]domain.Tenant, error) {
	if s.tenants != nil {
		return s.tenants, nil
	}
	return []domain.Tenant{{ID: 11, Name: "Example Org", Slug: "example", Status: "active"}}, nil
}

func (s *fakeStore) UpdateTenant(ctx context.Context, tenant domain.Tenant) (domain.Tenant, error) {
	s.tenant = tenant
	return tenant, nil
}

func (s *fakeStore) DeleteTenant(ctx context.Context, tenantID uint64) error {
	return nil
}

func (s *fakeStore) CreateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error) {
	s.mailDomain = mailDomain
	mailDomain.ID = 22
	mailDomain.Status = "active"
	return mailDomain, nil
}

func (s *fakeStore) ListDomains(ctx context.Context) ([]domain.Domain, error) {
	if s.domains != nil {
		return s.domains, nil
	}
	return []domain.Domain{{ID: 22, TenantID: 11, Name: "example.com", Status: "active", DKIMSelector: "mail"}}, nil
}

func (s *fakeStore) UpdateDomain(ctx context.Context, mailDomain domain.Domain) (domain.Domain, error) {
	s.mailDomain = mailDomain
	return mailDomain, nil
}

func (s *fakeStore) DeleteDomain(ctx context.Context, domainID uint64) error {
	return nil
}

func (s *fakeStore) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	s.user = user
	s.createdUsers = append(s.createdUsers, user)
	user.ID = 33
	user.Status = "active"
	if user.MailboxType == "" {
		user.MailboxType = "user"
	}
	return user, nil
}

func (s *fakeStore) ListUsers(ctx context.Context) ([]domain.User, error) {
	if s.users != nil {
		return s.users, nil
	}
	return []domain.User{{ID: 33, TenantID: 11, PrimaryDomainID: 22, LocalPart: "marko", DisplayName: "Marko", MailboxType: "user", Status: "active"}}, nil
}

func (s *fakeStore) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	s.user = user
	return user, nil
}

func (s *fakeStore) DeleteUser(ctx context.Context, userID uint64) error {
	return nil
}

func (s *fakeStore) UnlockUser(ctx context.Context, userID uint64) (domain.User, error) {
	s.unlockedUserID = userID
	user := s.user
	if user.ID == 0 {
		user.ID = userID
	}
	user.Status = "active"
	return user, nil
}

func (s *fakeStore) ResetUserMFA(ctx context.Context, userID uint64) error {
	s.resetMFAUserID = userID
	return nil
}

func (s *fakeStore) CreateTenantAdmin(ctx context.Context, admin domain.TenantAdmin) (domain.TenantAdmin, error) {
	s.tenantAdmin = admin
	if admin.ID == 0 {
		admin.ID = 99
	}
	if admin.Status == "" {
		admin.Status = "active"
	}
	s.tenantAdmins = append(s.tenantAdmins, admin)
	return admin, nil
}

func (s *fakeStore) ListTenantAdmins(ctx context.Context) ([]domain.TenantAdmin, error) {
	if s.tenantAdmins != nil {
		return s.tenantAdmins, nil
	}
	return []domain.TenantAdmin{{ID: 99, TenantID: 11, UserID: 33, Role: "tenant_admin", Status: "active"}}, nil
}

func (s *fakeStore) GetTenantAdminGrants(ctx context.Context, email string) ([]domain.TenantAdmin, error) {
	return s.ListTenantAdmins(ctx)
}

func (s *fakeStore) DeleteTenantAdmin(ctx context.Context, adminID uint64) error {
	s.deletedTenantAdminID = adminID
	return nil
}

func (s *fakeStore) ListLoginRateLimits(ctx context.Context) ([]domain.LoginRateLimit, error) {
	return s.loginRateLimits, nil
}

func (s *fakeStore) ClearLoginRateLimit(ctx context.Context, limitID uint64) error {
	for i, item := range s.loginRateLimits {
		if item.ID == limitID {
			s.loginRateLimits = append(s.loginRateLimits[:i], s.loginRateLimits[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *fakeStore) GetAdminMFASettings(ctx context.Context) (domain.AdminMFASettings, error) {
	settings := s.adminMFASettings
	if settings.ProIdentityTimeoutSeconds == 0 {
		settings.ProIdentityTimeoutSeconds = defaultProIdentityTimeoutSeconds
	}
	return settings, nil
}

func (s *fakeStore) SaveAdminMFASettings(ctx context.Context, settings domain.AdminMFASettings) (domain.AdminMFASettings, error) {
	s.adminMFASettings = settings
	if s.adminMFASettings.ProIdentityTimeoutSeconds == 0 {
		s.adminMFASettings.ProIdentityTimeoutSeconds = defaultProIdentityTimeoutSeconds
	}
	return s.adminMFASettings, nil
}

func (s *fakeStore) CreateAdminMFAChallenge(ctx context.Context, challenge domain.AdminMFAChallenge) error {
	if s.adminMFAChallenges == nil {
		s.adminMFAChallenges = map[string]domain.AdminMFAChallenge{}
	}
	s.adminMFAChallenges[challenge.Token] = challenge
	return nil
}

func (s *fakeStore) GetAdminMFAChallenge(ctx context.Context, token string) (domain.AdminMFAChallenge, error) {
	challenge, ok := s.adminMFAChallenges[token]
	if !ok {
		return domain.AdminMFAChallenge{}, errors.New("not found")
	}
	return challenge, nil
}

func (s *fakeStore) DeleteAdminMFAChallenge(ctx context.Context, token string) error {
	delete(s.adminMFAChallenges, token)
	return nil
}

func (s *fakeStore) ListAdminWebAuthnCredentials(ctx context.Context) ([]domain.AdminWebAuthnCredential, error) {
	return append([]domain.AdminWebAuthnCredential(nil), s.adminWebAuthnCredentials...), nil
}

func (s *fakeStore) CreateAdminWebAuthnCredential(ctx context.Context, credential domain.AdminWebAuthnCredential) (domain.AdminWebAuthnCredential, error) {
	if credential.ID == 0 {
		credential.ID = uint64(len(s.adminWebAuthnCredentials) + 1)
	}
	if credential.Name == "" {
		credential.Name = "Hardware key"
	}
	s.adminWebAuthnCredentials = append(s.adminWebAuthnCredentials, credential)
	s.adminMFASettings.NativeWebAuthnEnabled = true
	return credential, nil
}

func (s *fakeStore) UpdateAdminWebAuthnCredential(ctx context.Context, credential domain.AdminWebAuthnCredential) error {
	for i, item := range s.adminWebAuthnCredentials {
		if item.ID == credential.ID {
			s.adminWebAuthnCredentials[i] = credential
			return nil
		}
	}
	return nil
}

func (s *fakeStore) CreateAdminWebAuthnSession(ctx context.Context, session domain.AdminWebAuthnSession) error {
	if s.adminWebAuthnSessions == nil {
		s.adminWebAuthnSessions = map[string]domain.AdminWebAuthnSession{}
	}
	s.adminWebAuthnSessions[session.Token] = session
	return nil
}

func (s *fakeStore) GetAdminWebAuthnSession(ctx context.Context, token string) (domain.AdminWebAuthnSession, error) {
	session, ok := s.adminWebAuthnSessions[token]
	if !ok {
		return domain.AdminWebAuthnSession{}, errors.New("not found")
	}
	return session, nil
}

func (s *fakeStore) DeleteAdminWebAuthnSession(ctx context.Context, token string) error {
	delete(s.adminWebAuthnSessions, token)
	return nil
}

func (s *fakeStore) CreateAlias(ctx context.Context, alias domain.Alias) (domain.Alias, error) {
	s.alias = alias
	alias.ID = 66
	return alias, nil
}

func (s *fakeStore) ListAliases(ctx context.Context) ([]domain.Alias, error) {
	return []domain.Alias{{ID: 66, TenantID: 11, DomainID: 22, SourceLocalPart: "sales", Destination: "marko@example.com"}}, nil
}

func (s *fakeStore) UpdateAlias(ctx context.Context, alias domain.Alias) (domain.Alias, error) {
	s.alias = alias
	return alias, nil
}

func (s *fakeStore) DeleteAlias(ctx context.Context, aliasID uint64) error {
	return nil
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

func (s *fakeStore) UpdateCatchAllRoute(ctx context.Context, route domain.CatchAllRoute) (domain.CatchAllRoute, error) {
	s.catchAll = route
	return route, nil
}

func (s *fakeStore) DeleteCatchAllRoute(ctx context.Context, routeID uint64) error {
	return nil
}

func (s *fakeStore) CreateSharedMailboxPermission(ctx context.Context, permission domain.SharedMailboxPermission) (domain.SharedMailboxPermission, error) {
	s.sharedPermission = permission
	permission.ID = 88
	return permission, nil
}

func (s *fakeStore) ListSharedMailboxPermissions(ctx context.Context) ([]domain.SharedMailboxPermission, error) {
	return []domain.SharedMailboxPermission{{ID: 88, TenantID: 11, SharedMailboxID: 44, UserID: 33, CanRead: true, CanSendAs: true}}, nil
}

func (s *fakeStore) UpdateSharedMailboxPermission(ctx context.Context, permission domain.SharedMailboxPermission) (domain.SharedMailboxPermission, error) {
	s.sharedPermission = permission
	return permission, nil
}

func (s *fakeStore) DeleteSharedMailboxPermission(ctx context.Context, permissionID uint64) error {
	return nil
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
	if s.auditEvents != nil {
		return s.auditEvents, nil
	}
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

func (s *fakeStore) GetMailServerSettings(ctx context.Context) (domain.MailServerSettings, error) {
	if s.mailServerSettings.HostnameMode == "" {
		return domain.MailServerSettings{HostnameMode: "shared", MailHostname: "mail.example.com", TLSMode: "system", ForceHTTPS: true, EffectiveHostname: "mail.example.com", DefaultLanguage: "en"}, nil
	}
	return s.mailServerSettings, nil
}

func (s *fakeStore) UpdateMailServerSettings(ctx context.Context, settings domain.MailServerSettings) (domain.MailServerSettings, error) {
	s.mailServerSettings = settings
	if s.mailServerSettings.TLSMode == "" {
		s.mailServerSettings.TLSMode = "system"
	}
	if s.mailServerSettings.DefaultLanguage == "" {
		s.mailServerSettings.DefaultLanguage = "en"
	}
	if s.mailServerSettings.EffectiveHostname == "" {
		s.mailServerSettings.EffectiveHostname = settings.MailHostname
	}
	return s.mailServerSettings, nil
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
		MailHost: "mail.example.com",
		Records: []domain.DNSRecord{
			{Type: "MX", Name: "example.com", Value: "mail.example.com", Priority: &priority},
			{Type: "TXT", Name: "example.com", Value: "v=spf1 mx -all"},
			{Type: "TXT", Name: "_dmarc.example.com", Value: "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"},
			{Type: "TXT", Name: "mail._domainkey.example.com", Value: "v=DKIM1; k=rsa; p=test"},
			{Type: "CNAME", Name: "autoconfig.example.com", Value: "mail.example.com"},
			{Type: "CNAME", Name: "autodiscover.example.com", Value: "mail.example.com"},
			{Type: "SRV", Name: "_imaps._tcp.example.com", Value: "0 1 993 mail.example.com"},
			{Type: "SRV", Name: "_submission._tcp.example.com", Value: "0 1 587 mail.example.com"},
		},
		ClientSetup: []domain.ClientSetup{
			{Client: "Thunderbird", Method: "Autoconfig XML", Status: "supported"},
			{Client: "Outlook", Method: "Autodiscover XML", Status: "supported"},
			{Client: "Gmail app", Method: "Manual IMAP/SMTP", Status: "manual"},
		},
		Provisionable: true,
	}, nil
}

func (s *fakeStore) GetDomainTLS(ctx context.Context, domainID uint64) (domain.DomainTLS, error) {
	if s.tlsState.DomainID != 0 {
		return s.tlsState, nil
	}
	settings := domain.DomainTLSSettings{
		DomainID:               domainID,
		DNSWebmailAliasEnabled: true,
		DNSAdminAliasEnabled:   true,
		TLSMode:                "inherit",
		ChallengeType:          "dns-cloudflare",
		UseForHTTPS:            true,
		UseForMailSNI:          true,
		IncludeMailHostname:    true,
		IncludeWebmailHostname: true,
		IncludeAdminHostname:   true,
		DesiredHostnames:       []string{"mail.example.com", "webmail.example.com", "madmin.example.com"},
	}
	return domain.DomainTLS{
		DomainID: domainID,
		Domain:   "example.com",
		Settings: settings,
		Certificates: []domain.TLSCertificate{{
			ID:             91,
			DomainID:       domainID,
			Source:         "letsencrypt",
			Status:         "active",
			CommonName:     "mail.example.com",
			SANs:           []string{"mail.example.com", "webmail.example.com"},
			CertPath:       "/etc/letsencrypt/live/mail.example.com/fullchain.pem",
			KeyPath:        "/etc/letsencrypt/live/mail.example.com/privkey.pem",
			DaysRemaining:  72,
			UsedForHTTPS:   true,
			UsedForMailSNI: true,
		}},
		Jobs: []domain.TLSCertificateJob{{
			ID:            92,
			DomainID:      domainID,
			JobType:       "issue",
			ChallengeType: "dns-cloudflare",
			Status:        "queued",
			Step:          "queued",
			Progress:      0,
			Hostnames:     []string{"mail.example.com"},
		}},
	}, nil
}

func (s *fakeStore) ListAvailableTLSCertificates(ctx context.Context) ([]domain.TLSCertificate, error) {
	return []domain.TLSCertificate{{
		ID:            77,
		DomainID:      22,
		DomainName:    "example.com",
		Source:        "letsencrypt",
		Status:        "active",
		CommonName:    "mail.example.com",
		SANs:          []string{"mail.example.com", "webmail.example.com"},
		CertPath:      "/etc/letsencrypt/live/mail.example.com/fullchain.pem",
		KeyPath:       "/etc/letsencrypt/live/mail.example.com/privkey.pem",
		DaysRemaining: 82,
		UsedForHTTPS:  true,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}}, nil
}

func (s *fakeStore) UpdateDomainTLSSettings(ctx context.Context, settings domain.DomainTLSSettings) (domain.DomainTLSSettings, error) {
	s.tlsSettings = settings
	if s.tlsSettings.TLSMode == "" {
		s.tlsSettings.TLSMode = "inherit"
	}
	if s.tlsSettings.ChallengeType == "" {
		s.tlsSettings.ChallengeType = "dns-cloudflare"
	}
	s.tlsSettings.DesiredHostnames = []string{"mail.example.com"}
	return s.tlsSettings, nil
}

func (s *fakeStore) CreateTLSCertificateJob(ctx context.Context, job domain.TLSCertificateJob) (domain.TLSCertificateJob, error) {
	s.tlsJob = job
	s.tlsJob.ID = 92
	s.tlsJob.Status = "queued"
	s.tlsJob.Step = "queued"
	if s.tlsJob.ChallengeType == "" {
		s.tlsJob.ChallengeType = "dns-cloudflare"
	}
	if len(s.tlsJob.Hostnames) == 0 {
		s.tlsJob.Hostnames = []string{"mail.example.com"}
	}
	return s.tlsJob, nil
}

func (s *fakeStore) GetCloudflareConfig(ctx context.Context, domainID uint64) (domain.CloudflareConfig, error) {
	if s.cloudflareConfig.DomainID == 0 {
		return domain.CloudflareConfig{DomainID: domainID, ZoneID: s.savedCloudflareZoneID, Status: "configured", TokenConfigured: s.savedCloudflareToken != ""}, nil
	}
	return s.cloudflareConfig, nil
}

func (s *fakeStore) SaveCloudflareConfig(ctx context.Context, domainID uint64, zoneID, apiToken string) (domain.CloudflareConfig, error) {
	s.savedCloudflareDomainID = domainID
	s.savedCloudflareZoneID = zoneID
	s.savedCloudflareToken = apiToken
	s.cloudflareConfig = domain.CloudflareConfig{DomainID: domainID, ZoneID: zoneID, Status: "configured", TokenConfigured: apiToken != ""}
	return s.cloudflareConfig, nil
}

func (s *fakeStore) CheckCloudflareDNS(ctx context.Context, domainID uint64) (domain.DNSProvisionPlan, error) {
	if s.cloudflarePlan.DomainID == 0 {
		return domain.DNSProvisionPlan{DomainID: domainID, Domain: "example.com", Provider: "cloudflare", Status: "ok", Summary: "Cloudflare DNS already matches desired mail records."}, nil
	}
	return s.cloudflarePlan, nil
}

func (s *fakeStore) ApplyCloudflareDNS(ctx context.Context, domainID uint64, replace bool) (domain.DNSProvisionResult, error) {
	s.appliedCloudflareReplace = replace
	if s.cloudflareResult.Plan.DomainID == 0 {
		s.cloudflareResult.Plan = s.cloudflarePlan
	}
	if s.cloudflareResult.Plan.DomainID == 0 {
		s.cloudflareResult.Plan = domain.DNSProvisionPlan{DomainID: domainID, Domain: "example.com", Provider: "cloudflare", Status: "ok"}
	}
	return s.cloudflareResult, nil
}
