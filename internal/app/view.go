package app

import "htui/internal/ui"

func (m App) View() string {
	if !m.ready() {
		return "Loading...\n"
	}

	sidebarView := m.renderPanel(m.sidebar.View(), m.focus == PanelSidebar, m.sidebar.Width(), m.sidebar.Height())
	editorView := m.renderPanel(m.editor.View(), m.focus == PanelEditor, m.editor.Width(), m.editor.Height())
	responseView := m.renderPanel(m.response.View(), m.focus == PanelResponse, m.response.Width(), m.response.Height())

	layout := ui.RenderLayout(m.layoutMode, sidebarView, editorView, responseView)

	if m.showHelp {
		layout = ui.RenderHelpOverlay(layout, m.help.View(), m.width, m.height)
	}

	return layout
}

func (m App) renderPanel(content string, active bool, w, h int) string {
	style := ui.Theme.InactiveBorder
	if active {
		style = ui.Theme.ActiveBorder
	}
	return style.Width(w).Height(h).Render(content)
}
