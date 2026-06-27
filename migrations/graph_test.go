package migrations

import (
	"errors"
	"reflect"
	"testing"
)

func TestMigrationGraphPlans(t *testing.T) {
	graph := NewGraph()
	m1 := testMigration("blog", "0001_initial")
	m2 := testMigration("blog", "0002_post")
	m2.Dependencies = []Dependency{{AppLabel: "blog", Name: "0001_initial"}}
	auth := testMigration("auth", "0001_initial")
	m2.Dependencies = append(m2.Dependencies, Dependency{AppLabel: "auth", Name: "0001_initial"})
	for _, migration := range []Migration{m1, auth, m2} {
		if err := graph.Add(migration); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}
	plan, err := graph.ForwardsPlan(Dependency{AppLabel: "blog", Name: "0002_post"})
	if err != nil {
		t.Fatalf("ForwardsPlan() error = %v", err)
	}
	got := identities(plan)
	want := []string{"auth.0001_initial", "blog.0001_initial", "blog.0002_post"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("plan = %#v, want %#v", got, want)
	}
	backwards, err := graph.BackwardsPlan(Dependency{AppLabel: "blog", Name: "0002_post"})
	if err != nil {
		t.Fatalf("BackwardsPlan() error = %v", err)
	}
	if !reflect.DeepEqual(identities(backwards), []string{"blog.0002_post"}) {
		t.Fatalf("backwards = %#v", identities(backwards))
	}
}

func TestMigrationGraphFailuresAndReplacements(t *testing.T) {
	graph := NewGraph()
	if err := graph.Add(testMigration("blog", "0001_initial")); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := graph.Add(testMigration("blog", "0001_initial")); !errors.Is(err, ErrDuplicateMigration) {
		t.Fatalf("duplicate error = %v", err)
	}

	missing := testMigration("blog", "0002_missing")
	missing.Dependencies = []Dependency{{AppLabel: "blog", Name: "0009_missing"}}
	if err := NewGraph().Add(missing); !errors.Is(err, ErrMissingDependency) {
		t.Fatalf("missing dependency error = %v", err)
	}

	cyclic := NewGraph()
	a := testMigration("blog", "0001_a")
	b := testMigration("blog", "0002_b")
	a.Dependencies = []Dependency{{AppLabel: "blog", Name: "0002_b"}}
	b.Dependencies = []Dependency{{AppLabel: "blog", Name: "0001_a"}}
	_ = cyclic.Add(a)
	if err := cyclic.Add(b); !errors.Is(err, ErrMigrationCycle) {
		t.Fatalf("cycle error = %v", err)
	}

	conflict := NewGraph()
	_ = conflict.Add(testMigration("blog", "0001_initial"))
	leafA := testMigration("blog", "0002_a")
	leafA.Dependencies = []Dependency{{AppLabel: "blog", Name: "0001_initial"}}
	leafB := testMigration("blog", "0002_b")
	leafB.Dependencies = []Dependency{{AppLabel: "blog", Name: "0001_initial"}}
	_ = conflict.Add(leafA)
	_ = conflict.Add(leafB)
	if conflicts := conflict.ConflictingLeaves(); len(conflicts["blog"]) != 2 {
		t.Fatalf("conflicts = %#v", conflicts)
	}

	squashed := testMigration("blog", "0003_squashed")
	squashed.Replaces = []Dependency{{AppLabel: "blog", Name: "0002_a"}, {AppLabel: "blog", Name: "0002_b"}}
	if !squashed.ReplacesMigration(Dependency{AppLabel: "blog", Name: "0002_a"}) {
		t.Fatalf("replacement metadata missing")
	}
}

func testMigration(app, name string) Migration {
	return Migration{AppLabel: app, Name: name, Operations: []Operation{NoopOperation{NameValue: "op"}}}
}

func identities(migrations []Migration) []string {
	values := make([]string, len(migrations))
	for i, migration := range migrations {
		values[i] = migration.Identity()
	}
	return values
}
