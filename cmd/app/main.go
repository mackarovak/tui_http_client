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
		fmt.Fprintf(os.Stderr, "startup: %v\n", err)
		os.Exit(1)
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run: %v\n", err)
		os.Exit(1)
	}
}
