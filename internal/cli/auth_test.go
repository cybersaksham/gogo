package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/cybersaksham/gogo/auth"
)

func TestCreateSuperuserNonInteractiveCreatesUser(t *testing.T) {
	store, _ := auth.NewMemoryUserStore()
	command := NewCreateSuperuserCommand(store)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := command.(interface {
		runWithIO(context.Context, []string, io.Writer, io.Writer) error
	}).runWithIO(context.Background(), []string{
		"--username", "Admin",
		"--email", "Admin@Example.COM",
		"--password", "CorrectHorseBatteryStaple42",
		"--database", "default",
		"--noinput",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("createsuperuser error = %v", err)
	}

	user, ok, err := store.FindByUsername(context.Background(), "admin")
	if err != nil || !ok {
		t.Fatalf("FindByUsername(admin) = %#v, %v, %v", user, ok, err)
	}
	if !user.IsSuperuser || !user.IsStaff || !user.IsActive || user.Email != "admin@example.com" {
		t.Fatalf("created user = %#v", user)
	}
	if ok, _ := auth.CheckPassword("CorrectHorseBatteryStaple42", user.Password); !ok {
		t.Fatalf("created password does not verify")
	}
	if !strings.Contains(stdout.String(), "created superuser admin on database default") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestCreateSuperuserRejectsDuplicateInvalidPasswordAndUnsafePasswordFlag(t *testing.T) {
	store, _ := auth.NewMemoryUserStore()
	command := NewCreateSuperuserCommand(store)
	run := func(args ...string) error {
		return command.(interface {
			runWithIO(context.Context, []string, io.Writer, io.Writer) error
		}).runWithIO(context.Background(), args, io.Discard, io.Discard)
	}

	if err := run("--username", "admin", "--password", "CorrectHorseBatteryStaple42", "--noinput"); err != nil {
		t.Fatalf("first createsuperuser error = %v", err)
	}
	if err := run("--username", "admin", "--password", "AnotherStrongPassword42", "--noinput"); !errors.Is(err, auth.ErrDuplicateUser) {
		t.Fatalf("duplicate error = %v, want ErrDuplicateUser", err)
	}
	if err := run("--username", "weak", "--password", "password", "--noinput"); !errors.Is(err, auth.ErrPasswordValidation) {
		t.Fatalf("invalid password error = %v, want ErrPasswordValidation", err)
	}
	if err := run("--username", "unsafe", "--password", "CorrectHorseBatteryStaple42"); !errors.Is(err, ErrCommandFailed) {
		t.Fatalf("unsafe password flag error = %v, want ErrCommandFailed", err)
	}
}

func TestChangePasswordUpdatesExistingUser(t *testing.T) {
	hash, err := auth.EncodePBKDF2PasswordWithIterations("old-secret", "salt", 1)
	if err != nil {
		t.Fatalf("EncodePBKDF2PasswordWithIterations() error = %v", err)
	}
	store, _ := auth.NewMemoryUserStore(auth.User{AbstractUser: auth.AbstractUser{
		AbstractBaseUser: auth.AbstractBaseUser{ID: 1, Password: hash, IsActive: true},
		Username:         "admin",
	}})
	command := NewChangePasswordCommand(store)
	var stdout bytes.Buffer

	err = command.(interface {
		runWithIO(context.Context, []string, io.Writer, io.Writer) error
	}).runWithIO(context.Background(), []string{
		"--username", "admin",
		"--password", "CorrectHorseBatteryStaple42",
		"--database", "default",
		"--noinput",
	}, &stdout, io.Discard)
	if err != nil {
		t.Fatalf("changepassword error = %v", err)
	}

	user, ok, err := store.FindByUsername(context.Background(), "admin")
	if err != nil || !ok {
		t.Fatalf("FindByUsername(admin) = %#v, %v, %v", user, ok, err)
	}
	if ok, _ := auth.CheckPassword("CorrectHorseBatteryStaple42", user.Password); !ok {
		t.Fatalf("changed password does not verify")
	}
	if !strings.Contains(stdout.String(), "changed password for admin on database default") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
