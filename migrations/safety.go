package migrations

import (
	"fmt"
	"strings"
)

// SafetyOptions controls unsafe migration confirmation.
type SafetyOptions struct {
	NonInteractive bool
	Confirmed      bool
}

// SafetyCheck describes one risky operation.
type SafetyCheck struct {
	Operation string
	Message   string
}

// SafetyCheckedOperation can report unsafe behavior.
type SafetyCheckedOperation interface {
	SafetyChecks() []SafetyCheck
}

// CheckSafety validates unsafe operations against confirmation options.
func CheckSafety(operations []Operation, options SafetyOptions) error {
	var checks []SafetyCheck
	for _, operation := range operations {
		if checked, ok := operation.(SafetyCheckedOperation); ok {
			checks = append(checks, checked.SafetyChecks()...)
		}
	}
	if len(checks) == 0 || !options.NonInteractive || options.Confirmed {
		return nil
	}
	messages := make([]string, len(checks))
	for i, check := range checks {
		messages[i] = check.Operation + ": " + check.Message
	}
	return fmt.Errorf("%w: %s", ErrUnsafeMigration, strings.Join(messages, "; "))
}
