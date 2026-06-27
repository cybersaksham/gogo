package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrFormValidation = errors.New("form validation failed")

// LoginForm validates username/email and password credentials.
type LoginForm struct {
	Store    UserStore
	Username string
	Email    string
	Password string
	User     User
}

// Validate authenticates the submitted credentials.
func (f *LoginForm) Validate(ctx context.Context) (bool, error) {
	user, ok, err := Authenticate(ctx, f.Store, Credentials{
		Username: f.Username,
		Email:    f.Email,
		Password: f.Password,
	})
	if err != nil || !ok {
		return false, err
	}
	f.User = user
	return true, nil
}

// PasswordChangeForm validates old password and stores a new password hash.
type PasswordChangeForm struct {
	User        User
	OldPassword string
	NewPassword string
}

// Validate checks the old password and built-in validators.
func (f *PasswordChangeForm) Validate() (bool, error) {
	ok, err := CheckPassword(f.OldPassword, f.User.Password)
	if err != nil || !ok {
		return false, err
	}
	if err := ValidatePassword(f.NewPassword, f.User); err != nil {
		return false, err
	}
	return true, nil
}

// Save hashes and stores the new password on the form user.
func (f *PasswordChangeForm) Save() error {
	hash, err := MakePassword(f.NewPassword)
	if err != nil {
		return err
	}
	f.User.Password = hash
	return nil
}

// PasswordResetRequestForm validates a reset request by email.
type PasswordResetRequestForm struct {
	Store UserStore
	Email string
	User  User
}

// Validate finds an active user for the submitted email.
func (f *PasswordResetRequestForm) Validate(ctx context.Context) (bool, error) {
	if f.Store == nil {
		return false, ErrUserStoreRequired
	}
	user, ok, err := f.Store.FindByEmail(ctx, NormalizeEmail(f.Email))
	if err != nil || !ok || !user.IsActive {
		return false, err
	}
	f.User = user
	return true, nil
}

// PasswordResetConfirmForm validates a reset token and new password.
type PasswordResetConfirmForm struct {
	Signer      PasswordResetTokenSigner
	User        User
	Token       string
	NewPassword string
}

// Validate checks reset token validity and password strength.
func (f *PasswordResetConfirmForm) Validate() (bool, error) {
	ok, err := f.Signer.CheckToken(f.User, f.Token)
	if err != nil || !ok {
		return false, err
	}
	if err := ValidatePassword(f.NewPassword, f.User); err != nil {
		return false, err
	}
	return true, nil
}

// Save stores the reset password hash on the form user.
func (f *PasswordResetConfirmForm) Save() error {
	hash, err := MakePassword(f.NewPassword)
	if err != nil {
		return err
	}
	f.User.Password = hash
	return nil
}

// SetPasswordForm validates and stores a new password without old-password check.
type SetPasswordForm struct {
	User        User
	NewPassword string
}

// Validate checks password strength.
func (f *SetPasswordForm) Validate() (bool, error) {
	if err := ValidatePassword(f.NewPassword, f.User); err != nil {
		return false, err
	}
	return true, nil
}

// Save hashes and stores the new password on the form user.
func (f *SetPasswordForm) Save() error {
	hash, err := MakePassword(f.NewPassword)
	if err != nil {
		return err
	}
	f.User.Password = hash
	return nil
}

// UserCreationForm validates and builds the built-in user model.
type UserCreationForm struct {
	Username  string
	Email     string
	Password1 string
	Password2 string
	Now       func() time.Time
}

// Validate checks required identity fields and password confirmation.
func (f *UserCreationForm) Validate() (bool, error) {
	username := NormalizeUsername(f.Username)
	if username == "" {
		return false, fmt.Errorf("%w: username is required", ErrFormValidation)
	}
	if f.Password1 != f.Password2 {
		return false, fmt.Errorf("%w: passwords do not match", ErrFormValidation)
	}
	user := User{AbstractUser: AbstractUser{Username: username, Email: NormalizeEmail(f.Email)}}
	if err := ValidatePassword(f.Password1, user); err != nil {
		return false, err
	}
	return true, nil
}

// Save returns a new active user with normalized fields and hashed password.
func (f *UserCreationForm) Save() (User, error) {
	hash, err := MakePassword(f.Password1)
	if err != nil {
		return User{}, err
	}
	now := time.Now().UTC()
	if f.Now != nil {
		now = f.Now().UTC()
	}
	return User{AbstractUser: AbstractUser{
		AbstractBaseUser: AbstractBaseUser{Password: hash, IsActive: true, DateJoined: now},
		Username:         NormalizeUsername(f.Username),
		Email:            NormalizeEmail(f.Email),
	}}, nil
}

// UserChangeForm validates and applies editable built-in user fields.
type UserChangeForm struct {
	User      User
	Username  string
	Email     string
	FirstName string
	LastName  string
	IsStaff   *bool
	IsActive  *bool
}

// Validate checks basic editable field shape.
func (f *UserChangeForm) Validate() (bool, error) {
	if strings.TrimSpace(f.Email) != "" && !strings.Contains(f.Email, "@") {
		return false, fmt.Errorf("%w: invalid email", ErrFormValidation)
	}
	return true, nil
}

// Save applies normalized values and returns the changed user.
func (f *UserChangeForm) Save() User {
	user := f.User
	if strings.TrimSpace(f.Username) != "" {
		user.Username = NormalizeUsername(f.Username)
	}
	if strings.TrimSpace(f.Email) != "" {
		user.Email = NormalizeEmail(f.Email)
	}
	if f.FirstName != "" {
		user.FirstName = f.FirstName
	}
	if f.LastName != "" {
		user.LastName = f.LastName
	}
	if f.IsStaff != nil {
		user.IsStaff = *f.IsStaff
	}
	if f.IsActive != nil {
		user.IsActive = *f.IsActive
	}
	return user
}
