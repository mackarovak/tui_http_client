package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestComputeLayoutAlwaysUsesStableArrangement(t *testing.T) {
	widths := []int{48, 80, 120, 200}
	for _, width := range widths {
		if got := ComputeLayout(width); got != LayoutNarrow {
			t.Fatalf("width %d: expected LayoutNarrow, got %v", width, got)
		}
	}
}

func TestComputeShellDimensionsSwitchesHeaderMode(t *testing.T) {
	large := ComputeShellDimensions(120, 36)
	if !large.UseASCIIHeader {
		t.Fatalf("expected ASCII header for spacious terminal")
	}

	small := ComputeShellDimensions(36, 16)
	if small.UseASCIIHeader {
		t.Fatalf("expected compact header for small terminal")
	}
}

func TestComputePanelDimensionsFillStableGrid(t *testing.T) {
	shell := ComputeShellDimensions(140, 42)
	dims := ComputePanelDimensions(shell.BodyWidth, shell.BodyHeight, LayoutNarrow)

	sidebarFrameW := dims.Sidebar.Width + panelBorderSize
	rightFrameW := dims.Editor.Width + panelBorderSize
	if sidebarFrameW+panelGap+rightFrameW != shell.BodyWidth {
		t.Fatalf("unexpected horizontal split: got %d, want %d", sidebarFrameW+panelGap+rightFrameW, shell.BodyWidth)
	}

	editorFrameH := dims.Editor.Height + panelBorderSize
	responseFrameH := dims.Response.Height + panelBorderSize
	if editorFrameH+panelGap+responseFrameH != shell.BodyHeight {
		t.Fatalf("unexpected vertical split: got %d, want %d", editorFrameH+panelGap+responseFrameH, shell.BodyHeight)
	}

	if dims.Sidebar.Height+panelBorderSize != shell.BodyHeight {
		t.Fatalf("sidebar should span full shell body height")
	}
}

func TestRenderShellMatchesRequestedWindowSize(t *testing.T) {
	shell := ComputeShellDimensions(120, 36)
	panel := Theme.InactiveBorder.Width(10).Height(3).Render("demo")
	rendered := RenderShell(shell, RenderLayout(LayoutNarrow, panel, panel, panel))

	if got := lipgloss.Width(rendered); got != 120 {
		t.Fatalf("unexpected rendered width: got %d, want 120", got)
	}
	if got := lipgloss.Height(rendered); got != 36 {
		t.Fatalf("unexpected rendered height: got %d, want 36", got)
	}
}
