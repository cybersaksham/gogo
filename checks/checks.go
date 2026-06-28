package checks

import (
	"context"
	"sort"
)

type Severity int

const (
	SeverityDebug Severity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARN"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

type Check struct {
	ID       string
	Tags     []string
	Severity Severity
	Message  string
	Hint     string
	Object   string
	Run      func(context.Context) Result
}

type Result struct {
	ID       string
	Tags     []string
	Severity Severity
	Message  string
	Hint     string
	Object   string
}

type Options struct {
	Tags            []string
	MinimumSeverity Severity
}

type Registry struct {
	checks []Check
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(check Check) {
	r.checks = append(r.checks, check)
}

func (r *Registry) Run(ctx context.Context, options Options) []Result {
	var results []Result
	for _, check := range r.checks {
		if !matchesTags(check.Tags, options.Tags) || check.Severity < options.MinimumSeverity {
			continue
		}
		result := Result{ID: check.ID, Tags: append([]string(nil), check.Tags...), Severity: check.Severity, Message: check.Message, Hint: check.Hint, Object: check.Object}
		if check.Run != nil {
			result = check.Run(ctx)
		}
		if result.ID == "" {
			result.ID = check.ID
		}
		if len(result.Tags) == 0 {
			result.Tags = append([]string(nil), check.Tags...)
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results
}

func HasFailures(results []Result, failLevel Severity) bool {
	for _, result := range results {
		if result.Severity >= failLevel {
			return true
		}
	}
	return false
}

func DefaultRegistry() *Registry {
	registry := NewRegistry()
	defaults := []Check{
		{ID: "settings.I001", Tags: []string{"settings"}, Severity: SeverityInfo, Message: "settings loaded"},
		{ID: "apps.I001", Tags: []string{"apps"}, Severity: SeverityInfo, Message: "app registry checks registered"},
		{ID: "models.I001", Tags: []string{"models"}, Severity: SeverityInfo, Message: "model metadata checks registered"},
		{ID: "migrations.I001", Tags: []string{"migrations"}, Severity: SeverityInfo, Message: "migration graph checks registered"},
		{ID: "auth.I001", Tags: []string{"auth"}, Severity: SeverityInfo, Message: "auth checks registered"},
		{ID: "admin.I001", Tags: []string{"admin"}, Severity: SeverityInfo, Message: "admin checks registered"},
		{ID: "static.I001", Tags: []string{"static"}, Severity: SeverityInfo, Message: "static files checks registered"},
		{ID: "security.I001", Tags: []string{"security"}, Severity: SeverityInfo, Message: "security checks registered"},
		{ID: "database.I001", Tags: []string{"database"}, Severity: SeverityInfo, Message: "database checks registered"},
		{ID: "queue.I001", Tags: []string{"queue"}, Severity: SeverityInfo, Message: "queue checks registered"},
	}
	for _, check := range defaults {
		registry.Register(check)
	}
	return registry
}

func matchesTags(checkTags []string, wanted []string) bool {
	if len(wanted) == 0 {
		return true
	}
	for _, want := range wanted {
		for _, tag := range checkTags {
			if tag == want {
				return true
			}
		}
	}
	return false
}
