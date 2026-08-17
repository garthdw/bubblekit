# bubblekit

Shared, dependency-light Go helpers for Bubble Tea v2 TUI apps.

## Configurable keybindings (`keymap`)

An app defines its keybindings as a tree of named, parent-chained
`*keymap.Map` values — child contexts fall through to their parent for any
key they don't bind themselves:

```go
// keys.go
var globalMap = keymap.New("global", nil, map[string]string{
	"q":      "quit",
	"ctrl+c": "quit",
	"?":      "help",
})

var paneMap = keymap.New("pane", globalMap, map[string]string{
	"enter": "select",
	"esc":   "close-pane",
})

var baseMaps = map[string]*keymap.Map{
	"global": globalMap,
	"pane":   paneMap,
}
```

At startup, load an optional user override on top of these defaults with
`keymap.Loader`:

```go
loader := keymap.Loader{ConfigSubdir: "myapp"}
maps, err := loader.Load(*keymapFlag, baseMaps) // *keymapFlag may be "" for the XDG default
```

`Load` resolves `*keymapFlag` as a file path, falling back to
`$XDG_CONFIG_HOME/myapp/keymap.json` when empty. A missing file yields
`baseMaps` unchanged; a malformed one warns to stderr and does the same.

A user's `~/.config/myapp/keymap.json` can override individual bindings per
context, without having to redeclare the whole map:

```json
{
  "global": {
    "q": "quit-app"
  },
  "pane": {
    "esc": null,
    "backspace": "close-pane"
  }
}
```

This remaps `q` in the `global` context, removes (`null`) the `esc` binding
from `pane` so it falls through to whatever (if anything) `global` binds
for `esc`, and adds a new `backspace` binding to `pane`. Every other
binding — `ctrl+c`, `?`, `enter` — is untouched and keeps its default. Since
`pane`'s parent is `global`, the merged tree still cascades: looking up a
key on `pane` that neither map overrides continues to resolve through
`global` exactly as it did before loading the override.
