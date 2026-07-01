package operations

import (
	"context"
	"fmt"
	"strings"

	"github.com/cybersaksham/gogo/migrations"
)

type AddField struct {
	AppLabel, ModelName string
	TableName           string
	Field               migrations.FieldState
	HasDefault          bool
	UnsafeAcknowledged  bool
}

type RemoveField struct {
	AppLabel, ModelName string
	TableName           string
	Field               migrations.FieldState
}

type AlterField struct {
	AppLabel, ModelName string
	TableName           string
	OldField            migrations.FieldState
	NewField            migrations.FieldState
}

type RenameField struct {
	AppLabel, ModelName string
	TableName           string
	OldName, NewName    string
}

func (o AddField) Name() string { return "AddField" }
func (o AddField) ValidateSafety() error {
	if !o.Field.Null && !o.HasDefault && !o.UnsafeAcknowledged {
		return fmt.Errorf("%w: adding non-null field %s requires a default or acknowledgement", migrations.ErrUnsafeMigration, o.Field.Name)
	}
	return nil
}
func (o AddField) StateForwards(state *migrations.ProjectState) error {
	if err := o.ValidateSafety(); err != nil {
		return err
	}
	state.AddField(o.AppLabel, o.ModelName, o.Field)
	return nil
}
func (o AddField) DatabaseForwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.AddColumn(table, o.Field))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, columnName(o.Field), fieldKind(o.Field)))
}
func (o AddField) DatabaseBackwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.DropColumn(table, columnName(o.Field)))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, columnName(o.Field)))
}
func (o AddField) Describe() string { return "Add field " + o.Field.Name }
func (o AddField) Reversible() bool { return true }
func (o AddField) ReferencesModel(appLabel, modelName string) bool {
	return o.AppLabel == appLabel && o.ModelName == modelName
}
func (o AddField) ReferencesField(appLabel, modelName, fieldName string) bool {
	return o.ReferencesModel(appLabel, modelName) && o.Field.Name == fieldName
}
func (o AddField) SafetyChecks() []migrations.SafetyCheck {
	if !o.Field.Null && !o.HasDefault && !o.UnsafeAcknowledged {
		return []migrations.SafetyCheck{{Operation: o.Name(), Message: "adds non-null field without default"}}
	}
	return nil
}

func (o RemoveField) Name() string { return "RemoveField" }
func (o RemoveField) StateForwards(state *migrations.ProjectState) error {
	model := state.Models[key(o.AppLabel, o.ModelName)]
	model.Fields = removeField(model.Fields, o.Field.Name)
	state.Models[key(o.AppLabel, o.ModelName)] = model
	return nil
}
func (o RemoveField) DatabaseForwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.DropColumn(table, columnName(o.Field)))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, columnName(o.Field)))
}
func (o RemoveField) DatabaseBackwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.AddColumn(table, o.Field))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, columnName(o.Field), fieldKind(o.Field)))
}
func (o RemoveField) Describe() string { return "Remove field " + o.Field.Name }
func (o RemoveField) Reversible() bool { return true }
func (o RemoveField) ReferencesModel(appLabel, modelName string) bool {
	return o.AppLabel == appLabel && o.ModelName == modelName
}
func (o RemoveField) ReferencesField(appLabel, modelName, fieldName string) bool {
	return o.ReferencesModel(appLabel, modelName) && o.Field.Name == fieldName
}
func (o RemoveField) SafetyChecks() []migrations.SafetyCheck {
	return []migrations.SafetyCheck{{Operation: o.Name(), Message: "drops column " + o.Field.Name}}
}

