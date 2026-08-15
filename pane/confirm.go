package pane

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/garthdw/bubblekit/keymap"
	"github.com/garthdw/bubblekit/numeric"
)

// ConfirmPopup is a small reusable yes/no modal, for delete and
// discard-unsaved-changes confirmations. Callers supply the keymap
// resolving "confirm"/"cancel" actions, the help text to show under the
// message, and a Render func for the app's own titled-pane chrome.
type ConfirmPopup[M any] struct {
	Message  string
	HelpText string
	Keymap   *keymap.Map
	OnYes    func(m *M) tea.Cmd
	OnNo     func(m *M) tea.Cmd
	// Render draws the popup body inside the app's titled-pane chrome,
	// e.g. func(w, h int, title, body string) string.
	Render func(width, height int, title, body string) string
	// MaxWidth returns the widest the popup may be for the given model,
	// e.g. derived from terminal width.
	MaxWidth func(m M) int
}

// Default popup sizing: messagePadding accounts for the surrounding chrome
// added around c.Message's own width, minWidth/defaultMaxWidth bound the
// result, and height is fixed since the popup body is always message +
// blank line + help text.
const (
	defaultMaxWidth = 34
	messagePadding  = 8
	minWidth        = 30
	popupHeight     = 5
)

func (c *ConfirmPopup[M]) Update(m *M, msg tea.KeyPressMsg) (tea.Cmd, bool) {
	action, _ := c.Keymap.Lookup(msg.String())
	switch action {
	case "confirm":
		return callCmd(c.OnYes, m), true
	case "cancel":
		return callCmd(c.OnNo, m), true
	}
	return nil, false
}

func callCmd[M any](fn func(m *M) tea.Cmd, m *M) tea.Cmd {
	if fn == nil {
		return nil
	}
	return fn(m)
}

func (c *ConfirmPopup[M]) View(m M, width, height int) string {
	body := c.Message + "\n\n" + c.HelpText
	return c.Render(width-2, height-2, "Confirm", body)
}

func (c *ConfirmPopup[M]) Size(m M) (int, int) {
	maxW := defaultMaxWidth
	if c.MaxWidth != nil {
		maxW = c.MaxWidth(m)
	}
	w := numeric.Clamp(lipgloss.Width(c.Message)+messagePadding, minWidth, maxW)
	return w, popupHeight
}
