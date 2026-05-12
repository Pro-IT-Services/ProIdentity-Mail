package webmail

import (
	"strings"
	"testing"
)

func TestWebmailSidebarScrollsOnlyOnOverflow(t *testing.T) {
	for _, want := range []string{
		"overflow-y: auto;",
		"overscroll-behavior: contain;",
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail sidebar CSS missing %q", want)
		}
	}
}

func TestWebmailFocusedOtherTabsExplainAutomatedMail(t *testing.T) {
	for _, want := range []string{
		`title="Focused shows normal person-to-person inbox mail"`,
		`title="Other shows newsletters, notifications, receipts, and automated mail"`,
		"auto-submitted",
		"automatic",
		"microsoft outlook",
		"list-unsubscribe",
	} {
		if !strings.Contains(strings.ToLower(webmailIndexHTML), strings.ToLower(want)) {
			t.Fatalf("webmail focused/other behavior missing %q", want)
		}
	}
}

func TestWebmailComposeSupportsAttachmentsSignatureAndProfile(t *testing.T) {
	for _, want := range []string{
		`type="file"`,
		`name="attachments"`,
		`id="profile-card"`,
		`id="profile-modal"`,
		`profile-settings-grid`,
		`name="signature_html"`,
		`name="signature_auto_add"`,
		`name="language"`,
		`id="profile-language"`,
		`id="app-password-summary"`,
		`<strong>Mailbox 2FA</strong>`,
		`<strong>App passwords</strong>`,
		`supportedLanguages`,
		`Slovenčina`,
		`i18nCatalog`,
		`translateUI`,
		`const applyLanguage`,
		`id="password-current"`,
		`Name and mailbox address are managed by an administrator`,
		`data-editor-command="justifyLeft"`,
		`data-editor-command="formatBlock"`,
		`api("/api/v1/profile"`,
		`box.textContent = t(message)`,
		`confirm(t("Revoke this app password? Devices using it will stop syncing."))`,
		`confirm(t("Discard this message draft?"))`,
		`t("Set up mailbox 2FA")`,
		`t("Scan this QR code with an authenticator app, then enter the generated code.")`,
		`t("Enter the code from your authenticator app to open this mailbox.")`,
		`t("No app passwords yet.")`,
		`t("Copy this app password now.")`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail profile/compose UI missing %q", want)
		}
	}
	if strings.Contains(webmailIndexHTML, "Mailbox 2FA and app passwords") {
		t.Fatal("webmail profile still combines mailbox 2FA and app passwords under one heading")
	}
	for _, forbidden := range []string{
		`name="first_name"`,
		`name="last_name"`,
		`name="display_name"`,
		`id="profile-login-email"`,
	} {
		if strings.Contains(webmailIndexHTML, forbidden) {
			t.Fatalf("webmail profile UI still exposes admin-owned identity field %q", forbidden)
		}
	}
}

func TestWebmailSecondaryViewsUseTranslatableChrome(t *testing.T) {
	for _, want := range []string{
		`updateViewChrome("Search contacts...")`,
		`translateAttributes(document.querySelector("#search"))`,
		`People available to webmail and CardDAV clients.`,
		`Rules saved for this mailbox. Delivery-time execution is the next mail pipeline step.`,
		`No contacts found`,
		`No filters yet`,
		`Add Contact`,
		`Add Filter`,
		`Sync info`,
		`Destination folder`,
		`Move to folder`,
		`Mailbox security badges are shown only when message authentication data exists for the selected message.`,
		`t("Edit Contact")`,
		`t("Edit Event")`,
		`t("Add at least one recipient")`,
		`t("Send failed")`,
		`t("No messages")`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail secondary view translation hook missing %q", want)
		}
	}
}

func TestWebmailLoginIncludesProIdentityPushUI(t *testing.T) {
	for _, want := range []string{
		`id="mailbox-push-card"`,
		`id="mailbox-push-status"`,
		`id="mailbox-push-manual"`,
		`id="mailbox-push-check"`,
		`pollMailboxProIdentityMFA`,
		`showMailboxPushManualCode`,
		`provider === "proidentity"`,
		`Check your phone`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail ProIdentity push UI missing %q", want)
		}
	}
}

