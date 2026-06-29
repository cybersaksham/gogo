package app

import (
	"fmt"
	"sync"
)

// Registry stores installed app state.
type Registry struct {
	lifecycleMu sync.Mutex
	mu          sync.RWMutex
	apps        []Config
	ordered     []Config
	byName      map[string]Config
	byLabel     map[string]Config
	models      []ModelResource
	admin       []AdminResource
	routes      []RouteResource
	apiRoutes   []APIRouteResource
	forms       []FormResource
	templates   []TemplateResource
	staticRoots []StaticResource
	tasks       []TaskResource
	commands    []CommandResource
	mgmtCommand []ManagementCommand
	mgmtByName  map[string]ManagementCommand
	migrations  []MigrationResource
	preparing   bool
	ready       bool
}

// NewRegistry creates an empty app registry.
func NewRegistry() *Registry {
	return &Registry{
		byName:     make(map[string]Config),
		byLabel:    make(map[string]Config),
		mgmtByName: make(map[string]ManagementCommand),
	}
}

// Register adds an app config to the registry.
func (r *Registry) Register(config Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ready || r.preparing {
		return fmt.Errorf("%w: cannot register %s after Ready", ErrRegistryReady, config.Name())
	}
	if err := ValidateConfig(config); err != nil {
		return err
	}
	if _, exists := r.byName[config.Name()]; exists {
		return fmt.Errorf("%w: name %q", ErrDuplicateApp, config.Name())
	}
	if _, exists := r.byLabel[config.Label()]; exists {
		return fmt.Errorf("%w: label %q", ErrDuplicateApp, config.Label())
	}

	r.apps = append(r.apps, config)
	r.byName[config.Name()] = config
	r.byLabel[config.Label()] = config
	return nil
}

// Apps returns registered apps in registration order before readiness and
// dependency order after readiness.
func (r *Registry) Apps() []Config {
	r.mu.RLock()
	defer r.mu.RUnlock()

	source := r.apps
	if r.ready {
		source = r.ordered
	}

	apps := make([]Config, len(source))
	copy(apps, source)
	return apps
}

// Get looks up an app by full name or label.
func (r *Registry) Get(nameOrLabel string) (Config, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if config, ok := r.byName[nameOrLabel]; ok {
		return config, true
	}
	config, ok := r.byLabel[nameOrLabel]
	return config, ok
}

// MustGet looks up an app by full name or label and panics if it is missing.
func (r *Registry) MustGet(nameOrLabel string) Config {
	config, ok := r.Get(nameOrLabel)
	if !ok {
		panic(fmt.Sprintf("app %q is not registered", nameOrLabel))
	}
	return config
}

// Labels returns app labels in registration order.
func (r *Registry) Labels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	labels := make([]string, 0, len(r.apps))
	for _, config := range r.apps {
		labels = append(labels, config.Label())
	}
	return labels
}

func (r *Registry) resolveDependencyOrder() ([]Config, error) {
	visiting := make(map[string]bool, len(r.apps))
	visited := make(map[string]bool, len(r.apps))
	ordered := make([]Config, 0, len(r.apps))

	var visit func(Config) error
	visit = func(config Config) error {
		name := config.Name()
		if visited[name] {
			return nil
		}
		if visiting[name] {
			return fmt.Errorf("%w: %s", ErrDependencyCycle, name)
		}

		visiting[name] = true
		for _, dependencyName := range config.Dependencies() {
			dependency, exists := r.byName[dependencyName]
			if !exists {
				return fmt.Errorf("%w: %s depends on %s", ErrMissingDependency, name, dependencyName)
			}
			if err := visit(dependency); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		ordered = append(ordered, config)
		return nil
	}

	for _, config := range r.apps {
		if err := visit(config); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}
