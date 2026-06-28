package i18n

import (
	"context"
	"time"
)

type timezoneContextKey struct{}

func WithTimezone(ctx context.Context, location *time.Location) context.Context {
	if location == nil {
		location = time.UTC
	}
	return context.WithValue(ctx, timezoneContextKey{}, location)
}

func TimezoneFromContext(ctx context.Context) *time.Location {
	location, _ := ctx.Value(timezoneContextKey{}).(*time.Location)
	if location == nil {
		return time.UTC
	}
	return location
}

func StoreUTC(value time.Time) time.Time {
	return value.UTC()
}

func FormatLocalDate(ctx context.Context, value time.Time, layout string) string {
	return value.In(TimezoneFromContext(ctx)).Format(layout)
}
