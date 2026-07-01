package migrations

import "context"

// ManifestOperation is a lightweight operation restored from manifests.
type ManifestOperation struct {
	NameValue string
	Spec      OperationSpec
	SpecJSON  string
}

func (o ManifestOperation) Name() string {
	if spec, ok := o.OperationSpec(); ok && spec.Type != "" {
		return spec.Type
	}
	return o.NameValue
}
func (o ManifestOperation) StateForwards(*ProjectState) error {
	return nil
}
func (o ManifestOperation) DatabaseForwards(context.Context, SchemaEditor) error {
	return nil
}
func (o ManifestOperation) DatabaseBackwards(context.Context, SchemaEditor) error {
	return nil
}
func (o ManifestOperation) Describe() string { return o.Name() }
func (o ManifestOperation) Reversible() bool { return true }
func (o ManifestOperation) ReferencesModel(string, string) bool {
	return false
}
func (o ManifestOperation) ReferencesField(string, string, string) bool {
	return false
}

func (o ManifestOperation) MigrationOperationSpec() OperationSpec {
	if spec, ok := o.OperationSpec(); ok {
		return spec
	}
	return OperationSpec{Type: o.NameValue}
}

func (o ManifestOperation) OperationSpec() (OperationSpec, bool) {
	if o.Spec.Type != "" {
		return o.Spec, true
	}
	if o.SpecJSON == "" {
		return OperationSpec{}, false
	}
	spec, err := OperationSpecFromJSON(o.SpecJSON)
	if err != nil {
		return OperationSpec{}, false
	}
	return spec, spec.Type != ""
}
