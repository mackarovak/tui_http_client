package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"htui/internal/app"
)

func main() {
	m, err := app.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error initializing app:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error running app:", err)
		os.Exit(1)
	}
}
