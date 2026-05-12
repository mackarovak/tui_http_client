package sidebar

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"htui/internal/types"
)

func TestViewFitsAssignedHeight(t *testing.T) {
	m := New(types.DemoRequests(), nil).SelectFirst()
	m, _ = m.SetSize(32, 18)

	if got := lipgloss.Height(m.View()); got > m.Height() {
		t.Fatalf("sidebar view overflowed: got %d lines, limit %d", got, m.Height())
	}
}
