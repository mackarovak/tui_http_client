package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"htui/internal/httpclient"
	"htui/internal/requesteditor"
	"htui/internal/responsedisplay"
	"htui/internal/sidebar"
	"htui/internal/store"
	"htui/internal/types"
	"htui/internal/ui"
)

// FocusedPanel определяет активную панель.
type FocusedPanel int

const (
	PanelSidebar FocusedPanel = iota
	PanelEditor
	PanelResponse
)

// Сообщения от дочерних компонентов вверх.
type ResponseReceivedMsg struct{ Data types.ResponseData }
type RequestSavedMsg struct{ Request types.SavedRequest }
type RequestDeletedMsg struct{ ID string }

// Заглушки под будущий стриминг.
type BodyChunkMsg struct{ Chunk []byte }
type ResponseCompleteMsg struct{ Data types.ResponseData }

// App — корневая модель.
type App struct {
	sidebar  sidebar.Model
	editor   requesteditor.Model
	response responsedisplay.Model
	help     ui.HelpModel

	focus      FocusedPanel
	loading    bool
	showHelp   bool
	width      int
	height     int
	layoutMode ui.LayoutMode

	store  store.Store
	client *httpclient.Client
	keys   KeyMap
}

// New инициализирует приложение.
func New() (App, error) {
	s, err := store.New()
	if err != nil {
		return App{}, err
	}

	firstRun, err := s.IsFirstRun()
	if err != nil {
		return App{}, err
	}
	if firstRun {
		for _, r := range types.DemoRequests() {
			if err := s.Save(r); err != nil {
				return App{}, err
			}
		}
		if err := s.MarkSeeded(); err != nil {
			return App{}, err
		}
	}

	requests, err := s.List()
	if err != nil {
		return App{}, err
	}

	m := App{
		store:    s,
		client:   httpclient.New(),
		keys:     DefaultKeyMap,
		focus:    PanelSidebar,
	}

	m.sidebar = sidebar.New(requests)
	m.editor = requesteditor.New()
	m.response = responsedisplay.New()
	m.help = ui.NewHelpModel(DefaultKeyMap.Bindings())

	if len(requests) > 0 {
		m.editor = m.editor.LoadRequest(requests[0])
		m.sidebar = m.sidebar.SelectFirst()
	}

	return m, nil
}

// Init запускает спиннер ответа.
func (m App) Init() tea.Cmd {
	return m.response.Init()
}

func (m App) ready() bool {
	return m.width > 0 && m.height > 0
}

func (m App) shiftFocus(delta int) App {
	panels := []FocusedPanel{PanelSidebar, PanelEditor, PanelResponse}
	idx := int(m.focus)
	idx = (idx + delta + len(panels)) % len(panels)
	m.focus = panels[idx]
	return m
}
