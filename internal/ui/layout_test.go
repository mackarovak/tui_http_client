package ui

import "testing"

// ТЗ §8.2 — пороги ширины терминала (WideThreshold, MinimalThreshold).

func TestComputeLayout_TZ82(t *testing.T) {
	t.Parallel()
	tests := []struct {
		width int
		want  LayoutMode
	}{
		{79, LayoutMinimal},
		{80, LayoutNarrow},
		{119, LayoutNarrow},
		{120, LayoutWide},
		{200, LayoutWide},
	}
	for _, tc := range tests {
		if got := ComputeLayout(tc.width); got != tc.want {
			t.Errorf("ComputeLayout(%d) = %v, want %v", tc.width, got, tc.want)
		}
	}
}

func TestComputePanelDimensions_NonZeroInnerSize(t *testing.T) {
	t.Parallel()
	d := ComputePanelDimensions(120, 40, LayoutWide)
	if d.Sidebar.Width <= 0 || d.Editor.Width <= 0 || d.Response.Width <= 0 {
		t.Fatalf("wide: %+v", d)
	}
	d2 := ComputePanelDimensions(100, 30, LayoutNarrow)
	if d2.Sidebar.Width <= 0 || d2.Editor.Width <= 0 {
		t.Fatalf("narrow editor: %+v", d2)
	}
}
