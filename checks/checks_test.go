package checks

import (
	"context"
	"reflect"
	"testing"
)

func TestCheckRegistrationFilteringAndSeverityThresholds(t *testing.T) {
	registry := NewRegistry()
	registry.Register(Check{ID: "models.E001", Tags: []string{"models"}, Severity: SeverityError, Message: "model error"})
	registry.Register(Check{ID: "security.W001", Tags: []string{"security"}, Severity: SeverityWarning, Message: "security warning"})
	results := registry.Run(context.Background(), Options{Tags: []string{"models"}, MinimumSeverity: SeverityDebug})
	if len(results) != 1 || results[0].ID != "models.E001" {
		t.Fatalf("filtered results = %#v", results)
	}
	if !HasFailures(results, SeverityError) || HasFailures(results, SeverityCritical) {
		t.Fatalf("failure threshold mismatch")
	}
	if got := resultIDs(DefaultRegistry().Run(context.Background(), Options{})); !contains(got, "settings.I001") || !contains(got, "queue.I001") {
		t.Fatalf("default check IDs = %#v", got)
	}
}

func TestResultOrdering(t *testing.T) {
	registry := NewRegistry()
	registry.Register(Check{ID: "b.W001", Severity: SeverityWarning})
	registry.Register(Check{ID: "a.W001", Severity: SeverityWarning})
	got := resultIDs(registry.Run(context.Background(), Options{}))
	if !reflect.DeepEqual(got, []string{"a.W001", "b.W001"}) {
		t.Fatalf("ordered IDs = %#v", got)
	}
}

func resultIDs(results []Result) []string {
	ids := make([]string, len(results))
	for i, result := range results {
		ids[i] = result.ID
	}
	return ids
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
