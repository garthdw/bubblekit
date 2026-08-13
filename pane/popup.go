package pane

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Popup is anything that can take over keyboard input and render itself as
// a modal layer on top of the main view. Popups are meant to be held in a
// Stack, so a popup (e.g. an editor) can itself push another popup on top
// (e.g. a confirm dialog); the topmost popup owns all key input until it
// pops itself.
type Popup[M any] interface {
	// Update handles one key press while this popup is topmost. Returning
	// done=true pops this popup off the stack.
	Update(m *M, msg tea.KeyPressMsg) (cmd tea.Cmd, done bool)
	// View renders the popup's content at the given size.
	View(m M, width, height int) string
	// Size returns the popup's desired (width, height) given the current
	// terminal size.
	Size(m M) (width, height int)
}

// Stack is a LIFO stack of popups: the top of the stack owns all key
// input, and can itself push further popups on top (e.g. a confirm dialog
// over an editor).
type Stack[M any] struct {
	popups []Popup[M]
}

// Push adds p to the top of the stack.
func (s *Stack[M]) Push(p Popup[M]) {
	s.popups = append(s.popups, p)
}

// Pop removes the top of the stack, if any.
func (s *Stack[M]) Pop() {
	if len(s.popups) == 0 {
		return
	}
	s.popups = s.popups[:len(s.popups)-1]
}

// Top returns the top of the stack, or nil if empty.
func (s *Stack[M]) Top() Popup[M] {
	if len(s.popups) == 0 {
		return nil
	}
	return s.popups[len(s.popups)-1]
}

// Len returns the number of popups currently on the stack.
func (s *Stack[M]) Len() int {
	return len(s.popups)
}

// CompositeCenteredPopup renders popupStr centered over mainContent using
// lipgloss layers.
func CompositeCenteredPopup(mainContent, popupStr string) string {
	baseW, baseH := lipgloss.Size(mainContent)
	popW, popH := lipgloss.Size(popupStr)

	x := (baseW - popW) / 2
	y := (baseH - popH) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	base := lipgloss.NewLayer(mainContent)
	pop := lipgloss.NewLayer(popupStr).X(x).Y(y).Z(1)
	return lipgloss.NewCompositor(base, pop).Render()
}
