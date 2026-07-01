package operations

import "github.com/cybersaksham/gogo/migrations"

func (o CreateModel) MigrationOperationSpec() migrations.OperationSpec {
	model := cloneModelState(o.Model)
	return migrations.OperationSpec{Type: o.Name(), Model: &model}
}

func (o DeleteModel) MigrationOperationSpec() migrations.OperationSpec {
	model := cloneModelState(o.Model)
	return migrations.OperationSpec{Type: o.Name(), Model: &model}
}

func (o RenameModel) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, OldName: o.OldName, NewName: o.NewName}
}

func (o AlterModelTable) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, OldTable: o.OldTable, NewTable: o.NewTable}
}

func (o AlterModelTableComment) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, Comment: o.Comment}
}

func (o AlterModelOptions) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, Options: cloneOptions(o.Options)}
}

func (o AlterModelManagers) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, Managers: append([]string(nil), o.Managers...)}
}

func (o AlterOrderWithRespectTo) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, FieldName: o.Field}
}

func (o AlterTogether) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{
		Type:           o.Name(),
		AppLabel:       o.AppLabel,
		ModelName:      o.ModelName,
		UniqueTogether: cloneStringMatrix(o.UniqueTogether),
		IndexTogether:  cloneStringMatrix(o.IndexTogether),
	}
}

func (o AddField) MigrationOperationSpec() migrations.OperationSpec {
	field := o.Field
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, Field: &field, HasDefault: o.HasDefault, UnsafeAcknowledged: o.UnsafeAcknowledged}
}

func (o RemoveField) MigrationOperationSpec() migrations.OperationSpec {
	field := o.Field
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, Field: &field}
}

func (o AlterField) MigrationOperationSpec() migrations.OperationSpec {
	oldField := o.OldField
	newField := o.NewField
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, OldField: &oldField, NewField: &newField}
}

func (o RenameField) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, OldName: o.OldName, NewName: o.NewName}
}

func (o AddIndex) MigrationOperationSpec() migrations.OperationSpec {
	index := migrations.IndexState{Name: o.Index.Name, Fields: append([]string(nil), o.Index.Fields...)}
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, Index: &index}
}

func (o RemoveIndex) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, IndexName: o.IndexName}
}

func (o RenameIndex) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, OldName: o.OldName, NewName: o.NewName}
}

func (o AddConstraint) MigrationOperationSpec() migrations.OperationSpec {
	constraint := migrations.ConstraintState{Name: o.Constraint.Name, Type: o.Constraint.Type, Fields: append([]string(nil), o.Constraint.Fields...), Check: o.Constraint.Check}
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, Constraint: &constraint}
}

func (o RemoveConstraint) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), AppLabel: o.AppLabel, ModelName: o.ModelName, ConstraintName: o.ConstraintName}
}

func (o RunSQL) MigrationOperationSpec() migrations.OperationSpec {
	return migrations.OperationSpec{Type: o.Name(), SQL: o.SQL, ReverseSQL: o.ReverseSQL, Elidable: o.ElidableOp}
}

func (o SeparateDatabaseAndState) MigrationOperationSpec() migrations.OperationSpec {
	spec := migrations.OperationSpec{Type: o.Name()}
	for _, operation := range o.DatabaseOperations {
		spec.DatabaseOperations = append(spec.DatabaseOperations, migrations.OperationSpecFor(operation))
	}
	for _, operation := range o.StateOperations {
		spec.StateOperations = append(spec.StateOperations, migrations.OperationSpecFor(operation))
	}
	return spec
}

func cloneModelState(model migrations.ModelState) migrations.ModelState {
	model.Fields = append([]migrations.FieldState(nil), model.Fields...)
	indexes := model.Indexes
	model.Indexes = make([]migrations.IndexState, len(indexes))
	for index, value := range indexes {
		model.Indexes[index] = migrations.IndexState{Name: value.Name, Fields: append([]string(nil), value.Fields...)}
	}
	constraints := model.Constraints
	model.Constraints = make([]migrations.ConstraintState, len(constraints))
	for index, value := range constraints {
		model.Constraints[index] = migrations.ConstraintState{Name: value.Name, Type: value.Type, Fields: append([]string(nil), value.Fields...), Check: value.Check}
	}
	model.Options = cloneOptions(model.Options)
	return model
}

func cloneOptions(options map[string]any) map[string]any {
	if options == nil {
		return nil
	}
	clone := make(map[string]any, len(options))
	for key, value := range options {
		clone[key] = value
	}
	return clone
}

func cloneStringMatrix(values [][]string) [][]string {
	if values == nil {
		return nil
	}
	clone := make([][]string, len(values))
	for index, value := range values {
		clone[index] = append([]string(nil), value...)
	}
	return clone
}
