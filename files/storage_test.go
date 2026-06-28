package files

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestLocalStorageEveryMethod(t *testing.T) {
	ctx := context.Background()
	storage := NewLocalStorage(t.TempDir(), LocalOptions{BaseURL: "/media/"})

	name, err := storage.Save(ctx, "docs/report.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if name != "docs/report.txt" {
		t.Fatalf("Save() name = %q", name)
	}

	exists, err := storage.Exists(ctx, name)
	if err != nil || !exists {
		t.Fatalf("Exists() = %v, %v", exists, err)
	}
	reader, err := storage.Open(ctx, name)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	content, err := io.ReadAll(reader)
	reader.Close()
	if err != nil || string(content) != "hello" {
		t.Fatalf("ReadAll() = %q, %v", content, err)
	}
	size, err := storage.Size(ctx, name)
	if err != nil || size != 5 {
		t.Fatalf("Size() = %d, %v", size, err)
	}
	modified, err := storage.ModifiedTime(ctx, name)
	if err != nil || modified.IsZero() {
		t.Fatalf("ModifiedTime() = %v, %v", modified, err)
	}
	path, err := storage.Path(name)
	if err != nil || !strings.HasSuffix(path, "docs/report.txt") {
		t.Fatalf("Path() = %q, %v", path, err)
	}
	url, err := storage.URL(name)
	if err != nil || url != "/media/docs/report.txt" {
		t.Fatalf("URL() = %q, %v", url, err)
	}
	listed, err := storage.List(ctx, "docs")
	if err != nil || len(listed) != 1 || listed[0] != "docs/report.txt" {
		t.Fatalf("List() = %#v, %v", listed, err)
	}
	if err := storage.Delete(ctx, name); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	exists, err = storage.Exists(ctx, name)
	if err != nil || exists {
		t.Fatalf("Exists(deleted) = %v, %v", exists, err)
	}
}

func TestLocalStorageRejectsTraversalAndAvoidsOverwrite(t *testing.T) {
	ctx := context.Background()
	storage := NewLocalStorage(t.TempDir(), LocalOptions{})

	if _, err := storage.Save(ctx, "../bad.txt", strings.NewReader("bad")); !errors.Is(err, ErrUnsafeName) {
		t.Fatalf("Save(traversal) error = %v, want ErrUnsafeName", err)
	}
	first, err := storage.Save(ctx, "same.txt", strings.NewReader("first"))
	if err != nil {
		t.Fatalf("first Save() error = %v", err)
	}
	second, err := storage.Save(ctx, "same.txt", strings.NewReader("second"))
	if err != nil {
		t.Fatalf("second Save() error = %v", err)
	}
	if first != "same.txt" || second != "same_1.txt" {
		t.Fatalf("collision names = %q, %q", first, second)
	}
	reader, err := storage.Open(ctx, first)
	if err != nil {
		t.Fatalf("Open(first) error = %v", err)
	}
	content, _ := io.ReadAll(reader)
	reader.Close()
	if string(content) != "first" {
		t.Fatalf("first content overwritten: %q", content)
	}
}
