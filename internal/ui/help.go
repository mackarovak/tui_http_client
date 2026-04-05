package ui

import "github.com/charmbracelet/lipgloss"

// HelpModel — простая заглушка для help overlay.
type HelpModel struct{}

// NewHelpModel принимает KeyMap из app, но мы не используем его внутри.
func NewHelpModel(_ interface{}) HelpModel {
    return HelpModel{}
}

func (h HelpModel) View() string {
    return "Help\n\nPress ? or Esc to close."
}

// RenderHelpOverlay рисует рамку поверх layout.
func RenderHelpOverlay(bg, helpView string, width, height int) string {
    box := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62")).
        Padding(1, 2).
        Width(width/2).
        Height(height/2).
        Render(helpView)

    // Очень простая версия — просто вертикально склеиваем
    return lipgloss.JoinVertical(lipgloss.Left, bg, box)
}