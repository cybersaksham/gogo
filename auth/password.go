package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

const (
	pbkdf2Algorithm          = "pbkdf2_sha256"
	argon2IDAlgorithm        = "argon2id"
	DefaultPBKDF2Iterations  = 720000
	defaultSaltLength        = 16
	defaultPasswordKeyLength = 32
	unusablePasswordPrefix   = "!"
)

var (
	// ErrInvalidPasswordHash is returned for malformed or unsupported hashes.
	ErrInvalidPasswordHash = errors.New("invalid password hash")
	// ErrPasswordValidation is returned when a password validator rejects input.
	ErrPasswordValidation = errors.New("password validation failed")
)

// PasswordHasher hashes, verifies, and evaluates stored password hashes.
type PasswordHasher interface {
	Algorithm() string
	Encode(password string) (string, error)
	Verify(password, encoded string) (bool, error)
	MustUpdate(encoded string) bool
}

// PBKDF2SHA256Hasher implements Django-compatible PBKDF2-SHA256 hashes.
type PBKDF2SHA256Hasher struct {
	Iterations int
	SaltLength int
	KeyLength  int
}

// Algorithm returns the encoded hash prefix.
func (h PBKDF2SHA256Hasher) Algorithm() string { return pbkdf2Algorithm }

// Encode hashes a password with a generated salt.
func (h PBKDF2SHA256Hasher) Encode(password string) (string, error) {
	params := h.normalized()
	salt, err := randomToken(params.SaltLength)
	if err != nil {
		return "", err
	}
	return encodePBKDF2(password, salt, params.Iterations, params.KeyLength)
}

// Verify validates a password against a PBKDF2-SHA256 encoded hash.
func (h PBKDF2SHA256Hasher) Verify(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != pbkdf2Algorithm {
		return false, fmt.Errorf("%w: expected %s", ErrInvalidPasswordHash, pbkdf2Algorithm)
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false, fmt.Errorf("%w: invalid PBKDF2 iterations", ErrInvalidPasswordHash)
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("%w: invalid PBKDF2 digest", ErrInvalidPasswordHash)
	}
	candidate, err := encodePBKDF2(password, parts[2], iterations, len(decoded))
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(encoded)) == 1, nil
}

// MustUpdate reports whether the hash is below configured PBKDF2 parameters.
func (h PBKDF2SHA256Hasher) MustUpdate(encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != pbkdf2Algorithm {
		return true
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return true
	}
	return iterations < h.normalized().Iterations
}

func (h PBKDF2SHA256Hasher) normalized() PBKDF2SHA256Hasher {
	if h.Iterations <= 0 {
		h.Iterations = DefaultPBKDF2Iterations
	}
	if h.SaltLength <= 0 {
		h.SaltLength = defaultSaltLength
	}
	if h.KeyLength <= 0 {
		h.KeyLength = defaultPasswordKeyLength
	}
	return h
}

// Argon2IDHasher implements Argon2id password hashes.
type Argon2IDHasher struct {
	MemoryKiB  uint32
	Time       uint32
	Threads    uint8
	SaltLength int
	KeyLength  uint32
}

// Algorithm returns the encoded hash prefix.
func (h Argon2IDHasher) Algorithm() string { return argon2IDAlgorithm }

// Encode hashes a password with Argon2id and a generated salt.
func (h Argon2IDHasher) Encode(password string) (string, error) {
	params := h.normalized()
	salt, err := randomBytes(params.SaltLength)
	if err != nil {
		return "", err
	}
	digest := argon2.IDKey([]byte(password), salt, params.Time, params.MemoryKiB, params.Threads, params.KeyLength)
	return fmt.Sprintf("%s$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2IDAlgorithm,
		params.MemoryKiB,
		params.Time,
		params.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(digest),
	), nil
}

// Verify validates a password against an Argon2id encoded hash.
func (h Argon2IDHasher) Verify(password, encoded string) (bool, error) {
	params, salt, digest, err := parseArgon2ID(encoded)
	if err != nil {
		return false, err
	}
	candidate := argon2.IDKey([]byte(password), salt, params.Time, params.MemoryKiB, params.Threads, uint32(len(digest)))
	return subtle.ConstantTimeCompare(candidate, digest) == 1, nil
}

