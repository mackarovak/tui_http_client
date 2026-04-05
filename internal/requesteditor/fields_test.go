package requesteditor

import (
	"testing"

	"htui/internal/types"
)

// ТЗ §9.4 — заголовки как пары ключ-значение; пустые ключи не попадают в запрос.

func TestKVTable_ToHeaders_SkipsEmptyKey(t *testing.T) {
	t.Parallel()
	tb := KVTable{
		rows: []kvRow{
			{key: "X-Test", value: "1", enabled: true},
			{key: "", value: "skip", enabled: true},
		},
	}
	h := tb.ToHeaders()
	if len(h) != 1 || h[0].Key != "X-Test" {
		t.Fatalf("ToHeaders: %+v", h)
	}
}

func TestKVTable_FromHeaders_RoundTrip(t *testing.T) {
	t.Parallel()
	src := []types.Header{
		{Key: "A", Value: "1", Enabled: true},
	}
	tb := FromHeaders(src)
	if len(tb.rows) != 1 || tb.rows[0].key != "A" {
		t.Fatalf("%+v", tb.rows)
	}
	back := tb.ToHeaders()
	if len(back) != 1 || back[0].Value != "1" {
		t.Fatalf("%+v", back)
	}
}
