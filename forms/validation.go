package forms

// FormCleanFunc validates cross-field state after individual fields clean.
type FormCleanFunc func(*Form) error

// Media contains static assets required by a form or widget.
type Media struct {
	CSS []string
	JS  []string
}

// MediaProvider is implemented by widgets that need CSS or JavaScript assets.
type MediaProvider interface {
	Media() Media
}

func (m Media) Merge(other Media) Media {
	return Media{
		CSS: mergeMediaValues(m.CSS, other.CSS),
		JS:  mergeMediaValues(m.JS, other.JS),
	}
}

func mergeMediaValues(left, right []string) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	merged := make([]string, 0, len(left)+len(right))
	for _, value := range append(append([]string(nil), left...), right...) {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		merged = append(merged, value)
	}
	return merged
}
