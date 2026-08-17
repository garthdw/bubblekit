package keymap

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func testBase() map[string]*Map {
	global := New("global", nil, map[string]string{"q": "quit", "?": "help"})
	pane := New("pane", global, map[string]string{"enter": "select", "esc": "close-pane"})
	return map[string]*Map{"global": global, "pane": pane}
}

func TestMergeKeyByKeyOverride(t *testing.T) {
	base := testBase()
	got, err := Merge([]byte(`{"pane": {"enter": "confirm"}}`), base, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["pane"].Actions["enter"] != "confirm" {
		t.Errorf("pane.enter = %q, want confirm", got["pane"].Actions["enter"])
	}
	if got["pane"].Actions["esc"] != "close-pane" {
		t.Errorf("pane.esc = %q, want close-pane (unchanged)", got["pane"].Actions["esc"])
	}
	if got["global"].Actions["q"] != "quit" {
		t.Errorf("global.q = %q, want quit (unchanged)", got["global"].Actions["q"])
	}
}

func TestMergeParentReresolution(t *testing.T) {
	base := testBase()
	got, err := Merge([]byte(`{"global": {"q": "custom-quit"}}`), base, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	action, ok := got["pane"].Lookup("q")
	if !ok || action != "custom-quit" {
		t.Errorf("pane.Lookup(q) = (%q, %v), want (custom-quit, true)", action, ok)
	}
}

func TestMergeUnsetBinding(t *testing.T) {
	base := testBase()
	got, err := Merge([]byte(`{"pane": {"esc": null}}`), base, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := got["pane"].Actions["esc"]; ok {
		t.Error("expected pane.esc to be unset")
	}
	if _, ok := got["pane"].Lookup("esc"); ok {
		t.Error("expected Lookup(esc) to fail after unset (global has no esc binding)")
	}
}

func TestMergeUnsetWithNoBaseBindingWarns(t *testing.T) {
	base := testBase()
	var buf bytes.Buffer
	got, err := Merge([]byte(`{"pane": {"x": null}}`), base, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected a warning for unsetting a key with no base binding")
	}
	if _, ok := got["pane"].Actions["x"]; ok {
		t.Error("expected pane.x to remain absent")
	}
}

func TestMergeEmptyActionWarns(t *testing.T) {
	base := testBase()
	var buf bytes.Buffer
	got, err := Merge([]byte(`{"pane": {"enter": ""}}`), base, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected a warning for empty action")
	}
	if got["pane"].Actions["enter"] != "select" {
		t.Errorf("pane.enter = %q, want select (base preserved)", got["pane"].Actions["enter"])
	}
}

func TestMergeUnknownContextCreatedRootless(t *testing.T) {
	base := testBase()
	got, err := Merge([]byte(`{"popup": {"esc": "close-popup", "bad": null}}`), base, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	popup, ok := got["popup"]
	if !ok {
		t.Fatal("expected a new popup context")
	}
	if popup.Parent != nil {
		t.Error("expected popup.Parent to be nil")
	}
	if popup.Actions["esc"] != "close-popup" {
		t.Errorf("popup.esc = %q, want close-popup", popup.Actions["esc"])
	}
	if _, ok := popup.Actions["bad"]; ok {
		t.Error("expected popup.bad to be skipped")
	}
}

func TestMergeMalformedJSON(t *testing.T) {
	base := testBase()
	if _, err := Merge([]byte(`{ not valid json`), base, &bytes.Buffer{}); err == nil {
		t.Fatal("expected Merge to return an error for malformed JSON")
	}
}

func TestMergeDoesNotMutateBase(t *testing.T) {
	base := testBase()
	got, err := Merge([]byte(`{"global": {"q": "custom-quit"}}`), base, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base["global"].Actions["q"] != "quit" {
		t.Errorf("base global.q mutated: got %q, want quit", base["global"].Actions["q"])
	}
	if base["global"] == got["global"] {
		t.Error("expected Merge to return a distinct *Map, not the original base pointer")
	}
}

func TestLoaderLoadMissingFile(t *testing.T) {
	base := testBase()
	loader := Loader{}
	got, err := loader.Load(filepath.Join(t.TempDir(), "does-not-exist.json"), base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["global"].Actions["q"] != "quit" {
		t.Errorf("expected base keymap returned, got %q", got["global"].Actions["q"])
	}
}

func TestLoaderLoadUsesConfigFile(t *testing.T) {
	base := testBase()
	dir := t.TempDir()
	path := filepath.Join(dir, "keymap.json")
	if err := os.WriteFile(path, []byte(`{"global": {"q": "custom-quit"}}`), 0o644); err != nil {
		t.Fatalf("writing temp keymap.json: %v", err)
	}
	loader := Loader{}
	got, err := loader.Load(path, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["global"].Actions["q"] != "custom-quit" {
		t.Errorf("global.q = %q, want custom-quit", got["global"].Actions["q"])
	}
}

func TestLoaderLoadMalformedFileFallsBack(t *testing.T) {
	base := testBase()
	dir := t.TempDir()
	path := filepath.Join(dir, "keymap.json")
	if err := os.WriteFile(path, []byte(`{ not valid json`), 0o644); err != nil {
		t.Fatalf("writing temp keymap.json: %v", err)
	}
	loader := Loader{}
	got, err := loader.Load(path, base)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["global"].Actions["q"] != "quit" {
		t.Errorf("expected base keymap returned, got %q", got["global"].Actions["q"])
	}
}
