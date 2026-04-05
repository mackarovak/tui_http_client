package types

import (
	"strings"
	"testing"
)

// Критерии ТЗ §17 (Demo collection) и §9.1 (HTTP methods).

func TestHTTPMethods_MVPMethods(t *testing.T) {
	t.Parallel()
	want := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	if len(HTTPMethods) != len(want) {
		t.Fatalf("HTTPMethods: len %d, want %d", len(HTTPMethods), len(want))
	}
	for i, m := range want {
		if HTTPMethods[i] != m {
			t.Errorf("HTTPMethods[%d] = %q, want %q", i, HTTPMethods[i], m)
		}
	}
}

func TestDemoRequests_TZSection17(t *testing.T) {
	t.Parallel()
	demos := DemoRequests()
	if len(demos) < 3 {
		t.Fatalf("DemoRequests: need at least 3 demos per TZ §17, got %d", len(demos))
	}

	// Demo 1: Simple GET, jsonplaceholder posts/1
	g1 := demos[0]
	if g1.Name != "Simple GET" || g1.Method != "GET" {
		t.Errorf("demo[0]: name=%q method=%q", g1.Name, g1.Method)
	}
	if !strings.Contains(g1.URL, "jsonplaceholder.typicode.com") || !strings.Contains(g1.URL, "posts/1") {
		t.Errorf("demo[0] URL: %q", g1.URL)
	}

	// Demo 2: Create Post, POST JSON
	p2 := demos[1]
	if p2.Name != "Create Post" || p2.Method != "POST" || p2.BodyMode != BodyModeJSON {
		t.Errorf("demo[1]: %+v", p2)
	}
	if !strings.Contains(p2.Body, "hello") || !strings.Contains(p2.Body, "userId") {
		t.Errorf("demo[1] body should contain JSON from TZ: %q", p2.Body)
	}

	// Demo 3: Bearer example
	b3 := demos[2]
	if b3.Auth.Type != AuthBearer {
		t.Errorf("demo[2] auth: %+v", b3.Auth)
	}
}

func TestNewSavedRequest_Defaults(t *testing.T) {
	t.Parallel()
	r := NewSavedRequest()
	if r.Method != "GET" {
		t.Errorf("Method = %q, want GET", r.Method)
	}
	if r.BodyMode != BodyModeNone {
		t.Errorf("BodyMode = %q, want none", r.BodyMode)
	}
	if r.Auth.Type != AuthNone {
		t.Errorf("Auth.Type = %q", r.Auth.Type)
	}
	if r.ID == "" || r.Name == "" {
		t.Errorf("ID/Name should be set: %#v", r)
	}
}

func TestEnsureID_GeneratesWhenEmpty(t *testing.T) {
	t.Parallel()
	r := SavedRequest{}
	r.EnsureID()
	if r.ID == "" {
		t.Fatal("EnsureID should set ID")
	}
}
