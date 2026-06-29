package gogo

import (
	"os"
	"strings"
	"testing"
)

func TestTutorialDocsCoverRequiredFlows(t *testing.T) {
	checkTutorial(t, "docs/code/tutorials/quickstart.md", []string{
		"gogo startproject", "go run manage.go startapp", "models.Metadata", "go run manage.go makemigrations", "go run manage.go migrate", "go run manage.go createsuperuser", "admin.ModelAdmin", "api.NewRouter", "go run manage.go runserver",
	})
	checkTutorial(t, "docs/code/tutorials/blog.md", []string{
		"Author", "Post", "Tag", "Comment", "forms.NewForm", "ListFilter", "PageNumberPagination", "email", "queue.NewSignature",
	})
	checkTutorial(t, "docs/code/tutorials/admin.md", []string{
		"ListDisplay", "SearchFields", "ListFilter", "ReadonlyFields", "Actions", "Inlines", "AutocompleteFields", "HasChangePermission",
	})
	checkTutorial(t, "docs/code/tutorials/tasks.md", []string{
		"MaxRetries", "RetryBackoff", "go run manage.go beat", "Chain", "Group", "Chord", "SoftTimeout", "HardTimeout",
	})
}

func checkTutorial(t *testing.T, path string, required []string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(body)
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("%s does not contain %s", path, want)
		}
	}
}
