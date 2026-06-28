package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthRegistryPassingFailingAndTimeout(t *testing.T) {
	registry := NewRegistry()
	registry.Register(Check{Name: "database", Required: true, Timeout: time.Second, Run: func(context.Context) error { return nil }})
	wantErr := errors.New("down")
	registry.Register(Check{Name: "cache", Required: true, Timeout: time.Second, Run: func(context.Context) error { return wantErr }})
	report := registry.Run(context.Background(), Readiness)
	if report.OK || len(report.Results) != 2 || report.Results[1].Error == "" {
		t.Fatalf("report = %#v", report)
	}
	timeoutRegistry := NewRegistry()
	timeoutRegistry.Register(Check{Name: "slow", Required: true, Timeout: time.Millisecond, Run: func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}})
	report = timeoutRegistry.Run(context.Background(), Readiness)
	if report.OK || !strings.Contains(report.Results[0].Error, "deadline") {
		t.Fatalf("timeout report = %#v", report)
	}
}

func TestReadinessAndLivenessHandlers(t *testing.T) {
	registry := NewRegistry()
	registry.Register(Check{Name: "optional", Required: false, Run: func(context.Context) error { return errors.New("optional down") }})
	recorder := httptest.NewRecorder()
	registry.ReadinessHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("optional readiness status = %d", recorder.Code)
	}
	registry.Register(Check{Name: "required", Required: true, Run: func(context.Context) error { return errors.New("required down") }})
	recorder = httptest.NewRecorder()
	registry.ReadinessHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("required readiness status = %d", recorder.Code)
	}
	recorder = httptest.NewRecorder()
	registry.LivenessHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/live", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("liveness status = %d", recorder.Code)
	}
}
