package sidebar

import (
	"strings"

	"htui/internal/store"
	"htui/internal/types"
	"htui/internal/ui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Сообщения, эмитируемые sidebar вверх в App ---

type RequestSelectedMsg struct{ Request types.SavedRequest }
type TemplateSelectedMsg struct{ Template types.SavedRequest }
type NewRequestMsg struct{}
type NewRequestWithMethodMsg struct{ Method string }
type DeleteRequestMsg struct{ ID string }
type DuplicateRequestMsg struct{ Request types.SavedRequest }

// --- requestItem адаптирует SavedRequest для bubbles/list ---

type requestItem struct {
	req types.SavedRequest
}

func (i requestItem) Title() string {
	return i.req.Name
}

func (i requestItem) Description() string {
	return ui.MethodStyle(i.req.Method).Render(i.req.Method) +
		"  " + shortURL(i.req.URL)
}

func (i requestItem) FilterValue() string {
	return i.req.Name + " " + i.req.URL
}

// templateItem — элемент списка шаблонов (со звёздочкой в заголовке).
type templateItem struct {
	req types.SavedRequest
}

func (i templateItem) Title() string {
	return "★ " + i.req.Name
}

func (i templateItem) Description() string {
	return ui.MethodStyle(i.req.Method).Render(i.req.Method) +
		"  " + shortURL(i.req.URL)
}

func (i templateItem) FilterValue() string {
	return i.req.Name + " " + i.req.URL
}

// shortURL обрезает URL до отображаемой длины (убирает схему).
func shortURL(raw string) string {
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	if len(raw) > 40 {
		return raw[:37] + "..."
	}
	return raw
}

// --- Model ---

type Model struct {
	list         list.Model
	search       textinput.Model
	searching    bool
	templateMode bool
	requests     []types.SavedRequest // запросы (IsTemplate=false)
	templates    []types.SavedRequest // шаблоны (IsTemplate=true)
	focused      bool
	width        int
	height       int
}

// New создаёт Sidebar с начальным набором запросов и шаблонов.
func New(requests []types.SavedRequest, templates []types.SavedRequest) Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("236")).
		Bold(true).
		BorderLeftForeground(lipgloss.Color("62"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236"))

	l := list.New(toListItems(requests), delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // используем свою фильтрацию

	si := textinput.New()
	si.Placeholder = "Search..."
	si.CharLimit = 100

	return Model{
		list:      l,
		search:    si,
		requests:  requests,
		templates: templates,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Template mode: навигация по шаблонам
	if m.templateMode {
		if m.searching {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "esc":
					m.searching = false
					m.search.SetValue("")
					m.applyFilter("")
					return m, nil
				case "enter":
					m.searching = false
					return m, nil
				default:
					var c tea.Cmd
					m.search, c = m.search.Update(msg)
					cmds = append(cmds, c)
					m.applyFilter(m.search.Value())
					return m, tea.Batch(cmds...)
				}
			}
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "t", "esc":
				m.templateMode = false
				m.searching = false
				m.search.SetValue("")
				m.list.SetItems(toListItems(m.requests))
				return m, nil
			case "enter":
				if item, ok := m.list.SelectedItem().(templateItem); ok {
					return m, func() tea.Msg { return TemplateSelectedMsg{Template: item.req} }
				}
			case "delete", "backspace":
				if item, ok := m.list.SelectedItem().(templateItem); ok {
					return m, func() tea.Msg { return DeleteRequestMsg{ID: item.req.ID} }
				}
			case "/":
				m.searching = true
				m.search.Focus()
				return m, textinput.Blink
			}
		}

		var c tea.Cmd
		m.list, c = m.list.Update(msg)
		cmds = append(cmds, c)
		return m, tea.Batch(cmds...)
	}

	// Search mode
	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searching = false
				m.search.SetValue("")
				m.applyFilter("")
				return m, nil
			case "enter":
				m.searching = false
				return m, nil
			default:
				var c tea.Cmd
				m.search, c = m.search.Update(msg)
				cmds = append(cmds, c)
				m.applyFilter(m.search.Value())
				return m, tea.Batch(cmds...)
			}
		}
	}

	// Нормальный режим
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(requestItem); ok {
				return m, func() tea.Msg { return RequestSelectedMsg{Request: item.req} }
			}
		case "t":
			m.templateMode = true
			m.list.SetItems(toTemplateListItems(m.templates))
			m.list.Select(0)
			return m, nil
		case "n":
			return m, func() tea.Msg { return NewRequestMsg{} }
		case "p":
			return m, func() tea.Msg { return NewRequestWithMethodMsg{Method: "POST"} }
		case "u":
			return m, func() tea.Msg { return NewRequestWithMethodMsg{Method: "PUT"} }
		case "h":
			return m, func() tea.Msg { return NewRequestWithMethodMsg{Method: "PATCH"} }
		case "o":
			return m, func() tea.Msg { return NewRequestWithMethodMsg{Method: "OPTIONS"} }
		case "e":
			return m, func() tea.Msg { return NewRequestWithMethodMsg{Method: "HEAD"} }
		case "d":
			if item, ok := m.list.SelectedItem().(requestItem); ok {
				return m, func() tea.Msg { return DuplicateRequestMsg{Request: item.req} }
			}
		case "delete", "backspace":
			if item, ok := m.list.SelectedItem().(requestItem); ok {
				return m, func() tea.Msg { return DeleteRequestMsg{ID: item.req.ID} }
			}
		case "/":
			m.searching = true
			m.search.Focus()
			return m, textinput.Blink
		}
	}

	var c tea.Cmd
	m.list, c = m.list.Update(msg)
	cmds = append(cmds, c)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.templateMode {
		var header string
		if m.searching {
			header = ui.Theme.Highlight.Render("Search: ") + m.search.View()
		} else {
			header = ui.Theme.Highlight.Render("★ TEMPLATES") +
				ui.Theme.Muted.Render("  [t/esc] назад")
		}

		if len(m.templates) == 0 && !m.searching {
			emptyMsg := lipgloss.NewStyle().Padding(2, 2).
				Render(ui.Theme.Muted.Render("Нет шаблонов.\n\nВ редакторе нажмите ") +
					ui.Theme.Highlight.Render("[ctrl+t]") +
					ui.Theme.Muted.Render(" чтобы сохранить."))
			helpText := ui.Theme.Muted.Render("\n  [t/esc] назад")
			return lipgloss.JoinVertical(lipgloss.Left, header, emptyMsg, helpText)
		}

		helpText := ui.Theme.Muted.Render("\n  [enter] открыть  [del] удалить  [/] поиск  [t/esc] назад")
		return lipgloss.JoinVertical(lipgloss.Left, header, m.list.View(), helpText)
	}

	if m.searching {
		searchView := lipgloss.JoinVertical(lipgloss.Left,
			ui.Theme.Highlight.Render("Search: ")+m.search.View(),
			m.list.View(),
		)
		helpText := ui.Theme.Muted.Render("\n  [esc] cancel  [/] search")
		return lipgloss.JoinVertical(lipgloss.Left, searchView, helpText)
	}

	if len(m.requests) == 0 {
		return lipgloss.NewStyle().Padding(2, 2).
			Render(ui.Theme.Muted.Render("No saved requests.\n\nPress ") +
				ui.Theme.Highlight.Render("[n]") +
				ui.Theme.Muted.Render(" to create one."))
	}

	listView := m.list.View()
	helpText := ui.Theme.Muted.Render("\n  [n] new  [p] POST  [u] PUT  [h] PATCH\n  [o] OPTIONS  [e] HEAD\n  [d] dup  [del] delete  [/] search  [t] templates")
	return lipgloss.JoinVertical(lipgloss.Left, listView, helpText)
}

