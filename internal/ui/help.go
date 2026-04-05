package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel — текст для overlay помощи (без импорта app: биндинги передаются снаружи).
type HelpModel struct {
	bindings []key.Binding
}

// NewHelpModel собирает модель из списка горячих клавиш.
func NewHelpModel(bindings []key.Binding) HelpModel {
	return HelpModel{bindings: bindings}
}

// View рисует таблицу подсказок.
func (h HelpModel) View() string {
	var b strings.Builder
	b.WriteString(Theme.Bold.Render("Keyboard shortcuts"))
	b.WriteString("\n\n")
	for _, bind := range h.bindings {
		help := bind.Help()
		line := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(20).Render(help.Key),
			help.Desc,
		)
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(Theme.Muted.Render("esc or ? — close"))
	return b.String()
}

// RenderHelpOverlay центрирует окно помощи поверх текущей вёрстки.
func RenderHelpOverlay(base, help string, w, h int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		MaxWidth(w - 4).
		Render(help)

	if w <= 0 || h <= 0 {
		return base + "\n" + box
	}
	_ = base
	// lipgloss v1.0.0: без PlaceOverlay — полноэкранное центрирование окна помощи
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}
