package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Mode string

const (
	Readiness Mode = "readiness"
	Liveness  Mode = "liveness"
)

type Check struct {
	Name     string
	Kind     string
	Required bool
	Timeout  time.Duration
	Run      func(context.Context) error
}

type Result struct {
	Name     string        `json:"name"`
	Kind     string        `json:"kind,omitempty"`
	Required bool          `json:"required"`
	OK       bool          `json:"ok"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

type Report struct {
	Mode    Mode      `json:"mode"`
	OK      bool      `json:"ok"`
	Results []Result  `json:"results"`
	Checked time.Time `json:"checked"`
}

type Registry struct {
	mu     sync.RWMutex
	checks []Check
	now    func() time.Time
}

func NewRegistry() *Registry {
	return &Registry{now: time.Now}
}

func (r *Registry) Register(check Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks = append(r.checks, check)
}

func (r *Registry) Run(ctx context.Context, mode Mode) Report {
	checks := r.checksSnapshot()
	report := Report{Mode: mode, OK: true, Checked: r.now().UTC()}
	for _, check := range checks {
		result := runCheck(ctx, check)
		if mode == Readiness && check.Required && !result.OK {
			report.OK = false
		}
		report.Results = append(report.Results, result)
	}
	return report
}

func (r *Registry) ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		report := r.Run(req.Context(), Readiness)
		writeReport(w, report)
	})
}

func (r *Registry) LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		report := r.Run(req.Context(), Liveness)
		writeReport(w, report)
	})
}

func (r *Registry) checksSnapshot() []Check {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]Check(nil), r.checks...)
}

func runCheck(ctx context.Context, check Check) Result {
	start := time.Now()
	result := Result{Name: check.Name, Kind: check.Kind, Required: check.Required, OK: true}
	runCtx := ctx
	cancel := func() {}
	if check.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, check.Timeout)
	}
	defer cancel()
	if check.Run != nil {
		if err := check.Run(runCtx); err != nil {
			result.OK = false
			result.Error = err.Error()
		}
	}
	result.Duration = time.Since(start)
	return result
}

func writeReport(w http.ResponseWriter, report Report) {
	w.Header().Set("Content-Type", "application/json")
	if !report.OK && report.Mode == Readiness {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(report)
}

type HealthChecker interface {
	HealthCheck(context.Context) error
}

func DependencyCheck(name string, kind string, required bool, timeout time.Duration, checker HealthChecker) Check {
	return Check{Name: name, Kind: kind, Required: required, Timeout: timeout, Run: checker.HealthCheck}
}
