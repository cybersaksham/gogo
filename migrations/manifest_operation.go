package migrations

import "context"

// ManifestOperation is a lightweight operation restored from manifests.
type ManifestOperation struct {
	NameValue string
}

func (o ManifestOperation) Name() string { return o.NameValue }
func (o ManifestOperation) StateForwards(*ProjectState) error {
	return nil
}
func (o ManifestOperation) DatabaseForwards(context.Context, SchemaEditor) error {
	return nil
}
func (o ManifestOperation) DatabaseBackwards(context.Context, SchemaEditor) error {
	return nil
}
func (o ManifestOperation) Describe() string { return o.NameValue }
func (o ManifestOperation) Reversible() bool { return true }
func (o ManifestOperation) ReferencesModel(string, string) bool {
	return false
}
func (o ManifestOperation) ReferencesField(string, string, string) bool {
	return false
}
