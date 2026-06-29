package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileUserStore persists built-in users as JSON for generated projects and simple deployments.
type FileUserStore struct {
	mu   sync.Mutex
	path string
}

// NewFileUserStore creates a file-backed user store at path.
func NewFileUserStore(path string) *FileUserStore {
	return &FileUserStore{path: path}
}

// Add inserts a user and persists the full user set.
func (s *FileUserStore) Add(user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.load()
	if err != nil {
		return err
	}
	if err := store.Add(user); err != nil {
		return err
	}
	return s.save(store.Users())
}

// FindByUsername returns a user by normalized username.
func (s *FileUserStore) FindByUsername(ctx context.Context, username string) (User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.load()
	if err != nil {
		return User{}, false, err
	}
	return store.FindByUsername(ctx, username)
}

// FindByEmail returns a user by normalized email.
func (s *FileUserStore) FindByEmail(ctx context.Context, email string) (User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.load()
	if err != nil {
		return User{}, false, err
	}
	return store.FindByEmail(ctx, email)
}

// FindByID returns a user by primary key.
func (s *FileUserStore) FindByID(ctx context.Context, id int64) (User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.load()
	if err != nil {
		return User{}, false, err
	}
	return store.FindByID(ctx, id)
}

// UpdateUser replaces an existing user while preserving unique username/email indexes.
func (s *FileUserStore) UpdateUser(ctx context.Context, user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.load()
	if err != nil {
		return err
	}
	if err := store.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.save(store.Users())
}

// UpdateLastLogin stores the latest successful login timestamp.
func (s *FileUserStore) UpdateLastLogin(ctx context.Context, userID int64, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.load()
	if err != nil {
		return err
	}
	if err := store.UpdateLastLogin(ctx, userID, at); err != nil {
		return err
	}
	return s.save(store.Users())
}

func (s *FileUserStore) load() (*MemoryUserStore, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return NewMemoryUserStore()
	}
	if err != nil {
		return nil, err
	}
	var users []User
	if len(data) > 0 {
		if err := json.Unmarshal(data, &users); err != nil {
			return nil, err
		}
	}
	return NewMemoryUserStore(users...)
}

func (s *FileUserStore) save(users []User) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, s.path)
}
