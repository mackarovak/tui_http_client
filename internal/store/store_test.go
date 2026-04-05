package store

import (
	"path/filepath"
	"testing"

	"htui/internal/types"
)

// ТЗ §19.1, §20 п.10 — сохранение, удаление, перезагрузка с диска.

func TestNewAt_SaveListDelete_Reload(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	base := filepath.Join(dir, "htui")

	s, err := NewAt(base)
	if err != nil {
		t.Fatal(err)
	}

	first, err := s.IsFirstRun()
	if err != nil || !first {
		t.Fatalf("IsFirstRun = %v, err %v", first, err)
	}

	r := types.NewSavedRequest()
	r.Name = "Test GET"
	r.URL = "https://example.com"
	if err := s.Save(r); err != nil {
		t.Fatal(err)
	}

	list, err := s.List()
	if err != nil || len(list) != 1 {
		t.Fatalf("List: %v err %v", list, err)
	}
	if list[0].Name != "Test GET" {
		t.Errorf("Name = %q", list[0].Name)
	}

	if err := s.Delete(list[0].ID); err != nil {
		t.Fatal(err)
	}
	list2, _ := s.List()
	if len(list2) != 0 {
		t.Fatalf("after Delete: len %d", len(list2))
	}

	// Новый экземпляр store — читает тот же файл с диска (ТЗ §20.10)
	s2, err := NewAt(base)
	if err != nil {
		t.Fatal(err)
	}
	list3, _ := s2.List()
	if len(list3) != 0 {
		t.Errorf("reopened store should see empty list, got %d", len(list3))
	}
}

func TestMarkSeeded_IsFirstRunFalse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := NewAt(filepath.Join(dir, "htui"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.MarkSeeded(); err != nil {
		t.Fatal(err)
	}
	ok, err := s.IsFirstRun()
	if err != nil || ok {
		t.Fatalf("after MarkSeeded IsFirstRun = %v err %v", ok, err)
	}
}
