package auth

import (
	"context"
	"testing"
	"time"
)

func TestAuthenticateSupportsUsernameAndEmailLogin(t *testing.T) {
	password, err := MakePassword("secret")
	if err != nil {
		t.Fatalf("MakePassword() error = %v", err)
	}
	store, err := NewMemoryUserStore(User{AbstractUser: AbstractUser{
		AbstractBaseUser: AbstractBaseUser{ID: 1, Password: password, IsActive: true},
		Username:         "Saksham",
		Email:            "SAKSHAM@Example.COM",
	}})
	if err != nil {
		t.Fatalf("NewMemoryUserStore() error = %v", err)
	}
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)

	byUsername, ok, err := Authenticate(context.Background(), store, Credentials{
		Username: "  saksham ",
		Password: "secret",
		Now:      func() time.Time { return now },
	})
	if err != nil || !ok {
		t.Fatalf("Authenticate(username) = %#v, %v, %v", byUsername, ok, err)
	}
	if byUsername.ID != 1 || !byUsername.LastLogin.Equal(now) {
		t.Fatalf("username result = %#v", byUsername)
	}

	byEmail, ok, err := Authenticate(context.Background(), store, Credentials{
		Email:    "saksham@example.com",
		Password: "secret",
		Now:      func() time.Time { return now.Add(time.Minute) },
	})
	if err != nil || !ok {
		t.Fatalf("Authenticate(email) = %#v, %v, %v", byEmail, ok, err)
	}
	if byEmail.ID != 1 || !byEmail.LastLogin.Equal(now.Add(time.Minute)) {
		t.Fatalf("email result = %#v", byEmail)
	}
}

func TestAuthenticateRejectsInvalidPasswordInactiveAndMissingUsers(t *testing.T) {
	password, err := MakePassword("secret")
	if err != nil {
		t.Fatalf("MakePassword() error = %v", err)
	}
	store, err := NewMemoryUserStore(
		User{AbstractUser: AbstractUser{AbstractBaseUser: AbstractBaseUser{ID: 1, Password: password, IsActive: true}, Username: "active", Email: "active@example.com"}},
		User{AbstractUser: AbstractUser{AbstractBaseUser: AbstractBaseUser{ID: 2, Password: password, IsActive: false}, Username: "inactive", Email: "inactive@example.com"}},
	)
	if err != nil {
		t.Fatalf("NewMemoryUserStore() error = %v", err)
	}

	if user, ok, err := Authenticate(context.Background(), store, Credentials{Username: "active", Password: "wrong"}); err != nil || ok || user.ID != 0 {
		t.Fatalf("Authenticate(wrong password) = %#v, %v, %v", user, ok, err)
	}
	if user, ok, err := Authenticate(context.Background(), store, Credentials{Username: "inactive", Password: "secret"}); err != nil || ok || user.ID != 0 {
		t.Fatalf("Authenticate(inactive) = %#v, %v, %v", user, ok, err)
	}
	if user, ok, err := Authenticate(context.Background(), store, Credentials{Username: "missing", Password: "secret"}); err != nil || ok || user.ID != 0 {
		t.Fatalf("Authenticate(missing) = %#v, %v, %v", user, ok, err)
	}
}

func TestNormalizeUsernameAndEmail(t *testing.T) {
	if got := NormalizeUsername("  Saksham  "); got != "saksham" {
		t.Fatalf("NormalizeUsername() = %q", got)
	}
	if got := NormalizeEmail("  Saksham@Example.COM  "); got != "saksham@example.com" {
		t.Fatalf("NormalizeEmail() = %q", got)
	}
}
