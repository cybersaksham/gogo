package models

import (
	"reflect"
	"strings"
	"time"
	"unicode"
)

// State describes the lifecycle state of a model instance.
type State int

const (
	StateNew State = iota
	StateLoaded
	StateDirty
	StateDeleted
)

// String returns a stable state label.
func (s State) String() string {
	switch s {
	case StateNew:
		return "new"
	case StateLoaded:
		return "loaded"
	case StateDirty:
		return "dirty"
	case StateDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

// Model exposes stable metadata for a model declaration.
type Model interface {
	ModelMeta() Metadata
}

// BaseModel is embeddable common model state.
type BaseModel struct {
	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time

	state State
}

// ModelMeta returns empty metadata for default resolution.
func (m BaseModel) ModelMeta() Metadata {
	return Metadata{}
}

// ModelState returns the instance lifecycle state.
func (m BaseModel) ModelState() State {
	return m.state
}

// SetModelState updates the instance lifecycle state.
func (m *BaseModel) SetModelState(state State) {
	m.state = state
}

// CompositePrimaryKey describes a multi-column primary key.
type CompositePrimaryKey struct {
	Columns []string
}

// Metadata contains the core model metadata available from task one.
type Metadata struct {
	DBTable             string
	DBTableComment      string
	AppLabel            string
	ModelName           string
	TableName           string
	VerboseName         string
	VerboseNamePlural   string
	Ordering            []string
	OrderWithRespectTo  string
	GetLatestBy         []string
	DefaultRelatedName  string
	DefaultManagerName  string
	BaseManagerName     string
	Abstract            bool
	Proxy               bool
	Managed             *bool
	RequiredDBVendor    string
	RequiredDBFeatures  []string
	Indexes             []Index
	Constraints         []Constraint
	Permissions         []Permission
	DefaultPermissions  []string
	SelectOnSave        bool
	GenerateMigrations  bool
	CompositePrimaryKey *CompositePrimaryKey
}

// Clone returns an immutable copy of metadata.
func (m Metadata) Clone() Metadata {
	copied := m
	if m.Managed != nil {
		value := *m.Managed
		copied.Managed = &value
	}
	copied.Ordering = append([]string(nil), m.Ordering...)
	copied.GetLatestBy = append([]string(nil), m.GetLatestBy...)
	copied.RequiredDBFeatures = append([]string(nil), m.RequiredDBFeatures...)
	copied.Indexes = cloneIndexes(m.Indexes)
	copied.Constraints = cloneConstraints(m.Constraints)
	copied.Permissions = append([]Permission(nil), m.Permissions...)
	copied.DefaultPermissions = append([]string(nil), m.DefaultPermissions...)
	if m.CompositePrimaryKey != nil {
		copied.CompositePrimaryKey = &CompositePrimaryKey{
			Columns: append([]string(nil), m.CompositePrimaryKey.Columns...),
		}
	}
	return copied
}

// ResolveMetadata resolves explicit and default metadata for a model.
func ResolveMetadata(model Model) Metadata {
	meta := Metadata{}
	if model != nil {
		meta = model.ModelMeta().Clone()
	}
	if meta.ModelName == "" {
		meta.ModelName = modelTypeName(model)
	}

	modelSlug := snakeCase(meta.ModelName)
	if meta.DBTable != "" && meta.TableName == "" {
		meta.TableName = meta.DBTable
	}
	if meta.TableName != "" && meta.DBTable == "" {
		meta.DBTable = meta.TableName
	}
	if meta.TableName == "" {
		if meta.AppLabel != "" {
			meta.TableName = meta.AppLabel + "_" + modelSlug
		} else {
			meta.TableName = modelSlug
		}
		meta.DBTable = meta.TableName
	}
	if meta.VerboseName == "" {
		meta.VerboseName = strings.ReplaceAll(modelSlug, "_", " ")
	}
	if meta.VerboseNamePlural == "" {
		meta.VerboseNamePlural = meta.VerboseName + "s"
	}
	if meta.DefaultManagerName == "" {
		meta.DefaultManagerName = "objects"
	}
	if meta.DefaultPermissions == nil {
		meta.DefaultPermissions = []string{"add", "change", "delete", "view"}
	}
	return meta
}

func modelTypeName(model Model) string {
	if model == nil {
		return ""
	}
	value := reflect.TypeOf(model)
	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	return value.Name()
}

func snakeCase(value string) string {
	var builder strings.Builder
	var previousLower bool
	for _, char := range value {
		if unicode.IsUpper(char) {
			if builder.Len() > 0 && previousLower {
				builder.WriteByte('_')
			}
			builder.WriteRune(unicode.ToLower(char))
			previousLower = false
			continue
		}
		if char == '-' || char == ' ' {
			builder.WriteByte('_')
			previousLower = false
			continue
		}
		builder.WriteRune(char)
		previousLower = unicode.IsLetter(char) || unicode.IsDigit(char)
	}
	return builder.String()
}
