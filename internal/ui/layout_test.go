package ui

import (
	"strings"
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

func TestRenderHeaderFitsShellWidth(t *testing.T) {
	header := RenderHeader(32, true)
	for _, line := range strings.Split(header, "\n") {
		if got := lipgloss.Width(line); got != 32 {
			t.Fatalf("header line width mismatch: got %d, want 32; line %q", got, line)
		}
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

func TestFitBlockConstrainsContent(t *testing.T) {
	rendered := FitBlock("this line is longer than the box\nline 2\nline 3", 8, 2)

	if got := lipgloss.Width(rendered); got != 8 {
		t.Fatalf("unexpected block width: got %d, want 8", got)
	}
	if got := lipgloss.Height(rendered); got != 2 {
		t.Fatalf("unexpected block height: got %d, want 2", got)
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
