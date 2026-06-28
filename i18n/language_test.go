package i18n

import (
	"context"
	"testing"
)

func TestLanguageContextAndNegotiation(t *testing.T) {
	ctx := WithLanguage(context.Background(), "fr")
	if got := LanguageFromContext(ctx); got != "fr" {
		t.Fatalf("LanguageFromContext() = %q, want fr", got)
	}

	got := NegotiateLanguage("fr-CA,fr;q=0.9,en;q=0.8", []string{"en", "fr"}, "en")
	if got != "fr" {
		t.Fatalf("NegotiateLanguage() = %q, want fr", got)
	}
}

func TestTranslationCatalogLazyValuesAndDefaultLanguage(t *testing.T) {
	catalog := NewMemoryCatalog(map[string]map[string]string{
		"fr": {"hello": "bonjour"},
	})
	ctx := WithLanguage(context.Background(), "fr")
	if got := Translate(ctx, catalog, "hello"); got != "bonjour" {
		t.Fatalf("Translate() = %q", got)
	}
	lazy := Lazy("hello", catalog)
	if got := lazy.String(ctx); got != "bonjour" {
		t.Fatalf("Lazy.String() = %q", got)
	}
	if got := LanguageFromContext(WithDefaultLanguage(context.Background(), "en")); got != "en" {
		t.Fatalf("default language = %q", got)
	}
}
