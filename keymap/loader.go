package keymap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Loader resolves and loads a user's keymap.json override for a specific
// app, mirroring theme.Loader's degrade-gracefully behavior.
type Loader struct {
	// ConfigSubdir is the directory under the user's config dir holding
	// keymap.json, e.g. "httgo" for ~/.config/httgo/keymap.json.
	ConfigSubdir string
}

// Load resolves and reads a keymap override, merging it over base. base is
// supplied per call rather than held on Loader since it's the caller's own
// hardcoded *Map tree. overridePath is treated as a file path, falling
// back to $XDG_CONFIG_HOME/<ConfigSubdir>/keymap.json (or
// ~/.config/<ConfigSubdir>/keymap.json when unset) if empty. A missing
// file silently yields base unchanged. An unreadable or malformed file
// falls back to base with a warning on stderr; invalid individual entries
// fall back per-entry with their own warning. Load only returns an error
// for conditions unrelated to the keymap file itself (e.g. no resolvable
// config directory).
func (l Loader) Load(overridePath string, base map[string]*Map) (map[string]*Map, error) {
	path := overridePath
	if path == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("resolving config directory: %w", err)
		}
		path = filepath.Join(configDir, l.ConfigSubdir, "keymap.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return base, nil
		}
		fmt.Fprintf(os.Stderr, "keymap: could not read %q (%v), using built-in defaults\n", path, err)
		return base, nil
	}

	merged, err := Merge(data, base, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keymap: could not parse %q (%v), using built-in defaults\n", path, err)
		return base, nil
	}

	return merged, nil
}
