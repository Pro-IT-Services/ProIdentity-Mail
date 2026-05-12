package i18n

import (
	"strings"
	"testing"
)

func TestSupportedLanguagesMatchProductLocales(t *testing.T) {
	want := []string{"bg", "hr", "cs", "da", "nl", "en", "fi", "fr", "de", "el", "hu", "ga", "it", "lv", "lt", "pl", "pt", "ro", "sk", "sl", "es", "sv"}
	if len(SupportedLanguages) != len(want) {
		t.Fatalf("supported languages = %d, want %d", len(SupportedLanguages), len(want))
	}
	for i, code := range want {
		if SupportedLanguages[i].Code != code {
			t.Fatalf("language[%d] = %q, want %q", i, SupportedLanguages[i].Code, code)
		}
	}
}

func TestNormalizeLanguage(t *testing.T) {
	for input, want := range map[string]string{
		"":       "en",
		"EN_us":  "en",
		"sk-SK":  "sk",
		"pt-PT":  "pt",
		"xx":     "",
		"gmail":  "",
		" sv-SE": "sv",
	} {
		if got := NormalizeLanguage(input); got != want {
			t.Fatalf("NormalizeLanguage(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestTranslationCatalogIsCompleteForEverySupportedLanguage(t *testing.T) {
	catalog := TranslationCatalog()
	keys := CatalogKeys()
	if len(keys) < 60 {
		t.Fatalf("translation catalog is too small for the admin/webmail chrome: %d keys", len(keys))
	}
	for _, language := range SupportedLanguages {
		entries, ok := catalog[language.Code]
		if !ok {
			t.Fatalf("missing catalog for %s", language.Code)
		}
		if len(entries) != len(keys) {
			t.Fatalf("catalog %s has %d entries, want %d", language.Code, len(entries), len(keys))
		}
		for _, key := range keys {
			value := strings.TrimSpace(entries[key])
			if value == "" {
				t.Fatalf("catalog %s has empty translation for %q", language.Code, key)
			}
		}
	}
}

func TestTranslationCatalogSmokeChecksWordOrder(t *testing.T) {
	catalog := TranslationCatalog()
	for language, checks := range map[string]map[string]string{
		"de": {"Default interface language": "Standardsprache der Oberfläche", "Always trust this sender": "Diesem Absender immer vertrauen"},
		"fr": {"Default interface language": "Langue par défaut de l'interface", "Secure mailbox login": "Connexion sécurisée à la boîte mail"},
		"sk": {"Default interface language": "Predvolený jazyk rozhrania", "Do you really want to remove this message?": "Naozaj chcete odstrániť túto správu?"},
		"es": {"Default interface language": "Idioma predeterminado de la interfaz", "Send Message": "Enviar mensaje"},
	} {
		for key, want := range checks {
			if got := catalog[language][key]; got != want {
				t.Fatalf("catalog[%s][%q] = %q, want %q", language, key, got, want)
			}
		}
	}
}

func TestWebmailWorkspaceTranslationsCoverSecondaryViews(t *testing.T) {
	catalog := TranslationCatalog()
	for key, want := range map[string]string{
		"People available to webmail and CardDAV clients.":                                      "Ľudia dostupní vo webmaile a klientoch CardDAV.",
		"Rules saved for this mailbox. Delivery-time execution is the next mail pipeline step.": "Pravidlá uložené pre túto schránku. Spúšťanie pri doručení je ďalší krok poštového spracovania.",
		"Search contacts...": "Hľadať kontakty...",
		"Sync info":          "Informácie o synchronizácii",
		"Add Contact":        "Pridať kontakt",
		"Add Filter":         "Pridať filter",
		"No contacts found":  "Nenašli sa žiadne kontakty",
		"No filters yet":     "Zatiaľ žiadne filtre",
		"Destination folder": "Cieľový priečinok",
		"Move to folder":     "Presunúť do priečinka",
		"Mailbox security badges are shown only when message authentication data exists for the selected message.": "Bezpečnostné štítky schránky sa zobrazujú iba vtedy, keď pre vybranú správu existujú údaje o overení.",
	} {
		if got := catalog["sk"][key]; got != want {
			t.Fatalf("catalog[sk][%q] = %q, want %q", key, got, want)
		}
	}
}

func TestRecentAdminAndWebmailFeaturesAreTranslated(t *testing.T) {
	catalog := TranslationCatalog()
	for key, want := range map[string]string{
		"Admin MFA":            "Správcovské MFA",
		"ProIdentity Auth":     "ProIdentity overenie",
		"Native hardware keys": "Natívne hardvérové kľúče",
		"Hosted TOTP":          "Hostované TOTP",
		"Service URL":          "URL služby",
		"Real client IP":       "Skutočná IP klienta",
		"Behind Cloudflare proxy, use CF-Connecting-IP": "Za Cloudflare proxy používať CF-Connecting-IP",
		"Login protection":                  "Ochrana prihlásenia",
		"Mailbox two-factor authentication": "Dvojfaktorové overenie schránky",
		"Mailbox 2FA":                       "2FA schránky",
		"App passwords":                     "Heslá aplikácií",
		"Use app passwords for IMAP, SMTP, POP3, CalDAV, and CardDAV clients.": "Používajte heslá aplikácií pre klientov IMAP, SMTP, POP3, CalDAV a CardDAV.",
		"Check your phone": "Skontrolujte telefón",
		"We sent a sign-in request to the ProIdentity app on your registered device.": "Odoslali sme žiadosť o prihlásenie do aplikácie ProIdentity na vašom registrovanom zariadení.",
		"Revoke this app password? Devices using it will stop syncing.":               "Zrušiť toto heslo aplikácie? Zariadenia, ktoré ho používajú, sa prestanú synchronizovať.",
		"Two-factor verification": "Dvojfaktorové overenie",
		"Waiting for approval...": "Čaká sa na schválenie...",
		"Set up mailbox 2FA":      "Nastaviť 2FA schránky",
		"Scan this QR code with an authenticator app, then enter the generated code.": "Naskenujte tento QR kód v autentifikačnej aplikácii a potom zadajte vygenerovaný kód.",
		"Enter the code from your authenticator app to open this mailbox.":            "Zadajte kód z autentifikačnej aplikácie na otvorenie tejto schránky.",
		"Enter the hosted TOTP code or approve the push request.":                     "Zadajte hostovaný TOTP kód alebo schváľte push žiadosť.",
		"2FA enabled":                 "2FA zapnuté",
		"Set up 2FA":                  "Nastaviť 2FA",
		"No app passwords yet.":       "Zatiaľ žiadne heslá aplikácií.",
		"Scan setup QR":               "Naskenovať nastavovací QR kód",
		"Authenticator code":          "Kód autentifikátora",
		"Verify and enable":           "Overiť a zapnúť",
		"Copy this app password now.": "Skopírujte toto heslo aplikácie teraz.",
		"Discard this message draft?": "Zahodiť tento koncept správy?",
		"Draft saved to mailbox":      "Koncept uložený do schránky",
		"Move to Trash":               "Presunúť do koša",
		"Delete forever":              "Natrvalo odstrániť",
		"Do you really want to permanently remove this message?":   "Naozaj chcete natrvalo odstrániť túto správu?",
		"Do you really want to permanently remove these messages?": "Naozaj chcete natrvalo odstrániť tieto správy?",
		"Do you really want to remove these messages?":             "Naozaj chcete odstrániť tieto správy?",
		"selected messages":          "vybrané správy",
		"Edit Contact":               "Upraviť kontakt",
		"Edit Event":                 "Upraviť udalosť",
		"No messages":                "Žiadne správy",
		"Add at least one recipient": "Pridajte aspoň jedného príjemcu",
		"Send failed":                "Odoslanie zlyhalo",
	} {
		if got := catalog["sk"][key]; got != want {
			t.Fatalf("catalog[sk][%q] = %q, want %q", key, got, want)
		}
	}
}
