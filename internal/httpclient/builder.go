package httpclient

import (
	"encoding/json"
	"fmt"
	"htui/internal/types"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Build конструирует *http.Request из SavedRequest.
// Выполняет: валидацию URL, сборку query-параметров,
// кодирование body по BodyMode, инжект заголовков и auth.
func Build(r types.SavedRequest) (*http.Request, error) {
	// 1. Нормализация URL: добавить схему если отсутствует
	rawURL := r.URL
	if rawURL == "" {
		return nil, fmt.Errorf("URL is required")
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	// 2. Собрать итоговый URL с query-параметрами
	finalURL, err := buildURL(rawURL, r.Params)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// 3. Подготовить body
	body, contentType, err := buildBody(r)
	if err != nil {
		return nil, err
	}

	// 4. Создать запрос
	req, err := http.NewRequest(r.Method, finalURL, body)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	// 5. Content-Type (если тело есть и пользователь не переопределил)
	if contentType != "" {
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	// 6. Пользовательские заголовки
	for _, h := range r.Headers {
		if !h.Enabled || h.Key == "" {
			continue
		}
		req.Header.Set(h.Key, h.Value)
	}

	// 7. Auth — перезаписывает Authorization если установлен Bearer
	// (auth имеет приоритет над ручным заголовком)
	if r.Auth.Type == types.AuthBearer && r.Auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+r.Auth.Token)
	}

	return req, nil
}

// buildURL парсит base URL и добавляет к нему включённые Params.
func buildURL(base string, params []types.Param) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for _, p := range params {
		if !p.Enabled || p.Key == "" {
			continue
		}
		q.Set(p.Key, p.Value)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// buildBody кодирует тело запроса согласно BodyMode.
// Возвращает (reader, contentType, error).
// При BodyModeNone возвращает (nil, "", nil).
func buildBody(r types.SavedRequest) (io.Reader, string, error) {
	switch r.BodyMode {
	case types.BodyModeNone, "":
		return nil, "", nil

	case types.BodyModeRawText:
		return strings.NewReader(r.Body), "text/plain", nil

	case types.BodyModeJSON:
		// Валидируем перед отправкой
		if r.Body != "" && !json.Valid([]byte(r.Body)) {
			return nil, "", fmt.Errorf("Invalid JSON in request body")
		}
		body := r.Body
		if body == "" {
			body = "{}"
		}
		return strings.NewReader(body), "application/json", nil

	case types.BodyModeForm:
		vals := url.Values{}
		for _, p := range r.Params {
			if p.Enabled && p.Key != "" {
				vals.Set(p.Key, p.Value)
			}
		}
		// Form body использует отдельный набор пар (не URL-params).
		// В SavedRequest для form body используется поле Body как encoded string
		// ИЛИ отдельный список. В MVP: декодируем r.Body как form-encoded строку.
		if r.Body != "" {
			return strings.NewReader(r.Body), "application/x-www-form-urlencoded", nil
		}
		return strings.NewReader(vals.Encode()), "application/x-www-form-urlencoded", nil

	default:
		return nil, "", fmt.Errorf("unknown body mode: %s", r.BodyMode)
	}
}
