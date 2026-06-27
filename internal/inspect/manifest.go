package inspect

import "sort"

// AppManifest stores generated discovery metadata for one app.
type AppManifest struct {
	Name       string              `json:"name"`
	Label      string              `json:"label"`
	Path       string              `json:"path"`
	Models     []ModelManifest     `json:"models,omitempty"`
	Admin      []AdminManifest     `json:"admin,omitempty"`
	Routes     []RouteManifest     `json:"routes,omitempty"`
	Tasks      []TaskManifest      `json:"tasks,omitempty"`
	Commands   []CommandManifest   `json:"commands,omitempty"`
	Migrations []MigrationManifest `json:"migrations,omitempty"`
}

// ModelManifest describes an app-owned model.
type ModelManifest struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	Type    string `json:"type"`
}

// AdminManifest describes an admin registration.
type AdminManifest struct {
	Model string `json:"model"`
	Type  string `json:"type"`
}

// RouteManifest describes a route registration.
type RouteManifest struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

// TaskManifest describes a queue task registration.
type TaskManifest struct {
	Name    string `json:"name"`
	Handler string `json:"handler"`
}

// CommandManifest describes a management command registration.
type CommandManifest struct {
	Name    string `json:"name"`
	Handler string `json:"handler"`
}

// MigrationManifest describes a migration package entry.
type MigrationManifest struct {
	Name    string `json:"name"`
	Package string `json:"package"`
}

// Sort makes manifest resource order deterministic.
func (m *AppManifest) Sort() {
	sort.Slice(m.Models, func(i, j int) bool {
		return m.Models[i].Name < m.Models[j].Name
	})
	sort.Slice(m.Admin, func(i, j int) bool {
		return m.Admin[i].Model < m.Admin[j].Model
	})
	sort.Slice(m.Routes, func(i, j int) bool {
		return m.Routes[i].Name < m.Routes[j].Name
	})
	sort.Slice(m.Tasks, func(i, j int) bool {
		return m.Tasks[i].Name < m.Tasks[j].Name
	})
	sort.Slice(m.Commands, func(i, j int) bool {
		return m.Commands[i].Name < m.Commands[j].Name
	})
	sort.Slice(m.Migrations, func(i, j int) bool {
		return m.Migrations[i].Name < m.Migrations[j].Name
	})
}
