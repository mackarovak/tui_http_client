package requesteditor

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"htui/internal/types"
	"htui/internal/ui"
)

type EditorTab int

const (
	TabBody EditorTab = iota
	TabHeaders
	TabParams
	TabAuth
)

var tabNames = []string{"Body", "Headers", "Params", "Auth"}

type FocusField int

const (
	FieldMethod FocusField = iota
	FieldURL
	FieldTabContent
)

type SendRequestMsg struct{ Request types.SavedRequest }
type SaveRequestMsg struct{ Request types.SavedRequest }

type Model struct {
	current  types.SavedRequest
	dirty    bool

	urlInput    textinput.Model
	bodyInput   textarea.Model
	headerTable KVTable
	paramTable  KVTable
	bodyModeIdx int

	authTypeIdx  int
	tokenInput   textinput.Model
	tokenVisible bool

	activeTab      EditorTab
	focusField     FocusField
	methodIdx      int
	focused        bool
	width          int
	height         int
	validationErrs []string
}

var bodyModes = []types.BodyMode{
	types.BodyModeNone,
	types.BodyModeRawText,
	types.BodyModeJSON,
	types.BodyModeForm,
}

var bodyModeLabels = []string{"none", "raw", "json", "form"}

func New() Model {
	url := textinput.New()
	url.Placeholder = "https://api.example.com/endpoint"
	url.CharLimit = 2000
	url.Focus()

	body := textarea.New()
	body.Placeholder = "Request body..."
	body.CharLimit = 0
	body.SetHeight(8)

	token := textinput.New()
	token.Placeholder = "Enter bearer token..."
	token.EchoMode = textinput.EchoPassword
	token.CharLimit = 500

	return Model{
		current:     types.NewSavedRequest(),
		urlInput:    url,
		bodyInput:   body,
		headerTable: NewKVTable(),
		paramTable:  NewKVTable(),
		tokenInput:  token,
		focusField:  FieldURL,
	}
}

func (m Model) LoadRequest(r types.SavedRequest) Model {
	m.current = r
	m.dirty = false

	m.urlInput.SetValue(r.URL)

	m.methodIdx = 0
	for i, method := range types.HTTPMethods {
		if method == r.Method {
			m.methodIdx = i
			break
		}
	}

	m.bodyModeIdx = 0
	for i, bm := range bodyModes {
		if bm == r.BodyMode {
			m.bodyModeIdx = i
			break
		}
	}
	m.bodyInput.SetValue(r.Body)

	m.headerTable = FromHeaders(r.Headers)
	m.paramTable = FromParams(r.Params)

	if r.Auth.Type == types.AuthBearer {
		m.authTypeIdx = 1
	} else {
		m.authTypeIdx = 0
	}
	m.tokenInput.SetValue(r.Auth.Token)
	m.tokenVisible = r.Auth.TokenVisible

	m.validationErrs = nil
	return m
}

func (m Model) Clear() Model {
	return New()
}

func (m Model) CurrentID() string {
	return m.current.ID
}

