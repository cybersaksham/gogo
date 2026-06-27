package operations

import (
	"context"
	"fmt"
	"strings"

	"github.com/cybersaksham/gogo/migrations"
)

type AddConstraint struct {
	AppLabel, ModelName string
	Constraint          migrations.ConstraintState
}
type RemoveConstraint struct {
	AppLabel, ModelName string
	ConstraintName      string
}

func (o AddConstraint) Name() string { return "AddConstraint" }
func (o AddConstraint) StateForwards(state *migrations.ProjectState) error {
	model := state.Models[key(o.AppLabel, o.ModelName)]
	model.Constraints = append(model.Constraints, o.Constraint)
	state.Models[key(o.AppLabel, o.ModelName)] = model
	return nil
}
func (o AddConstraint) DatabaseForwards(ctx context.Context, editor migrations.SchemaEditor) error {
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", tableName(o.AppLabel, o.ModelName), o.Constraint.Name, constraintSQL(o.Constraint)))
}
func (o AddConstraint) DatabaseBackwards(ctx context.Context, editor migrations.SchemaEditor) error {
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tableName(o.AppLabel, o.ModelName), o.Constraint.Name))
}
func (o AddConstraint) Describe() string { return "Add constraint " + o.Constraint.Name }
func (o AddConstraint) Reversible() bool { return true }
func (o AddConstraint) ReferencesModel(appLabel, modelName string) bool {
	return o.AppLabel == appLabel && o.ModelName == modelName
}
func (o AddConstraint) ReferencesField(appLabel, modelName, fieldName string) bool {
	return o.ReferencesModel(appLabel, modelName) && contains(o.Constraint.Fields, fieldName)
}

func (o RemoveConstraint) Name() string { return "RemoveConstraint" }
func (o RemoveConstraint) StateForwards(state *migrations.ProjectState) error {
	model := state.Models[key(o.AppLabel, o.ModelName)]
	constraints := model.Constraints[:0]
	for _, constraint := range model.Constraints {
		if constraint.Name != o.ConstraintName {
			constraints = append(constraints, constraint)
		}
	}
	model.Constraints = constraints
	state.Models[key(o.AppLabel, o.ModelName)] = model
	return nil
}
func (o RemoveConstraint) DatabaseForwards(ctx context.Context, editor migrations.SchemaEditor) error {
	return editor.Execute(ctx, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", tableName(o.AppLabel, o.ModelName), o.ConstraintName))
}
func (o RemoveConstraint) DatabaseBackwards(context.Context, migrations.SchemaEditor) error {
	return nil
}
func (o RemoveConstraint) Describe() string {
	return "Remove constraint " + o.ConstraintName
}
func (o RemoveConstraint) Reversible() bool { return false }
func (o RemoveConstraint) ReferencesModel(appLabel, modelName string) bool {
	return o.AppLabel == appLabel && o.ModelName == modelName
}
func (o RemoveConstraint) ReferencesField(string, string, string) bool { return false }
func (o RemoveConstraint) SafetyChecks() []migrations.SafetyCheck {
	if strings.HasPrefix(o.ConstraintName, "uniq") {
		return []migrations.SafetyCheck{{Operation: o.Name(), Message: "removes unique constraint " + o.ConstraintName}}
	}
	return nil
}

func constraintSQL(constraint migrations.ConstraintState) string {
	switch constraint.Type {
	case "check":
		return "CHECK (" + constraint.Check + ")"
	case "exclusion":
		return "EXCLUDE (" + strings.Join(constraint.Fields, ", ") + ")"
	default:
		return "UNIQUE (" + strings.Join(constraint.Fields, ", ") + ")"
	}
}
