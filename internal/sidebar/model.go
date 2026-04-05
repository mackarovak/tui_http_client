package sidebar

import (
    tea "github.com/charmbracelet/bubbletea"
    "htui/internal/store"
    "htui/internal/types"
)

// Model — заглушка модели sidebar для Task 05.
type Model struct {
    w, h   int
    items  []types.SavedRequest
}

// Сообщения, которые использует app.update.go
type RequestSelectedMsg struct{ Request types.SavedRequest }
type NewRequestMsg struct{}
type DeleteRequestMsg struct{ ID string }
type DuplicateRequestMsg struct{ Request types.SavedRequest }

// New создаёт модель sidebar с начальным списком запросов.
func New(requests []types.SavedRequest) Model {
    return Model{items: requests}
}

func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
    m.w, m.h = w, h
    return m, nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    // В Task 05 логика не нужна, заглушка
    return m, nil
}

func (m Model) View() string {
    return "Sidebar"
}

func (m Model) Width() int  { return m.w }
func (m Model) Height() int { return m.h }

// SelectFirst — заглушка выбора первого элемента.
func (m Model) SelectFirst() Model {
    return m
}

// Reload — загружает данные из store, но в заглушке ничего не делает.
func (m Model) Reload(_ store.Store) Model {
    return m
}