func (m Model) BuildRequest() types.SavedRequest {
	r := m.current
	r.Method = types.HTTPMethods[m.methodIdx]
	r.URL = m.urlInput.Value()
	r.BodyMode = bodyModes[m.bodyModeIdx]
	r.Body = m.bodyInput.Value()
	r.Headers = m.headerTable.ToHeaders()
	r.Params = m.paramTable.ToParams()

	if m.authTypeIdx == 1 {
		r.Auth = types.AuthConfig{
			Type:         types.AuthBearer,
			Token:        m.tokenInput.Value(),
			TokenVisible: m.tokenVisible,
		}
	} else {
		r.Auth = types.AuthConfig{Type: types.AuthNone}
	}

	return r
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+enter":
			req := m.BuildRequest()
			result := Validate(req)
			if !result.Valid {
				m.validationErrs = result.Errors
				return m, nil
			}
			m.validationErrs = nil
			return m, func() tea.Msg { return SendRequestMsg{Request: req} }

		case "ctrl+s":
			req := m.BuildRequest()
			m.dirty = false
			return m, func() tea.Msg { return SaveRequestMsg{Request: req} }

		case "left", "right":
			if m.focusField == FieldMethod {
				delta := 1
				if msg.String() == "left" {
					delta = -1
				}
				m.methodIdx = (m.methodIdx + delta + len(types.HTTPMethods)) % len(types.HTTPMethods)
				m.dirty = true
				return m, nil
			}

		case "tab":
			switch m.focusField {
			case FieldMethod:
				m.focusField = FieldURL
				m.urlInput.Focus()
			case FieldURL:
				m.urlInput.Blur()
				m.focusField = FieldTabContent
				m.focusTabContent()
			case FieldTabContent:
				m.activeTab = EditorTab((int(m.activeTab) + 1) % len(tabNames))
				m.focusTabContent()
			}
			return m, nil

		case "shift+tab":
			switch m.focusField {
			case FieldTabContent:
				m.focusField = FieldURL
				m.urlInput.Focus()
			case FieldURL:
				m.urlInput.Blur()
				m.focusField = FieldMethod
			}
			return m, nil

		case "1", "2", "3", "4", "5":
			idx := int(msg.String()[0] - '1')
			if idx < len(types.HTTPMethods) {
				m.methodIdx = idx
				m.dirty = true
			}
			return m, nil
		}
	}

	return m.routeToActiveField(msg)
}

func (m Model) routeToActiveField(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focusField {
	case FieldURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
		m.dirty = true

	case FieldTabContent:
		switch m.activeTab {
		case TabBody:
			if bodyModes[m.bodyModeIdx] == types.BodyModeJSON ||
				bodyModes[m.bodyModeIdx] == types.BodyModeRawText {
				m.bodyInput, cmd = m.bodyInput.Update(msg)
				m.dirty = true
			} else if bodyModes[m.bodyModeIdx] == types.BodyModeForm {
				m.paramTable, cmd = m.paramTable.Update(msg)
				m.dirty = true
			} else {
				switch msg := msg.(type) {
				case tea.KeyMsg:
					switch msg.String() {
					case "left":
						if m.bodyModeIdx > 0 {
							m.bodyModeIdx--
						}
					case "right":
						if m.bodyModeIdx < len(bodyModes)-1 {
							m.bodyModeIdx++
						}
					}
				}
			}

		case TabHeaders:
			m.headerTable, cmd = m.headerTable.Update(msg)
			m.dirty = true

		case TabParams:
			m.paramTable, cmd = m.paramTable.Update(msg)
			m.dirty = true

		case TabAuth:
			var c tea.Cmd
			m, c = m.updateAuthTab(msg)
			cmd = c
		}
	}
	return m, cmd
}

func (m Model) updateAuthTab(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "right":
			m.authTypeIdx = (m.authTypeIdx + 1) % 2
			if m.authTypeIdx == 1 {
				m.tokenInput.Focus()
			} else {
				m.tokenInput.Blur()
			}
			m.dirty = true
			return m, nil
		case "ctrl+h":
			m.tokenVisible = !m.tokenVisible
			if m.tokenVisible {
				m.tokenInput.EchoMode = textinput.EchoNormal
			} else {
				m.tokenInput.EchoMode = textinput.EchoPassword
			}
			return m, nil
		}
	}
	if m.authTypeIdx == 1 {
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		m.dirty = true
	}
	return m, cmd
}

