package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
)

// helpRow — одна строка горячей клавиши.
type helpRow struct {
    key  string
    desc string
}

// helpSection — секция в таблице помощи.
type helpSection struct {
    title string
    rows  []helpRow
}

// HelpModel — модель экрана помощи.
// Принимает биндинги как данные (не импортирует app.KeyMap напрямую).
type HelpModel struct {
    sections []helpSection
}

// NewHelpModel создаёт HelpModel из стандартных биндингов.
// Вызывается из app.New() с конкретными строками биндингов.
func NewHelpModel(bindings map[string]string) HelpModel {
    // bindings пока не используем, берём захардкоженные секции
    return HelpModel{
        sections: defaultSections(),
    }
}

// defaultSections возвращает секции по умолчанию.
// Содержимое синхронизировано с app.DefaultKeyMap.
func defaultSections() []helpSection {
    return []helpSection{
        {
            title: "Navigation",
            rows: []helpRow{
                {"tab", "next panel"},
                {"shift+tab", "prev panel"},
                {"q / ctrl+c", "quit"},
                {"?", "close help"},
            },
        },
        {
            title: "Request Editor",
            rows: []helpRow{
                {"ctrl+enter", "send request"},
                {"ctrl+s", "save request"},
                {"n", "new request"},
                {"d", "duplicate"},
                {"del", "delete request"},
                {"/", "search requests"},
            },
        },
        {
            title: "Sidebar",
            rows: []helpRow{
                {"↑/↓, j/k", "navigate list"},
                {"enter", "open request"},
                {"/", "search/filter"},
                {"esc", "cancel search"},
            },
        },
        {
            title: "Response",
            rows: []helpRow{
                {"j/k, ↑/↓", "scroll"},
                {"PgUp/PgDn", "fast scroll"},
                {"tab", "body ↔ headers"},
            },
        },
        {
            title: "Request Body (KV Table)",
            rows: []helpRow{
                {"a", "add row"},
                {"del", "delete row"},
                {"enter", "edit cell"},
                {"tab", "next column"},
                {"space", "toggle enabled"},
            },
        },
    }
}

// View рендерит содержимое overlay (без рамки — рамка добавляется в RenderHelpOverlay).
func (m HelpModel) View() string {
    var sb strings.Builder

    sb.WriteString(Theme.Bold.Render("Keyboard Shortcuts") + "\n\n")

    // Рендерим секции в два столбца
    half := (len(m.sections) + 1) / 2
    left := m.sections[:half]
    right := m.sections[half:]

    leftStr := renderSections(left, 38)
    rightStr := renderSections(right, 38)

    combined := lipgloss.JoinHorizontal(lipgloss.Top,
        lipgloss.NewStyle().Width(42).Render(leftStr),
        lipgloss.NewStyle().Width(42).Render(rightStr),
    )
    sb.WriteString(combined)

    sb.WriteString("\n\n" + Theme.Muted.Render("Press ? or esc to close"))
    return sb.String()
}

func renderSections(sections []helpSection, width int) string {
    var sb strings.Builder
    for _, sec := range sections {
        sb.WriteString(Theme.Highlight.Render(sec.title) + "\n")
        for _, row := range sec.rows {
            keyStr := Theme.Bold.Render(fmt.Sprintf("  %-16s", row.key))
            descStr := Theme.Muted.Render(row.desc)
            sb.WriteString(keyStr + descStr + "\n")
        }
        sb.WriteString("\n")
    }
    return sb.String()
}

// RenderHelpOverlay накладывает overlay поверх base layout.
// Использует lipgloss.Place для центрирования, затемняет фон.
func RenderHelpOverlay(base, content string, w, h int) string {
    box := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 3).
        Background(lipgloss.Color("235")).
        Render(content)

    // Накладываем overlay по центру экрана
    return lipgloss.Place(
        w, h,
        lipgloss.Center,
        lipgloss.Center,
        box,
        lipgloss.WithWhitespaceBackground(lipgloss.Color("236")),
    )
}