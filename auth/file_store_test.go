package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileUserStorePersistsUsersAndUpdatesLogin(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".gogo", "auth_users.json")
	store := NewFileUserStore(path)
	user := User{AbstractUser: AbstractUser{
		AbstractBaseUser: AbstractBaseUser{IsActive: true},
		Username:         "Admin",
		Email:            "ADMIN@example.com",
		IsStaff:          true,
	}}
	if err := store.Add(user); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("store file mode = %v, %v", info, err)
	}

	reopened := NewFileUserStore(path)
	got, ok, err := reopened.FindByUsername(context.Background(), "admin")
	if err != nil || !ok || got.ID == 0 || got.Username != "admin" || got.Email != "admin@example.com" {
		t.Fatalf("FindByUsername() = %#v, %v, %v", got, ok, err)
	}
	if _, ok, err := reopened.FindByEmail(context.Background(), "admin@example.com"); err != nil || !ok {
		t.Fatalf("FindByEmail() ok=%v err=%v", ok, err)
	}
	at := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	if err := reopened.UpdateLastLogin(context.Background(), got.ID, at); err != nil {
		t.Fatalf("UpdateLastLogin() error = %v", err)
	}
	updated, ok, err := reopened.FindByID(context.Background(), got.ID)
	if err != nil || !ok || !updated.LastLogin.Equal(at) {
		t.Fatalf("FindByID() after login = %#v, %v, %v", updated, ok, err)
	}
}
