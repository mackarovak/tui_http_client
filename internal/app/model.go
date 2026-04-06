package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"htui/internal/store"
	"htui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

// App — корневая модель приложения.
type App struct {
	requests []types.SavedRequest // все запросы приложения
	editor   types.SavedRequest   // текущий выбранный запрос
	sidebar  []string             // имена запросов для бокового меню
	response string               // последний полученный ответ
}

// New инициализирует приложение.
func New() (App, error) {
	// создаём store
	s, err := store.New()
	if err != nil {
		return App{}, err
	}

	// проверяем первый запуск
	firstRun, err := s.IsFirstRun()
	if err != nil {
		return App{}, err
	}

	if firstRun {
		// сидируем демо-запросы
		for _, r := range types.DemoRequests() {
			if err := s.Save(r); err != nil {
				fmt.Println("Ошибка сохранения демо-запроса:", err)
			}
		}
		s.MarkSeeded() // создаём файл .seeded
	}

	// загружаем все запросы из store
	requests, err := s.List()
	if err != nil {
		return App{}, err
	}

	// создаём sidebar — просто имена запросов
	sidebar := make([]string, len(requests))
	for i, r := range requests {
		sidebar[i] = r.Name
	}

	// выбираем первый запрос в editor
	var editor types.SavedRequest
	if len(requests) > 0 {
		editor = requests[0]
	}

	return App{
		requests: requests,
		editor:   editor,
		sidebar:  sidebar,
	}, nil
}

func (m App) Init() tea.Cmd {
	return nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// 📥 Получили ответ от HTTP-запроса
	case string:
		m.response = msg
		return m, nil

	// ⌨️ Обработка клавиш
	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

			// асинхронный запрос (не блокирует UI)
			//case "ctrl+enter":
		case "enter":
			return m, func() tea.Msg {
				respText, err := sendRequest(m.editor)
				if err != nil {
					return fmt.Sprintf("Ошибка: %v", err)
				}
				return respText
			}

		}
	}

	return m, nil
}

// sendRequest выполняет HTTP-запрос для переданного SavedRequest
func sendRequest(r types.SavedRequest) (string, error) {
	var body io.Reader
	if r.BodyMode == "json" && r.Body != "" {
		body = bytes.NewBuffer([]byte(r.Body))
	}

	req, err := http.NewRequest(r.Method, r.URL, body)
	if err != nil {
		return "", err
	}

	// заголовки
	for _, h := range r.Headers {
		req.Header.Set(h.Key, h.Value)
	}

	// auth — только Bearer токен
	if strings.ToLower(string(r.Auth.Type)) == "bearer" && r.Auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Auth.Token)
	}

	// параметры GET/POST
	q := req.URL.Query()
	for _, p := range r.Params {
		q.Add(p.Key, p.Value)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBytes), nil
}

func (m App) View() string {
	view := "htui — TUI HTTP Client\n\n"

	// Sidebar
	view += "Requests:\n"
	for _, name := range m.sidebar {
		if m.editor.Name == name {
			view += fmt.Sprintf(" → %s\n", name) // текущий выбранный запрос
		} else {
			view += fmt.Sprintf("   %s\n", name)
		}
	}

	// Editor
	view += "\nEditor:\n"
	view += fmt.Sprintf("Name: %s\n", m.editor.Name)
	view += fmt.Sprintf("Method: %s\n", m.editor.Method)
	view += fmt.Sprintf("URL: %s\n", m.editor.URL)
	view += fmt.Sprintf("BodyMode: %s\n", m.editor.BodyMode)
	if m.editor.BodyMode == "json" && m.editor.Body != "" {
		view += fmt.Sprintf("Body: %s\n", m.editor.Body)
	}

	// Response
	if m.response != "" {
		view += "\nResponse:\n"
		view += m.response + "\n"
	}

	view += "\nPress q to quit | enter to send request\n"
	//view += "\nPress q to quit | ctrl+enter to send request\n"
	return view
}
