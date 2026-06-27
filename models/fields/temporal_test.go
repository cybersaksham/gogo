package fields

import (
	"errors"
	"testing"
	"time"
)

func TestDateFieldRequiresDateOnlyValue(t *testing.T) {
	field := NewDateField(Options{Name: "published_on"})

	if err := field.Validate(time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("Validate(date) error = %v", err)
	}
	if err := field.Validate(time.Date(2026, 6, 27, 1, 0, 0, 0, time.UTC)); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(datetime) error = %v, want ErrValidation", err)
	}
	if got := field.ColumnType("postgres"); got != "date" {
		t.Fatalf("ColumnType(postgres) = %q, want date", got)
	}
}

func TestDateTimeFieldNormalizesTimezoneAndAutoValues(t *testing.T) {
	location := time.FixedZone("IST", 5*60*60+30*60)
	now := time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)
	field := NewDateTimeField(Options{Name: "created_at"}, TemporalConfig{
		AutoNowAdd: true,
		Location:   location,
		Now:        func() time.Time { return now },
	})

	value, ok := field.AutoValue(true)
	if !ok || value.Location() != location || value.Hour() != 15 || value.Minute() != 30 {
		t.Fatalf("AutoValue(create) = (%v, %v), want IST 15:30", value, ok)
	}
	if _, ok := field.AutoValue(false); ok {
		t.Fatalf("AutoValue(update) returned value for AutoNowAdd")
	}

	dbValue, err := field.ToDB(time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ToDB() error = %v", err)
	}
	converted := dbValue.(time.Time)
	if converted.Location() != location || converted.Hour() != 5 || converted.Minute() != 30 {
		t.Fatalf("ToDB() = %v, want IST 05:30", converted)
	}
}

func TestTimeFieldRequiresTimeOnlyValue(t *testing.T) {
	field := NewTimeField(Options{Name: "starts_at"})

	if err := field.Validate(time.Date(0, 1, 1, 9, 30, 0, 0, time.UTC)); err != nil {
		t.Fatalf("Validate(time-only) error = %v", err)
	}
	if err := field.Validate(time.Date(2026, 6, 27, 9, 30, 0, 0, time.UTC)); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(date+time) error = %v, want ErrValidation", err)
	}
}

func TestDurationFieldConvertsDuration(t *testing.T) {
	field := NewDurationField(Options{Name: "elapsed"})
	value, err := field.ToDB(2 * time.Hour)
	if err != nil {
		t.Fatalf("ToDB() error = %v", err)
	}
	if value != int64((2 * time.Hour).Nanoseconds()) {
		t.Fatalf("ToDB() = %#v, want duration nanoseconds", value)
	}
	if err := field.Validate("slow"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(slow) error = %v, want ErrValidation", err)
	}
}
