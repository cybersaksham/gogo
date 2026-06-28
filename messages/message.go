package messages

type Level int

const (
	LevelDebug   Level = 10
	LevelInfo    Level = 20
	LevelSuccess Level = 25
	LevelWarning Level = 30
	LevelError   Level = 40
)

type Message struct {
	Level     Level    `json:"level"`
	Text      string   `json:"text"`
	Tags      []string `json:"tags,omitempty"`
	ExtraTags []string `json:"extra_tags,omitempty"`
}

func (m Message) AllTags() []string {
	tags := append([]string(nil), m.Tags...)
	tags = append(tags, m.ExtraTags...)
	return tags
}

func TagsForLevel(level Level) []string {
	switch level {
	case LevelDebug:
		return []string{"debug"}
	case LevelInfo:
		return []string{"info"}
	case LevelSuccess:
		return []string{"success"}
	case LevelWarning:
		return []string{"warning"}
	case LevelError:
		return []string{"error"}
	default:
		return nil
	}
}
