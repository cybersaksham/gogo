package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var goIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// NewStartprojectCommand creates the project generator command.
func NewStartprojectCommand() Command {
	return startprojectCommand{}
}

type startprojectCommand struct{}

func (c startprojectCommand) Name() string {
	return "startproject"
}

func (c startprojectCommand) Summary() string {
	return "Create a new Gogo project"
}

func (c startprojectCommand) Run(_ context.Context, args []string) error {
	flags := flag.NewFlagSet("startproject", flag.ContinueOnError)
	force := flags.Bool("force", false, "allow generation into a non-empty directory")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidArguments, err)
	}

	remaining := flags.Args()
	if len(remaining) < 1 || len(remaining) > 2 {
		return fmt.Errorf("%w: usage startproject [--force] <name> [path]", ErrInvalidArguments)
	}

	projectName := remaining[0]
	if !goIdentifierPattern.MatchString(projectName) {
		return fmt.Errorf("%w: project name %q must be a valid Go identifier", ErrInvalidArguments, projectName)
	}

	target := projectName
	if len(remaining) == 2 {
		target = remaining[1]
	}

	if err := ensureProjectTarget(target, *force); err != nil {
		return err
	}

	files := projectFiles(projectName)
	for relativePath, contents := range files {
		path := filepath.Join(target, relativePath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("%w: create directory for %s: %v", ErrCommandFailed, relativePath, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("%w: write %s: %v", ErrCommandFailed, relativePath, err)
		}
	}

	return nil
}

func ensureProjectTarget(target string, force bool) error {
	entries, err := os.ReadDir(target)
	if err == nil {
		if len(entries) > 0 && !force {
			return fmt.Errorf("%w: target directory %s is not empty", ErrCommandFailed, target)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("%w: inspect target directory %s: %v", ErrCommandFailed, target, err)
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("%w: create target directory %s: %v", ErrCommandFailed, target, err)
	}
	return nil
}

func projectFiles(projectName string) map[string]string {
	return map[string]string{
		"go.mod":                             "module " + projectName + "\n\ngo 1.26\n",
		"manage.go":                          manageGo(projectName),
		".gitignore":                         generatedGitignore(),
		".env.example":                       generatedEnvExample(),
		"Makefile":                           generatedMakefile(),
		"README.md":                          "# " + projectName + "\n\nGenerated Gogo project.\n",
		filepath.Join(projectName, "app.go"): packageFile(projectName, "Project bootstrap."),
		filepath.Join(projectName, "settings", "base.go"):       packageFile("settings", "Base settings."),
		filepath.Join(projectName, "settings", "local.go"):      packageFile("settings", "Local settings."),
		filepath.Join(projectName, "settings", "test.go"):       packageFile("settings", "Test settings."),
		filepath.Join(projectName, "settings", "production.go"): packageFile("settings", "Production settings."),
		filepath.Join(projectName, "urls.go"):                   packageFile(projectName, "Root routes."),
		filepath.Join(projectName, "admin.go"):                  packageFile(projectName, "Admin customization."),
		filepath.Join(projectName, "middleware.go"):             packageFile(projectName, "Project middleware."),
		filepath.Join(projectName, "queue.go"):                  packageFile(projectName, "Queue setup."),
		filepath.Join("apps", ".keep"):                          "",
		filepath.Join("templates", "base.html"):                 "<!doctype html>\n<html><body>{{ block \"content\" . }}{{ end }}</body></html>\n",
		filepath.Join("static", ".keep"):                        "",
		filepath.Join("fixtures", ".keep"):                      "",
		filepath.Join("tests", "integration", ".keep"):          "",
		filepath.Join("deploy", "docker", "Dockerfile"):         generatedDockerfile(),
		filepath.Join("deploy", "docker", "docker-compose.yml"): generatedCompose(),
	}
}

func manageGo(projectName string) string {
	return `package main

func main() {
	// Project CLI wiring is generated in the client template phase.
}
`
}

func packageFile(packageName string, comment string) string {
	return fmt.Sprintf("// %s\npackage %s\n", comment, packageName)
}

func generatedGitignore() string {
	return strings.TrimLeft(`
# Go build output
bin/
dist/
*.test
*.out

# Coverage
coverage/
coverage.out

# Local environment
.env
.env.*.local

# Local databases
*.sqlite
*.sqlite3
*.db

# Uploads and generated media
media/
uploads/

# Editor files
.idea/
.vscode/
*.swp
`, "\n")
}

func generatedEnvExample() string {
	return strings.TrimLeft(`
# Framework
GOGO_ENV=development
GOGO_SECRET_KEY=
GOGO_DEBUG=
GOGO_INSTALLED_APPS=
GOGO_MIDDLEWARE=
GOGO_ROOT_URLCONF=
GOGO_DEFAULT_AUTO_FIELD=BigAutoField
GOGO_TIME_ZONE=UTC
GOGO_LANGUAGE_CODE=en-us

# Database
DATABASE_URL=

# Server
GOGO_HTTP_ADDR=:8000

# Static and media files
GOGO_STATIC_URL=/static/
GOGO_STATIC_ROOT=
GOGO_MEDIA_URL=/media/
GOGO_MEDIA_ROOT=
GOGO_TEMPLATE_DIRS=

# Queue
GOGO_BROKER_URL=
GOGO_RESULT_BACKEND=

# Cache
GOGO_CACHE_URL=

# Email
GOGO_EMAIL_URL=

# Sessions and CSRF
GOGO_SESSION_COOKIE_NAME=gogo_sessionid
GOGO_CSRF_COOKIE_NAME=gogo_csrftoken

# Security
GOGO_ALLOWED_HOSTS=localhost,127.0.0.1
`, "\n")
}

func generatedMakefile() string {
	return strings.TrimLeft(`
.PHONY: test run check

test:
	go test ./...

run:
	go run manage.go runserver

check:
	go run manage.go check
`, "\n")
}

func generatedDockerfile() string {
	return strings.TrimLeft(`
FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/app ./manage.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/app
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
`, "\n")
}

func generatedCompose() string {
	return strings.TrimLeft(`
services:
  app:
    build:
      context: ../..
      dockerfile: deploy/docker/Dockerfile
    env_file:
      - ../../.env
    depends_on:
      - db
      - redis

  db:
    image: postgres:17
    environment:
      POSTGRES_DB: gogo
      POSTGRES_USER: gogo
      POSTGRES_PASSWORD: gogo
    volumes:
      - postgres-data:/var/lib/postgresql/data

  redis:
    image: redis:8
    volumes:
      - redis-data:/data

volumes:
  postgres-data:
  redis-data:
`, "\n")
}
