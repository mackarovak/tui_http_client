package app

import (
	"htui/internal/ui"
)

func (m App) View() string {
	if !m.ready() {
		return "Loading...\n"
	}

	shell := ui.ComputeShellDimensions(m.width, m.height)

	// Каждая панель рендерит себя с рамкой нужного цвета
	sidebarView := m.renderPanel(
		m.sidebar.View(),
		m.focus == PanelSidebar,
		m.sidebar.Width(),
		m.sidebar.Height(),
	)

	editorView := m.renderPanel(
		m.editor.View(),
		m.focus == PanelEditor,
		m.editor.Width(),
		m.editor.Height(),
	)

	responseView := m.renderPanel(
		m.response.View(),
		m.focus == PanelResponse,
		m.response.Width(),
		m.response.Height(),
	)

	layout := ui.RenderShell(
		shell,
		ui.RenderLayout(m.layoutMode, sidebarView, editorView, responseView),
	)

	if m.showHelp {
		layout = ui.RenderHelpOverlay(layout, m.help.View(), m.width, m.height)
	}

	return layout
}

// renderPanel оборачивает содержимое панели в рамку (active/inactive).
func (m App) renderPanel(content string, active bool, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}


content = ui.FitBlock(content, w, h)

style := ui.Theme.InactiveBorder
if active {
    style = ui.Theme.ActiveBorder
}
return style.Render(content) } 