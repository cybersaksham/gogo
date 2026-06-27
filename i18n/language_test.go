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
