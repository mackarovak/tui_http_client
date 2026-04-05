package types

// ResponseData — результат выполнения HTTP-запроса.
type ResponseData struct {
	StatusCode int      `json:"status_code"`
	StatusText string   `json:"status_text"`
	DurationMs int64    `json:"duration_ms"`
	SizeBytes  int      `json:"size_bytes"`
	Headers    []Header `json:"headers"`
	Body       string   `json:"body"`
	IsBinary   bool     `json:"is_binary"`
	// Error заполняется вместо Body при ошибке сети/таймаута.
	Error string `json:"error,omitempty"`
}

// IsError возвращает true если запрос завершился ошибкой (не HTTP-ошибкой,
// а сетевой — нет соединения, таймаут и т.д.).
func (r ResponseData) IsError() bool {
	return r.Error != ""
}
