package layout

import (
	"charm.land/lipgloss/v2"

	"github.com/garthdw/bubblekit/numeric"
)

// VisibleRange determines which contiguous run of indices [lo, hi) fits in
// width starting the scan from offset, and whether left/right overflow
// indicators are needed. slotWidths[i] is each item's rendered width
// including any trailing gap. leftIndicator/rightIndicator are the
// (already-rendered) strings shown when items are scrolled off that edge;
// pass "" if a strip never truncates on that side. At least one item is
// always included, even if it alone overflows width, so the strip is
// never empty.
func VisibleRange(slotWidths []int, offset, width int, leftIndicator, rightIndicator string) (lo, hi int, showLeft, showRight bool) {
	n := len(slotWidths)
	if n == 0 {
		return 0, 0, false, false
	}
	offset = numeric.Clamp(offset, 0, n-1)
	showLeft = offset > 0

	avail := width
	if showLeft {
		avail -= lipgloss.Width(leftIndicator)
	}
	rightW := lipgloss.Width(rightIndicator)

	hi = offset
	for hi < n {
		w := slotWidths[hi]
		reserveRight := 0
		if hi+1 < n {
			reserveRight = rightW
		}
		if avail-w-reserveRight < 0 && hi > offset {
			break
		}
		avail -= w
		hi++
	}
	if hi == offset {
		hi = offset + 1
	}
	showRight = hi < n
	lo = offset
	return lo, hi, showLeft, showRight
}

// ClampOffset adjusts offset (an index into slotWidths) so activeIndex
// stays within the visible window VisibleRange would compute for width,
// scanning by index since item widths vary.
func ClampOffset(slotWidths []int, activeIndex, offset, width int, leftIndicator, rightIndicator string) int {
	n := len(slotWidths)
	if n == 0 || activeIndex < 0 {
		return 0
	}
	offset = numeric.Clamp(offset, 0, activeIndex)
	for {
		lo, hi, _, _ := VisibleRange(slotWidths, offset, width, leftIndicator, rightIndicator)
		if activeIndex >= lo && activeIndex < hi {
			return offset
		}
		if activeIndex < lo {
			offset = activeIndex
			continue
		}
		offset++
		if offset >= n {
			return n - 1
		}
	}
}
