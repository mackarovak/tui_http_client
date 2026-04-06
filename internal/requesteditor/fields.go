package requesteditor

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "htui/internal/types"
    "htui/internal/ui"
)

type KVTable struct {
    rows    []kvRow
    cursor  int
    editCol int
    editing bool
    input   textinput.Model
    width   int
}

type kvRow struct {
    key     string
    value   string
    enabled bool
}

func NewKVTable() KVTable {
    ti := textinput.New()
    ti.CharLimit = 500
    return KVTable{
        rows:  []kvRow{{enabled: true}},
        input: ti,
    }
}

func FromHeaders(headers []types.Header) KVTable {
    t := NewKVTable()
    if len(headers) == 0 {
        return t
    }
    t.rows = make([]kvRow, len(headers))
    for i, h := range headers {
        t.rows[i] = kvRow{key: h.Key, value: h.Value, enabled: h.Enabled}
    }
    return t
}

func FromParams(params []types.Param) KVTable {
    t := NewKVTable()
    if len(params) == 0 {
        return t
    }
    t.rows = make([]kvRow, len(params))
    for i, p := range params {
        t.rows[i] = kvRow{key: p.Key, value: p.Value, enabled: p.Enabled}
    }
    return t
}

func (t KVTable) ToHeaders() []types.Header {
    var result []types.Header
    for _, r := range t.rows {
        if r.key != "" {
            result = append(result, types.Header{Key: r.key, Value: r.value, Enabled: r.enabled})
        }
    }
    return result
}

func (t KVTable) ToParams() []types.Param {
    var result []types.Param
    for _, r := range t.rows {
        if r.key != "" {
            result = append(result, types.Param{Key: r.key, Value: r.value, Enabled: r.enabled})
        }
    }
    return result
}

func (t KVTable) Update(msg tea.Msg) (KVTable, tea.Cmd) {
    if t.editing {
        switch msg := msg.(type) {
        case tea.KeyMsg:
            switch msg.String() {
            case "esc":
                t = t.commitEdit()
                t.editing = false
                return t, nil
            case "tab", "enter":
                t = t.commitEdit()
                if msg.String() == "tab" {
                    t.editCol = (t.editCol + 1) % 2
                    t = t.startEdit()
                } else {
                    t.editing = false
                }
                return t, nil
            }
        }
        var c tea.Cmd
        t.input, c = t.input.Update(msg)
        return t, c
    }

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if t.cursor > 0 {
                t.cursor--
            }
        case "down", "j":
            if t.cursor < len(t.rows)-1 {
                t.cursor++
            }
        case "enter":
            t.editCol = 0
            t = t.startEdit()
        case "tab":
            t.editCol = 1
            t = t.startEdit()
        case "a":
            t.rows = append(t.rows, kvRow{enabled: true})
            t.cursor = len(t.rows) - 1
        case "delete", "backspace":
            if len(t.rows) > 1 {
                t.rows = append(t.rows[:t.cursor], t.rows[t.cursor+1:]...)
                if t.cursor >= len(t.rows) {
                    t.cursor = len(t.rows) - 1
                }
            } else {
                t.rows[0] = kvRow{enabled: true}
            }
        case " ":
            t.rows[t.cursor].enabled = !t.rows[t.cursor].enabled
        }
    }
    return t, nil
}

func (t KVTable) View() string {
    var sb strings.Builder

    keyW := t.width/2 - 2
    valW := t.width - keyW - 4
    header := ui.Theme.Muted.Render(
        fmt.Sprintf("  %-*s  %-*s", keyW, "Key", valW, "Value"),
    )
    sb.WriteString(header + "\n")

    for i, row := range t.rows {
        prefix := "  "
        if i == t.cursor {
            prefix = ui.Theme.Highlight.Render("> ")
        }

        enabled := ui.Theme.Muted.Render("✓")
        if !row.enabled {
            enabled = ui.Theme.Muted.Render("○")
        }

        var keyStr, valStr string
        if t.editing && i == t.cursor {
            if t.editCol == 0 {
                keyStr = t.input.View()
                valStr = ui.Theme.Muted.Render(row.value)
            } else {
                keyStr = ui.Theme.Muted.Render(row.key)
                valStr = t.input.View()
            }
        } else {
            keyStr = row.key
            if keyStr == "" {
                keyStr = ui.Theme.Muted.Render("(key)")
            }
            valStr = row.value
            if valStr == "" {
                valStr = ui.Theme.Muted.Render("(value)")
            }
        }

        line := fmt.Sprintf("%s%s %-*s  %-*s", prefix, enabled, keyW, keyStr, valW, valStr)
        sb.WriteString(line + "\n")
    }

    sb.WriteString(ui.Theme.Muted.Render("\n  [a] add  [del] remove  [space] toggle  [enter] edit"))
    return sb.String()
}

func (t KVTable) SetWidth(w int) KVTable {
    t.width = w
    t.input.Width = w/2 - 4
    return t
}

func (t *KVTable) startEdit() KVTable {
    t.editing = true
    var val string
    if t.editCol == 0 {
        val = t.rows[t.cursor].key
    } else {
        val = t.rows[t.cursor].value
    }
    t.input.SetValue(val)
    t.input.Focus()
    t.input.CursorEnd()
    return *t
}

func (t *KVTable) commitEdit() KVTable {
    val := t.input.Value()
    if t.editCol == 0 {
        t.rows[t.cursor].key = val
    } else {
        t.rows[t.cursor].value = val
    }
    t.input.Blur()
    return *t
}