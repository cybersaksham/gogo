package migrations

import (
	"fmt"
	"sort"
)

// Graph stores migration dependency nodes.
type Graph struct {
	nodes map[string]Migration
}

// NewGraph creates an empty migration graph.
func NewGraph() *Graph {
	return &Graph{nodes: make(map[string]Migration)}
}

// Add inserts a migration and validates graph consistency.
func (g *Graph) Add(migration Migration) error {
	key := migration.Identity()
	if _, exists := g.nodes[key]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateMigration, key)
	}
	g.nodes[key] = migration
	if err := g.validateDependencies(migration); err != nil {
		return err
	}
	return g.detectCycles()
}

// ForwardsPlan returns dependencies before the target migration.
func (g *Graph) ForwardsPlan(target Dependency) ([]Migration, error) {
	seen := map[string]bool{}
	var plan []Migration
	var visit func(string) error
	visit = func(key string) error {
		if seen[key] {
			return nil
		}
		migration, ok := g.nodes[key]
		if !ok {
			return fmt.Errorf("%w: %s", ErrMissingDependency, key)
		}
		deps := append([]Dependency(nil), migration.Dependencies...)
		sort.Slice(deps, func(i, j int) bool { return deps[i].Identity() < deps[j].Identity() })
		for _, dependency := range deps {
			if err := visit(dependency.Identity()); err != nil {
				return err
			}
		}
		seen[key] = true
		plan = append(plan, migration)
		return nil
	}
	if err := visit(target.Identity()); err != nil {
		return nil, err
	}
	return plan, nil
}

// BackwardsPlan returns migrations to unapply for the target.
func (g *Graph) BackwardsPlan(target Dependency) ([]Migration, error) {
	migration, ok := g.nodes[target.Identity()]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingDependency, target.Identity())
	}
	return []Migration{migration}, nil
}

// ConflictingLeaves returns apps with multiple leaf migrations.
func (g *Graph) ConflictingLeaves() map[string][]Migration {
	dependents := map[string]int{}
	for _, migration := range g.nodes {
		for _, dependency := range migration.Dependencies {
			dependents[dependency.Identity()]++
		}
	}
	leaves := map[string][]Migration{}
	for key, migration := range g.nodes {
		if dependents[key] == 0 {
			leaves[migration.AppLabel] = append(leaves[migration.AppLabel], migration)
		}
	}
	for app, appLeaves := range leaves {
		if len(appLeaves) < 2 {
			delete(leaves, app)
			continue
		}
		sort.Slice(appLeaves, func(i, j int) bool { return appLeaves[i].Name < appLeaves[j].Name })
		leaves[app] = appLeaves
	}
	return leaves
}

func (g *Graph) validateDependencies(migration Migration) error {
	for _, dependency := range migration.Dependencies {
		if _, ok := g.nodes[dependency.Identity()]; !ok {
			return fmt.Errorf("%w: %s", ErrMissingDependency, dependency.Identity())
		}
	}
	return nil
}

func (g *Graph) detectCycles() error {
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string) error
	visit = func(key string) error {
		if visiting[key] {
			return fmt.Errorf("%w: %s", ErrMigrationCycle, key)
		}
		if visited[key] {
			return nil
		}
		visiting[key] = true
		for _, dependency := range g.nodes[key].Dependencies {
			if _, ok := g.nodes[dependency.Identity()]; ok {
				if err := visit(dependency.Identity()); err != nil {
					return err
				}
			}
		}
		visiting[key] = false
		visited[key] = true
		return nil
	}
	for key := range g.nodes {
		if err := visit(key); err != nil {
			return err
		}
	}
	return nil
}
