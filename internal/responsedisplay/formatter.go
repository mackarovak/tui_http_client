package responsedisplay

import (
	"encoding/json"
	"fmt"
	"strings"

	"htui/internal/types"
)

// FormatBody форматирует тело ответа для отображения.
// JSON → pretty-print. Бинарный → информационное сообщение. Иначе → как есть.
//
// АРХИТЕКТУРНАЯ ЗАКЛАДКА:
// Функция принимает готовый ResponseData.Body (string).
// В будущей версии со стримингом эта же функция будет вызываться
// для каждого накопленного чанка из chan []byte — сигнатура не изменится,
// изменится только момент вызова (не после полного получения, а пошагово).
func FormatBody(data types.ResponseData) string {
	if data.IsBinary {
		return fmt.Sprintf("[Binary response received: %s]", FormatBytes(data.SizeBytes))
	}
	if data.Body == "" {
		return "(empty response body)"
	}
	if isJSON(data) {
		if pretty, err := prettyJSON(data.Body); err == nil {
			return pretty
		}
	}
	return data.Body
}

// FormatHeaders форматирует заголовки ответа для отображения.
func FormatHeaders(headers []types.Header) string {
	if len(headers) == 0 {
		return "(no headers)"
	}
	var sb strings.Builder
	for _, h := range headers {
		sb.WriteString(fmt.Sprintf("%-30s %s\n", h.Key+":", h.Value))
	}
	return sb.String()
}

// FormatBytes форматирует размер в читаемый вид.
func FormatBytes(n int) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}

// FormatDuration форматирует время выполнения в читаемый вид.
func FormatDuration(ms int64) string {
	switch {
	case ms < 1000:
		return fmt.Sprintf("%d ms", ms)
	default:
		return fmt.Sprintf("%.2f s", float64(ms)/1000)
	}
}

// isJSON определяет является ли ответ JSON по Content-Type или попытке парсинга.
func isJSON(data types.ResponseData) bool {
	for _, h := range data.Headers {
		if strings.EqualFold(h.Key, "Content-Type") {
			if strings.Contains(h.Value, "application/json") ||
				strings.Contains(h.Value, "text/json") {
				return true
			}
		}
	}
	// Fallback: попробовать распарсить
	return json.Valid([]byte(strings.TrimSpace(data.Body)))
}

// prettyJSON форматирует JSON строку с отступами.
func prettyJSON(s string) (string, error) {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
