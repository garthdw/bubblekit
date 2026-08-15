// Package numeric provides small numeric helpers with no other
// dependencies, shared across bubblekit's packages (and safe for
// consuming apps to use directly).
package numeric

// Clamp floors v to lo, then ceils the result to hi (applied in that
// order, so it degrades gracefully to hi even when hi < lo, e.g. a
// terminal too narrow to fit the preferred minimum width).
func Clamp(v, lo, hi int) int {
	if v < lo {
		v = lo
	}
	if v > hi {
		v = hi
	}
	return v
}
