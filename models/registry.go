package models

import (
	"fmt"
	"sync"
)

// Registry stores model metadata by app label and model name.
type Registry struct {
	mu     sync.RWMutex
	models []Metadata
	byKey  map[string]Metadata
}

// NewRegistry creates an empty model registry.
func NewRegistry() *Registry {
	return &Registry{byKey: make(map[string]Metadata)}
}

// Register resolves and registers model metadata.
func (r *Registry) Register(model Model) error {
	return r.RegisterMetadata(ResolveMetadata(model))
}

// RegisterMetadata registers already-resolved metadata.
func (r *Registry) RegisterMetadata(meta Metadata) error {
	resolved := meta.Clone()
	if err := ValidateMetadata(resolved); err != nil {
		return err
	}
	if resolved.AppLabel == "" || resolved.ModelName == "" {
		return fmt.Errorf("%w: app label and model name are required", ErrInvalidMetadata)
	}

	key := modelKey(resolved.AppLabel, resolved.ModelName)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byKey[key]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateModel, key)
	}

	r.models = append(r.models, resolved.Clone())
	r.byKey[key] = resolved.Clone()
	return nil
}

// Lookup returns metadata by app_label.ModelName.
func (r *Registry) Lookup(label string) (Metadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, ok := r.byKey[label]
	if !ok {
		return Metadata{}, false
	}
	return meta.Clone(), true
}

// Models returns registered model metadata in registration order.
func (r *Registry) Models() []Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return cloneMetadataSlice(r.models)
}

// AppMetadata returns metadata for app registry integration.
func (r *Registry) AppMetadata() []Metadata {
	return r.Models()
}

// MigrationMetadata returns metadata for migration generation.
func (r *Registry) MigrationMetadata() []Metadata {
	return r.Models()
}

// ORMMetadata returns metadata for the ORM.
func (r *Registry) ORMMetadata() []Metadata {
	return r.Models()
}

// AdminMetadata returns metadata for the admin.
func (r *Registry) AdminMetadata() []Metadata {
	return r.Models()
}

// SerializerMetadata returns metadata for serializers and APIs.
func (r *Registry) SerializerMetadata() []Metadata {
	return r.Models()
}

// ContentTypeMetadata returns metadata for content type creation.
func (r *Registry) ContentTypeMetadata() []Metadata {
	return r.Models()
}

func modelKey(appLabel, modelName string) string {
	return appLabel + "." + modelName
}

func cloneMetadataSlice(values []Metadata) []Metadata {
	copied := make([]Metadata, len(values))
	for i, value := range values {
		copied[i] = value.Clone()
	}
	return copied
}
