package theme

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func testBase() Theme {
	return Theme{
		Chrome: ChromeTokens{
			Accent: Token{Color: "#25A065"},
			Error:  Token{Color: "#EF5350"},
		},
		Groups: map[string]map[string]Token{
			"methods": {
				"GET":  {Color: "#4CAF50"},
				"POST": {Color: "#FF9800"},
			},
			"syntax": {
				"comment": {Color: "#757575", Italic: boolPtr(true)},
				"keyword": {Color: "#CE93D8", Bold: boolPtr(true)},
			},
		},
	}
}

func TestMergeMissingFile(t *testing.T) {
	loader := Loader{Base: testBase()}
	got, err := loader.Load(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Chrome.Accent.Color != testBase().Chrome.Accent.Color {
		t.Errorf("expected base theme, got accent %q", got.Chrome.Accent.Color)
	}
}

func TestMergeMalformedJSON(t *testing.T) {
	var buf bytes.Buffer
	base := testBase()
	if _, err := Merge([]byte(`{ not valid json`), base, &buf); err == nil {
		t.Fatal("expected Merge to return an error for malformed JSON")
	}
}

func TestMergePartialOverride(t *testing.T) {
	base := testBase()
	got, err := Merge([]byte(`{"chrome": {"accent": {"color": "#00FF88"}}}`), base, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Chrome.Accent.Color != "#00FF88" {
		t.Errorf("chrome.accent = %q, want #00FF88", got.Chrome.Accent.Color)
	}
	if got.Chrome.Error.Color != base.Chrome.Error.Color {
		t.Errorf("chrome.error should remain default, got %q", got.Chrome.Error.Color)
	}
	if got.Groups["methods"]["GET"].Color != base.Groups["methods"]["GET"].Color {
		t.Errorf("groups.methods.GET should remain default, got %q", got.Groups["methods"]["GET"].Color)
	}
}

func TestMergeInvalidColorFallsBackPerField(t *testing.T) {
	var buf bytes.Buffer
	base := testBase()
	merged, err := Merge([]byte(`{
		"chrome": {"accent": {"color": "not-a-hex"}},
		"groups": {"methods": {"POST": {"color": "#00FF88"}}}
	}`), base, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Chrome.Accent.Color != base.Chrome.Accent.Color {
		t.Errorf("chrome.accent should fall back to default, got %q", merged.Chrome.Accent.Color)
	}
	if merged.Groups["methods"]["POST"].Color != "#00FF88" {
		t.Errorf("groups.methods.POST = %q, want #00FF88", merged.Groups["methods"]["POST"].Color)
	}
	if !strings.Contains(buf.String(), "chrome.accent") {
		t.Errorf("expected warning mentioning chrome.accent, got: %s", buf.String())
	}
}

func TestMergeAddsNewGroupKey(t *testing.T) {
	base := testBase()
	merged, err := Merge([]byte(`{"groups": {"methods": {"DELETE": {"color": "#F44336"}}}}`), base, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Groups["methods"]["DELETE"].Color != "#F44336" {
		t.Errorf("groups.methods.DELETE = %q, want #F44336", merged.Groups["methods"]["DELETE"].Color)
	}
	if merged.Groups["methods"]["GET"].Color != base.Groups["methods"]["GET"].Color {
		t.Errorf("groups.methods.GET should remain default, got %q", merged.Groups["methods"]["GET"].Color)
	}
}

func TestLoaderBuiltinAndLoad(t *testing.T) {
	loader := Loader{
		Base: testBase(),
		Builtins: map[string][]byte{
			"other": []byte(`{"chrome": {"accent": {"color": "#111111"}}}`),
		},
	}

	got, ok := loader.Builtin("OTHER")
	if !ok {
		t.Fatal("Builtin(\"OTHER\") not found")
	}
	if got.Chrome.Accent.Color != "#111111" {
		t.Errorf("Builtin(\"OTHER\") accent = %q, want #111111", got.Chrome.Accent.Color)
	}

	if names := loader.BuiltinNames(); len(names) != 1 || names[0] != "other" {
		t.Errorf("BuiltinNames() = %v, want [other]", names)
	}

	viaLoad, err := loader.Load("other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if viaLoad.Chrome.Accent.Color != got.Chrome.Accent.Color {
		t.Errorf("Load(\"other\") accent = %q, want %q", viaLoad.Chrome.Accent.Color, got.Chrome.Accent.Color)
	}
}

func TestLoaderLoadUsesConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.json")
	if err := os.WriteFile(path, []byte(`{"chrome": {"accent": {"color": "#00FF88"}}}`), 0o644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	loader := Loader{Base: testBase()}
	got, err := loader.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Chrome.Accent.Color != "#00FF88" {
		t.Errorf("chrome.accent = %q, want #00FF88", got.Chrome.Accent.Color)
	}
}
