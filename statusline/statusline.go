// Package statusline lays out a heirline-inspired component tree of
// already-styled content into one or more lines: leaves are hidden by a
// Condition, positioned into a Left/Center/Right zone, and reflowed onto
// additional lines (rather than truncated) when they don't fit the given
// width.
package statusline

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/garthdw/bubblekit/numeric"
)

// Align is which zone of a rendered line a leaf Node's content flows into.
type Align int

const (
	Left Align = iota
	Center
	Right
)

// Node is one entry in a statusline's component tree, heirline-style:
// Condition (if non-nil) gates the whole subtree as a unit — false hides
// this node and every descendant, so an app can group segments that only
// make sense together (e.g. a response status + duration pair) under one
// condition. A node is a leaf when Provider is set (Align/Priority apply
// to it) and a branch when Children is set; Provider is only called when
// Condition passes, so expensive formatting is skipped for hidden nodes.
type Node struct {
	Condition func() bool
	Provider  func() string
	Children  []Node
	Align     Align
	Priority  int
}

// leaf is a flattened, measured Node ready for packing.
type leaf struct {
	text     string
	width    int
	align    Align
	priority int
}

// flatten walks root depth-first, skipping any node (and its subtree)
// whose Condition reports false, and collecting every node with a
// Provider as a measured leaf in document order.
func flatten(root []Node) []leaf {
	var leaves []leaf
	var walk func(nodes []Node)
	walk = func(nodes []Node) {
		for _, n := range nodes {
			if n.Condition != nil && !n.Condition() {
				continue
			}
			if n.Provider != nil {
				text := n.Provider()
				leaves = append(leaves, leaf{
					text:     text,
					width:    lipgloss.Width(text),
					align:    n.Align,
					priority: n.Priority,
				})
			}
			if n.Children != nil {
				walk(n.Children)
			}
		}
	}
	walk(root)
	return leaves
}

// Render flattens root's visible leaves (skipping any whose Condition is
// false) and packs them, in document order, into lines of at most width
// display columns — like flowing words in a paragraph, each leaf wraps
// onto a new line individually rather than the whole statusline
// truncating. Each leaf lands in its line's Left/Center/Right zone; gap
// separates adjacent leaves within the same zone, and the remaining width
// on each line is distributed between zones (center is centered in
// whatever space is left between the left and right content, matching a
// classic 3-part statusline). At least one leaf is always placed per
// line, even if it alone overflows width, so a single very wide segment
// doesn't stall the wrap. If maxLines > 0 and the content needs more
// lines than that, whole leaves are dropped lowest-Priority-first (ties
// broken by document order) and the line is repacked until it fits;
// maxLines <= 0 means unlimited lines.
func Render(root []Node, width int, gap string, maxLines int) []string {
	leaves := flatten(root)
	lines := pack(leaves, width, gap)

	for maxLines > 0 && len(lines) > maxLines && len(leaves) > 0 {
		drop := 0
		for i, l := range leaves {
			if l.priority < leaves[drop].priority {
				drop = i
			}
		}
		leaves = append(leaves[:drop], leaves[drop+1:]...)
		lines = pack(leaves, width, gap)
	}
	return lines
}

// pack is Render's single-pass line-wrapping core, kept separate so
// Render's priority-drop loop can call it repeatedly without re-flattening
// the tree each time.
func pack(leaves []leaf, width int, gap string) []string {
	if len(leaves) == 0 {
		return nil
	}
	gapW := lipgloss.Width(gap)

	var lines []string
	var buckets [3][]leaf
	total := 0
	empty := true

	flush := func() {
		if empty {
			return
		}
		lines = append(lines, renderLine(buckets, width, gap))
		buckets = [3][]leaf{}
		total = 0
		empty = true
	}

	for _, l := range leaves {
		add := l.width
		if len(buckets[l.align]) > 0 {
			add += gapW
		}
		if total+add > width && !empty {
			flush()
			add = l.width
		}
		buckets[l.align] = append(buckets[l.align], l)
		total += add
		empty = false
	}
	flush()
	return lines
}

// renderLine composes one line from its zone buckets: left content at the
// start, right content at the end, and center content centered in
// whatever space remains between them.
func renderLine(buckets [3][]leaf, width int, gap string) string {
	left := joinLeaves(buckets[Left], gap)
	center := joinLeaves(buckets[Center], gap)
	right := joinLeaves(buckets[Right], gap)

	mid := numeric.Clamp(width-lipgloss.Width(left)-lipgloss.Width(right), 0, width)
	return left + lipgloss.PlaceHorizontal(mid, lipgloss.Center, center) + right
}

func joinLeaves(leaves []leaf, gap string) string {
	texts := make([]string, len(leaves))
	for i, l := range leaves {
		texts[i] = l.text
	}
	return strings.Join(texts, gap)
}
