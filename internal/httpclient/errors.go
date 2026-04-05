package httpclient

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
)

// MapError преобразует технические ошибки Go в понятные пользователю строки.
// Вызывается при любой сетевой ошибке до заполнения ResponseData.
func MapError(err error) string {
	if err == nil {
		return ""
	}

	// Таймаут / отмена контекста
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "Request timed out"
	}

	// url.Error оборачивает многие ошибки net и url
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// Невалидный URL (до отправки)
		if urlErr.Op == "parse" {
			return "Invalid URL"
		}
		// Разворачиваем для дальнейшей проверки
		err = urlErr.Err
	}

	// DNS / хост недоступен
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "Could not connect to server (host not found)"
	}

	// Соединение отклонено
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if strings.Contains(opErr.Error(), "connection refused") {
			return "Connection refused by server"
		}
		if opErr.Timeout() {
			return "Request timed out"
		}
		return "Network error: could not reach server"
	}

	// Общий таймаут через строку (fallback)
	msg := err.Error()
	if strings.Contains(msg, "timeout") {
		return "Request timed out"
	}
	if strings.Contains(msg, "no such host") {
		return "Could not connect to server (host not found)"
	}
	if strings.Contains(msg, "connection refused") {
		return "Connection refused by server"
	}

	return "Request failed: " + msg
}
