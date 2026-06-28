package i18n

import (
	"context"
	"testing"
	"time"
)

func TestTimezoneActivationFormattingAndUTCStorage(t *testing.T) {
	location := time.FixedZone("IST", 5*60*60+30*60)
	ctx := WithTimezone(context.Background(), location)
	if TimezoneFromContext(ctx) != location {
		t.Fatalf("timezone missing from context")
	}
	utc := StoreUTC(time.Date(2026, 6, 28, 12, 0, 0, 0, location))
	if utc.Location() != time.UTC {
		t.Fatalf("StoreUTC() location = %s", utc.Location())
	}
	formatted := FormatLocalDate(ctx, utc, "2006-01-02 15:04")
	if formatted != "2026-06-28 12:00" {
		t.Fatalf("FormatLocalDate() = %q", formatted)
	}
	if TimezoneFromContext(context.Background()) != time.UTC {
		t.Fatalf("default timezone should be UTC")
	}
}
