// Package pane provides a modal/pane composition architecture for Bubble
// Tea apps: a Pane interface for focusable regions, a Popup interface plus
// stack for modal layering, and small shared helpers (a centered-popup
// compositor, a generic confirm dialog, and the enter=commit/esc=cancel
// inline-edit contract).
//
// Both Pane and Popup are parametrized over the app's own model type M, so
// the app's panes/popups can take *M/M directly without bubblekit needing
// to know its shape.
package pane

import tea "charm.land/bubbletea/v2"

// Pane is a self-contained, focusable region of the UI: something that can
// handle key input, render itself at a given size, and report its desired
// size. Pane is deliberately identical in shape to Popup — a Popup already
// satisfies Pane structurally, with no code change required. A modal
// should hold a Pane value for its body and delegate to it, keeping
// modal-lifecycle concerns (dirty-close confirmation, stack push/pop,
// persistence) on the outer Popup type.
type Pane[M any] interface {
	// Update handles one key press while this pane has focus. Returning
	// done=true signals the pane is finished (e.g. wants to close); panes
	// that never close themselves always return done=false.
	Update(m *M, msg tea.KeyPressMsg) (cmd tea.Cmd, done bool)
	// View renders the pane's content at the given size.
	View(m M, width, height int) string
	// Size returns the pane's desired (width, height) given the current
	// terminal size.
	Size(m M) (width, height int)
	// IsEditing reports whether the pane is currently capturing raw text
	// input (e.g. an inline text field), so the caller can skip
	// leader/global key dispatch entirely rather than relying on each pane
	// to guard for itself.
	IsEditing(m M) bool
}

// EditingKeyPress implements the shared enter=commit / esc=cancel / else=
// passthrough contract used by every inline text editor.
func EditingKeyPress(msg tea.KeyPressMsg, commit, cancel func(), passthrough func(tea.KeyPressMsg) tea.Cmd) tea.Cmd {
	switch msg.String() {
	case "enter":
		commit()
		return nil
	case "esc":
		cancel()
		return nil
	}
	return passthrough(msg)
}