func (m *Model) focusTabContent() {
	switch m.activeTab {
	case TabBody:
		if bodyModes[m.bodyModeIdx] != types.BodyModeNone {
			m.bodyInput.Focus()
		}
	case TabAuth:
		if m.authTypeIdx == 1 {
			m.tokenInput.Focus()
		}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(m.renderMethodSelector() + "\n")
	sb.WriteString(m.renderURLField() + "\n")

	if len(m.validationErrs) > 0 {
		for _, e := range m.validationErrs {
			sb.WriteString(ui.Theme.Error.Render("  ✗ "+e) + "\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(m.renderTabBar() + "\n\n")
	sb.WriteString(m.renderTabContent())

	hint := ui.Theme.Muted.Render("  ctrl+enter send  ctrl+s save")
	if m.dirty {
		hint = ui.Theme.Muted.Render("  ctrl+enter send  ctrl+s save  ") +
			ui.Theme.Error.Render("● unsaved")
	}
	sb.WriteString("\n" + hint)

	return sb.String()
}

func (m Model) renderMethodSelector() string {
	var parts []string
	for i, method := range types.HTTPMethods {
		style := ui.Theme.Muted
		if i == m.methodIdx {
			style = ui.MethodStyle(method).Bold(true)
		}
		parts = append(parts, style.Render("["+method+"]"))
	}
	return "  " + strings.Join(parts, " ")
}

func (m Model) renderURLField() string {
	label := ui.Theme.Muted.Render("  URL: ")
	return label + m.urlInput.View()
}

func (m Model) renderTabBar() string {
	var parts []string
	for i, name := range tabNames {
		if EditorTab(i) == m.activeTab {
			parts = append(parts, ui.Theme.TabActive.Render(name))
		} else {
			parts = append(parts, ui.Theme.TabInactive.Render(name))
		}
	}
	return "  " + strings.Join(parts, "  ")
}

func (m Model) renderTabContent() string {
	switch m.activeTab {
	case TabBody:
		return m.renderBodyTab()
	case TabHeaders:
		return m.headerTable.View()
	case TabParams:
		return m.paramTable.View()
	case TabAuth:
		return m.renderAuthTab()
	}
	return ""
}

func (m Model) renderBodyTab() string {
	var modes []string
	for i, label := range bodyModeLabels {
		if i == m.bodyModeIdx {
			modes = append(modes, ui.Theme.TabActive.Render("["+label+"]"))
		} else {
			modes = append(modes, ui.Theme.TabInactive.Render("["+label+"]"))
		}
	}
	modeBar := "  Mode: " + strings.Join(modes, " ") + "\n\n"

	switch bodyModes[m.bodyModeIdx] {
	case types.BodyModeNone:
		return modeBar + ui.Theme.Muted.Render("  No body will be sent.")
	case types.BodyModeRawText, types.BodyModeJSON:
		return modeBar + m.bodyInput.View()
	case types.BodyModeForm:
		return modeBar + m.paramTable.View()
	}
	return modeBar
}

func (m Model) renderAuthTab() string {
	authTypes := []string{"None", "Bearer Token"}
	var parts []string
	for i, name := range authTypes {
		if i == m.authTypeIdx {
			parts = append(parts, ui.Theme.TabActive.Render("["+name+"]"))
		} else {
			parts = append(parts, ui.Theme.TabInactive.Render("["+name+"]"))
		}
	}
	header := "  Auth: " + strings.Join(parts, " ") + "\n\n"

	if m.authTypeIdx == 0 {
		return header + ui.Theme.Muted.Render("  No authentication.")
	}

	showHide := ui.Theme.Muted.Render("[ctrl+h toggle]")
	return header + "  Token: " + m.tokenInput.View() + "  " + showHide
}

func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
	m.width = w
	m.height = h
	m.urlInput.Width = w - 10
	m.bodyInput.SetWidth(w - 4)
	if h > 6 {
		m.bodyInput.SetHeight(h / 3)
	}
	m.headerTable = m.headerTable.SetWidth(w - 4)
	m.paramTable = m.paramTable.SetWidth(w - 4)
	m.tokenInput.Width = w - 20
	return m, nil
}

func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }
