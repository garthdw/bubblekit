package theme

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Loader resolves and loads themes for a specific app: its base (default)
// theme, its built-in named themes, and the config subdirectory a user's
// theme.json override lives under.
type Loader struct {
	// ConfigSubdir is the directory under the user's config dir holding
	// theme.json, e.g. "httgo" for ~/.config/httgo/theme.json.
	ConfigSubdir string
	// Base is the app's default theme, normally decoded from an embedded
	// default.json via Merge(data, Theme{}, io.Discard).
	Base Theme
	// Builtins maps built-in theme names (matched case-insensitively) to
	// raw theme JSON, each merged over Base.
	Builtins map[string][]byte
}

// BuiltinNames returns the loader's built-in theme names, sorted.
func (l Loader) BuiltinNames() []string {
	names := make([]string, 0, len(l.Builtins))
	for name := range l.Builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Builtin returns the named built-in theme (case-insensitive), merged
// over Base, and whether name matched a built-in. Embedded theme JSON is
// assumed to be a build-time invariant, so a decode failure panics rather
// than propagating an error a caller could handle.
func (l Loader) Builtin(name string) (Theme, bool) {
	data, ok := l.Builtins[strings.ToLower(name)]
	if !ok {
		return Theme{}, false
	}
	t, err := Merge(data, l.Base, io.Discard)
	if err != nil {
		panic("theme: embedded builtin " + strings.ToLower(name) + " is invalid: " + err.Error())
	}
	return t, true
}

// Load resolves and reads a theme, merging it over Base. overridePath is
// checked against BuiltinNames first, case-insensitively; if it matches,
// that built-in is returned directly. Otherwise it's treated as a file
// path, falling back to $XDG_CONFIG_HOME/<ConfigSubdir>/theme.json (or
// ~/.config/<ConfigSubdir>/theme.json when unset) if empty. A missing file
// silently yields Base. An unreadable or malformed file falls back to
// Base with a warning on stderr; invalid individual fields fall back
// per-field with their own warning. Load only returns an error for
// conditions unrelated to the theme file itself (e.g. no resolvable
// config directory).
func (l Loader) Load(overridePath string) (Theme, error) {
	if overridePath != "" {
		if t, ok := l.Builtin(overridePath); ok {
			return t, nil
		}
	}

	path := overridePath
	if path == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return Theme{}, fmt.Errorf("resolving config directory: %w", err)
		}
		path = filepath.Join(configDir, l.ConfigSubdir, "theme.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return l.Base, nil
		}
		fmt.Fprintf(os.Stderr, "theme: could not read %q (%v), using built-in defaults\n", path, err)
		return l.Base, nil
	}

	merged, err := Merge(data, l.Base, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "theme: could not parse %q (%v), using built-in defaults\n", path, err)
		return l.Base, nil
	}

	return merged, nil
}
