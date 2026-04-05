package responsedisplay

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/spinner"
    "htui/internal/types"
)

// Model — заглушка панели ответа.
type Model struct {
    w, h    int
    spin    spinner.Model
    loading bool
}

func New() Model {
    s := spinner.New()
    return Model{spin: s}
}

func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
    m.w, m.h = w, h
    return m, nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    // В Task 05 логика не важна
    _ = msg
    return m, nil
}

func (m Model) UpdateSpinner(msg spinner.TickMsg) (Model, tea.Cmd) {
    var cmd tea.Cmd
    m.spin, cmd = m.spin.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    if m.loading {
        return "Loading response..."
    }
    return "Response"
}

func (m Model) Width() int  { return m.w }
func (m Model) Height() int { return m.h }

// SetResponse — заглушка установки ответа.
func (m Model) SetResponse(_ types.ResponseData) Model {
    m.loading = false
    return m
}

// SetLoading — заглушка включения/выключения загрузки.
func (m Model) SetLoading(loading bool) Model {
    m.loading = loading
    return m
}