package keymap

import (
	"encoding/json"
	"fmt"
	"io"
)

// Merge decodes user-supplied keymap JSON — a map of context name to
// {key: action} overrides — and merges it over base's per-context Actions,
// key by key, the way theme.Merge merges token fields: a key absent from
// the override keeps base's binding, a non-empty action string overrides
// it, and a JSON null unsets it (so Lookup falls through to Parent
// instead). A context name absent from base becomes a new parentless
// (root) Map — warned about, since it's as likely a typo as an
// intentional new context — because its place in the parent chain can't
// be inferred from JSON. The returned map is entirely new — base's *Map
// values are never mutated — with every Parent pointer re-resolved into
// the returned map, so an override to a parent context is visible to its
// children's Lookup. Invalid/missing data degrades per-key with a
// warning; only a malformed JSON payload as a whole is a hard error.
func Merge(data []byte, base map[string]*Map, warnOut io.Writer) (map[string]*Map, error) {
	var raw map[string]map[string]*string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]*Map, len(base)+len(raw))
	for name := range base {
		result[name] = &Map{}
	}
	for name := range raw {
		if _, ok := result[name]; !ok {
			result[name] = &Map{}
		}
	}

	for name, m := range result {
		baseMap, known := base[name]
		m.Name = name

		var baseActions map[string]string
		if known {
			baseActions = baseMap.Actions
			if baseMap.Parent != nil {
				if parentMap, ok := result[baseMap.Parent.Name]; ok {
					m.Parent = parentMap
				} else {
					m.Parent = baseMap.Parent
				}
			}
		} else {
			fmt.Fprintf(warnOut, "keymap: %q is not a known context, creating a new root context with no parent\n", name)
		}
		m.Actions = mergeActions(name, raw[name], baseActions, warnOut)
	}

	return result, nil
}

// mergeActions applies raw's key-by-key overrides to base (nil for a
// context with no base counterpart, in which case every key is judged
// against an empty base). A key absent from raw keeps base's value; null
// unsets a key that had a base binding (warning if it didn't); an empty
// string is always invalid and warned about, distinct from a null unset.
func mergeActions(context string, raw map[string]*string, base map[string]string, warnOut io.Writer) map[string]string {
	result := make(map[string]string, len(base))
	for k, v := range base {
		result[k] = v
	}
	for key, val := range raw {
		_, hadBase := result[key]
		if val == nil {
			if !hadBase {
				fmt.Fprintf(warnOut, "keymap: %s.%s has no base binding to unset, ignoring\n", context, key)
				continue
			}
			delete(result, key)
			continue
		}
		if *val == "" {
			if hadBase {
				fmt.Fprintf(warnOut, "keymap: %s.%s has empty action, using default\n", context, key)
			} else {
				fmt.Fprintf(warnOut, "keymap: %s.%s has empty action, ignoring\n", context, key)
			}
			continue
		}
		result[key] = *val
	}
	return result
}
