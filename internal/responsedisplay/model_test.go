package responsedisplay

import (
	"strings"
	"testing"
)

// ТЗ §6.2 п.1, §8.5 — pretty-print JSON в ответе.

func TestFormatBody_JSONPrettyPrinted(t *testing.T) {
	t.Parallel()
	raw := `{"b":2,"a":1}`
	out := formatBody(raw)
	if !strings.Contains(out, "\n") {
		t.Fatalf("expected indented JSON, got %q", out)
	}
	if !strings.Contains(out, `"a"`) || !strings.Contains(out, `"b"`) {
		t.Fatalf("got %q", out)
	}
}

func TestFormatBody_NonJSON_Unchanged(t *testing.T) {
	t.Parallel()
	raw := "plain text\nline2"
	if got := formatBody(raw); got != raw {
		t.Fatalf("got %q", got)
	}
}

func TestFormatBody_Empty(t *testing.T) {
	t.Parallel()
	if formatBody("") != "" || formatBody("  \n") != "" {
		t.Fatal("empty input should yield empty")
	}
}
