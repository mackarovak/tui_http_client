package responsedisplay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"htui/internal/types"
	"htui/internal/ui"
)

// Model — панель ответа: статус, метаданные, тело.
type Model struct {
	spinner  spinner.Model
	loading  bool
	data     types.ResponseData
	width    int
	height   int
	bodyView string
}

// New создаёт модель с линейным спиннером.
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Line
	return Model{spinner: s, loading: false}
}

// Init запускает тик спиннера.
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// SetLoading включает/выключает индикатор загрузки.
func (m Model) SetLoading(on bool) Model {
	m.loading = on
	if on {
		m.data = types.ResponseData{}
		m.bodyView = ""
	}
	return m
}

// SetResponse заполняет данные ответа.
func (m Model) SetResponse(d types.ResponseData) Model {
	m.loading = false
	m.data = d
	m.bodyView = formatBody(d.Body)
	return m
}

// UpdateSpinner пробрасывает spinner.TickMsg.
func (m Model) UpdateSpinner(msg tea.Msg) (Model, tea.Cmd) {
	if !m.loading {
		return m, nil
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// Update — делегирование спиннеру при загрузке.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m.UpdateSpinner(msg)
}

// SetSize задаёт размер панели.
func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
	m.width = w
	m.height = h
	return m, nil
}

func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }

// View рисует ответ или спиннер.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.loading {
		return fmt.Sprintf("\n\n  %s sending request…", m.spinner.View())
	}

	if m.data.Error != "" {
		return ui.Theme.Error.Render("  "+m.data.Error) + "\n"
	}

	var b strings.Builder
	if m.data.StatusCode > 0 {
		line := fmt.Sprintf("  %s  %s  %d ms  %d B",
			ui.StatusStyle(m.data.StatusCode).Render(fmt.Sprintf("%d", m.data.StatusCode)),
			ui.Theme.Muted.Render(m.data.StatusText),
			m.data.DurationMs,
			m.data.SizeBytes,
		)
		b.WriteString(line)
		b.WriteString("\n\n")
	}

	if len(m.data.Headers) > 0 {
		b.WriteString(ui.Theme.Muted.Render("  Headers:"))
		b.WriteString("\n")
		for _, h := range m.data.Headers {
			if h.Key == "" {
				continue
			}
			b.WriteString(ui.Theme.Muted.Render("    "+h.Key+": "+h.Value) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(ui.Theme.Muted.Render("  Body:"))
	b.WriteString("\n")
	body := m.bodyView
	if body == "" {
		body = ui.Theme.Muted.Render("  (empty)")
	}
	b.WriteString(wrapBody(body, m.width))
	return b.String()
}

func formatBody(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !json.Valid([]byte(raw)) {
		return raw
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return raw
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func wrapBody(s string, maxW int) string {
	if maxW <= 8 {
		return s
	}
	lines := strings.Split(s, "\n")
	var out strings.Builder
	inner := maxW - 4
	if inner < 20 {
		inner = 20
	}
	for _, line := range lines {
		for len(line) > inner {
			out.WriteString("  ")
			out.WriteString(line[:inner])
			out.WriteString("\n")
			line = line[inner:]
		}
		out.WriteString("  ")
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}
