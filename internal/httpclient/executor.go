package httpclient

import (
	"context"
	"htui/internal/types"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	DefaultTimeout = 30 * time.Second
	MaxBodyBytes   = 5 * 1024 * 1024 // 5 MB
)

// Client выполняет HTTP-запросы.
type Client struct {
	http *http.Client
}

// New создаёт Client с разумными дефолтами.
func New() *Client {
	return &Client{
		http: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// Execute собирает и отправляет запрос, возвращает ResponseData.
// Никогда не паникует. При ошибке заполняет ResponseData.Error.
func (c *Client) Execute(ctx context.Context, req types.SavedRequest) types.ResponseData {
	start := time.Now()

	httpReq, err := Build(req)
	if err != nil {
		return types.ResponseData{Error: err.Error()}
	}
	httpReq = httpReq.WithContext(ctx)

	resp, err := c.http.Do(httpReq)
	durationMs := time.Since(start).Milliseconds()
	if err != nil {
		return types.ResponseData{
			DurationMs: durationMs,
			Error:      MapError(err),
		}
	}
	defer resp.Body.Close()

	// Читаем тело — выделено в отдельную функцию для будущего стриминга.
	bodyBytes, isBinary := readBodyChunked(resp.Body, MaxBodyBytes)

	// Заголовки ответа
	var headers []types.Header
	for k, vals := range resp.Header {
		headers = append(headers, types.Header{
			Key:     k,
			Value:   strings.Join(vals, ", "),
			Enabled: true,
		})
	}

	return types.ResponseData{
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		DurationMs: durationMs,
		SizeBytes:  len(bodyBytes),
		Headers:    headers,
		Body:       string(bodyBytes),
		IsBinary:   isBinary,
	}
}

// readBodyChunked читает тело ответа с ограничением по размеру.
// Определяет бинарность по первым 512 байтам (не UTF-8 → binary).
//
// АРХИТЕКТУРНАЯ ЗАКЛАДКА:
// Сейчас читает всё за один раз и возвращает []byte.
// В будущей версии эта функция будет запускаться в горутине,
// отправлять чанки через chan []byte, а форматтер будет обрабатывать
// каждый чанк независимо. Сигнатура Execute при этом не изменится —
// только внутренняя логика сборки ответа.
func readBodyChunked(r io.Reader, maxBytes int) ([]byte, bool) {
	limited := io.LimitReader(r, int64(maxBytes))
	data, err := io.ReadAll(limited)
	if err != nil || len(data) == 0 {
		return data, false
	}

	// Определить бинарность: первые 512 байт не являются валидным UTF-8
	probe := data
	if len(probe) > 512 {
		probe = probe[:512]
	}
	isBinary := !utf8.Valid(probe)

	return data, isBinary
}
