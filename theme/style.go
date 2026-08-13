package theme

import "charm.land/lipgloss/v2"

// ApplyTo sets base's foreground color and any bold/italic/underline
// attributes the token specifies, leaving everything else (borders,
// padding, background) as base already has it.
func (t Token) ApplyTo(base lipgloss.Style) lipgloss.Style {
	s := base
	if t.Color != "" {
		s = s.Foreground(lipgloss.Color(t.Color))
	}
	if t.Bold != nil {
		s = s.Bold(*t.Bold)
	}
	if t.Italic != nil {
		s = s.Italic(*t.Italic)
	}
	if t.Underline != nil {
		s = s.Underline(*t.Underline)
	}
	return s
}
