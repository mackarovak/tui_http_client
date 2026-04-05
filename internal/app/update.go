package app

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"htui/internal/httpclient"
	"htui/internal/requesteditor"
	"htui/internal/sidebar"
	"htui/internal/types"
	"htui/internal/ui"
)

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutMode = ui.ComputeLayout(msg.Width)
		dims := ui.ComputePanelDimensions(msg.Width, msg.Height, m.layoutMode)
		var cmds []tea.Cmd
		var c tea.Cmd
		m.sidebar, c = m.sidebar.SetSize(dims.Sidebar.Width, dims.Sidebar.Height)
		cmds = append(cmds, c)
		m.editor, c = m.editor.SetSize(dims.Editor.Width, dims.Editor.Height)
		cmds = append(cmds, c)
		m.response, c = m.response.SetSize(dims.Response.Width, dims.Response.Height)
		cmds = append(cmds, c)
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.showHelp {
			if msg.String() == "?" || msg.String() == "esc" {
				m.showHelp = false
			}
			return m, nil
		}

		switch msg.String() {
		case "tab":
			m = m.shiftFocus(+1)
			return m, nil
		case "shift+tab":
			m = m.shiftFocus(-1)
			return m, nil
		case "?":
			m.showHelp = true
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		return m.routeKeyToFocusedPanel(msg)

	case ResponseReceivedMsg:
		m.loading = false
		m.response = m.response.SetResponse(msg.Data)
		m.focus = PanelResponse
		return m, nil

	case sidebar.RequestSelectedMsg:
		m.editor = m.editor.LoadRequest(msg.Request)
		m.focus = PanelEditor
		return m, nil

	case sidebar.NewRequestMsg:
		m.editor = m.editor.LoadRequest(types.NewSavedRequest())
		m.focus = PanelEditor
		return m, nil

	case sidebar.DeleteRequestMsg:
		return m, deleteRequestCmd(m.store, msg.ID)

	case sidebar.DuplicateRequestMsg:
		dup := msg.Request
		dup.ID = fmt.Sprintf("%d", now())
		dup.Name = dup.Name + " (copy)"
		return m, saveRequestCmd(m.store, dup)

	case requesteditor.SendRequestMsg:
		if m.loading {
			return m, nil
		}
		m.loading = true
		m.response = m.response.SetLoading(true)
		return m, sendRequestCmd(m.client, msg.Request)

	case requesteditor.SaveRequestMsg:
		return m, saveRequestCmd(m.store, msg.Request)

	case RequestSavedMsg:
		m.sidebar = m.sidebar.Reload(m.store)
		return m, nil

	case RequestDeletedMsg:
		m.sidebar = m.sidebar.Reload(m.store)
		if m.editor.CurrentID() == msg.ID {
			m.editor = m.editor.Clear()
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.response, cmd = m.response.UpdateSpinner(msg)
		return m, cmd
	}

	return m.routeToFocusedPanel(msg)
}

func (m App) routeKeyToFocusedPanel(msg tea.KeyMsg) (App, tea.Cmd) {
	return m.routeToFocusedPanel(msg)
}

func (m App) routeToFocusedPanel(msg tea.Msg) (App, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case PanelSidebar:
		m.sidebar, cmd = m.sidebar.Update(msg)
	case PanelEditor:
		m.editor, cmd = m.editor.Update(msg)
	case PanelResponse:
		m.response, cmd = m.response.Update(msg)
	}
	return m, cmd
}

func sendRequestCmd(client *httpclient.Client, req types.SavedRequest) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), httpclient.DefaultTimeout)
		defer cancel()
		result := client.Execute(ctx, req)
		return ResponseReceivedMsg{Data: result}
	}
}

func saveRequestCmd(s interface{ Save(types.SavedRequest) error }, r types.SavedRequest) tea.Cmd {
	return func() tea.Msg {
		_ = s.Save(r)
		return RequestSavedMsg{Request: r}
	}
}

func deleteRequestCmd(s interface{ Delete(string) error }, id string) tea.Cmd {
	return func() tea.Msg {
		_ = s.Delete(id)
		return RequestDeletedMsg{ID: id}
	}
}

func now() int64 {
	return time.Now().UnixNano()
}