func TestWebmailSignatureAutolinksEmailAndPhoneNumbers(t *testing.T) {
	for _, want := range []string{
		`linkifySignatureHTML`,
		`linkifySignatureText`,
		`normalizeSignaturePhone`,
		`<a href=\"`,
		`mailto:`,
		`tel:`,
		`(?:phone|mobile|cell|tel|telephone)`,
		`(?:mail|email|e-mail)`,
		`startsInternational`,
		`raw.startsWith("0")`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail signature linkification missing %q", want)
		}
	}
}

func TestWebmailComposeProtectsDirtyDraftsAndSavesServerDrafts(t *testing.T) {
	for _, want := range []string{
		`function composeIsDirty`,
		`function requestComposeClose`,
		`confirm(t("Discard this message draft?"))`,
		`/api/v1/drafts`,
		`state.currentDraftId`,
		`data.set("draft_id"`,
		`Draft saved to mailbox`,
		`id: "drafts", name: "Drafts"`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail compose draft protection missing %q", want)
		}
	}
	if strings.Contains(webmailIndexHTML, `#compose-backdrop").addEventListener("click", closeCompose)`) {
		t.Fatal("compose backdrop still closes without dirty draft guard")
	}
}

func TestWebmailSpamActionsAndContextMenuAreFolderAware(t *testing.T) {
	for _, want := range []string{
		`function canMarkSpamInFolder`,
		`function canMarkNotSpamInFolder`,
		`#mark-spam`,
		`#mark-ham`,
		`message-context-menu`,
		`function openMessageContextMenu`,
		`function contextMenuActions`,
		`data-context-action`,
		`archive-message`,
		`delete-forever`,
		`move-folder:`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail folder-aware spam/context menu missing %q", want)
		}
	}
	if strings.Contains(webmailIndexHTML, `contextmenu", closeCompose`) {
		t.Fatal("context menu is wired to unrelated compose close behavior")
	}
}

func TestWebmailTrashRestoreDeleteModalAndNoTopRibbon(t *testing.T) {
	for _, want := range []string{
		`function restoreSelectedFromTrash`,
		`restore-trash`,
		`data-context-action`,
		`activeSelectionIDs()`,
		`delete-confirm-modal`,
		`function showDeleteConfirmation`,
		`deleteSelectionMode`,
		`event.key === "Delete"`,
		`message-delete-subject`,
		`sidebar-separator`,
		`t("Move to Trash")`,
		`t("Delete forever")`,
		`t("Do you really want to permanently remove this message?")`,
		`t("Do you really want to remove these messages?")`,
		`t("selected messages")`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail restore/delete/sidebar behavior missing %q", want)
		}
	}
	for _, forbidden := range []string{
		`ribbon-tabs`,
		`ribbon-tab`,
		`data-ribbon-action`,
		`function handleRibbonAction`,
	} {
		if strings.Contains(webmailIndexHTML, forbidden) {
			t.Fatalf("webmail top ribbon should be removed, found %q", forbidden)
		}
	}
}

func TestWebmailResponsiveMailShell(t *testing.T) {
	for _, want := range []string{
		`@media (max-width: 1180px)`,
		`@media (max-width: 760px)`,
		`mobile-switcher`,
		`mobile-compose-button`,
		`data-mobile-pane="sidebar"`,
		`data-mobile-pane="list"`,
		`data-mobile-pane="reader"`,
		`function setMobilePane`,
		`function isMobileLayout`,
		`mobile-pane-list`,
		`mobile-pane-reader`,
		`100dvh`,
		`env(safe-area-inset-bottom)`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail responsive shell missing %q", want)
		}
	}
	for _, forbidden := range []string{
		`body { overflow: auto; height: auto; }`,
		`.app { height: auto; grid-template-columns: 1fr; }`,
	} {
		if strings.Contains(webmailIndexHTML, forbidden) {
			t.Fatalf("webmail responsive shell still uses broken stacked layout %q", forbidden)
		}
	}
}

