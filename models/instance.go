package models

import "context"

// SaveOptions configures instance save behavior.
type SaveOptions struct {
	ForceInsert  bool
	ForceUpdate  bool
	UpdateFields []string
	Using        string
	Raw          bool
}

// SaveOption mutates save options.
type SaveOption func(*SaveOptions)

// ForceInsert forces an insert operation.
func ForceInsert() SaveOption {
	return func(options *SaveOptions) {
		options.ForceInsert = true
	}
}

// ForceUpdate forces an update operation.
func ForceUpdate() SaveOption {
	return func(options *SaveOptions) {
		options.ForceUpdate = true
	}
}

// UpdateFields limits saving to selected fields.
func UpdateFields(fields ...string) SaveOption {
	return func(options *SaveOptions) {
		options.UpdateFields = append([]string(nil), fields...)
	}
}

// UsingDatabase selects a database alias.
func UsingDatabase(alias string) SaveOption {
	return func(options *SaveOptions) {
		options.Using = alias
	}
}

// RawSave marks the save as raw data loading.
func RawSave() SaveOption {
	return func(options *SaveOptions) {
		options.Raw = true
	}
}

// DeleteOptions configures instance delete behavior.
type DeleteOptions struct {
	Using       string
	KeepParents bool
}

// DeleteOption mutates delete options.
type DeleteOption func(*DeleteOptions)

// DeleteUsingDatabase selects a database alias for delete.
func DeleteUsingDatabase(alias string) DeleteOption {
	return func(options *DeleteOptions) {
		options.Using = alias
	}
}

// KeepParentRows preserves parent rows for inheritance strategies.
func KeepParentRows() DeleteOption {
	return func(options *DeleteOptions) {
		options.KeepParents = true
	}
}

// RefreshOptions configures refresh behavior.
type RefreshOptions struct {
	Using  string
	Fields []string
}

// InstanceStore is the persistence contract used by instance operations.
type InstanceStore interface {
	Save(context.Context, Model, SaveOptions) error
	Delete(context.Context, Model, DeleteOptions) error
	RefreshFromDB(context.Context, Model, RefreshOptions) error
	FromDB(context.Context, Model, map[string]any) error
}

// Save persists a model through a store.
func Save(ctx context.Context, model Model, store InstanceStore, options ...SaveOption) error {
	if store == nil {
		return ErrMissingStore
	}
	resolved := SaveOptions{}
	for _, option := range options {
		option(&resolved)
	}
	if err := store.Save(ctx, model, resolved); err != nil {
		return err
	}
	setModelState(model, StateLoaded)
	return nil
}

// Delete deletes a model through a store.
func Delete(ctx context.Context, model Model, store InstanceStore, options ...DeleteOption) error {
	if store == nil {
		return ErrMissingStore
	}
	resolved := DeleteOptions{}
	for _, option := range options {
		option(&resolved)
	}
	if err := store.Delete(ctx, model, resolved); err != nil {
		return err
	}
	setModelState(model, StateDeleted)
	return nil
}

// FullClean runs all model validation steps.
func FullClean(ctx context.Context, model Model) error {
	if err := CleanFields(ctx, model); err != nil {
		return err
	}
	if err := Clean(ctx, model); err != nil {
		return err
	}
	if err := ValidateUnique(ctx, model); err != nil {
		return err
	}
	return ValidateConstraints(ctx, model)
}

// CleanFields validates individual fields when the model implements the hook.
func CleanFields(ctx context.Context, model Model) error {
	if hook, ok := model.(interface{ CleanFields(context.Context) error }); ok {
		return hook.CleanFields(ctx)
	}
	return nil
}

// Clean runs model-level validation when the model implements the hook.
func Clean(ctx context.Context, model Model) error {
	if hook, ok := model.(interface{ Clean(context.Context) error }); ok {
		return hook.Clean(ctx)
	}
	return nil
}

// ValidateUnique validates uniqueness when the model implements the hook.
func ValidateUnique(ctx context.Context, model Model) error {
	if hook, ok := model.(interface{ ValidateUnique(context.Context) error }); ok {
		return hook.ValidateUnique(ctx)
	}
	return nil
}

// ValidateConstraints validates constraints when the model implements the hook.
func ValidateConstraints(ctx context.Context, model Model) error {
	if hook, ok := model.(interface{ ValidateConstraints(context.Context) error }); ok {
		return hook.ValidateConstraints(ctx)
	}
	return nil
}

// RefreshFromDB refreshes a model from the store.
func RefreshFromDB(ctx context.Context, model Model, store InstanceStore, options RefreshOptions) error {
	if store == nil {
		return ErrMissingStore
	}
	if err := store.RefreshFromDB(ctx, model, options); err != nil {
		return err
	}
	setModelState(model, StateLoaded)
	return nil
}

// FromDB hydrates a model from database values.
func FromDB(ctx context.Context, model Model, store InstanceStore, values map[string]any) error {
	if store == nil {
		return ErrMissingStore
	}
	if err := store.FromDB(ctx, model, values); err != nil {
		return err
	}
	setModelState(model, StateLoaded)
	return nil
}

// GetAbsoluteURL returns the model URL when implemented.
func GetAbsoluteURL(model Model) string {
	if hook, ok := model.(interface{ AbsoluteURL() string }); ok {
		return hook.AbsoluteURL()
	}
	return ""
}

// GetFieldDisplay returns a display value for a field when implemented.
func GetFieldDisplay(model Model, field string) (string, bool) {
	if hook, ok := model.(interface{ FieldDisplay(string) (string, bool) }); ok {
		return hook.FieldDisplay(field)
	}
	return "", false
}

// SerializableValue returns a serialized field value when implemented.
func SerializableValue(model Model, field string) (any, bool) {
	if hook, ok := model.(interface{ SerializableValue(string) (any, bool) }); ok {
		return hook.SerializableValue(field)
	}
	return nil, false
}

// NaturalKey returns the model natural key when implemented.
func NaturalKey(model Model) []any {
	if hook, ok := model.(interface{ NaturalKey() []any }); ok {
		return append([]any(nil), hook.NaturalKey()...)
	}
	return nil
}

func setModelState(model Model, state State) {
	if setter, ok := model.(interface{ SetModelState(State) }); ok {
		setter.SetModelState(state)
	}
}