// --- Публичные методы для App ---

// SetSize обновляет размеры компонента.
func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
	m.width = w
	m.height = h
	// Leave space for help text at bottom (4 lines: 1 blank + 3 text)
	m.list.SetWidth(max(w, 0))
	m.list.SetHeight(max(h-4, 0))
	return m, nil
}

// Width / Height для renderPanel в app/view.go
func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }

// Focus / Blur управляют визуальным состоянием.
func (m Model) Focus() Model { m.focused = true; return m }
func (m Model) Blur() Model  { m.focused = false; return m }

// SelectFirst выбирает первый элемент списка.
func (m Model) SelectFirst() Model {
	if len(m.requests) > 0 {
		m.list.Select(0)
	}
	return m
}

// Reload перезагружает список запросов и шаблонов из store.
func (m Model) Reload(s store.Store) Model {
	requests, err := s.List()
	if err == nil {
		m.requests = requests
	}
	templates, err := s.ListTemplates()
	if err == nil {
		m.templates = templates
	}
	if m.templateMode {
		m.list.SetItems(toTemplateListItems(m.templates))
	} else {
		m.applyFilter(m.search.Value())
	}
	return m
}

// --- Внутренние хелперы ---

func (m *Model) applyFilter(query string) {
	var source []types.SavedRequest
	if m.templateMode {
		source = m.templates
	} else {
		source = m.requests
	}

	if query == "" {
		if m.templateMode {
			m.list.SetItems(toTemplateListItems(source))
		} else {
			m.list.SetItems(toListItems(source))
		}
		return
	}

	q := strings.ToLower(query)
	var filtered []list.Item
	for _, r := range source {
		if strings.Contains(strings.ToLower(r.Name), q) ||
			strings.Contains(strings.ToLower(r.URL), q) {
			if m.templateMode {
				filtered = append(filtered, templateItem{r})
			} else {
				filtered = append(filtered, requestItem{r})
			}
		}
	}
	m.list.SetItems(filtered)
}

func toListItems(requests []types.SavedRequest) []list.Item {
	items := make([]list.Item, len(requests))
	for i, r := range requests {
		items[i] = requestItem{r}
	}
	return items
}

func toTemplateListItems(templates []types.SavedRequest) []list.Item {
	items := make([]list.Item, len(templates))
	for i, r := range templates {
		items[i] = templateItem{r}
	}
	return items
}
