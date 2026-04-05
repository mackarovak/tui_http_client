package ui

import "github.com/charmbracelet/lipgloss"

var Theme = struct {
    // Рамки панелей
    ActiveBorder   lipgloss.Style
    InactiveBorder lipgloss.Style

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
    ActiveBorder:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")),
    InactiveBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")),

    MethodGET:    lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
    MethodPOST:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
    MethodPUT:    lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true),
    MethodPATCH:  lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true),
    MethodDELETE: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),

    Status2xx: lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
    Status3xx: lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
    Status4xx: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
    Status5xx: lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),

    Bold:        lipgloss.NewStyle().Bold(true),
    Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
    Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
    Highlight:   lipgloss.NewStyle().Foreground(lipgloss.Color("62")),
    TabActive:   lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).Underline(true),
    TabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
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