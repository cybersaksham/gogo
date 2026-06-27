package fields

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestBinaryFieldCopiesBytesAndRejectsNonBytes(t *testing.T) {
	field := NewBinaryField(Options{Name: "payload"})
	input := []byte("abc")
	value, err := field.ToDB(input)
	if err != nil {
		t.Fatalf("ToDB() error = %v", err)
	}
	input[0] = 'z'
	if string(value.([]byte)) != "abc" {
		t.Fatalf("ToDB() did not copy bytes: %q", value)
	}
	if err := field.Validate("abc"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(string) error = %v, want ErrValidation", err)
	}
}

func TestJSONFieldValidatesMarshalAndScansRawJSON(t *testing.T) {
	field := NewJSONField(Options{Name: "data"})
	if err := field.Validate(map[string]any{"ok": true}); err != nil {
		t.Fatalf("Validate(map) error = %v", err)
	}
	if err := field.Validate(func() {}); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(func) error = %v, want ErrValidation", err)
	}
	value, err := field.FromDB([]byte(`{"ok":true}`))
	if err != nil {
		t.Fatalf("FromDB() error = %v", err)
	}
	if !json.Valid(value.(json.RawMessage)) {
		t.Fatalf("FromDB() = %s, want valid raw JSON", value)
	}
}

func TestGeneratedFieldRejectsManualDatabaseValues(t *testing.T) {
	field := NewGeneratedField(Options{Name: "total"}, "price * quantity")
	if field.Expression() != "price * quantity" {
		t.Fatalf("Expression() = %q", field.Expression())
	}
	if _, err := field.ToDB(10); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("ToDB() error = %v, want ErrInvalidField", err)
	}
}

func TestFileFieldGeneratesSafeUploadPath(t *testing.T) {
	field := NewFileField(Options{Name: "avatar"}, FileConfig{UploadTo: "uploads"})

	path, err := field.GeneratePath("profile.png")
	if err != nil {
		t.Fatalf("GeneratePath() error = %v", err)
	}
	if path != filepath.Join("uploads", "profile.png") {
		t.Fatalf("path = %q, want uploads/profile.png", path)
	}
	if _, err := field.GeneratePath("../secret.txt"); !errors.Is(err, ErrInvalidField) {
		t.Fatalf("GeneratePath(traversal) error = %v, want ErrInvalidField", err)
	}
}

func TestImageFieldUsesMetadataInspector(t *testing.T) {
	field := NewImageField(Options{Name: "image"}, FileConfig{}, func(string) (ImageMetadata, error) {
		return ImageMetadata{Width: 640, Height: 480, Format: "png"}, nil
	})

	meta, err := field.Inspect("image.png")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if meta.Width != 640 || meta.Height != 480 || meta.Format != "png" {
		t.Fatalf("metadata = %#v", meta)
	}
}

func TestFilePathFieldValidatesRootContainment(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "allowed.txt")
	if err := os.WriteFile(file, []byte("ok"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	field := NewFilePathField(Options{Name: "path"}, FilePathConfig{Root: root})
	if err := field.Validate(file); err != nil {
		t.Fatalf("Validate(file) error = %v", err)
	}
	if err := field.Validate(filepath.Join(root, "..", "secret.txt")); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(outside) error = %v, want ErrValidation", err)
	}
}

func TestGenericIPAddressFieldProtocolsAndMappedIPv4(t *testing.T) {
	ipv4 := NewGenericIPAddressField(Options{Name: "ip"}, IPAddressConfig{Protocol: IPv4})
	if err := ipv4.Validate("192.0.2.1"); err != nil {
		t.Fatalf("Validate(ipv4) error = %v", err)
	}
	if err := ipv4.Validate("2001:db8::1"); !errors.Is(err, ErrValidation) {
		t.Fatalf("Validate(ipv6 as ipv4) error = %v, want ErrValidation", err)
	}

	both := NewGenericIPAddressField(Options{Name: "ip"}, IPAddressConfig{Protocol: BothIP, UnpackIPv4Mapped: true})
	mapped := net.IPv4(192, 0, 2, 1).To16().String()
	value, err := both.ToDB(mapped)
	if err != nil {
		t.Fatalf("ToDB(mapped) error = %v", err)
	}
	if value != "192.0.2.1" {
		t.Fatalf("ToDB(mapped) = %q, want 192.0.2.1", value)
	}
}
