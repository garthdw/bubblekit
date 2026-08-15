package statusline

import (
	"reflect"
	"testing"
)

func text(s string) func() string { return func() string { return s } }

func TestRenderSingleLineAllZones(t *testing.T) {
	root := []Node{
		{Provider: text("left"), Align: Left},
		{Provider: text("mid"), Align: Center},
		{Provider: text("right"), Align: Right},
	}
	lines := Render(root, 20, " ", 0)
	want := []string{"left    mid    right"}
	if !reflect.DeepEqual(lines, want) {
		t.Errorf("got %q, want %q", lines, want)
	}
}

func TestRenderConditionHidesSubtree(t *testing.T) {
	root := []Node{
		{
			Condition: func() bool { return false },
			Children: []Node{
				{Provider: text("hidden"), Align: Left},
			},
		},
		{Provider: text("shown"), Align: Left},
	}
	lines := Render(root, 20, " ", 0)
	want := []string{"shown               "}
	if !reflect.DeepEqual(lines, want) {
		t.Errorf("got %q, want %q", lines, want)
	}
}

func TestRenderForcesOverwideLeafOnOwnLine(t *testing.T) {
	root := []Node{
		{Provider: text("way-too-wide-for-the-line"), Align: Left},
		{Provider: text("next"), Align: Left},
	}
	lines := Render(root, 5, " ", 0)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %q", len(lines), lines)
	}
	if lines[0] != "way-too-wide-for-the-line" {
		t.Errorf("line 0 = %q", lines[0])
	}
	if lines[1] != "next " {
		t.Errorf("line 1 = %q", lines[1])
	}
}

func TestRenderWrapsOverflowInDocumentOrder(t *testing.T) {
	root := []Node{
		{Provider: text("aaaa"), Align: Left},
		{Provider: text("bbbb"), Align: Left},
		{Provider: text("cccc"), Align: Left},
	}
	lines := Render(root, 9, " ", 0)
	want := []string{"aaaa bbbb", "cccc     "}
	if !reflect.DeepEqual(lines, want) {
		t.Errorf("got %q, want %q", lines, want)
	}
}

func TestRenderMaxLinesDropsLowestPriorityFirst(t *testing.T) {
	root := []Node{
		{Provider: text("aaaa"), Align: Left, Priority: 2},
		{Provider: text("bbbb"), Align: Left, Priority: 0},
		{Provider: text("cccc"), Align: Left, Priority: 1},
	}
	lines := Render(root, 9, " ", 1)
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1: %q", len(lines), lines)
	}
	want := "aaaa cccc"
	if lines[0] != want {
		t.Errorf("got %q, want %q", lines[0], want)
	}
}

func TestRenderMaxLinesZeroNeverDrops(t *testing.T) {
	root := []Node{
		{Provider: text("aaaa"), Align: Left},
		{Provider: text("bbbb"), Align: Left},
		{Provider: text("cccc"), Align: Left},
	}
	lines := Render(root, 9, " ", 0)
	if len(lines) != 2 {
		t.Errorf("got %d lines, want 2: %q", len(lines), lines)
	}
}
