package sidebar

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"htui/internal/store"
	"htui/internal/types"
	"htui/internal/ui"
)

type RequestSelectedMsg struct{ Request types.SavedRequest }
type NewRequestMsg struct{}
type DeleteRequestMsg struct{ ID string }
type DuplicateRequestMsg struct{ Request types.SavedRequest }

type requestItem struct {
	req types.SavedRequest
}

func (i requestItem) Title() string     { return i.req.Name }
func (i requestItem) FilterValue() string { return i.req.Name + " " + i.req.URL }
func (i requestItem) Description() string {
	return ui.MethodStyle(i.req.Method).Render(i.req.Method) + "  " + shortURL(i.req.URL)
}

func shortURL(raw string) string {
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	if len(raw) > 40 {
		return raw[:37] + "..."
	}
	return raw
}

type Model struct {
	list      list.Model
	search    textinput.Model
	searching bool
	requests  []types.SavedRequest
	focused   bool
	width     int
	height    int
}

func New(requests []types.SavedRequest) Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("62")).
		BorderLeftForeground(lipgloss.Color("62"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("62"))

	l := list.New(toListItems(requests), delegate, 0, 0)
	l.Title = "Requests"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))

	si := textinput.New()
	si.Placeholder = "Search..."
	si.CharLimit = 100

	return Model{
		list:     l,
		search:   si,
		requests: requests,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searching = false
				m.search.SetValue("")
				m = applyFilter(m, "")
				return m, nil
			case "enter":
				m.searching = false
				return m, nil
			default:
				var c tea.Cmd
				m.search, c = m.search.Update(msg)
				cmds = append(cmds, c)
				m = applyFilter(m, m.search.Value())
				return m, tea.Batch(cmds...)
			}
		}
		var c tea.Cmd
		m.search, c = m.search.Update(msg)
		cmds = append(cmds, c)
		m = applyFilter(m, m.search.Value())
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(requestItem); ok {
				return m, func() tea.Msg { return RequestSelectedMsg{Request: item.req} }
			}
		case "n":
			return m, func() tea.Msg { return NewRequestMsg{} }
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
	if m.searching {
		return lipgloss.JoinVertical(lipgloss.Left,
			ui.Theme.Highlight.Render("Search: ")+m.search.View(),
			m.list.View(),
		)
	}

	if len(m.requests) == 0 {
		return lipgloss.NewStyle().Padding(2, 2).
			Render(ui.Theme.Muted.Render("No saved requests.\n\nPress ") +
				ui.Theme.Highlight.Render("[n]") +
				ui.Theme.Muted.Render(" to create one."))
	}

	return m.list.View()
}

func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
	m.width = w
	m.height = h
	m.list.SetWidth(w)
	m.list.SetHeight(h)
	return m, nil
}

func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }

func (m Model) Focus() Model { m.focused = true; return m }
func (m Model) Blur() Model  { m.focused = false; return m }
func (m Model) SelectFirst() Model {
	if len(m.requests) > 0 {
		m.list.Select(0)
	}
	return m
}

func (m Model) Reload(s store.Store) Model {
	requests, err := s.List()
	if err != nil {
		return m
	}
	m.requests = requests
	m = applyFilter(m, m.search.Value())
	return m
}

func applyFilter(m Model, query string) Model {
	if query == "" {
		m.list.SetItems(toListItems(m.requests))
		return m
	}
	q := strings.ToLower(query)
	var filtered []list.Item
	for _, r := range m.requests {
		if strings.Contains(strings.ToLower(r.Name), q) ||
			strings.Contains(strings.ToLower(r.URL), q) {
			filtered = append(filtered, requestItem{r})
		}
	}
	m.list.SetItems(filtered)
	return m
}

func toListItems(requests []types.SavedRequest) []list.Item {
	items := make([]list.Item, len(requests))
	for i, r := range requests {
		items[i] = requestItem{r}
	}
	return items
}
