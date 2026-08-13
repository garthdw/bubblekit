package theme

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

var hexColorRE = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type rawToken struct {
	Color     *string `json:"color"`
	Bold      *bool   `json:"bold"`
	Italic    *bool   `json:"italic"`
	Underline *bool   `json:"underline"`
}

type rawChromeTokens struct {
	Accent         *rawToken `json:"accent"`
	Background     *rawToken `json:"background"`
	Border         *rawToken `json:"border"`
	BorderInactive *rawToken `json:"borderInactive"`
	Text           *rawToken `json:"text"`
	MutedText      *rawToken `json:"mutedText"`
	Error          *rawToken `json:"error"`
	Loading        *rawToken `json:"loading"`
}

type rawTheme struct {
	Chrome *rawChromeTokens                      `json:"chrome"`
	Groups map[string]map[string]json.RawMessage `json:"groups"`
}

// Merge decodes user-supplied theme JSON, validates each field, and merges
// it over base (typically an app's default theme). Invalid or missing
// fields silently (missing) or with a warning (invalid) fall back to
// base's value for that field/token; nothing in data ever causes the
// whole merge to fail. Warnings are written to warnOut, one line per
// issue.
func Merge(data []byte, base Theme, warnOut io.Writer) (Theme, error) {
	var raw rawTheme
	if err := json.Unmarshal(data, &raw); err != nil {
		return Theme{}, err
	}

	result := base

	if raw.Chrome != nil {
		mergeChrome(raw.Chrome, &result.Chrome, base.Chrome, warnOut)
	}
	if raw.Groups != nil {
		result.Groups = mergeGroups(raw.Groups, base.Groups, warnOut)
	}

	return result, nil
}

func mergeToken(path string, raw *rawToken, def Token, warnOut io.Writer) Token {
	if raw == nil {
		return def
	}
	color := def.Color
	if raw.Color == nil || *raw.Color == "" {
		fmt.Fprintf(warnOut, "theme: %s missing color, using default\n", path)
	} else if !hexColorRE.MatchString(*raw.Color) {
		fmt.Fprintf(warnOut, "theme: %s has invalid color %q, using default\n", path, *raw.Color)
	} else {
		color = *raw.Color
	}
	tok := Token{Color: color, Bold: def.Bold, Italic: def.Italic, Underline: def.Underline}
	if raw.Bold != nil {
		tok.Bold = raw.Bold
	}
	if raw.Italic != nil {
		tok.Italic = raw.Italic
	}
	if raw.Underline != nil {
		tok.Underline = raw.Underline
	}
	return tok
}

func mergeChrome(raw *rawChromeTokens, out *ChromeTokens, def ChromeTokens, warnOut io.Writer) {
	out.Accent = mergeToken("chrome.accent", raw.Accent, def.Accent, warnOut)
	out.Background = mergeToken("chrome.background", raw.Background, def.Background, warnOut)
	out.Border = mergeToken("chrome.border", raw.Border, def.Border, warnOut)
	out.BorderInactive = mergeToken("chrome.borderInactive", raw.BorderInactive, def.BorderInactive, warnOut)
	out.Text = mergeToken("chrome.text", raw.Text, def.Text, warnOut)
	out.MutedText = mergeToken("chrome.mutedText", raw.MutedText, def.MutedText, warnOut)
	out.Error = mergeToken("chrome.error", raw.Error, def.Error, warnOut)
	out.Loading = mergeToken("chrome.loading", raw.Loading, def.Loading, warnOut)
}

// mergeGroups merges user-supplied token groups over base's groups. A
// group/key present in base but absent from raw keeps its base value; one
// present in raw but absent from base is added as a new entry. An app
// that wants a closed set of group/token names should validate raw itself
// before calling Merge (or post-filter result.Groups).
func mergeGroups(raw map[string]map[string]json.RawMessage, base map[string]map[string]Token, warnOut io.Writer) map[string]map[string]Token {
	result := make(map[string]map[string]Token, len(base))
	for group, toks := range base {
		copyMap := make(map[string]Token, len(toks))
		for k, v := range toks {
			copyMap[k] = v
		}
		result[group] = copyMap
	}
	for group, toks := range raw {
		out, ok := result[group]
		if !ok {
			out = make(map[string]Token, len(toks))
			result[group] = out
		}
		for key, rawVal := range toks {
			var tok rawToken
			if err := json.Unmarshal(rawVal, &tok); err != nil {
				fmt.Fprintf(warnOut, "theme: groups.%s.%s invalid, using default\n", group, key)
				continue
			}
			out[key] = mergeToken(fmt.Sprintf("groups.%s.%s", group, key), &tok, out[key], warnOut)
		}
	}
	return result
}
