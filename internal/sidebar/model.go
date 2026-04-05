package sidebar

import (
	tea "github.com/charmbracelet/bubbletea"

	"htui/internal/store"
	"htui/internal/types"
	"htui/internal/ui"
)

// Сообщения (контракт для app) — полная реализация в ветке step/06-sidebar.

type RequestSelectedMsg struct{ Request types.SavedRequest }
type NewRequestMsg struct{}
type DeleteRequestMsg struct{ ID string }
type DuplicateRequestMsg struct{ Request types.SavedRequest }

// Model — заглушка для задачи 05 (заменяется в ветке 06).
type Model struct {
	width, height int
}

func New(requests []types.SavedRequest) Model {
	_ = requests
	return Model{}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { return m, nil }

func (m Model) View() string {
	if m.width <= 0 {
		return ""
	}
	return ui.Theme.Muted.Render("  Sidebar — реализация в ветке step/06-sidebar\n") +
		ui.Theme.Muted.Render("  (список запросов, поиск, n/d/del)")
}

func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
	m.width, m.height = w, h
	return m, nil
}

func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }

func (m Model) SelectFirst() Model { return m }

func (m Model) Reload(store.Store) Model { return m }
