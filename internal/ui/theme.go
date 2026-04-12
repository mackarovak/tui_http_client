package ui

import "github.com/charmbracelet/lipgloss"

const (
	colorGreen     = lipgloss.Color("42")
	colorOrange    = lipgloss.Color("214")
	colorBlue      = lipgloss.Color("33")
	colorPurple    = lipgloss.Color("141")
	colorRed       = lipgloss.Color("196")
	colorHighlight = lipgloss.Color("62")
	colorMuted     = lipgloss.Color("240")
	colorBorder    = lipgloss.Color("238")
)

var Theme = struct {
	// Рамки панелей
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style
	ShellBorder    lipgloss.Style
	HeaderArt      lipgloss.Style
	HeaderSubtitle lipgloss.Style
	Selection      lipgloss.Style
	FocusLabel     lipgloss.Style

	// Method badges в sidebar
	MethodGET    lipgloss.Style
	MethodPOST   lipgloss.Style
	MethodPUT    lipgloss.Style
	MethodPATCH  lipgloss.Style
	MethodDELETE lipgloss.Style

	// HTTP status codes
	Status2xx lipgloss.Style
	Status3xx lipgloss.Style
	Status4xx lipgloss.Style
	Status5xx lipgloss.Style

	// Общие
	Bold        lipgloss.Style
	Muted       lipgloss.Style
	Error       lipgloss.Style
	Highlight   lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
}{
	ActiveBorder:   lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(colorHighlight),
	InactiveBorder: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(colorBorder),
	ShellBorder:    lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(colorBorder),
	HeaderArt:      lipgloss.NewStyle().Foreground(colorHighlight).Bold(true),
	HeaderSubtitle: lipgloss.NewStyle().Foreground(colorMuted),
	Selection:      lipgloss.NewStyle().Background(lipgloss.Color("236")).Bold(true),
	FocusLabel:     lipgloss.NewStyle().Foreground(colorHighlight).Bold(true),

	MethodGET:    lipgloss.NewStyle().Foreground(colorGreen).Bold(true),
	MethodPOST:   lipgloss.NewStyle().Foreground(colorOrange).Bold(true),
	MethodPUT:    lipgloss.NewStyle().Foreground(colorBlue).Bold(true),
	MethodPATCH:  lipgloss.NewStyle().Foreground(colorPurple).Bold(true),
	MethodDELETE: lipgloss.NewStyle().Foreground(colorRed).Bold(true),

	Status2xx: lipgloss.NewStyle().Foreground(colorGreen).Bold(true),
	Status3xx: lipgloss.NewStyle().Foreground(colorOrange).Bold(true),
	Status4xx: lipgloss.NewStyle().Foreground(colorRed).Bold(true),
	Status5xx: lipgloss.NewStyle().Foreground(colorRed).Bold(true),

	Bold:        lipgloss.NewStyle().Bold(true),
	Muted:       lipgloss.NewStyle().Foreground(colorMuted),
	Error:       lipgloss.NewStyle().Foreground(colorRed),
	Highlight:   lipgloss.NewStyle().Foreground(colorHighlight),
	TabActive:   lipgloss.NewStyle().Foreground(colorHighlight).Bold(true).Underline(true),
	TabInactive: lipgloss.NewStyle().Foreground(colorMuted),
}

// MethodStyle возвращает стиль для конкретного HTTP метода.
func MethodStyle(method string) lipgloss.Style {
	switch method {
	case "GET":
		return Theme.MethodGET
	case "POST":
		return Theme.MethodPOST
	case "PUT":
		return Theme.MethodPUT
	case "PATCH":
		return Theme.MethodPATCH
	case "DELETE":
		return Theme.MethodDELETE
	default:
		return Theme.Bold
	}
}

// StatusStyle возвращает стиль для HTTP status code.
func StatusStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return Theme.Status2xx
	case code >= 300 && code < 400:
		return Theme.Status3xx
	case code >= 400 && code < 500:
		return Theme.Status4xx
	case code >= 500:
		return Theme.Status5xx
	default:
		return Theme.Muted
	}
}