func (o AlterField) Name() string { return "AlterField" }
func (o AlterField) StateForwards(state *migrations.ProjectState) error {
	model := state.Models[key(o.AppLabel, o.ModelName)]
	for i, field := range model.Fields {
		if field.Name == o.OldField.Name {
			model.Fields[i] = o.NewField
			break
		}
	}
	state.Models[key(o.AppLabel, o.ModelName)] = model
	return nil
}
func (o AlterField) DatabaseForwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.AlterColumnType(table, columnName(o.NewField), fieldKind(o.NewField)))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", table, columnName(o.NewField), fieldKind(o.NewField)))
}
func (o AlterField) DatabaseBackwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.AlterColumnType(table, columnName(o.OldField), fieldKind(o.OldField)))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", table, columnName(o.OldField), fieldKind(o.OldField)))
}
func (o AlterField) Describe() string { return "Alter field " + o.NewField.Name }
func (o AlterField) Reversible() bool { return true }
func (o AlterField) ReferencesModel(appLabel, modelName string) bool {
	return o.AppLabel == appLabel && o.ModelName == modelName
}
func (o AlterField) ReferencesField(appLabel, modelName, fieldName string) bool {
	return o.ReferencesModel(appLabel, modelName) && (o.OldField.Name == fieldName || o.NewField.Name == fieldName)
}
func (o AlterField) SafetyChecks() []migrations.SafetyCheck {
	if isNarrowingType(o.OldField.Kind, o.NewField.Kind) {
		return []migrations.SafetyCheck{{Operation: o.Name(), Message: "narrows field type " + o.NewField.Name}}
	}
	return nil
}

func (o RenameField) Name() string { return "RenameField" }
func (o RenameField) StateForwards(state *migrations.ProjectState) error {
	model := state.Models[key(o.AppLabel, o.ModelName)]
	for i, field := range model.Fields {
		if field.Name == o.OldName {
			field.Name = o.NewName
			if field.Column == "" || field.Column == o.OldName {
				field.Column = o.NewName
			}
			model.Fields[i] = field
			break
		}
	}
	state.Models[key(o.AppLabel, o.ModelName)] = model
	return nil
}
func (o RenameField) DatabaseForwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.RenameColumn(table, o.OldName, o.NewName))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, o.OldName, o.NewName))
}
func (o RenameField) DatabaseBackwards(ctx context.Context, editor migrations.SchemaEditor) error {
	table := operationTableName(o.TableName, o.AppLabel, o.ModelName)
	if renderer, ok := editor.(migrations.SchemaRenderer); ok {
		return editor.Execute(ctx, renderer.RenameColumn(table, o.NewName, o.OldName))
	}
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, o.NewName, o.OldName))
}
func (o RenameField) Describe() string { return "Rename field " + o.OldName + " to " + o.NewName }
func (o RenameField) Reversible() bool { return true }
func (o RenameField) ReferencesModel(appLabel, modelName string) bool {
	return o.AppLabel == appLabel && o.ModelName == modelName
}
func (o RenameField) ReferencesField(appLabel, modelName, fieldName string) bool {
	return o.ReferencesModel(appLabel, modelName) && (o.OldName == fieldName || o.NewName == fieldName)
}
func (o RenameField) SafetyChecks() []migrations.SafetyCheck {
	return []migrations.SafetyCheck{{Operation: o.Name(), Message: "renames field with ambiguous data movement"}}
}

func removeField(fields []migrations.FieldState, name string) []migrations.FieldState {
	result := fields[:0]
	for _, field := range fields {
		if field.Name != name {
			result = append(result, field)
		}
	}
	return result
}

func tableName(appLabel, modelName string) string {
	return appLabel + "_" + strings.ToLower(modelName)
}

func operationTableName(explicit, appLabel, modelName string) string {
	if explicit != "" {
		return explicit
	}
	return tableName(appLabel, modelName)
}

func columnName(field migrations.FieldState) string {
	if field.Column != "" {
		return field.Column
	}
	return field.Name
}

func fieldKind(field migrations.FieldState) string {
	if field.Kind == "" {
		return "text"
	}
	return field.Kind
}

func isNarrowingType(oldKind, newKind string) bool {
	return oldKind == "text" && strings.HasPrefix(newKind, "varchar")
}
