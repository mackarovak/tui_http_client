package app

import tea "github.com/charmbracelet/bubbletea"

// App — корневая модель приложения (stub).
type App struct{}

// New инициализирует приложение.
func New() (App, error) {
	return App{}, nil
}

func (m App) Init() tea.Cmd {
	return nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m App) View() string {
	return "htui — TUI HTTP Client\n\nPress q to quit\n"
}
