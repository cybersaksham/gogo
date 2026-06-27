package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	// ErrDuplicateUser is returned when a user store sees duplicate identities.
	ErrDuplicateUser = errors.New("duplicate user")
	// ErrUserStoreRequired is returned when authentication has no user source.
	ErrUserStoreRequired = errors.New("user store is required")
	// ErrUserNotFound is returned when a controlled user update misses.
	ErrUserNotFound = errors.New("user not found")
)

// Credentials carries the fixed built-in username/email authentication inputs.
type Credentials struct {
	Username string
	Email    string
	Password string
	Now      func() time.Time
}

// UserStore is the minimal lookup/update contract for the built-in backend.
type UserStore interface {
	FindByUsername(ctx context.Context, username string) (User, bool, error)
	FindByEmail(ctx context.Context, email string) (User, bool, error)
	FindByID(ctx context.Context, id int64) (User, bool, error)
	UpdateLastLogin(ctx context.Context, userID int64, at time.Time) error
}

// Authenticate performs the non-configurable built-in username/email login.
func Authenticate(ctx context.Context, store UserStore, credentials Credentials) (User, bool, error) {
	if store == nil {
		return User{}, false, ErrUserStoreRequired
	}
	user, ok, err := lookupUser(ctx, store, credentials)
	if err != nil || !ok {
		return User{}, false, err
	}
	if !user.IsActive {
		return User{}, false, nil
	}
	valid, err := CheckPassword(credentials.Password, user.Password)
	if err != nil || !valid {
		return User{}, false, err
	}
	now := time.Now().UTC()
	if credentials.Now != nil {
		now = credentials.Now().UTC()
	}
	if err := store.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return User{}, false, err
	}
	user.LastLogin = now
	user.Authenticated = true
	user.Anonymous = false
	return user, true, nil
}

// NormalizeUsername returns a canonical username for lookups.
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// NormalizeEmail returns a canonical email for lookups.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func lookupUser(ctx context.Context, store UserStore, credentials Credentials) (User, bool, error) {
	if credentials.Username != "" {
		username := NormalizeUsername(credentials.Username)
		if strings.Contains(username, "@") && credentials.Email == "" {
			return store.FindByEmail(ctx, username)
		}
		return store.FindByUsername(ctx, username)
	}
	if credentials.Email != "" {
		return store.FindByEmail(ctx, NormalizeEmail(credentials.Email))
	}
	return User{}, false, nil
}

// MemoryUserStore stores users in memory for tests, bootstrapping, and examples.
type MemoryUserStore struct {
	mu         sync.RWMutex
	byID       map[int64]User
	byUsername map[string]int64
	byEmail    map[string]int64
}

// NewMemoryUserStore creates an in-memory user store.
func NewMemoryUserStore(users ...User) (*MemoryUserStore, error) {
	store := &MemoryUserStore{
		byID:       make(map[int64]User),
		byUsername: make(map[string]int64),
		byEmail:    make(map[string]int64),
	}
	for _, user := range users {
		if err := store.Add(user); err != nil {
			return nil, err
		}
	}
	return store, nil
}

// Add inserts a user into the memory store.
func (s *MemoryUserStore) Add(user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.ID == 0 {
		user.ID = int64(len(s.byID) + 1)
	}
	if _, exists := s.byID[user.ID]; exists {
		return fmt.Errorf("%w: id %d", ErrDuplicateUser, user.ID)
	}
	username := NormalizeUsername(user.Username)
	if username != "" {
		if _, exists := s.byUsername[username]; exists {
			return fmt.Errorf("%w: username %s", ErrDuplicateUser, username)
		}
		s.byUsername[username] = user.ID
		user.Username = username
	}
	email := NormalizeEmail(user.Email)
	if email != "" {
		if _, exists := s.byEmail[email]; exists {
			return fmt.Errorf("%w: email %s", ErrDuplicateUser, email)
		}
		s.byEmail[email] = user.ID
		user.Email = email
	}
	s.byID[user.ID] = user
	return nil
}

// FindByUsername returns a user by normalized username.
func (s *MemoryUserStore) FindByUsername(ctx context.Context, username string) (User, bool, error) {
	if err := ctx.Err(); err != nil {
		return User{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byUsername[NormalizeUsername(username)]
	if !ok {
		return User{}, false, nil
	}
	return s.byID[id], true, nil
}

// FindByEmail returns a user by normalized email.
func (s *MemoryUserStore) FindByEmail(ctx context.Context, email string) (User, bool, error) {
	if err := ctx.Err(); err != nil {
		return User{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byEmail[NormalizeEmail(email)]
	if !ok {
		return User{}, false, nil
	}
	return s.byID[id], true, nil
}

// FindByID returns a user by primary key.
func (s *MemoryUserStore) FindByID(ctx context.Context, id int64) (User, bool, error) {
	if err := ctx.Err(); err != nil {
		return User{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.byID[id]
	return user, ok, nil
}

// UpdateUser replaces an existing user while preserving uniqueness indexes.
func (s *MemoryUserStore) UpdateUser(ctx context.Context, user User) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.byID[user.ID]
	if !ok {
		return fmt.Errorf("%w: id %d", ErrUserNotFound, user.ID)
	}
	oldUsername := NormalizeUsername(existing.Username)
	oldEmail := NormalizeEmail(existing.Email)
	newUsername := NormalizeUsername(user.Username)
	newEmail := NormalizeEmail(user.Email)

	if newUsername != "" {
		if id, exists := s.byUsername[newUsername]; exists && id != user.ID {
			return fmt.Errorf("%w: username %s", ErrDuplicateUser, newUsername)
		}
	}
	if newEmail != "" {
		if id, exists := s.byEmail[newEmail]; exists && id != user.ID {
			return fmt.Errorf("%w: email %s", ErrDuplicateUser, newEmail)
		}
	}
	delete(s.byUsername, oldUsername)
	delete(s.byEmail, oldEmail)
	if newUsername != "" {
		s.byUsername[newUsername] = user.ID
		user.Username = newUsername
	}
	if newEmail != "" {
		s.byEmail[newEmail] = user.ID
		user.Email = newEmail
	}
	s.byID[user.ID] = user
	return nil
}

// UpdateLastLogin stores the latest successful login timestamp.
func (s *MemoryUserStore) UpdateLastLogin(ctx context.Context, userID int64, at time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.byID[userID]
	if !ok {
		return fmt.Errorf("%w: id %d", ErrUserNotFound, userID)
	}
	user.LastLogin = at.UTC()
	s.byID[userID] = user
	return nil
}
