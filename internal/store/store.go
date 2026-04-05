package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"htui/internal/types"
)

// Store — локальное хранилище сохранённых запросов.
type Store interface {
	List() ([]types.SavedRequest, error)
	Save(r types.SavedRequest) error
	Delete(id string) error
	IsFirstRun() (bool, error)
	MarkSeeded() error
}

// FileStore — JSON-файл + атомарная запись.
type FileStore struct {
	mu       sync.Mutex
	dataPath string
	seedPath string
	requests []types.SavedRequest
}

// New создаёт хранилище в user config dir (…/htui/).
func New() (Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return NewAt(filepath.Join(dir, "htui"))
}

// NewAt создаёт хранилище в указанной директории (удобно для тестов).
func NewAt(base string) (Store, error) {
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, err
	}
	fs := &FileStore{
		dataPath: filepath.Join(base, "requests.json"),
		seedPath: filepath.Join(base, ".seeded"),
	}
	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (f *FileStore) load() error {
	b, err := os.ReadFile(f.dataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f.requests = nil
			return nil
		}
		return err
	}
	var list []types.SavedRequest
	if len(b) == 0 {
		f.requests = nil
		return nil
	}
	if err := json.Unmarshal(b, &list); err != nil {
		return fmt.Errorf("parse requests: %w", err)
	}
	f.requests = list
	return nil
}

func (f *FileStore) flush() error {
	tmp := f.dataPath + ".tmp"
	b, err := json.MarshalIndent(f.requests, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, f.dataPath)
}

// List возвращает копию списка.
func (f *FileStore) List() ([]types.SavedRequest, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]types.SavedRequest, len(f.requests))
	copy(out, f.requests)
	return out, nil
}

// Save создаёт или обновляет запрос по ID.
func (f *FileStore) Save(r types.SavedRequest) error {
	r.EnsureID()
	f.mu.Lock()
	defer f.mu.Unlock()
	replaced := false
	for i := range f.requests {
		if f.requests[i].ID == r.ID {
			f.requests[i] = r
			replaced = true
			break
		}
	}
	if !replaced {
		f.requests = append(f.requests, r)
	}
	return f.flush()
}

// Delete удаляет запрос по ID.
func (f *FileStore) Delete(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	next := f.requests[:0]
	for _, r := range f.requests {
		if r.ID != id {
			next = append(next, r)
		}
	}
	f.requests = next
	return f.flush()
}

// IsFirstRun — true пока не создан маркер после сидирования демо.
func (f *FileStore) IsFirstRun() (bool, error) {
	_, err := os.Stat(f.seedPath)
	if err == nil {
		return false, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	return false, err
}

// MarkSeeded фиксирует, что демо уже выданы.
func (f *FileStore) MarkSeeded() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	seed, err := os.OpenFile(f.seedPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	return seed.Close()
}
