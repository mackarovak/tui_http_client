package requesteditor

import (
	"strings"
	"testing"

	"htui/internal/types"
)

// ТЗ §16.2, §19.4, §19.5 — валидация до отправки и понятные ошибки.

func TestValidate_EmptyURL(t *testing.T) {
	t.Parallel()
	res := Validate(types.SavedRequest{URL: ""})
	if res.Valid {
		t.Fatal("expected invalid")
	}
	if !containsError(res.Errors, "URL is required") {
		t.Fatalf("errors: %v", res.Errors)
	}
}

func TestValidate_InvalidURL(t *testing.T) {
	t.Parallel()
	res := Validate(types.SavedRequest{URL: "://broken"})
	if res.Valid {
		t.Fatal("expected invalid")
	}
	found := false
	for _, e := range res.Errors {
		if strings.Contains(strings.ToLower(e), "invalid") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected invalid URL message, got %v", res.Errors)
	}
}

func TestValidate_JSONBodyInvalid(t *testing.T) {
	t.Parallel()
	res := Validate(types.SavedRequest{
		URL:      "https://example.com",
		BodyMode: types.BodyModeJSON,
		Body:     `{not json`,
	})
	if res.Valid {
		t.Fatal("expected invalid JSON error")
	}
	if !containsError(res.Errors, "Invalid JSON") {
		t.Fatalf("errors: %v", res.Errors)
	}
}

func TestValidate_JSONBodyEmpty_OK(t *testing.T) {
	t.Parallel()
	res := Validate(types.SavedRequest{
		URL:      "https://example.com",
		BodyMode: types.BodyModeJSON,
		Body:     "",
	})
	if !res.Valid {
		t.Fatalf("empty JSON body should be valid: %v", res.Errors)
	}
}

func TestValidate_BearerEmptyToken_WarningNotBlocking(t *testing.T) {
	t.Parallel()
	res := Validate(types.SavedRequest{
		URL: "https://example.com",
		Auth: types.AuthConfig{
			Type: types.AuthBearer,
		},
	})
	if !res.Valid {
		t.Fatalf("Bearer without token should warn, not block: %v", res.Errors)
	}
	if len(res.Warnings) == 0 {
		t.Fatal("expected warning for empty bearer token")
	}
}

func TestValidate_SchemeOptional_HTTPSPrepended(t *testing.T) {
	t.Parallel()
	res := Validate(types.SavedRequest{URL: "example.com/path"})
	if !res.Valid {
		t.Fatalf("example.com/path should parse with implicit https: %v", res.Errors)
	}
}

func containsError(errs []string, substr string) bool {
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