// MustUpdate reports whether the hash is below configured Argon2id parameters.
func (h Argon2IDHasher) MustUpdate(encoded string) bool {
	params, _, _, err := parseArgon2ID(encoded)
	if err != nil {
		return true
	}
	current := h.normalized()
	return params.MemoryKiB < current.MemoryKiB || params.Time < current.Time || params.Threads < current.Threads || params.KeyLength < current.KeyLength
}

func (h Argon2IDHasher) normalized() Argon2IDHasher {
	if h.MemoryKiB == 0 {
		h.MemoryKiB = 64 * 1024
	}
	if h.Time == 0 {
		h.Time = 3
	}
	if h.Threads == 0 {
		h.Threads = 2
	}
	if h.SaltLength <= 0 {
		h.SaltLength = defaultSaltLength
	}
	if h.KeyLength == 0 {
		h.KeyLength = defaultPasswordKeyLength
	}
	return h
}

// MakePassword hashes a password with the default hasher.
func MakePassword(password string) (string, error) {
	return MakePasswordWithHasher(password, defaultPasswordHasher())
}

// MakePasswordWithHasher hashes a password with an explicit hasher.
func MakePasswordWithHasher(password string, hasher PasswordHasher) (string, error) {
	if hasher == nil {
		hasher = defaultPasswordHasher()
	}
	return hasher.Encode(password)
}

// CheckPassword verifies a raw password against an encoded hash.
func CheckPassword(password, encoded string) (bool, error) {
	if !IsPasswordUsable(encoded) {
		return false, nil
	}
	hasher, err := hasherForEncoded(encoded)
	if err != nil {
		return false, err
	}
	return hasher.Verify(password, encoded)
}

// IsPasswordUsable reports whether a stored password can be checked.
func IsPasswordUsable(encoded string) bool {
	return encoded != "" && !strings.HasPrefix(encoded, unusablePasswordPrefix)
}

// SetUnusablePassword returns a marker that can never validate a password.
func SetUnusablePassword() string {
	token, err := randomToken(defaultSaltLength)
	if err != nil {
		return unusablePasswordPrefix
	}
	return unusablePasswordPrefix + token
}

// MustUpdatePasswordHash reports whether the encoded hash should be upgraded.
func MustUpdatePasswordHash(encoded string) bool {
	if !IsPasswordUsable(encoded) {
		return false
	}
	hasher, err := hasherForEncoded(encoded)
	if err != nil {
		return true
	}
	defaultHasher := defaultPasswordHasher()
	if hasher.Algorithm() != defaultHasher.Algorithm() {
		return true
	}
	return defaultHasher.MustUpdate(encoded)
}

// EncodePBKDF2Password returns a PBKDF2-SHA256 test-vector-friendly hash.
func EncodePBKDF2Password(password, salt string, iterations int) (string, error) {
	return EncodePBKDF2PasswordWithIterations(password, salt, iterations)
}

// EncodePBKDF2PasswordWithIterations returns PBKDF2-SHA256 with explicit iterations.
func EncodePBKDF2PasswordWithIterations(password, salt string, iterations int) (string, error) {
	return encodePBKDF2(password, salt, iterations, defaultPasswordKeyLength)
}

// ValidatePassword runs the built-in password validators.
func ValidatePassword(password string, user User) error {
	var messages []string
	if len([]rune(password)) < 8 {
		messages = append(messages, "password is too short")
	}
	if isCommonPassword(password) {
		messages = append(messages, "password is too common")
	}
	if isNumericPassword(password) {
		messages = append(messages, "password is entirely numeric")
	}
	if isSimilarToUserAttribute(password, user) {
		messages = append(messages, "password is too similar to user attributes")
	}
	if len(messages) > 0 {
		return fmt.Errorf("%w: %s", ErrPasswordValidation, strings.Join(messages, "; "))
	}
	return nil
}

func defaultPasswordHasher() PBKDF2SHA256Hasher {
	return PBKDF2SHA256Hasher{Iterations: DefaultPBKDF2Iterations, SaltLength: defaultSaltLength, KeyLength: defaultPasswordKeyLength}
}

