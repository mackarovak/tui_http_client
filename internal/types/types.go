package types

import (
	"fmt"
	"strconv"
	"time"
)

// BodyMode задаёт способ отправки тела запроса.
type BodyMode string

const (
	BodyModeNone    BodyMode = "none"
	BodyModeRawText BodyMode = "raw"
	BodyModeJSON    BodyMode = "json"
	BodyModeForm    BodyMode = "form"
)

// AuthType — режим аутентификации.
type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthBearer AuthType = "bearer"
)

// HTTPMethods — поддерживаемые методы (порядок совпадает с UI).
var HTTPMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// Header — заголовок запроса.
type Header struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// Param — query / form пара.
type Param struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// AuthConfig — настройки auth.
type AuthConfig struct {
	Type         AuthType `json:"type"`
	Token        string   `json:"token,omitempty"`
	TokenVisible bool     `json:"token_visible,omitempty"`
}

// SavedRequest — сохранённый запрос.
type SavedRequest struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Method   string     `json:"method"`
	URL      string     `json:"url"`
	Params   []Param    `json:"params"`
	Headers  []Header   `json:"headers"`
	BodyMode BodyMode   `json:"body_mode"`
	Body     string     `json:"body"`
	Auth     AuthConfig `json:"auth"`
}

// ResponseData — результат выполнения HTTP-запроса.
type ResponseData struct {
	StatusCode int
	StatusText string
	DurationMs int64
	SizeBytes  int
	Headers    []Header
	Body       string
	Error      string
}

// NewSavedRequest — пустой черновик.
func NewSavedRequest() SavedRequest {
	return SavedRequest{
		ID:       strconv.FormatInt(time.Now().UnixNano(), 10),
		Name:     "New request",
		Method:   "GET",
		BodyMode: BodyModeNone,
		Auth:     AuthConfig{Type: AuthNone},
	}
}

// DemoRequests — три демо-запроса при первом запуске.
func DemoRequests() []SavedRequest {
	return []SavedRequest{
		{
			ID:     "demo-get-posts-1",
			Name:   "Simple GET",
			Method: "GET",
			URL:    "https://jsonplaceholder.typicode.com/posts/1",
		},
		{
			ID:       "demo-create-post",
			Name:     "Create Post",
			Method:   "POST",
			URL:      "https://jsonplaceholder.typicode.com/posts",
			BodyMode: BodyModeJSON,
			Body: `{
  "title": "hello",
  "body": "world",
  "userId": 1
}`,
			Headers: []Header{
				{Key: "Content-Type", Value: "application/json", Enabled: true},
			},
		},
		{
			ID:     "demo-bearer-example",
			Name:   "Example with token",
			Method: "GET",
			URL:    "https://jsonplaceholder.typicode.com/posts/1",
			Auth: AuthConfig{
				Type:  AuthBearer,
				Token: "replace-with-your-token",
			},
		},
	}
}

// EnsureID задаёт ID если пусто (после загрузки из файла).
func (r *SavedRequest) EnsureID() {
	if r.ID == "" {
		r.ID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
}

// DisplayName для списка.
func (r SavedRequest) DisplayName() string {
	if r.Name != "" {
		return r.Name
	}
	return fmt.Sprintf("Request %s", r.ID)
}
