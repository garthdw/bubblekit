// Package keymap provides a context-scoped keybinding table.
package keymap

// Map is a named, context-scoped set of key-string bindings. Lookups that
// miss fall through to Parent, letting a focused pane's bindings override a
// broader context's without redefining every key (Zed-style keybinding
// context cascade; neovim buffer-local-over-global as the mental model).
type Map struct {
	Name    string
	Parent  *Map
	Actions map[string]string // key string (e.g. "ctrl+e") -> action name
}

// New builds a Map with the given name, parent, and bindings.
func New(name string, parent *Map, actions map[string]string) *Map {
	return &Map{Name: name, Parent: parent, Actions: actions}
}

// Lookup resolves key against m, then each ancestor in turn, returning the
// action name and whether any context in the chain defined it.
func (m *Map) Lookup(key string) (action string, ok bool) {
	for cur := m; cur != nil; cur = cur.Parent {
		if a, found := cur.Actions[key]; found {
			return a, true
		}
	}
	return "", false
}