func hasherForEncoded(encoded string) (PasswordHasher, error) {
	switch {
	case strings.HasPrefix(encoded, pbkdf2Algorithm+"$"):
		return defaultPasswordHasher(), nil
	case strings.HasPrefix(encoded, argon2IDAlgorithm+"$"):
		return Argon2IDHasher{}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported algorithm", ErrInvalidPasswordHash)
	}
}

func encodePBKDF2(password, salt string, iterations, keyLength int) (string, error) {
	if iterations <= 0 || keyLength <= 0 {
		return "", fmt.Errorf("%w: invalid PBKDF2 parameters", ErrInvalidPasswordHash)
	}
	digest := pbkdf2SHA256([]byte(password), []byte(salt), iterations, keyLength)
	return fmt.Sprintf("%s$%d$%s$%s", pbkdf2Algorithm, iterations, salt, base64.StdEncoding.EncodeToString(digest)), nil
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLength int) []byte {
	hashLength := sha256.Size
	blockCount := int(math.Ceil(float64(keyLength) / float64(hashLength)))
	output := make([]byte, 0, blockCount*hashLength)
	for block := 1; block <= blockCount; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		var counter [4]byte
		binary.BigEndian.PutUint32(counter[:], uint32(block))
		mac.Write(counter[:])
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		output = append(output, t...)
	}
	return output[:keyLength]
}

func parseArgon2ID(encoded string) (Argon2IDHasher, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != argon2IDAlgorithm || parts[1] != "v=19" {
		return Argon2IDHasher{}, nil, nil, fmt.Errorf("%w: invalid Argon2id hash", ErrInvalidPasswordHash)
	}
	var memory, timeCost uint32
	var threads uint8
	for _, assignment := range strings.Split(parts[2], ",") {
		key, value, ok := strings.Cut(assignment, "=")
		if !ok {
			return Argon2IDHasher{}, nil, nil, fmt.Errorf("%w: invalid Argon2id parameters", ErrInvalidPasswordHash)
		}
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return Argon2IDHasher{}, nil, nil, fmt.Errorf("%w: invalid Argon2id parameter value", ErrInvalidPasswordHash)
		}
		switch key {
		case "m":
			memory = uint32(parsed)
		case "t":
			timeCost = uint32(parsed)
		case "p":
			threads = uint8(parsed)
		}
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return Argon2IDHasher{}, nil, nil, fmt.Errorf("%w: invalid Argon2id salt", ErrInvalidPasswordHash)
	}
	digest, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2IDHasher{}, nil, nil, fmt.Errorf("%w: invalid Argon2id digest", ErrInvalidPasswordHash)
	}
	return Argon2IDHasher{MemoryKiB: memory, Time: timeCost, Threads: threads, SaltLength: len(salt), KeyLength: uint32(len(digest))}, salt, digest, nil
}

func randomToken(length int) (string, error) {
	token, err := randomBytes(length)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(token), "="), nil
}

func randomBytes(length int) ([]byte, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func isCommonPassword(password string) bool {
	common := map[string]struct{}{
		"password":   {},
		"password1":  {},
		"qwerty":     {},
		"admin":      {},
		"letmein":    {},
		"12345678":   {},
		"123456789":  {},
		"1234567890": {},
	}
	_, ok := common[strings.ToLower(strings.TrimSpace(password))]
	return ok
}

func isNumericPassword(password string) bool {
	if password == "" {
		return false
	}
	for _, r := range password {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isSimilarToUserAttribute(password string, user User) bool {
	normalizedPassword := normalizePasswordComparison(password)
	attributes := []string{user.Username, user.FirstName, user.LastName, user.Email}
	for _, attribute := range attributes {
		value := normalizePasswordComparison(attribute)
		if len(value) < 4 {
			continue
		}
		if strings.Contains(normalizedPassword, value) || strings.Contains(value, normalizedPassword) || localEmailPart(value) != "" && strings.Contains(normalizedPassword, localEmailPart(value)) {
			return true
		}
	}
	return false
}

func normalizePasswordComparison(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '@' || r == '.' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func localEmailPart(value string) string {
	local, _, ok := strings.Cut(value, "@")
	if !ok || len(local) < 4 {
		return ""
	}
	return local
}
