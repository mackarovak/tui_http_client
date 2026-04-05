package requesteditor

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"htui/internal/types"
	"htui/internal/ui"
)

// Заглушка редактора для задачи 05 (полный UI — ветка step/07-request-editor).

type SendRequestMsg struct{ Request types.SavedRequest }
type SaveRequestMsg struct{ Request types.SavedRequest }

type Model struct {
	current  types.SavedRequest
	width    int
	height   int
	urlInput textinput.Model
}

func New() Model {
	url := textinput.New()
	url.Placeholder = "https://..."
	url.CharLimit = 2000
	url.Focus()
	return Model{
		current:  types.NewSavedRequest(),
		urlInput: url,
	}
}

func (m Model) LoadRequest(r types.SavedRequest) Model {
	m.current = r
	m.urlInput.SetValue(r.URL)
	return m
}

func (m Model) Clear() Model { return New() }

func (m Model) CurrentID() string { return m.current.ID }

func (m Model) BuildRequest() types.SavedRequest {
	r := m.current
	r.URL = strings.TrimSpace(m.urlInput.Value())
	if r.Method == "" {
		r.Method = "GET"
	}
	return r
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+enter":
			return m, func() tea.Msg { return SendRequestMsg{Request: m.BuildRequest()} }
		case "ctrl+s":
			return m, func() tea.Msg { return SaveRequestMsg{Request: m.BuildRequest()} }
		}
	}
	var cmd tea.Cmd
	m.urlInput, cmd = m.urlInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(ui.Theme.Muted.Render("  Editor — полный UI в step/07-request-editor\n\n"))
	sb.WriteString(ui.Theme.Muted.Render("  URL: "))
	sb.WriteString(m.urlInput.View())
	sb.WriteString("\n\n")
	sb.WriteString(ui.Theme.Muted.Render("  ctrl+enter send  ctrl+s save"))
	return sb.String()
}

func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
	m.width = w
	m.height = h
	m.urlInput.Width = w - 12
	return m, nil
}

func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }
