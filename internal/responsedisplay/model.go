package responsedisplay

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"htui/internal/types"
	"htui/internal/ui"
)

// ResponseTab — вкладка ответа.
type ResponseTab int

const (
	TabBody ResponseTab = iota
	TabHeaders
)

// Model — компонент отображения ответа.
type Model struct {
	data      *types.ResponseData // nil = пустое состояние
	loading   bool
	activeTab ResponseTab

	bodyVP    viewport.Model
	headersVP viewport.Model
	spinner   spinner.Model

	focused bool
	width   int
	height  int
	ready   bool // false пока не установлены размеры
}

// New создаёт пустой компонент ответа.
func New() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

	return Model{
		spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// SetSize устанавливает размеры и инициализирует viewport.
func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
	// Guard против нулевых размеров
	if w < 5 || h < 5 {
		m.width = w
		m.height = h
		m.ready = true
		return m, nil
	}

	m.width = w
	m.height = h
	m.ready = true

	// Статус-бар: 2 строки. Tab bar: 2 строки. Отступы: 2 строки.
	const reservedLines = 6
	vpHeight := h - reservedLines
	if vpHeight < 1 {
		vpHeight = 1
	}

	m.bodyVP = viewport.New(w, vpHeight)
	m.headersVP = viewport.New(w, vpHeight)

	// Перезаполнить viewport если данные уже есть
	if m.data != nil {
		m.bodyVP.SetContent(FormatBody(*m.data))
		m.headersVP.SetContent(FormatHeaders(m.data.Headers))
	}
	return m, nil
}

// SetLoading включает/выключает спиннер.
func (m Model) SetLoading(loading bool) Model {
	m.loading = loading
	return m
}

// SetResponse загружает данные ответа и обновляет viewport.
func (m Model) SetResponse(data types.ResponseData) Model {
	m.data = &data
	m.loading = false

	if m.ready {
		m.bodyVP.SetContent(FormatBody(data))
		m.headersVP.SetContent(FormatHeaders(data.Headers))
		m.bodyVP.GotoTop()
		m.headersVP.GotoTop()
	}
	return m
}

// UpdateSpinner обновляет спиннер (вызывается из root App при spinner.TickMsg).
func (m Model) UpdateSpinner(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// SpinnerCmd возвращает команду для запуска спиннера.
func (m Model) SpinnerCmd() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			// Переключение вкладки влево
			if m.activeTab == TabHeaders {
				m.activeTab = TabBody
			}
			return m, nil
		case "right":
			// Переключение вкладки вправо
			if m.activeTab == TabBody {
				m.activeTab = TabHeaders
			}
			return m, nil
		}

		// Прокрутка активного viewport
		var cmd tea.Cmd
		if m.activeTab == TabBody {
			m.bodyVP, cmd = m.bodyVP.Update(msg)
		} else {
			m.headersVP, cmd = m.headersVP.Update(msg)
		}
		return m, cmd
	}

	// Другие сообщения (например WindowSizeMsg переданный повторно)
	var cmd tea.Cmd
	if m.activeTab == TabBody {
		m.bodyVP, cmd = m.bodyVP.Update(msg)
	} else {
		m.headersVP, cmd = m.headersVP.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.Theme.Muted.Render("Loading..."),
		)
	}

	if m.loading {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.spinner.View()+"  Sending request...",
		)
	}

	if m.data == nil {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.Theme.Muted.Render("Press ")+
				ui.Theme.Highlight.Render("ctrl+enter")+
				ui.Theme.Muted.Render(" to send a request.\n\nResponse will appear here."),
		)
	}

	if m.data.IsError() {
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			ui.Theme.Error.Render("Request failed\n\n")+
				ui.Theme.Muted.Render(m.data.Error),
		)
	}

	var sb strings.Builder

	// Статус-бар
	sb.WriteString(m.renderStatusBar() + "\n\n")

	// Tab bar
	sb.WriteString(m.renderTabBar() + "\n\n")

	// Контент
	if m.activeTab == TabBody {
		sb.WriteString(m.bodyVP.View())
	} else {
		sb.WriteString(m.headersVP.View())
	}

	return sb.String()
}

func (m Model) renderStatusBar() string {
	statusStyle := ui.StatusStyle(m.data.StatusCode)
	status := statusStyle.Render(m.data.StatusText)
	duration := ui.Theme.Muted.Render(FormatDuration(m.data.DurationMs))
	size := ui.Theme.Muted.Render(FormatBytes(m.data.SizeBytes))
	sep := ui.Theme.Muted.Render("  │  ")
	shortcut := ui.Theme.Muted.Render("←/→: switch Body/Headers")

	return fmt.Sprintf("  %s%s%s%s%s%s%s",
		status, sep, duration, sep, size, sep, shortcut,
	)
}

func (m Model) renderTabBar() string {
	tabs := []string{"Body", "Headers"}
	var parts []string
	for i, name := range tabs {
		label := "[" + name + "]"
		if ResponseTab(i) == m.activeTab {
			parts = append(parts, ui.Theme.Selection.Render(ui.Theme.TabActive.Render(label)))
		} else {
			parts = append(parts, ui.Theme.TabInactive.Render(label))
		}
	}
	return "  " + strings.Join(parts, "  ")
}

// Width / Height для App.renderPanel
func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }
