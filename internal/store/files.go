package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"htui/internal/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileStore — реализация Store поверх JSON-файлов на диске.
type FileStore struct {
	dir     string // ~/.config/htui/requests/
	baseDir string // ~/.config/htui/
}

// New создаёт FileStore, создавая необходимые директории если их нет.
// Путь: os.UserConfigDir() / htui / requests /
func New() (*FileStore, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find config dir: %w", err)
	}

	baseDir := filepath.Join(configDir, "htui")
	reqDir := filepath.Join(baseDir, "requests")

	if err := os.MkdirAll(reqDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create requests dir: %w", err)
	}

	return &FileStore{dir: reqDir, baseDir: baseDir}, nil
}

func (s *FileStore) List() ([]types.SavedRequest, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read requests dir: %w", err)
	}

	var requests []types.SavedRequest
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			log.Printf("warn: cannot read %s: %v", e.Name(), err)
			continue
		}
		var r types.SavedRequest
		if err := json.Unmarshal(data, &r); err != nil {
			log.Printf("warn: corrupt JSON in %s, skipping: %v", e.Name(), err)
			continue
		}
		if r.IsTemplate {
			continue
		}
		requests = append(requests, r)
	}

	// Сортировка по UpdatedAt desc (новые сверху).
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].UpdatedAt.After(requests[j].UpdatedAt)
	})

	return requests, nil
}

func (s *FileStore) ListTemplates() ([]types.SavedRequest, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read requests dir: %w", err)
	}

	var templates []types.SavedRequest
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			log.Printf("warn: cannot read %s: %v", e.Name(), err)
			continue
		}
		var r types.SavedRequest
		if err := json.Unmarshal(data, &r); err != nil {
			log.Printf("warn: corrupt JSON in %s, skipping: %v", e.Name(), err)
			continue
		}
		if !r.IsTemplate {
			continue
		}
		templates = append(templates, r)
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].UpdatedAt.After(templates[j].UpdatedAt)
	})

	return templates, nil
}

func (s *FileStore) Get(id string) (types.SavedRequest, error) {
	path := s.reqPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return types.SavedRequest{}, fmt.Errorf("request %s not found", id)
		}
		return types.SavedRequest{}, err
	}
	var r types.SavedRequest
	if err := json.Unmarshal(data, &r); err != nil {
		return types.SavedRequest{}, fmt.Errorf("corrupt request file %s: %w", id, err)
	}
	return r, nil
}

func (s *FileStore) Save(r types.SavedRequest) error {
	r.UpdatedAt = time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = r.UpdatedAt
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal request: %w", err)
	}

	// Атомарная запись: сначала во временный файл, затем rename.
	path := s.reqPath(r.ID)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("cannot write tmp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp) // cleanup
		return fmt.Errorf("cannot rename tmp file: %w", err)
	}
	return nil
}

func (s *FileStore) Delete(id string) error {
	err := os.Remove(s.reqPath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil // idempotent
	}
	return err
}

func (s *FileStore) IsFirstRun() (bool, error) {
	_, err := os.Stat(s.seededPath())
	if err == nil {
		return false, nil // файл существует → не первый запуск
	}
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	return false, err
}

func (s *FileStore) MarkSeeded() error {
	return os.WriteFile(s.seededPath(), []byte("seeded\n"), 0644)
}

func (s *FileStore) reqPath(id string) string {
	return filepath.Join(s.dir, id+".json")
}

func (s *FileStore) seededPath() string {
	return filepath.Join(s.baseDir, ".seeded")
}
