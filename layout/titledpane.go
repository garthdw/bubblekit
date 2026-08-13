// Package layout provides small, dependency-light rendering/layout helpers
// for Bubble Tea apps that don't couple to any app's domain types: a
// lazygit-style bordered pane with an embedded title, display-width-aware
// truncation, viewport windowing over variable-width horizontal strips
// (e.g. a tab bar), and bin-packing titled entry groups into columns
// (e.g. a help popup).
package layout

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// RenderTitledPane renders content inside style's border, embedding title
// into the top border edge (e.g. lazygit-style "╭─ Files ──╮" corner
// labels). width and height are the desired content dimensions;
// lipgloss.Style.Width/Height treat their argument as the box's total
// size, so the border is added on top here.
func RenderTitledPane(style lipgloss.Style, width, height int, title, content string) string {
	rendered := style.Width(width + 2).Height(height + 2).Render(content)
	if title == "" || !style.GetBorderTop() {
		return rendered
	}

	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		return rendered
	}

	border := style.GetBorderStyle()
	label := " " + title + " "
	labelWidth := lipgloss.Width(label)

	// Measure the box's actual rendered width rather than trusting the
	// width argument: content can overflow the requested width, and the
	// top border must match whatever width the side borders actually
	// landed at.
	boxWidth := lipgloss.Width(lines[0])
	fill := boxWidth - lipgloss.Width(border.TopLeft) - lipgloss.Width(border.TopRight) - lipgloss.Width(border.Top) - labelWidth
	if fill < 0 {
		return rendered
	}

	top := border.TopLeft + border.Top + label + strings.Repeat(border.Top, fill) + border.TopRight
	topStyle := lipgloss.NewStyle().Foreground(style.GetBorderTopForeground()).Background(style.GetBackground())
	lines[0] = topStyle.Render(top)

	return strings.Join(lines, "\n")
}

// TruncateToWidth truncates s (a plain, unstyled string) to at most maxW
// display columns, appending "…" if truncation occurred.
func TruncateToWidth(s string, maxW int) string {
	if lipgloss.Width(s) <= maxW {
		return s
	}
	if maxW <= 1 {
		return "…"
	}
	budget := maxW - lipgloss.Width("…")

	runes := []rune(s)
	width, cut := 0, 0
	for i, r := range runes {
		rw := lipgloss.Width(string(r))
		if width+rw > budget {
			break
		}
		width += rw
		cut = i + 1
	}
	return string(runes[:cut]) + "…"
}
