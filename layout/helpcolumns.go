package layout

import "strings"

// HelpEntry is one key/label pair shown in a help listing, e.g.
// {"ctrl+e", "edit"}.
type HelpEntry struct {
	Key   string
	Label string
}

// HelpGroup is one titled section of a help popup, e.g. "Global" or
// "Tree", holding the entries active in that context.
type HelpGroup struct {
	Title   string
	Entries []HelpEntry
}

// GroupLines is how many rendered lines a group takes: title + entries +
// one blank separator line.
func GroupLines(g HelpGroup) int {
	return 1 + len(g.Entries) + 1
}

// SplitHelpColumns packs groups into columns of at most maxRows lines
// each, never splitting a group across columns.
func SplitHelpColumns(groups []HelpGroup, maxRows int) [][]HelpGroup {
	var cols [][]HelpGroup
	var cur []HelpGroup
	rows := 0
	for _, g := range groups {
		gl := GroupLines(g)
		if rows+gl > maxRows && len(cur) > 0 {
			cols = append(cols, cur)
			cur = nil
			rows = 0
		}
		cur = append(cur, g)
		rows += gl
	}
	if len(cur) > 0 {
		cols = append(cols, cur)
	}
	return cols
}

// RenderHelpColumn renders one column's groups and returns the text along
// with its natural (unpadded) width and line count. renderTitle and
// renderEntry apply the caller's own styling (e.g. theme-driven lipgloss
// styles) and must return plain strings whose display width RenderHelpColumn
// can measure.
func RenderHelpColumn(col []HelpGroup, renderTitle func(title string) string, renderEntry func(key, label string) string, width func(string) int) (text string, colWidth, lines int) {
	parts := make([]string, len(col))
	for i, g := range col {
		title := renderTitle(g.Title)
		if w := width(title); w > colWidth {
			colWidth = w
		}
		entryLines := make([]string, len(g.Entries))
		for j, e := range g.Entries {
			line := renderEntry(e.Key, e.Label)
			if w := width(line); w > colWidth {
				colWidth = w
			}
			entryLines[j] = line
		}
		parts[i] = title + "\n" + strings.Join(entryLines, "\n")
		lines += GroupLines(g)
	}
	return strings.Join(parts, "\n\n"), colWidth, lines
}
