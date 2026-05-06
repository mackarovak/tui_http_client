package requesteditor

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestViewFitsAssignedHeight(t *testing.T) {
	m := New()
	m, _ = m.SetSize(80, 20)

	if got := lipgloss.Height(m.View()); got > m.Height() {
		t.Fatalf("editor view overflowed: got %d lines, limit %d", got, m.Height())
	}
}
