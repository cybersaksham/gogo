package inspect

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAppManifestJSONRoundTrip(t *testing.T) {
	manifest := AppManifest{
		Name:  "example.blog",
		Label: "blog",
		Path:  "/srv/app/apps/blog",
		Models: []ModelManifest{
			{Name: "Post", Package: "example/blog", Type: "Post"},
		},
		Admin: []AdminManifest{
			{Model: "Post", Type: "PostAdmin"},
		},
		Routes: []RouteManifest{
			{Name: "blog:index", Path: "/blog/", Handler: "Index"},
		},
		Tasks: []TaskManifest{
			{Name: "blog.send_digest", Handler: "SendDigest"},
		},
		Commands: []CommandManifest{
			{Name: "blog.reindex", Handler: "Reindex"},
		},
		Migrations: []MigrationManifest{
			{Name: "0001_initial", Package: "example/blog/migrations"},
		},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got AppManifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(got, manifest) {
		t.Fatalf("round trip = %#v, want %#v", got, manifest)
	}
}

func TestAppManifestSortsResourcesDeterministically(t *testing.T) {
	manifest := AppManifest{
		Models: []ModelManifest{
			{Name: "Tag"},
			{Name: "Post"},
		},
		Admin: []AdminManifest{
			{Model: "Tag"},
			{Model: "Post"},
		},
		Routes: []RouteManifest{
			{Name: "blog:tag"},
			{Name: "blog:index"},
		},
		Tasks: []TaskManifest{
			{Name: "blog.z"},
			{Name: "blog.a"},
		},
		Commands: []CommandManifest{
			{Name: "z"},
			{Name: "a"},
		},
		Migrations: []MigrationManifest{
			{Name: "0002_second"},
			{Name: "0001_initial"},
		},
	}

	manifest.Sort()

	if got := manifest.Models[0].Name; got != "Post" {
		t.Fatalf("first model = %q, want Post", got)
	}
	if got := manifest.Admin[0].Model; got != "Post" {
		t.Fatalf("first admin = %q, want Post", got)
	}
	if got := manifest.Routes[0].Name; got != "blog:index" {
		t.Fatalf("first route = %q, want blog:index", got)
	}
	if got := manifest.Tasks[0].Name; got != "blog.a" {
		t.Fatalf("first task = %q, want blog.a", got)
	}
	if got := manifest.Commands[0].Name; got != "a" {
		t.Fatalf("first command = %q, want a", got)
	}
	if got := manifest.Migrations[0].Name; got != "0001_initial" {
		t.Fatalf("first migration = %q, want 0001_initial", got)
	}
}
