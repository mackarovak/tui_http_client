package requesteditor

import (
    tea "github.com/charmbracelet/bubbletea"
    "htui/internal/types"
)

// Model — заглушка редактора запроса.
type Model struct {
    w, h int
}

// Сообщения, которые использует app.update.go
type SendRequestMsg struct{ Request types.SavedRequest }
type SaveRequestMsg struct{ Request types.SavedRequest }

func New() Model {
    return Model{}
}

func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
    m.w, m.h = w, h
    return m, nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    // В Task 05 логика не нужна
    _ = msg
    return m, nil
}

func (m Model) View() string {
    return "Editor"
}

func (m Model) Width() int  { return m.w }
func (m Model) Height() int { return m.h }

// LoadRequest — заглушка загрузки сохранённого запроса.
func (m Model) LoadRequest(_ types.SavedRequest) Model {
    return m
}

// Clear — заглушка очистки редактора.
func (m Model) Clear() Model {
    return m
}

// CurrentID — ID текущего запроса, в заглушке пустой.
func (m Model) CurrentID() string {
    return ""
}