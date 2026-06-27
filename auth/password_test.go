package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestPBKDF2PasswordHashingUsesDjangoCompatibleFormat(t *testing.T) {
	encoded, err := EncodePBKDF2Password("password", "salt", 1)
	if err != nil {
		t.Fatalf("EncodePBKDF2Password() error = %v", err)
	}
	want := "pbkdf2_sha256$1$salt$Eg+2z/z4syxD5yJSVsT4N6hlSMkszDVICAWYfLcL4Xs="
	if encoded != want {
		t.Fatalf("encoded = %q, want %q", encoded, want)
	}

	hash, err := MakePassword("secret")
	if err != nil {
		t.Fatalf("MakePassword() error = %v", err)
	}
	if !strings.HasPrefix(hash, "pbkdf2_sha256$") {
		t.Fatalf("hash = %q, want pbkdf2_sha256 prefix", hash)
	}
	ok, err := CheckPassword("secret", hash)
	if err != nil || !ok {
		t.Fatalf("CheckPassword(valid) = %v, %v", ok, err)
	}
	ok, err = CheckPassword("wrong", hash)
	if err != nil || ok {
		t.Fatalf("CheckPassword(invalid) = %v, %v", ok, err)
	}
}

func TestArgon2IDPasswordHashingVerifiesAndReportsUpgrades(t *testing.T) {
	hasher := Argon2IDHasher{MemoryKiB: 1024, Time: 1, Threads: 1, SaltLength: 16, KeyLength: 32}
	hash, err := MakePasswordWithHasher("secret", hasher)
	if err != nil {
		t.Fatalf("MakePasswordWithHasher(argon2id) error = %v", err)
	}
	if !strings.HasPrefix(hash, "argon2id$") {
		t.Fatalf("hash = %q, want argon2id prefix", hash)
	}
	ok, err := CheckPassword("secret", hash)
	if err != nil || !ok {
		t.Fatalf("CheckPassword(argon2id) = %v, %v", ok, err)
	}
	if !MustUpdatePasswordHash(hash) {
		t.Fatalf("Argon2id should request update while PBKDF2 is configured as default")
	}
}

func TestUnusableInvalidAndOutdatedPasswordHashes(t *testing.T) {
	unusable := SetUnusablePassword()
	if IsPasswordUsable(unusable) {
		t.Fatalf("unusable password reported usable")
	}
	ok, err := CheckPassword("secret", unusable)
	if err != nil || ok {
		t.Fatalf("CheckPassword(unusable) = %v, %v", ok, err)
	}
	ok, err = CheckPassword("secret", "not-a-real-hash")
	if !errors.Is(err, ErrInvalidPasswordHash) || ok {
		t.Fatalf("CheckPassword(invalid) = %v, %v", ok, err)
	}

	old, err := EncodePBKDF2PasswordWithIterations("secret", "salt", 1)
	if err != nil {
		t.Fatalf("EncodePBKDF2PasswordWithIterations() error = %v", err)
	}
	if !MustUpdatePasswordHash(old) {
		t.Fatalf("old PBKDF2 hash should require an update")
	}

	current, err := MakePassword("secret")
	if err != nil {
		t.Fatalf("MakePassword() error = %v", err)
	}
	if MustUpdatePasswordHash(current) {
		t.Fatalf("current default hash should not require update")
	}
}

func TestPasswordValidatorsRejectWeakPasswords(t *testing.T) {
	user := User{AbstractUser: AbstractUser{Username: "saksham", FirstName: "Saksham", LastName: "Singh", Email: "saksham@example.com"}}

	tests := []string{"short", "password", "123456789", "saksham-2026"}
	for _, password := range tests {
		if err := ValidatePassword(password, user); !errors.Is(err, ErrPasswordValidation) {
			t.Fatalf("ValidatePassword(%q) error = %v, want ErrPasswordValidation", password, err)
		}
	}

	if err := ValidatePassword("CorrectHorseBatteryStaple42", user); err != nil {
		t.Fatalf("ValidatePassword(strong) error = %v", err)
	}
}
