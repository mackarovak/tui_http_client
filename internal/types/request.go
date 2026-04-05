package types

import (
	"fmt"
	"time"
)

// BodyMode определяет способ кодирования тела запроса.
type BodyMode string

const (
	BodyModeNone    BodyMode = "none"
	BodyModeRawText BodyMode = "raw"
	BodyModeJSON    BodyMode = "json"
	BodyModeForm    BodyMode = "form"
)

// HTTPMethods — допустимые методы в правильном порядке для UI.
var HTTPMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// Header — HTTP-заголовок.
type Header struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// Param — URL query-параметр.
type Param struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// SavedRequest — полное описание сохранённого HTTP-запроса.
type SavedRequest struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Method    string     `json:"method"`
	URL       string     `json:"url"`
	Params    []Param    `json:"params"`
	Headers   []Header   `json:"headers"`
	BodyMode  BodyMode   `json:"body_mode"`
	Body      string     `json:"body"`
	Auth      AuthConfig `json:"auth"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// NewSavedRequest возвращает пустой запрос с заполненным ID и дефолтами.
func NewSavedRequest() SavedRequest {
	now := time.Now()
	return SavedRequest{
		ID:        fmt.Sprintf("%d", now.UnixNano()),
		Name:      "New Request",
		Method:    "GET",
		BodyMode:  BodyModeNone,
		Auth:      AuthConfig{Type: AuthNone},
		Params:    []Param{},
		Headers:   []Header{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
