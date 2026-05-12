package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

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
			m.showHelp = true
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			m = m.cancelStream()
			return m, tea.Quit
		}

		// Роутинг к сфокусированной панели
		return m.routeKeyToFocusedPanel(msg)

	// --- Ответ получен полностью (ошибка сети) ---
	case ResponseReceivedMsg:
		m.loading = false
		m.response = m.response.SetResponse(msg.Data)
		m.focus = PanelResponse
		return m, nil

	// --- Потоковые сообщения ---

	case StreamStartMsg:
		m.loading = false
		m.stream = &streamState{body: msg.Body, meta: msg.Meta}
		m.response = m.response.SetStreamingMeta(msg.Meta)
		return m, readChunkCmd(m.stream)

	case BodyChunkMsg:
		if m.stream != nil {
			// Накапливаем тело (до MaxDisplayBytes) для финального форматирования
			if !m.stream.meta.IsBinary && len(m.stream.bodyAccum) < MaxDisplayBytes {
				m.stream.bodyAccum = append(m.stream.bodyAccum, msg.Chunk...)
			}
			m.response = m.response.AppendChunk(msg.Chunk, msg.TotalBytes)
		}
		if msg.Done {
			if m.stream != nil {
				data := buildResponseData(m.stream)
				m.response = m.response.FinalizeStream(data)
				m.stream = nil
			}
			return m, nil
		}
		if m.stream != nil {
			return m, readChunkCmd(m.stream)
		}
		return m, nil

	case ResponseCompleteMsg:
		// Зарезервировано для совместимости, не используется в текущей реализации
		return m, nil

	// --- Управление запросами ---

	case sidebar.RequestSelectedMsg:
		m = m.cancelStream()
		m.editor = m.editor.LoadRequest(msg.Request)
		m.focus = PanelEditor
		return m, nil

	case sidebar.NewRequestMsg:
		methods := types.HTTPMethods
		method := methods[m.nextMethodIdx%len(methods)]
		m.nextMethodIdx++

		req := types.SavedRequest{
			ID:     fmt.Sprintf("%d", now()),
			Name:   method + " request",
			Method: method,
			URL:    "https://example.com",
		}
		return m, createRequestCmd(m.store, req)

	case sidebar.NewRequestWithMethodMsg:
		req := types.SavedRequest{
			ID:     fmt.Sprintf("%d", now()),
			Name:   msg.Method + " request",
			Method: msg.Method,
			URL:    "https://example.com",
		}
		return m, createRequestCmd(m.store, req)

	case sidebar.TemplateSelectedMsg:
		m = m.cancelStream()
		newReq := msg.Template
		newReq.ID = fmt.Sprintf("%d", now())
		newReq.IsTemplate = false
		newReq.CreatedAt = time.Time{}
		newReq.UpdatedAt = time.Time{}
		m.editor = m.editor.LoadRequest(newReq)
		m.editor = m.editor.MarkDirty()
		m.focus = PanelEditor
		return m, nil

	case RequestCreatedMsg:
		m = m.cancelStream()
		m.sidebar = m.sidebar.Reload(m.store)
		m.editor = m.editor.LoadRequest(msg.Request)
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
		m = m.cancelStream()
		m.loading = true
		m.response = m.response.SetLoading(true)
		return m, tea.Batch(m.response.SpinnerCmd(), sendRequestCmd(m.client, msg.Request))

	case requesteditor.SaveRequestMsg:
		return m, saveRequestCmd(m.store, msg.Request)

	case requesteditor.SaveAsTemplateMsg:
		return m, saveTemplateCmd(m.store, msg.Request)

	case RequestSavedMsg:
		m.sidebar = m.sidebar.Reload(m.store)
		m.editor = m.editor.Clear()
		return m, nil

	case TemplateSavedMsg:
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

	// Остальные сообщения — только в сфокусированную панель
	return m.routeToFocusedPanel(msg)
}

// cancelStream закрывает активный стрим (если есть) и обнуляет состояние.
func (m App) cancelStream() App {
	if m.stream != nil {
		_ = m.stream.body.Close()
		m.stream = nil
		m.loading = false
		m.response = m.response.CancelStream()
	}
	return m
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

// --- tea.Cmd фабрики ---

// sendRequestCmd запускает HTTP запрос и возвращает StreamStartMsg с метаданными
// и открытым телом ответа. При сетевой ошибке возвращает ResponseReceivedMsg.
func sendRequestCmd(client *httpclient.Client, req types.SavedRequest) tea.Cmd {
	return func() tea.Msg {
		resp, dur, err := client.Start(context.Background(), req)
		if err != nil {
			return ResponseReceivedMsg{Data: types.ResponseData{
				DurationMs: dur.Milliseconds(),
				Error:      httpclient.MapError(err),
			}}
		}
		meta := httpclient.BuildResponseMeta(resp, dur)
		return StreamStartMsg{Meta: meta, Body: resp.Body}
	}
}

// readChunkCmd читает один чанк из тела ответа и возвращает BodyChunkMsg.
func readChunkCmd(s *streamState) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, httpclient.ChunkSize)
		n, err := s.body.Read(buf)
		s.total += n

		if n > 0 {
			done := errors.Is(err, io.EOF)
			if done {
				_ = s.body.Close()
			}
			return BodyChunkMsg{Chunk: buf[:n], TotalBytes: s.total, Done: done}
		}

		// n == 0: EOF или ошибка (в том числе body.Close() от cancelStream)
		_ = s.body.Close()
		return BodyChunkMsg{TotalBytes: s.total, Done: true}
	}
}

// buildResponseData собирает финальный ResponseData из накопленного стрима.
func buildResponseData(s *streamState) types.ResponseData {
	return types.ResponseData{
		StatusCode: s.meta.StatusCode,
		StatusText: s.meta.StatusText,
		DurationMs: s.meta.DurationMs,
		SizeBytes:  s.total,
		Headers:    s.meta.Headers,
		Body:       string(s.bodyAccum),
		IsBinary:   s.meta.IsBinary,
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

func saveTemplateCmd(s interface {
	Save(types.SavedRequest) error
}, r types.SavedRequest) tea.Cmd {
	return func() tea.Msg {
		_ = s.Save(r)
		return TemplateSavedMsg{Request: r}
	}
}

func createRequestCmd(s interface {
	Save(types.SavedRequest) error
}, r types.SavedRequest) tea.Cmd {
	return func() tea.Msg {
		_ = s.Save(r)
		return RequestCreatedMsg{Request: r}
	}
}

func deleteRequestCmd(s interface{ Delete(string) error }, id string) tea.Cmd {
	return func() tea.Msg {
		_ = s.Delete(id)
		return RequestDeletedMsg{ID: id}
	}
}
