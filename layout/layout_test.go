package layout

import "testing"

func TestTruncateToWidth(t *testing.T) {
	cases := []struct {
		s    string
		maxW int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello w…"},
		{"hello", 1, "…"},
		{"hello", 0, "…"},
	}
	for _, c := range cases {
		if got := TruncateToWidth(c.s, c.maxW); got != c.want {
			t.Errorf("TruncateToWidth(%q, %d) = %q, want %q", c.s, c.maxW, got, c.want)
		}
	}
}

func TestVisibleRangeAlwaysIncludesOne(t *testing.T) {
	lo, hi, showLeft, showRight := VisibleRange([]int{100}, 0, 10, "<|", "|>")
	if lo != 0 || hi != 1 {
		t.Errorf("VisibleRange = (%d, %d), want (0, 1)", lo, hi)
	}
	if showLeft || showRight {
		t.Errorf("showLeft=%v showRight=%v, want both false for a single item", showLeft, showRight)
	}
}

func TestVisibleRangeShowsIndicators(t *testing.T) {
	widths := []int{5, 5, 5, 5, 5}
	lo, hi, showLeft, showRight := VisibleRange(widths, 2, 12, "<|", "|>")
	if showLeft != true {
		t.Errorf("showLeft = %v, want true (offset > 0)", showLeft)
	}
	if lo != 2 {
		t.Errorf("lo = %d, want 2", lo)
	}
	if hi <= lo {
		t.Errorf("hi = %d, want > lo", hi)
	}
	if hi < len(widths) && !showRight {
		t.Errorf("showRight = %v, want true when hi < n", showRight)
	}
}

func TestClampOffsetKeepsActiveVisible(t *testing.T) {
	widths := []int{5, 5, 5, 5, 5}
	offset := ClampOffset(widths, 4, 0, 12, "<|", "|>")
	lo, hi, _, _ := VisibleRange(widths, offset, 12, "<|", "|>")
	if 4 < lo || 4 >= hi {
		t.Errorf("active index 4 not in visible range [%d, %d) at offset %d", lo, hi, offset)
	}
}

func TestSplitHelpColumnsNeverSplitsGroup(t *testing.T) {
	groups := []HelpGroup{
		{Title: "A", Entries: []HelpEntry{{"a", "one"}, {"b", "two"}}},
		{Title: "B", Entries: []HelpEntry{{"c", "three"}}},
	}
	cols := SplitHelpColumns(groups, 3)
	total := 0
	for _, col := range cols {
		total += len(col)
	}
	if total != len(groups) {
		t.Fatalf("expected all %d groups distributed across columns, got %d", len(groups), total)
	}
}

func TestRenderHelpColumn(t *testing.T) {
	col := []HelpGroup{{Title: "A", Entries: []HelpEntry{{"a", "one"}}}}
	text, w, lines := RenderHelpColumn(col,
		func(title string) string { return title },
		func(key, label string) string { return key + "  " + label },
		func(s string) int { return len(s) },
	)
	if text == "" {
		t.Error("expected non-empty text")
	}
	if w != len("a  one") {
		t.Errorf("width = %d, want %d", w, len("a  one"))
	}
	if lines != GroupLines(col[0]) {
		t.Errorf("lines = %d, want %d", lines, GroupLines(col[0]))
	}
}
