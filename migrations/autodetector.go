package migrations

import "reflect"

type ChangeType string

const (
	ChangeCreateModel      ChangeType = "create_model"
	ChangeDeleteModel      ChangeType = "delete_model"
	ChangeRenameModel      ChangeType = "rename_model"
	ChangeAlterModel       ChangeType = "alter_model"
	ChangeAddField         ChangeType = "add_field"
	ChangeRemoveField      ChangeType = "remove_field"
	ChangeRenameField      ChangeType = "rename_field"
	ChangeAlterField       ChangeType = "alter_field"
	ChangeAddIndex         ChangeType = "add_index"
	ChangeRemoveIndex      ChangeType = "remove_index"
	ChangeRenameIndex      ChangeType = "rename_index"
	ChangeAddConstraint    ChangeType = "add_constraint"
	ChangeRemoveConstraint ChangeType = "remove_constraint"
	ChangeManyToManyChange ChangeType = "many_to_many_change"
)

// DetectedChange describes one autodetected migration change.
type DetectedChange struct {
	Type      ChangeType
	AppLabel  string
	ModelName string
	OldName   string
	NewName   string
}

// RenameQuestioner asks whether a delete/create pair is a rename.
type RenameQuestioner interface {
	IsRename(oldModel, newModel ModelState) bool
}

// RenameQuestionerFunc adapts a function into RenameQuestioner.
type RenameQuestionerFunc func(oldModel, newModel ModelState) bool

func (f RenameQuestionerFunc) IsRename(oldModel, newModel ModelState) bool {
	return f(oldModel, newModel)
}

// Autodetector compares historical and current project state.
type Autodetector struct {
	From       ProjectState
	To         ProjectState
	Questioner RenameQuestioner
}

// NewAutodetector creates a state comparator.
func NewAutodetector(from, to ProjectState) Autodetector {
	return Autodetector{From: from.Clone(), To: to.Clone()}
}

// Changes returns detected state changes.
func (a Autodetector) Changes() []DetectedChange {
	var changes []DetectedChange
	handledDeletes := map[string]bool{}
	handledCreates := map[string]bool{}

	if a.Questioner != nil {
		for oldKey, oldModel := range a.From.Models {
			for newKey, newModel := range a.To.Models {
				if oldModel.AppLabel == newModel.AppLabel && oldModel.Name != newModel.Name && a.Questioner.IsRename(oldModel, newModel) {
					changes = append(changes, DetectedChange{Type: ChangeRenameModel, AppLabel: oldModel.AppLabel, OldName: oldModel.Name, NewName: newModel.Name})
					handledDeletes[oldKey] = true
					handledCreates[newKey] = true
				}
			}
		}
	}

	for key, newModel := range a.To.Models {
		if handledCreates[key] {
			continue
		}
		oldModel, exists := a.From.Models[key]
		if !exists {
			changes = append(changes, DetectedChange{Type: ChangeCreateModel, AppLabel: newModel.AppLabel, ModelName: newModel.Name})
			continue
		}
		changes = append(changes, compareModel(oldModel, newModel)...)
	}
	for key, oldModel := range a.From.Models {
		if handledDeletes[key] {
			continue
		}
		if _, exists := a.To.Models[key]; !exists {
			changes = append(changes, DetectedChange{Type: ChangeDeleteModel, AppLabel: oldModel.AppLabel, ModelName: oldModel.Name})
		}
	}
	return changes
}

// MergeAutodetectedOperations preserves manual operations and appends autodetected ones.
func MergeAutodetectedOperations(migration Migration, operations []Operation) Migration {
	copied := migration
	copied.Operations = append(append([]Operation(nil), migration.Operations...), operations...)
	return copied
}

func compareModel(oldModel, newModel ModelState) []DetectedChange {
	var changes []DetectedChange
	if !reflect.DeepEqual(oldModel.Options, newModel.Options) || oldModel.TableName != newModel.TableName {
		changes = append(changes, DetectedChange{Type: ChangeAlterModel, AppLabel: newModel.AppLabel, ModelName: newModel.Name})
	}
	oldFields := fieldMap(oldModel.Fields)
	newFields := fieldMap(newModel.Fields)
	for name, field := range newFields {
		old, exists := oldFields[name]
		if !exists {
			changes = append(changes, DetectedChange{Type: ChangeAddField, AppLabel: newModel.AppLabel, ModelName: newModel.Name, NewName: name})
		} else if old != field {
			changes = append(changes, DetectedChange{Type: ChangeAlterField, AppLabel: newModel.AppLabel, ModelName: newModel.Name, NewName: name})
		}
	}
	for name := range oldFields {
		if _, exists := newFields[name]; !exists {
			changes = append(changes, DetectedChange{Type: ChangeRemoveField, AppLabel: newModel.AppLabel, ModelName: newModel.Name, OldName: name})
		}
	}
	changes = append(changes, compareNamedStates(newModel, oldModel.Indexes, newModel.Indexes, ChangeRemoveIndex, ChangeAddIndex)...)
	changes = append(changes, compareNamedStates(newModel, oldModel.Constraints, newModel.Constraints, ChangeRemoveConstraint, ChangeAddConstraint)...)
	return changes
}

func fieldMap(fields []FieldState) map[string]FieldState {
	values := make(map[string]FieldState, len(fields))
	for _, field := range fields {
		values[field.Name] = field
	}
	return values
}

type namedState interface{ comparableName() string }

func compareNamedStates[T interface{ comparableName() string }](model ModelState, oldValues, newValues []T, removeType, addType ChangeType) []DetectedChange {
	oldMap := map[string]T{}
	newMap := map[string]T{}
	for _, value := range oldValues {
		oldMap[value.comparableName()] = value
	}
	for _, value := range newValues {
		newMap[value.comparableName()] = value
	}
	var changes []DetectedChange
	for name := range oldMap {
		if _, exists := newMap[name]; !exists {
			changes = append(changes, DetectedChange{Type: removeType, AppLabel: model.AppLabel, ModelName: model.Name, OldName: name})
		}
	}
	for name := range newMap {
		if _, exists := oldMap[name]; !exists {
			changes = append(changes, DetectedChange{Type: addType, AppLabel: model.AppLabel, ModelName: model.Name, NewName: name})
		}
	}
	return changes
}

func (i IndexState) comparableName() string      { return i.Name }
func (c ConstraintState) comparableName() string { return c.Name }
