package app

import (
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"htui/internal/httpclient"
	"htui/internal/requesteditor"
	"htui/internal/responsedisplay"
	"htui/internal/sidebar"
	"htui/internal/store"
	"htui/internal/types"
	"htui/internal/ui"
)

// FocusedPanel определяет, какая панель сейчас активна.
type FocusedPanel int

const (
	PanelSidebar FocusedPanel = iota
	PanelEditor
	PanelResponse
)

// Сообщения от дочерних компонентов вверх.
type ResponseReceivedMsg struct{ Data types.ResponseData }
type RequestSavedMsg struct{ Request types.SavedRequest }
type RequestCreatedMsg struct{ Request types.SavedRequest }
type RequestDeletedMsg struct{ ID string }

// Сообщения потокового (streaming) получения ответа.
type StreamStartMsg struct {
	Meta types.ResponseMeta
	Body io.ReadCloser
}
type BodyChunkMsg struct {
	Chunk      []byte
	TotalBytes int
	Done       bool
}
type ResponseCompleteMsg struct{ Data types.ResponseData }

// MaxDisplayBytes — лимит накопления тела ответа для отображения в вьюпорте.
const MaxDisplayBytes = 10 * 1024 * 1024 // 10 MB

// streamState хранит состояние активного потокового ответа.
// Указатель шарится между App и Cmd-замыканиями (goroutine-safe: читает только Cmd).
type streamState struct {
	body      io.ReadCloser
	meta      types.ResponseMeta
	bodyAccum []byte // накопленное тело для финального форматирования (до MaxDisplayBytes)
	total     int    // суммарно прочитано байт
}

// App — корневая модель приложения.
type App struct {
	// Глобальные хоткеи
	keys KeyMap

	// Подмодели
	sidebar  sidebar.Model
	editor   requesteditor.Model
	response responsedisplay.Model
	help     ui.HelpModel

	// Состояние UI
	focus      FocusedPanel
	loading    bool
	showHelp   bool
	width      int
	height     int
	layoutMode ui.LayoutMode

	// Хранилище и HTTP-клиент
	store         store.Store
	client        *httpclient.Client
	nextMethodIdx int

	// Активный стрим (nil если нет)
	stream *streamState
}

// New инициализирует приложение:
// создаёт store, сидирует демо если первый запуск, загружает запросы.
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
		store:  s,
		client: httpclient.New(),
		keys:   DefaultKeyMap, // <-- используем KeyMap из keymap.go
		focus:  PanelSidebar,
	}

	m.sidebar = sidebar.New(requests)
	m.editor = requesteditor.New()
	m.response = responsedisplay.New()
	// nil — использовать defaultSections() внутри ui.NewHelpModel,
	// но ты можешь передать сюда и конкретный map, если будешь использовать bindings.
	m.help = ui.NewHelpModel(nil)

	// Автовыбор первого запроса
	if len(requests) > 0 {
		m.editor = m.editor.LoadRequest(requests[0])
		m.sidebar = m.sidebar.SelectFirst()
	}

	return m, nil
}

func (m App) Init() tea.Cmd {
	return nil
}

// ready возвращает false пока не получены размеры терминала.
func (m App) ready() bool {
	return m.width > 0 && m.height > 0
}

// shiftFocus циклично переключает фокус между панелями.
func (m App) shiftFocus(delta int) App {
	panels := []FocusedPanel{PanelSidebar, PanelEditor, PanelResponse}
	idx := int(m.focus)
	idx = (idx + delta + len(panels)) % len(panels)
	m.focus = panels[idx]
	return m
}

// now вынесен сюда, чтобы использовать в update.go
func now() int64 {
	return time.Now().UnixNano()
}
