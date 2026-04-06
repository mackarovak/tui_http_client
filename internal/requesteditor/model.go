package requesteditor

import (
    "strings"

    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "htui/internal/types"
    "htui/internal/ui"
)

// EditorTab — вкладка в редакторе запроса.
type EditorTab int

const (
    TabBody EditorTab = iota
    TabHeaders
    TabParams
    TabAuth
)

var tabNames = []string{"Body", "Headers", "Params", "Auth"}

// FocusField — текущий активный элемент внутри editor.
type FocusField int

const (
    FieldMethod FocusField = iota
    FieldURL
    FieldTabContent // всё что ниже tab bar
)

// --- Сообщения вверх в App ---

type SendRequestMsg struct{ Request types.SavedRequest }
type SaveRequestMsg struct{ Request types.SavedRequest }

// --- Модель ---

type Model struct {
    current types.SavedRequest
    dirty   bool

    // Поля
    urlInput    textinput.Model
    bodyInput   textarea.Model
    headerTable KVTable
    paramTable  KVTable
    bodyModeIdx int // индекс в []BodyMode

    // Auth
    authTypeIdx  int // 0=none, 1=bearer
    tokenInput   textinput.Model
    tokenVisible bool

    // UI-состояние
    activeTab      EditorTab
    focusField     FocusField
    methodIdx      int // индекс в types.HTTPMethods
    focused        bool
    width          int
    height         int
    validationErrs []string // показываются под URL при ошибке
}

var bodyModes = []types.BodyMode{
    types.BodyModeNone,
    types.BodyModeRawText,
    types.BodyModeJSON,
    types.BodyModeForm,
}

var bodyModeLabels = []string{"none", "raw", "json", "form"}

