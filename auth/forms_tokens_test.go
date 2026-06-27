package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLoginPasswordChangeAndSetPasswordForms(t *testing.T) {
	hash, err := EncodePBKDF2PasswordWithIterations("old-secret", "salt", 1)
	if err != nil {
		t.Fatalf("EncodePBKDF2PasswordWithIterations() error = %v", err)
	}
	store, err := NewMemoryUserStore(User{AbstractUser: AbstractUser{
		AbstractBaseUser: AbstractBaseUser{ID: 1, Password: hash, IsActive: true},
		Username:         "saksham",
		Email:            "saksham@example.com",
	}})
	if err != nil {
		t.Fatalf("NewMemoryUserStore() error = %v", err)
	}

	login := LoginForm{Store: store, Username: "saksham", Password: "old-secret"}
	if ok, err := login.Validate(context.Background()); err != nil || !ok || login.User.ID != 1 {
		t.Fatalf("LoginForm.Validate() = %v, %v, user:%#v", ok, err, login.User)
	}

	change := PasswordChangeForm{User: login.User, OldPassword: "bad", NewPassword: "CorrectHorseBatteryStaple42"}
	if ok, err := change.Validate(); err != nil || ok {
		t.Fatalf("PasswordChangeForm should reject wrong old password = %v, %v", ok, err)
	}
	change.OldPassword = "old-secret"
	if ok, err := change.Validate(); err != nil || !ok {
		t.Fatalf("PasswordChangeForm.Validate() = %v, %v", ok, err)
	}
	if err := change.Save(); err != nil {
		t.Fatalf("PasswordChangeForm.Save() error = %v", err)
	}
	if ok, _ := CheckPassword("CorrectHorseBatteryStaple42", change.User.Password); !ok {
		t.Fatalf("changed password does not verify")
	}

	set := SetPasswordForm{User: login.User, NewPassword: "AnotherStrongPassword42"}
	if ok, err := set.Validate(); err != nil || !ok {
		t.Fatalf("SetPasswordForm.Validate() = %v, %v", ok, err)
	}
	if err := set.Save(); err != nil {
		t.Fatalf("SetPasswordForm.Save() error = %v", err)
	}
	if ok, _ := CheckPassword("AnotherStrongPassword42", set.User.Password); !ok {
		t.Fatalf("set password does not verify")
	}
}

func TestPasswordResetTokensExpireAndInvalidateAfterPasswordChange(t *testing.T) {
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	user := User{AbstractUser: AbstractUser{AbstractBaseUser: AbstractBaseUser{ID: 7, Password: "hash", IsActive: true}}}
	signer := PasswordResetTokenSigner{Secret: "reset-secret", MaxAge: time.Hour, Now: func() time.Time { return now }}

	token, err := signer.MakeToken(user)
	if err != nil {
		t.Fatalf("MakeToken() error = %v", err)
	}
	if ok, err := signer.CheckToken(user, token); err != nil || !ok {
		t.Fatalf("CheckToken(valid) = %v, %v", ok, err)
	}

	expiredSigner := signer
	expiredSigner.Now = func() time.Time { return now.Add(2 * time.Hour) }
	if ok, err := expiredSigner.CheckToken(user, token); !errors.Is(err, ErrInvalidPasswordResetToken) || ok {
		t.Fatalf("CheckToken(expired) = %v, %v", ok, err)
	}

	changed := user
	changed.Password = "new-hash"
	if ok, err := signer.CheckToken(changed, token); !errors.Is(err, ErrInvalidPasswordResetToken) || ok {
		t.Fatalf("CheckToken(changed password) = %v, %v", ok, err)
	}
}

func TestPasswordResetAndUserCreationForms(t *testing.T) {
	user := User{AbstractUser: AbstractUser{AbstractBaseUser: AbstractBaseUser{ID: 3, Password: "hash", IsActive: true}, Email: "reset@example.com"}}
	signer := PasswordResetTokenSigner{Secret: "reset-secret", MaxAge: time.Hour}
	token, err := signer.MakeToken(user)
	if err != nil {
		t.Fatalf("MakeToken() error = %v", err)
	}

	reset := PasswordResetConfirmForm{Signer: signer, User: user, Token: token, NewPassword: "CorrectHorseBatteryStaple42"}
	if ok, err := reset.Validate(); err != nil || !ok {
		t.Fatalf("PasswordResetConfirmForm.Validate() = %v, %v", ok, err)
	}
	if err := reset.Save(); err != nil {
		t.Fatalf("PasswordResetConfirmForm.Save() error = %v", err)
	}
	if ok, _ := CheckPassword("CorrectHorseBatteryStaple42", reset.User.Password); !ok {
		t.Fatalf("reset password does not verify")
	}

	create := UserCreationForm{Username: " NewUser ", Email: " New@Example.COM ", Password1: "CreatedStrongPassword42", Password2: "CreatedStrongPassword42"}
	if ok, err := create.Validate(); err != nil || !ok {
		t.Fatalf("UserCreationForm.Validate() = %v, %v", ok, err)
	}
	created, err := create.Save()
	if err != nil {
		t.Fatalf("UserCreationForm.Save() error = %v", err)
	}
	if created.Username != "newuser" || created.Email != "new@example.com" || !created.IsActive {
		t.Fatalf("created user = %#v", created)
	}

	mismatch := UserCreationForm{Username: "bad", Password1: "one", Password2: "two"}
	if ok, err := mismatch.Validate(); !errors.Is(err, ErrFormValidation) || ok {
		t.Fatalf("UserCreationForm mismatch = %v, %v", ok, err)
	}

	change := UserChangeForm{User: created, Email: " Changed@Example.COM ", FirstName: "Changed"}
	if ok, err := change.Validate(); err != nil || !ok {
		t.Fatalf("UserChangeForm.Validate() = %v, %v", ok, err)
	}
	changed := change.Save()
	if changed.Email != "changed@example.com" || changed.FirstName != "Changed" {
		t.Fatalf("changed user = %#v", changed)
	}
}
