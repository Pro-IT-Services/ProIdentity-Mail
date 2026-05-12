package i18n

import "strings"

type Language struct {
	Code string
	Name string
}

var SupportedLanguages = []Language{
	{Code: "bg", Name: "Bulgarian"},
	{Code: "hr", Name: "Croatian"},
	{Code: "cs", Name: "Czech"},
	{Code: "da", Name: "Danish"},
	{Code: "nl", Name: "Dutch"},
	{Code: "en", Name: "English"},
	{Code: "fi", Name: "Finnish"},
	{Code: "fr", Name: "French"},
	{Code: "de", Name: "German"},
	{Code: "el", Name: "Greek"},
	{Code: "hu", Name: "Hungarian"},
	{Code: "ga", Name: "Irish"},
	{Code: "it", Name: "Italian"},
	{Code: "lv", Name: "Latvian"},
	{Code: "lt", Name: "Lithuanian"},
	{Code: "pl", Name: "Polish"},
	{Code: "pt", Name: "Portuguese"},
	{Code: "ro", Name: "Romanian"},
	{Code: "sk", Name: "Slovak"},
	{Code: "sl", Name: "Slovenian"},
	{Code: "es", Name: "Spanish"},
	{Code: "sv", Name: "Swedish"},
}

func NormalizeLanguage(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	if value == "" {
		return "en"
	}
	if idx := strings.Index(value, "-"); idx > 0 {
		value = value[:idx]
	}
	for _, language := range SupportedLanguages {
		if language.Code == value {
			return value
		}
	}
	return ""
}

func IsSupportedLanguage(value string) bool {
	return NormalizeLanguage(value) != ""
}