// New создаёт пустой редактор.
func New() Model {
    url := textinput.New()
    url.Placeholder = "https://api.example.com/endpoint"
    url.CharLimit = 2000
    url.Focus()

    body := textarea.New()
    body.Placeholder = "Request body..."
    body.CharLimit = 0 // без ограничений
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

// LoadRequest загружает существующий запрос в редактор.
func (m Model) LoadRequest(r types.SavedRequest) Model {
    m.current = r
    m.dirty = false

    // URL
    m.urlInput.SetValue(r.URL)

    // Method
    m.methodIdx = 0
    for i, method := range types.HTTPMethods {
        if method == r.Method {
            m.methodIdx = i
            break
        }
    }

    // Body mode
    m.bodyModeIdx = 0
    for i, bm := range bodyModes {
        if bm == r.BodyMode {
            m.bodyModeIdx = i
            break
        }
    }
    m.bodyInput.SetValue(r.Body)

    // Headers / Params
    m.headerTable = FromHeaders(r.Headers)
    m.paramTable = FromParams(r.Params)

    // Auth
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

// Clear сбрасывает редактор в пустое состояние.
func (m Model) Clear() Model {
    return New()
}

// CurrentID возвращает ID текущего запроса (для сравнения при удалении).
func (m Model) CurrentID() string {
    return m.current.ID
}

// BuildRequest собирает types.SavedRequest из текущего состояния полей.
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
        // Отправка по Enter
        case "enter":
            req := m.BuildRequest()
            result := Validate(req)
            if !result.Valid {
                m.validationErrs = result.Errors
                return m, nil
            }
            m.validationErrs = nil
            return m, func() tea.Msg { return SendRequestMsg{Request: req} }

        // Сохранение всегда как новый
        case "ctrl+s":
            req := m.BuildRequest()
            // сбрасываем ID, чтобы storage создал новый запрос
            req.ID = ""
            m.dirty = false
            return m, func() tea.Msg { return SaveRequestMsg{Request: req} }

        case "up":
            switch m.focusField {
            case FieldURL:
                m.focusField = FieldMethod
                return m, nil
            case FieldTabContent:
                m.activeTab = EditorTab((int(m.activeTab) - 1 + len(tabNames)) % len(tabNames))
                m = m.focusTabContent()
                return m, nil
            }
        case "down":
            switch m.focusField {
            case FieldMethod:
                m.focusField = FieldURL
                m.urlInput.Focus()
                return m, nil
            case FieldURL:
                m.urlInput.Blur()
                m.focusField = FieldTabContent
                m = m.focusTabContent()
                return m, nil
            case FieldTabContent:
                m.activeTab = EditorTab((int(m.activeTab) + 1) % len(tabNames))
                m = m.focusTabContent()
                return m, nil
            }
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
            if m.focusField == FieldTabContent && m.activeTab == TabBody {
                switch msg.String() {
                case "left":
                    if m.bodyModeIdx > 0 {
                        m.bodyModeIdx--
                        m.dirty = true
                        m = m.focusTabContent()
                    }
                    return m, nil
                case "right":
                    if m.bodyModeIdx < len(bodyModes)-1 {
                        m.bodyModeIdx++
                        m.dirty = true
                        m = m.focusTabContent()
                    }
                    return m, nil
                }
            }
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

// routeToActiveField передаёт сообщение в текущий активный subcomponent.
func (m Model) routeToActiveField(msg tea.Msg) (Model, tea.Cmd) {
    var cmd tea.Cmd

    switch m.focusField {
    case FieldURL:
        m.urlInput, cmd = m.urlInput.Update(msg)
        m.dirty = true

    case FieldTabContent:
        switch m.activeTab {
        case TabBody:
            if msg, ok := msg.(tea.KeyMsg); ok {
                switch msg.String() {
                case "down":
                    if bodyModes[m.bodyModeIdx] != types.BodyModeNone {
                        if bodyModes[m.bodyModeIdx] == types.BodyModeJSON ||
                            bodyModes[m.bodyModeIdx] == types.BodyModeRawText {
                            m.bodyInput.Focus()
                        } else if bodyModes[m.bodyModeIdx] == types.BodyModeForm {
                            // Фокус на таблицу параметров
                        }
                    }
                    return m, nil
                }
            }
            if bodyModes[m.bodyModeIdx] == types.BodyModeJSON ||
                bodyModes[m.bodyModeIdx] == types.BodyModeRawText {
                m.bodyInput, cmd = m.bodyInput.Update(msg)
                m.dirty = true
            } else if bodyModes[m.bodyModeIdx] == types.BodyModeForm {
                m.paramTable, cmd = m.paramTable.Update(msg)
                m.dirty = true
            }

        case TabHeaders:
            m.headerTable, cmd = m.headerTable.Update(msg)
            m.dirty = true

        case TabParams:
            m.paramTable, cmd = m.paramTable.Update(msg)
            m.dirty = true

        case TabAuth:
            m = m.updateAuthTab(msg)
        }
    }
    return m, cmd
}

func (m Model) updateAuthTab(msg tea.Msg) Model {
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
            return m
        case "ctrl+h":
            m.tokenVisible = !m.tokenVisible
            if m.tokenVisible {
                m.tokenInput.EchoMode = textinput.EchoNormal
            } else {
                m.tokenInput.EchoMode = textinput.EchoPassword
            }
            return m
        }
    }
    if m.authTypeIdx == 1 {
        m.tokenInput, cmd = m.tokenInput.Update(msg)
        _ = cmd
        m.dirty = true
    }
    return m
}

func (m Model) focusTabContent() Model {
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
    return m
}

func (m Model) View() string {
    if m.width == 0 {
        return ""
    }

    var sb strings.Builder

    methodLine := m.renderMethodSelector()
    sb.WriteString(methodLine + "\n")

    urlLine := m.renderURLField()
    sb.WriteString(urlLine + "\n")

    if len(m.validationErrs) > 0 {
        for _, e := range m.validationErrs {
            sb.WriteString(ui.Theme.Error.Render("  ✗ "+e) + "\n")
        }
    }

    sb.WriteString("\n")

    sb.WriteString(m.renderTabBar() + "\n\n")

    sb.WriteString(m.renderTabContent())

    hint := ui.Theme.Muted.Render("  enter send  ctrl+s save as new")
    if m.dirty {
        hint = ui.Theme.Muted.Render("  enter send  ctrl+s save as new  ") +
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
    return header +
        "  Token: " + m.tokenInput.View() + "  " + showHide
}

// SetSize устанавливает размеры компонента.
func (m Model) SetSize(w, h int) (Model, tea.Cmd) {
    m.width = w
    m.height = h
    m.urlInput.Width = w - 10
    m.bodyInput.SetWidth(w - 4)
    m.bodyInput.SetHeight(h / 3)
    m.headerTable = m.headerTable.SetWidth(w - 4)
    m.paramTable = m.paramTable.SetWidth(w - 4)
    m.tokenInput.Width = w - 20
    return m, nil
}

func (m Model) Width() int  { return m.width }
func (m Model) Height() int { return m.height }