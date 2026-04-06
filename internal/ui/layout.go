package ui

import "github.com/charmbracelet/lipgloss"

// LayoutMode определяет режим раскладки.
type LayoutMode int

const (
    LayoutWide   LayoutMode = iota // три колонки рядом
    LayoutNarrow                   // sidebar + editor/response стек
    LayoutMinimal                  // sidebar скрыт, только editor/response
)

const (
    WideThreshold    = 120 // columns
    MinimalThreshold = 80  // columns
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

// ComputeLayout определяет режим раскладки по ширине терминала.
func ComputeLayout(width int) LayoutMode {
    switch {
    case width >= WideThreshold:
        return LayoutWide
    case width >= MinimalThreshold:
        return LayoutNarrow
    default:
        return LayoutMinimal
    }
}

// ComputePanelDimensions рассчитывает размеры панелей с учётом рамок (2px border).
func ComputePanelDimensions(w, h int, mode LayoutMode) PanelDimensions {
    // Рамка занимает 2 символа (1 с каждой стороны)
    const borderSize = 2

    switch mode {
    case LayoutWide:
        sidebarW := w * 25 / 100
        editorW := w * 40 / 100
        responseW := w - sidebarW - editorW
        return PanelDimensions{
            Sidebar:  PanelSize{sidebarW - borderSize, h - borderSize},
            Editor:   PanelSize{editorW - borderSize, h - borderSize},
            Response: PanelSize{responseW - borderSize, h - borderSize},
        }

    case LayoutNarrow:
        sidebarW := w * 30 / 100
        contentW := w - sidebarW
        contentH := h / 2
        return PanelDimensions{
            Sidebar:  PanelSize{sidebarW - borderSize, h - borderSize},
            Editor:   PanelSize{contentW - borderSize, contentH - borderSize},
            Response: PanelSize{contentW - borderSize, h - contentH - borderSize},
        }

    default: // LayoutMinimal
        half := h / 2
        return PanelDimensions{
            Sidebar:  PanelSize{0, 0},
            Editor:   PanelSize{w - borderSize, half - borderSize},
            Response: PanelSize{w - borderSize, h - half - borderSize},
        }
    }
}

// RenderLayout собирает финальный вид из трёх строк-компонентов.
func RenderLayout(mode LayoutMode, sidebar, editor, response string) string {
    switch mode {
    case LayoutWide:
        return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, editor, response)
    case LayoutNarrow:
        right := lipgloss.JoinVertical(lipgloss.Left, editor, response)
        return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, right)
    default:
        return lipgloss.JoinVertical(lipgloss.Left, editor, response)
    }
}