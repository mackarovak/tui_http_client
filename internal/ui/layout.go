package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LayoutMode определяет режим раскладки.
type LayoutMode int

const (
	LayoutWide    LayoutMode = iota // три колонки рядом
	LayoutNarrow                    // sidebar + editor/response стек
	LayoutMinimal                   // sidebar скрыт, только editor/response
)

const (
	WideThreshold    = 120 // columns
	MinimalThreshold = 80  // columns
)

const (
	shellBorderSize        = 2
	panelBorderSize        = 2
	panelGap               = 0
	minBodyHeight          = 1
	minSidebarFrameWidth   = 18
	maxSidebarFrameWidth   = 34
	minRightFrameWidth     = 30
	minEditorFrameHeight   = 10
	minResponseFrameHeight = 8
)

// PanelSize — ширина и высота панели.
type PanelSize struct {
	Width  int
	Height int
}

// PanelDimensions — размеры всех трёх панелей.
type PanelDimensions struct {
	Sidebar  PanelSize
	Editor   PanelSize
	Response PanelSize
}

// ShellDimensions stores the size of the outer application frame.
type ShellDimensions struct {
	Width          int
	Height         int
	BodyWidth      int
	BodyHeight     int
	UseASCIIHeader bool
}

var asciiHeader = strings.TrimSpace(`
 _   _  _____  _   _  ___
| | | ||_   _|| | | ||_ _|
| |_| |  | |  | | | | | |
|  _  |  | |  | |_| | | |
|_| |_|  |_|   \___/ |___|
`)
var asciiHeaderWidth = lipgloss.Width(asciiHeader)
// ComputeLayout определяет режим раскладки по ширине терминала.
func ComputeLayout(width int) LayoutMode {
	_ = width
	return LayoutNarrow
}

// ComputeShellDimensions returns the inner size of the shared outer frame.
func ComputeShellDimensions(w, h int) ShellDimensions {
	innerW := max(w-shellBorderSize, 1)
	innerH := max(h-shellBorderSize, 1)

	useASCIIHeader := innerW >=  asciiHeaderWidth && innerH >= 18
	headerHeight := lipgloss.Height(RenderHeader(innerW, useASCIIHeader))
	bodyHeight := max(innerH-headerHeight, minBodyHeight)

	return ShellDimensions{
		Width:          innerW,
		Height:         innerH,
		BodyWidth:      innerW,
		BodyHeight:     bodyHeight,
		UseASCIIHeader: useASCIIHeader,
	}
}

// ComputePanelDimensions рассчитывает размеры панелей с учётом рамок (2px border).
func ComputePanelDimensions(w, h int, mode LayoutMode) PanelDimensions {
	// Рамка занимает 2 символа (1 с каждой стороны)
	_ = mode

	sidebarFrameW, rightFrameW := splitHorizontalFrames(w)
	editorFrameH, responseFrameH := splitVerticalFrames(h)

	return PanelDimensions{
		Sidebar: PanelSize{
			Width:  innerFrameSize(sidebarFrameW),
			Height: innerFrameSize(h),
		},
		Editor: PanelSize{
			Width:  innerFrameSize(rightFrameW),
			Height: innerFrameSize(editorFrameH),
		},
		Response: PanelSize{
			Width:  innerFrameSize(rightFrameW),
			Height: innerFrameSize(responseFrameH),
		},
	}
}

// RenderLayout собирает финальный вид из трёх строк-компонентов.
// RenderHeader renders the HTUI header inside the shared frame.
func RenderHeader(width int, useASCIIHeader bool) string {
	if width <= 0 {
		return ""
	}
	if useASCIIHeader {
		
		art := Theme.HeaderArt.Render(lipgloss.PlaceHorizontal(width, lipgloss.Center, asciiHeader))
		subtitle := Theme.HeaderSubtitle.Render(lipgloss.PlaceHorizontal(width, lipgloss.Center, "HTTP terminal UI"))
		return lipgloss.JoinVertical(lipgloss.Left, art, subtitle)
	}

	title := Theme.HeaderArt.Render(lipgloss.PlaceHorizontal(width, lipgloss.Center, "HTUI"))
	subtitle := Theme.HeaderSubtitle.Render(lipgloss.PlaceHorizontal(width, lipgloss.Center, "HTTP terminal UI"))
	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle)
}

// RenderShell wraps the header and panels into one outer window.
func RenderShell(shell ShellDimensions, body string) string {
	header := RenderHeader(shell.Width, shell.UseASCIIHeader)
	body = lipgloss.NewStyle().
		Width(shell.BodyWidth).
		Height(shell.BodyHeight).
		MaxWidth(shell.BodyWidth).
		MaxHeight(shell.BodyHeight).
		Render(body)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
	)
	return Theme.ShellBorder.Width(shell.Width).Height(shell.Height).Render(content)
}
// FitBlock constrains content to an exact terminal-cell rectangle.
func FitBlock(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxWidth(width).
		MaxHeight(height).
		Render(content)
}

func RenderLayout(mode LayoutMode, sidebar, editor, response string) string {
	_ = mode

	right := lipgloss.JoinVertical(lipgloss.Left, editor, response)
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, right)
}

func splitHorizontalFrames(total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}

	usable := max(total-panelGap, 1)
	sidebar := usable * 30 / 100

	if usable >= minSidebarFrameWidth+minRightFrameWidth {
		sidebar = clamp(sidebar, minSidebarFrameWidth, maxSidebarFrameWidth)
		if usable-sidebar < minRightFrameWidth {
			sidebar = usable - minRightFrameWidth
		}
	} else {
		sidebar = max(usable/3, 0)
	}

	if sidebar > usable-1 {
		sidebar = max(usable-1, 0)
	}
	if sidebar < 0 {
		sidebar = 0
	}

	return sidebar, usable - sidebar
}

func splitVerticalFrames(total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}

	usable := max(total-panelGap, 1)
	editor := usable * 57 / 100

	if usable >= minEditorFrameHeight+minResponseFrameHeight {
		if editor < minEditorFrameHeight {
			editor = minEditorFrameHeight
		}
		if usable-editor < minResponseFrameHeight {
			editor = usable - minResponseFrameHeight
		}
	}

	if editor > usable-1 {
		editor = max(usable-1, 0)
	}
	if editor < 0 {
		editor = 0
	}

	return editor, usable - editor
}

func innerFrameSize(frame int) int {
	return max(frame-panelBorderSize, 0)
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