func TestWebmailReaderMetadataAndTrustedSenderBanner(t *testing.T) {
	for _, want := range []string{
		`reader-content mail-reader`,
		`message-display-body`,
		`message-auth-panel`,
		`message-trust-banner`,
		`message-local-delivery`,
		`message-summary`,
		`function isLocalServerMessage`,
		`function messageAuthStatus`,
		`Local server delivery`,
		`External sender checks are not required for local mail`,
		`Authentication details are not available for this message`,
		`TRUSTED SENDER IDENTITY VERIFIED`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail reader metadata/trust UI missing %q", want)
		}
	}
	if strings.Contains(webmailIndexHTML, `No SPF, DKIM, TLS, or identity verification claim is being made`) {
		t.Fatal("webmail reader auth copy is too scary for normal missing/local auth state")
	}
}

func TestWebmailReaderAuthPanelSitsUnderSenderInfo(t *testing.T) {
	senderRow := strings.Index(webmailIndexHTML, `"<div class=\"sender-row\"`)
	authPanel := strings.Index(webmailIndexHTML, `authPanel + "<div class=\"body message-display-body\">`)
	body := strings.Index(webmailIndexHTML, `"<div class=\"body message-display-body\">`)
	summaryMeta := strings.Index(webmailIndexHTML, `<div class=\"message-meta\"><div class=\"recommend message-summary\"`)
	for name, index := range map[string]int{
		"sender row":   senderRow,
		"auth panel":   authPanel,
		"message body": body,
		"summary meta": summaryMeta,
	} {
		if index < 0 {
			t.Fatalf("webmail reader markup missing %s", name)
		}
	}
	if !(senderRow < authPanel && authPanel < body) {
		t.Fatalf("auth panel must render under sender info and above message body: sender=%d auth=%d body=%d", senderRow, authPanel, body)
	}
	if strings.Contains(webmailIndexHTML, `<div class=\"message-meta\">" + authPanel`) {
		t.Fatal("auth panel still renders inside bottom message metadata block")
	}
}

func TestWebmailReaderMetadataDoesNotOverlayMessageBody(t *testing.T) {
	for _, want := range []string{
		`.message-display-body { flex: 0 0 auto;`,
		`.message-meta { margin-top: auto;`,
		`padding: 18px 0 32px;`,
		`const height = Math.min(Math.max(doc.documentElement.scrollHeight, doc.body.scrollHeight, 260) + 12, 12000);`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail reader metadata flow missing %q", want)
		}
	}
	if strings.Contains(webmailIndexHTML, `.message-display-body { flex: 1 1 auto;`) {
		t.Fatal("message body can shrink and let metadata overlay rendered email content")
	}
}

func TestWebmailExternalContentTrustControlsAreUserScoped(t *testing.T) {
	for _, want := range []string{
		`data-trust-external-sender`,
		`Always trust this sender`,
		`data-trust-external-domain`,
		`Always trust this domain`,
		`/api/v1/content-trust`,
		`state.contentTrust`,
		`function loadContentTrust`,
		`function messageContentTrust`,
		`function addContentTrust`,
		`function isBlockedPublicTrustDomain`,
		`gmail.com`,
		`outlook.com`,
		`outlook.xyz`,
		`azet.sk`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail external content trust controls missing %q", want)
		}
	}
	if strings.Contains(webmailIndexHTML, `globalContentTrust`) {
		t.Fatal("external content trust must be per user, not global")
	}
}

func TestWebmailComposeBottomBarsStayPinned(t *testing.T) {
	for _, want := range []string{
		`.compose-body { display: flex; min-height: 0; overflow: hidden; flex-direction: column; background: white; }`,
		`.compose-editor-shell { flex: 1 1 auto; min-height: 0; overflow: auto; padding: 24px 30px; }`,
		`.compose-attachments {`,
		`flex: 0 0 auto;`,
		`.compose-footer {`,
	} {
		if !strings.Contains(webmailIndexHTML, want) {
			t.Fatalf("webmail compose pinned bottom layout missing %q", want)
		}
	}
}
