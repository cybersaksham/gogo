package app

// ModelResource describes a registered model resource.
type ModelResource struct {
	AppLabel string
	Name     string
}

// AdminResource describes an admin registration resource.
type AdminResource struct {
	AppLabel  string
	ModelName string
	Handler   string
}

// RouteResource describes an HTTP route resource.
type RouteResource struct {
	AppLabel string
	Name     string
	Path     string
	Handler  string
}

// APIRouteResource describes an API route resource.
type APIRouteResource struct {
	AppLabel string
	Name     string
	Path     string
	Handler  string
}

// FormResource describes a form resource.
type FormResource struct {
	AppLabel string
	Name     string
	Handler  string
}

// TemplateResource describes an app template root or file.
type TemplateResource struct {
	AppLabel string
	Path     string
}

// StaticResource describes an app static asset root.
type StaticResource struct {
	AppLabel string
	Path     string
}

// TaskResource describes a queue task resource.
type TaskResource struct {
	AppLabel string
	Name     string
	Handler  string
}

// CommandResource describes a management command resource.
type CommandResource struct {
	AppLabel string
	Name     string
	Handler  string
}

// MigrationResource describes a migration resource.
type MigrationResource struct {
	AppLabel string
	Name     string
}

func (r *Registry) RegisterModel(resource ModelResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.models = append(r.models, resource)
}

func (r *Registry) Models() []ModelResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.models)
}

func (r *Registry) RegisterAdmin(resource AdminResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.admin = append(r.admin, resource)
}

func (r *Registry) Admin() []AdminResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.admin)
}

func (r *Registry) RegisterRoute(resource RouteResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, resource)
}

func (r *Registry) Routes() []RouteResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.routes)
}

func (r *Registry) RegisterAPIRoute(resource APIRouteResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.apiRoutes = append(r.apiRoutes, resource)
}

func (r *Registry) APIRoutes() []APIRouteResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.apiRoutes)
}

func (r *Registry) RegisterForm(resource FormResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.forms = append(r.forms, resource)
}

func (r *Registry) Forms() []FormResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.forms)
}

func (r *Registry) RegisterTemplate(resource TemplateResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.templates = append(r.templates, resource)
}

func (r *Registry) Templates() []TemplateResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.templates)
}

func (r *Registry) RegisterStaticRoot(resource StaticResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.staticRoots = append(r.staticRoots, resource)
}

func (r *Registry) StaticRoots() []StaticResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.staticRoots)
}

func (r *Registry) RegisterTask(resource TaskResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = append(r.tasks, resource)
}

func (r *Registry) Tasks() []TaskResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.tasks)
}

func (r *Registry) RegisterCommand(resource CommandResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = append(r.commands, resource)
}

func (r *Registry) Commands() []CommandResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.commands)
}

func (r *Registry) RegisterMigration(resource MigrationResource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.migrations = append(r.migrations, resource)
}

func (r *Registry) Migrations() []MigrationResource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return copySlice(r.migrations)
}

func copySlice[T any](values []T) []T {
	copied := make([]T, len(values))
	copy(copied, values)
	return copied
}
