package api

import (
	"reflect"
	"testing"
	"time"
)

func TestSerializerFieldsParsePrimitiveRelationshipAndFileTypes(t *testing.T) {
	image := UploadedFile{Filename: "avatar.png", Content: []byte("png")}
	fields := []SerializerField{
		BooleanField("active", FieldOptions{}),
		IntegerField("count", FieldOptions{}),
		FloatField("ratio", FieldOptions{}),
		DecimalField("price", FieldOptions{}),
		StringField("title", FieldOptions{}),
		EmailField("email", FieldOptions{}),
		URLField("site", FieldOptions{}),
		SlugField("slug", FieldOptions{}),
		UUIDField("uid", FieldOptions{}),
		DateField("date", FieldOptions{}),
		DateTimeField("created_at", FieldOptions{}),
		TimeField("clock", FieldOptions{}),
		DurationField("duration", FieldOptions{}),
		ChoiceField("status", FieldOptions{}, []string{"draft", "published"}),
		MultipleChoiceField("tags", FieldOptions{}, []string{"go", "api", "admin"}),
		JSONField("payload", FieldOptions{}),
		ListField("items", FieldOptions{}, StringField("item", FieldOptions{})),
		DictField("meta", FieldOptions{}),
		NestedObjectField("profile", FieldOptions{}, NewSerializer(StringField("name", FieldOptions{}))),
		PrimaryKeyRelatedField("author", FieldOptions{}),
		SlugRelatedField("category", FieldOptions{}),
		HyperlinkedRelatedField("url", FieldOptions{}),
		FileField("file", FieldOptions{}),
		ImageField("image", FieldOptions{}),
	}
	input := map[string]any{
		"active": true, "count": "42", "ratio": "3.5", "price": "10.25", "title": "Gogo",
		"email": "dev@example.com", "site": "https://example.com", "slug": "gogo-api",
		"uid": "550e8400-e29b-41d4-a716-446655440000", "date": "2026-06-28",
		"created_at": "2026-06-28T12:30:00Z", "clock": "12:30:00", "duration": "2h",
		"status": "draft", "tags": []any{"go", "api"}, "payload": map[string]any{"ok": true},
		"items": []any{"a", "b"}, "meta": map[string]any{"page": "1"}, "profile": map[string]any{"name": "Saksham"},
		"author": "7", "category": "framework", "url": "/api/categories/framework/", "file": UploadedFile{Filename: "doc.txt"}, "image": image,
	}
	serializer := NewSerializer(fields...)
	validated, fieldErrors, ok := serializer.Validate(input)
	if !ok {
		t.Fatalf("Validate() errors = %#v", fieldErrors)
	}

	if validated["count"] != int64(42) || validated["price"] != "10.25" || validated["author"] != int64(7) {
		t.Fatalf("validated primitives = %#v", validated)
	}
	if validated["duration"] != 2*time.Hour || validated["image"].(UploadedFile).Filename != "avatar.png" {
		t.Fatalf("validated duration/image = %#v / %#v", validated["duration"], validated["image"])
	}
	if !reflect.DeepEqual(validated["profile"], map[string]any{"name": "Saksham"}) {
		t.Fatalf("nested profile = %#v", validated["profile"])
	}
}

func TestSerializerSupportsDefaultsSourceReadWriteOnlyAndMethodFields(t *testing.T) {
	serializer := NewSerializer(
		StringField("title", FieldOptions{Source: "headline"}),
		StringField("state", FieldOptions{Default: "draft"}),
		StringField("secret", FieldOptions{WriteOnly: true}),
		StringField("id", FieldOptions{ReadOnly: true}),
		MethodField("display", func(obj map[string]any) any { return obj["headline"].(string) + "!" }),
	)
	validated, errors, ok := serializer.Validate(map[string]any{"title": "Gogo", "secret": "token"})
	if !ok {
		t.Fatalf("Validate() errors = %#v", errors)
	}
	if !reflect.DeepEqual(validated, map[string]any{"headline": "Gogo", "state": "draft", "secret": "token"}) {
		t.Fatalf("validated = %#v", validated)
	}

	rendered := serializer.Render(map[string]any{"id": 1, "headline": "Gogo", "secret": "token"})
	if !reflect.DeepEqual(rendered, map[string]any{"display": "Gogo!", "id": 1, "state": nil, "title": "Gogo"}) {
		t.Fatalf("rendered = %#v", rendered)
	}
}

func TestSerializerValidationErrorsAreNestedAndDeterministic(t *testing.T) {
	serializer := NewSerializer(
		StringField("title", FieldOptions{Required: true, AllowBlank: false}),
		EmailField("email", FieldOptions{}),
		NestedObjectField("profile", FieldOptions{}, NewSerializer(StringField("name", FieldOptions{Required: true}))),
	)
	_, errors, ok := serializer.Validate(map[string]any{"title": "", "email": "bad", "profile": map[string]any{}})
	if ok {
		t.Fatalf("Validate() ok = true, want errors")
	}
	want := map[string][]string{
		"email":        {"invalid email"},
		"profile.name": {"required"},
		"title":        {"blank"},
	}
	if !reflect.DeepEqual(errors, want) {
		t.Fatalf("errors = %#v, want %#v", errors, want)
	}
}
