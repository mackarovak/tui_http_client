package curlexport

import (
	"strings"
	"testing"

	"htui/internal/types"
)

func TestBuild_SimpleGET(t *testing.T) {
	r := types.NewSavedRequest()
	r.Method = "GET"
	r.URL = "https://api.example.com/users"

	got := Build(r)

	if !strings.Contains(got, "curl \"https://api.example.com/users\"") {
		t.Fatalf("unexpected output: %s", got)
	}
	if strings.Contains(got, "-X GET") {
		t.Fatalf("GET should not include -X flag: %s", got)
	}
}

func TestBuild_POSTWithJSON(t *testing.T) {
	r := types.NewSavedRequest()
	r.Method = "POST"
	r.URL = "https://api.example.com/posts"
	r.BodyMode = types.BodyModeJSON
	r.Body = `{"title":"hello"}`

	got := Build(r)

	if !strings.Contains(got, "-X POST") {
		t.Fatalf("expected -X POST: %s", got)
	}
	if !strings.Contains(got, "Content-Type: application/json") {
		t.Fatalf("expected Content-Type header: %s", got)
	}
	if !strings.Contains(got, `{"title":"hello"}`) {
		t.Fatalf("expected body: %s", got)
	}
}

func TestBuild_BearerAuth(t *testing.T) {
	r := types.NewSavedRequest()
	r.Method = "GET"
	r.URL = "https://api.example.com/me"
	r.Auth = types.AuthConfig{Type: types.AuthBearer, Token: "mi-token-secreto"}

	got := Build(r)

	if !strings.Contains(got, "Authorization: Bearer mi-token-secreto") {
		t.Fatalf("expected bearer token: %s", got)
	}
}

func TestBuild_WithQueryParams(t *testing.T) {
	r := types.NewSavedRequest()
	r.Method = "GET"
	r.URL = "https://api.example.com/search"
	r.Params = []types.Param{
		{Key: "q", Value: "golang", Enabled: true},
		{Key: "page", Value: "1", Enabled: true},
		{Key: "disabled", Value: "x", Enabled: false},
	}

	got := Build(r)

	if !strings.Contains(got, "q=golang") {
		t.Fatalf("expected query param q: %s", got)
	}
	if !strings.Contains(got, "page=1") {
		t.Fatalf("expected query param page: %s", got)
	}
	if strings.Contains(got, "disabled") {
		t.Fatalf("disabled param should not appear: %s", got)
	}
}

func TestBuild_WithCustomHeaders(t *testing.T) {
	r := types.NewSavedRequest()
	r.Method = "GET"
	r.URL = "https://api.example.com"
	r.Headers = []types.Header{
		{Key: "X-App-ID", Value: "123", Enabled: true},
		{Key: "X-Ignored", Value: "abc", Enabled: false},
	}

	got := Build(r)

	if !strings.Contains(got, "X-App-ID: 123") {
		t.Fatalf("expected custom header: %s", got)
	}
	if strings.Contains(got, "X-Ignored") {
		t.Fatalf("disabled header should not appear: %s", got)
	}
}

func TestBuild_URLWithoutScheme(t *testing.T) {
	r := types.NewSavedRequest()
	r.Method = "GET"
	r.URL = "api.example.com/users"

	got := Build(r)

	if !strings.Contains(got, "https://api.example.com/users") {
		t.Fatalf("expected https scheme to be added: %s", got)
	}
}
