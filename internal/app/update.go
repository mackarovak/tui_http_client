package app

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
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
		shell := ui.ComputeShellDimensions(msg.Width, msg.Height)
		dims := ui.ComputePanelDimensions(shell.BodyWidth, shell.BodyHeight, m.layoutMode)

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
		// Help overlay перехватывает все клавиши пока открыт
		if m.showHelp {
			if msg.String() == "?" || msg.String() == "esc" {
				m.showHelp = false
			}
			return m, nil
		}

		// Глобальные клавиши (до роутинга к панелям)
		switch {
		case key.Matches(msg, m.keys.NextPanel):
			m = m.shiftFocus(+1)
			return m, nil
		case key.Matches(msg, m.keys.PrevPanel):
			m = m.shiftFocus(-1)
			return m, nil
		case key.Matches(msg, m.keys.Help):
			// при повторном нажатии "?" удобно тоже закрывать,
			// но по ТЗ достаточно открыть, а закрытие уже перехватит блок выше
			m.showHelp = true
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

		// Роутинг к сфокусированной панели
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
		// список методов, которые хотим перебирать
		methods := []string{"GET", "POST", "PUT", "DELETE"}

		method := methods[m.nextMethodIdx%len(methods)]
		m.nextMethodIdx++

		req := types.SavedRequest{
			ID:     fmt.Sprintf("%d", now()),
			Name:   method + " request",
			Method: method,
			URL:    "https://example.com",
		}
		return m, saveRequestCmd(m.store, req)

	case sidebar.NewRequestWithMethodMsg:
		// новый запрос с заданным методом (POST/PUT/PATCH и т.п.)
		req := types.SavedRequest{
			ID:     fmt.Sprintf("%d", now()),
			Name:   msg.Method + " request",
			Method: msg.Method,
			URL:    "https://example.com",
		}
		return m, saveRequestCmd(m.store, req)

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
		return m, tea.Batch(m.response.SpinnerCmd(), sendRequestCmd(m.client, msg.Request))

	case requesteditor.SaveRequestMsg:
		return m, saveRequestCmd(m.store, msg.Request)

	case RequestSavedMsg:
		// после любого Save — перезагружать только сайдбар
		m.sidebar = m.sidebar.Reload(m.store)
		// Очистить редактор для нового запроса
		m.editor = m.editor.Clear()
		return m, nil

	case RequestDeletedMsg:
		m.sidebar = m.sidebar.Reload(m.store)
		// Если удалили текущий открытый запрос — очистить editor
		if m.editor.CurrentID() == msg.ID {
			m.editor = m.editor.Clear()
		}
		return m, nil

	case spinner.TickMsg:
		// Spinner всегда обновляется (независимо от фокуса)
		var cmd tea.Cmd
		m.response, cmd = m.response.UpdateSpinner(msg)
		return m, cmd
	}

	// Остальные сообщения — только в сфокусированную панель
	return m.routeToFocusedPanel(msg)
}

// routeKeyToFocusedPanel направляет KeyMsg в активную панель.
func (m App) routeKeyToFocusedPanel(msg tea.KeyMsg) (App, tea.Cmd) {
	return m.routeToFocusedPanel(msg)
}

// routeToFocusedPanel направляет любое сообщение в активную панель.
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

// --- tea.Cmd фабрики (единственное место где httpclient встречает bubbletea) ---

func sendRequestCmd(client *httpclient.Client, req types.SavedRequest) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), httpclient.DefaultTimeout)
		defer cancel()
		result := client.Execute(ctx, req)
		return ResponseReceivedMsg{Data: result}
	}
}

func saveRequestCmd(s interface {
	Save(types.SavedRequest) error
}, r types.SavedRequest) tea.Cmd {
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
