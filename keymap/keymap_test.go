package keymap

import "testing"

func TestLookupFallsThroughToParent(t *testing.T) {
	global := New("global", nil, map[string]string{"q": "quit", "?": "help"})
	pane := New("pane", global, map[string]string{"enter": "select", "q": "close-pane"})

	cases := []struct {
		key    string
		action string
		ok     bool
	}{
		{"enter", "select", true},
		{"q", "close-pane", true}, // pane shadows global
		{"?", "help", true},       // falls through to global
		{"x", "", false},
	}

	for _, c := range cases {
		action, ok := pane.Lookup(c.key)
		if action != c.action || ok != c.ok {
			t.Errorf("Lookup(%q) = (%q, %v), want (%q, %v)", c.key, action, ok, c.action, c.ok)
		}
	}
}

func TestLookupNoParent(t *testing.T) {
	m := New("root", nil, map[string]string{"q": "quit"})
	if _, ok := m.Lookup("x"); ok {
		t.Error("expected Lookup to fail for unbound key with no parent")
	}
}
