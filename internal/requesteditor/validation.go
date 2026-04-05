package requesteditor

import (
	"encoding/json"
	"net/url"
	"htui/internal/types"
)

// ValidationResult — результат проверки запроса перед отправкой.
type ValidationResult struct {
	Valid    bool
	Errors   []string // блокирующие ошибки (не отправляем)
	Warnings []string // предупреждения (отправляем, но предупреждаем)
}

// Validate проверяет запрос перед отправкой.
// Errors — блокируют отправку. Warnings — нет.
func Validate(r types.SavedRequest) ValidationResult {
	var res ValidationResult
	res.Valid = true

	// 1. URL не пустой
	if r.URL == "" {
		res.Valid = false
		res.Errors = append(res.Errors, "URL is required")
		return res
	}

	// 2. URL парсится (с учётом авто-схемы)
	testURL := r.URL
	if len(testURL) > 0 && !hasScheme(testURL) {
		testURL = "https://" + testURL
	}
	if _, err := url.ParseRequestURI(testURL); err != nil {
		res.Valid = false
		res.Errors = append(res.Errors, "Invalid URL: "+err.Error())
	}

	// 3. JSON body валиден
	if r.BodyMode == types.BodyModeJSON && r.Body != "" {
		if !json.Valid([]byte(r.Body)) {
			res.Valid = false
			res.Errors = append(res.Errors, "Invalid JSON in request body")
		}
	}

	// 4. Bearer без токена — предупреждение (не блокирующее)
	if r.Auth.Type == types.AuthBearer && r.Auth.Token == "" {
		res.Warnings = append(res.Warnings, "Bearer token is empty")
	}

	return res
}

func hasScheme(u string) bool {
	return len(u) > 4 &&
		(u[:7] == "http://" || u[:8] == "https://")
}