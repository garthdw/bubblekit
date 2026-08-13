// Package theme provides a themeable color/style schema for Bubble Tea
// apps: a fixed set of chrome tokens common to any TUI, plus arbitrary
// named token groups an app defines for its own domain (e.g. HTTP method
// colors, syntax-highlighting captures).
package theme

// Token is a single themeable style: a color plus optional text
// attributes. Attributes are *bool (not bool) so "unset" is distinguishable
// from "explicitly false" during merge with a partial user override.
type Token struct {
	Color     string `json:"color,omitempty"`
	Bold      *bool  `json:"bold,omitempty"`
	Italic    *bool  `json:"italic,omitempty"`
	Underline *bool  `json:"underline,omitempty"`
}

// ChromeTokens covers the styling any Bubble Tea app is likely to need for
// its own frame: accents, backgrounds, borders, and body/muted text.
// App-specific chrome (e.g. a "folder" or "auth" color) belongs in Groups
// instead.
type ChromeTokens struct {
	Accent         Token `json:"accent"`
	Background     Token `json:"background"`
	Border         Token `json:"border"`
	BorderInactive Token `json:"borderInactive"`
	Text           Token `json:"text"`
	MutedText      Token `json:"mutedText"`
	Error          Token `json:"error"`
	Loading        Token `json:"loading"`
}

// Theme is Chrome plus any number of app-defined token groups, each a flat
// name-to-Token map (e.g. Groups["methods"]["GET"]).
type Theme struct {
	Chrome ChromeTokens                `json:"chrome"`
	Groups map[string]map[string]Token `json:"groups,omitempty"`
}
