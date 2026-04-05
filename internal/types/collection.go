package types

import (
	"fmt"
	"time"
)

// DemoRequests возвращает начальный набор запросов для первого запуска.
// Используется store-слоем для сидирования при отсутствии .seeded.
func DemoRequests() []SavedRequest {
	base := time.Now()
	return []SavedRequest{
		{
			ID:        fmt.Sprintf("%d", base.UnixNano()),
			Name:      "Simple GET",
			Method:    "GET",
			URL:       "https://jsonplaceholder.typicode.com/posts/1",
			BodyMode:  BodyModeNone,
			Auth:      AuthConfig{Type: AuthNone},
			Params:    []Param{},
			Headers:   []Header{},
			CreatedAt: base,
			UpdatedAt: base,
		},
		{
			ID:        fmt.Sprintf("%d", base.Add(time.Millisecond).UnixNano()),
			Name:      "Create Post",
			Method:    "POST",
			URL:       "https://jsonplaceholder.typicode.com/posts",
			BodyMode:  BodyModeJSON,
			Body:      "{\n  \"title\": \"hello\",\n  \"body\": \"world\",\n  \"userId\": 1\n}",
			Auth:      AuthConfig{Type: AuthNone},
			Params:    []Param{},
			Headers:   []Header{},
			CreatedAt: base.Add(time.Millisecond),
			UpdatedAt: base.Add(time.Millisecond),
		},
		{
			ID:        fmt.Sprintf("%d", base.Add(2*time.Millisecond).UnixNano()),
			Name:      "Bearer Auth Example",
			Method:    "GET",
			URL:       "https://httpbin.org/bearer",
			BodyMode:  BodyModeNone,
			Auth:      AuthConfig{Type: AuthBearer, Token: "your-token-here"},
			Params:    []Param{},
			Headers:   []Header{},
			CreatedAt: base.Add(2 * time.Millisecond),
			UpdatedAt: base.Add(2 * time.Millisecond),
		},
	}
}
