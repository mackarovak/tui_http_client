package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"htui/internal/types"
)

// ТЗ §16.1, §19.2, §20 п.2–4, §9.2–9.3 — сборка URL/body и выполнение запроса.

func TestBuildURL_AddsHTTPSWhenNoScheme(t *testing.T) {
	t.Parallel()
	u, err := buildURL(types.SavedRequest{URL: "api.example.com/x", Method: "GET"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u, "https://") {
		t.Fatalf("got %q", u)
	}
}

func TestBuildURL_QueryParams(t *testing.T) {
	t.Parallel()
	u, err := buildURL(types.SavedRequest{
		URL:    "https://example.com/path",
		Method: "GET",
		Params: []types.Param{
			{Key: "q", Value: "a b", Enabled: true},
			{Key: "empty", Value: "x", Enabled: false},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "q=") || !strings.Contains(u, "path?") {
		t.Fatalf("expected query in URL: %q", u)
	}
}

func TestBuildURL_FormMode_NoQueryFromParams(t *testing.T) {
	t.Parallel()
	u, err := buildURL(types.SavedRequest{
		URL:      "https://example.com/p",
		BodyMode: types.BodyModeForm,
		Params:   []types.Param{{Key: "a", Value: "1", Enabled: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(u, "a=1") {
		t.Fatalf("form mode should not put params in query: %q", u)
	}
}

func TestBuildBody_JSON_SetsContentType(t *testing.T) {
	t.Parallel()
	req, _ := http.NewRequest("POST", "https://x", nil)
	r := types.SavedRequest{BodyMode: types.BodyModeJSON, Body: `{"a":1}`}
	rc, err := buildBody(r, req)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	if string(b) != `{"a":1}` {
		t.Fatalf("body %q", b)
	}
	if ct := req.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("Content-Type: %q", ct)
	}
}

func TestBuildBody_Form_URLEncoded(t *testing.T) {
	t.Parallel()
	req, _ := http.NewRequest("POST", "https://x", nil)
	r := types.SavedRequest{
		BodyMode: types.BodyModeForm,
		Params: []types.Param{
			{Key: "title", Value: "hello", Enabled: true},
		},
	}
	rc, err := buildBody(r, req)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	if !strings.Contains(string(b), "title=hello") {
		t.Fatalf("body %q", b)
	}
}

func TestExecute_GET_JSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New()
	out := c.Execute(context.Background(), types.SavedRequest{
		Method:   "GET",
		URL:      srv.URL,
		BodyMode: types.BodyModeNone,
	})
	if out.Error != "" {
		t.Fatal(out.Error)
	}
	if out.StatusCode != 200 {
		t.Fatalf("status %d", out.StatusCode)
	}
	if out.SizeBytes != len(`{"ok":true}`) {
		t.Fatalf("size %d", out.SizeBytes)
	}
	if out.Body != `{"ok":true}` {
		t.Fatalf("body %q", out.Body)
	}
}

func TestExecute_POST_JSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(b) != `{"x":1}` {
			t.Errorf("body %q", b)
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			t.Errorf("ct %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()

	c := New()
	out := c.Execute(context.Background(), types.SavedRequest{
		Method:   "POST",
		URL:      srv.URL,
		BodyMode: types.BodyModeJSON,
		Body:     `{"x":1}`,
	})
	if out.Error != "" || out.StatusCode != 201 {
		t.Fatalf("err=%q code=%d", out.Error, out.StatusCode)
	}
}

func TestExecute_BearerHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("Authorization: %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := New()
	out := c.Execute(context.Background(), types.SavedRequest{
		Method: "GET",
		URL:    srv.URL,
		Auth: types.AuthConfig{
			Type:  types.AuthBearer,
			Token: "secret-token",
		},
	})
	if out.Error != "" {
		t.Fatal(out.Error)
	}
}

func TestExecute_InvalidURL(t *testing.T) {
	t.Parallel()
	c := New()
	out := c.Execute(context.Background(), types.SavedRequest{URL: ""})
	if out.Error != "Invalid URL" {
		t.Fatalf("Error = %q", out.Error)
	}
}

func TestExecute_EmptyJSONBody_NoPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := New()
	out := c.Execute(context.Background(), types.SavedRequest{
		Method:   "POST",
		URL:      srv.URL,
		BodyMode: types.BodyModeJSON,
		Body:     "",
	})
	if out.Error != "" {
		t.Fatal(out.Error)
	}
}
