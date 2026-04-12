//go:build ignore

package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	s := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Width(10).Height(5).Render("x")
	fmt.Println("w", lipgloss.Width(s), "h", lipgloss.Height(s))
}
