# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`bubblekit` (module `github.com/garthdw/bubblekit`) is a shared Go library of small, dependency-light helpers extracted for reuse across Bubble Tea v2 TUI apps. It is not an application — there is no `main` package. Each package is independently useful and intentionally decoupled from any specific app's domain types (generics are used where an app's own model type needs to flow through, e.g. `pane.Pane[M any]`).

## Commands

```sh
go build ./...              # compile all packages
go test ./...                # run all tests
go test ./layout/...         # test a single package
go test ./layout/ -run TestName   # run a single test
go vet ./...                 # static checks
```

## Architecture

Six independent packages, each covering one concern:

- **`numeric`** — dependency-free numeric helpers shared by the rest of the module (and safe for consuming apps to use directly): `Clamp` floors then ceils a value into a range, degrading gracefully to `hi` even when `hi < lo`.

- **`keymap`** — a context-scoped keybinding table (`Map`). Lookups fall through to `Parent` when a key isn't bound in the current context (Zed-style context cascade / neovim buffer-local-over-global). Used to resolve a raw key string to an action name; callers own the actual `switch` on action names.

- **`layout`** — rendering/layout math that doesn't touch any app's domain types:
  - `titledpane.go`: `RenderTitledPane` draws a lipgloss-bordered box with the title embedded in the top border edge (lazygit-style `╭─ Files ──╮`), plus `TruncateToWidth` for display-width-aware truncation with an ellipsis.
  - `strip.go`: `VisibleRange`/`ClampOffset` do viewport windowing over a horizontal strip of variable-width items (e.g. a tab bar) — which items are visible, whether to show left/right overflow indicators, and how to adjust the scroll offset to keep an active index in view.
  - `helpcolumns.go`: `SplitHelpColumns`/`RenderHelpColumn` bin-pack titled help entry groups into columns of at most N rows each, never splitting a group across columns (e.g. a `?`-triggered help popup).

- **`pane`** — modal/pane composition architecture:
  - `Pane[M any]` — a focusable region: handles one keypress, renders at a given size, reports desired size, and reports whether it's mid-text-edit (`IsEditing`) so callers can skip global/leader key dispatch while typing.
  - `Popup[M any]` — structurally identical to `Pane` (deliberately, so a `Popup` already satisfies `Pane` with no adapter) but meant to live in `Stack[M]`, a LIFO stack of modal layers where the topmost popup owns all input (e.g. a confirm dialog pushed on top of an editor popup).
  - `ConfirmPopup[M any]` — a ready-made yes/no modal; callers inject their own keymap, render function (for app-specific titled-pane chrome), and yes/no callbacks.
  - `CompositeCenteredPopup` — centers a rendered popup string over the main view using lipgloss layers.
  - `EditingKeyPress` — the shared enter=commit / esc=cancel / else=passthrough contract every inline text editor in a consuming app should use.
  - Both `Pane` and `Popup` are parametrized over the app's own model type `M`, so `bubblekit` never needs to know the app's model shape.

- **`theme`** — themeable color/style schema:
  - `theme.go`: `Theme` = fixed `ChromeTokens` (accent, background, border, text, error, etc. — the styling any TUI app needs for its own frame) + `Groups map[string]map[string]Token` for arbitrary app-defined token groups (e.g. HTTP method colors, syntax highlight captures). `Token.Bold`/`Italic`/`Underline` are `*bool` (not `bool`) specifically so "unset" is distinguishable from "explicitly false" during merge.
  - `merge.go`: `Merge` decodes user-supplied theme JSON and merges it field-by-field over a base theme. Invalid/missing fields never fail the whole merge — they silently (missing) or with a stderr warning (invalid, e.g. bad hex color) fall back to the base value for just that field.
  - `loader.go`: `Loader` resolves an app's theme from (in order) a built-in name match, an explicit file path, or `$XDG_CONFIG_HOME/<ConfigSubdir>/theme.json`. A missing/unreadable/malformed file always degrades to `Base` rather than erroring.
  - `style.go`: `Token.ApplyTo(lipgloss.Style)` applies just the token's color/attributes onto an existing style, leaving borders/padding/background untouched.

- **`statusline`** — a heirline-inspired component tree for status bars: `Node` is a leaf (`Provider` renders already-styled text, `Align` places it in a Left/Center/Right zone, `Priority` orders it for dropping) or a branch (`Children`, gated as a unit by `Condition`). `Render` flattens the visible tree in document order and packs it into one or more lines of at most `width` columns, reflowing individual leaves onto additional lines instead of truncating when they overflow; `maxLines` (if > 0) caps the line count by dropping the lowest-`Priority` leaves first.

## Conventions specific to this repo

- Every exported type/func has a doc comment explaining *why*, not just what — match that style in new code (see e.g. the reasoning comments in `keymap.Map`, `theme.Token`, `pane.Pane`).
- Packages depend on `charm.land/bubbletea/v2` and `charm.land/lipgloss/v2` (not the older `github.com/charmbracelet/bubbletea` / `lipgloss` v1 import paths) — keep new code on the `charm.land/*` v2 modules.
- No package here should import another package's app-specific consumer or introduce a dependency on a concrete app's model type; generics (`[M any]`) are how app state flows through `pane`.
- Run `gofmt` before considering a change done. Keep exported surface area minimal — unexported helpers stay unexported unless a consuming app actually needs them.
- `theme.Loader.Builtin` panics on a malformed embedded builtin theme rather than returning an error — that's deliberate (the doc comment explains it's treated as a build-time invariant), not a pattern to generalize elsewhere; runtime-supplied input (theme files, user JSON) always degrades gracefully instead.

## Code comments

Comments should only exist to help when the code is confusing, or to explain purpose/rationale that isn't obvious from reading it. Don't add a comment if it's easily understood from the code itself, and don't leave comments describing how the code used to work. The exception is a comment explaining *why* code is written in an unusual way to handle a specific edge case or issue — that kind of non-obvious rationale is worth keeping (see e.g. `theme.Loader.Builtin`'s panic-vs-error rationale above, or the doc comments on `keymap.Map`/`theme.Token`).

## What's NOT here

No CI config, lint config, or Justfile/Makefile — `go build ./...`, `go test ./...`, and `go vet ./...` are the only checks currently in place.

## Surface mid-implementation decisions; decide them together

Planning can't anticipate everything. When a decision surfaces while implementing — a design choice, a tradeoff, a scope cut, a "this turned out harder than expected, so maybe X" — don't quietly make the call and keep going, even with a clear recommendation and even if the call seems small. Stop, lay out the options and a recommendation, and let the user weigh in.

This isn't a request to stop and ask about every trivial detail; obvious mechanical choices with one sensible answer don't need a checkpoint. It's about genuine forks — the ones where a reasonable person might pick differently, or where the call trades away something assumed going in (scope, API shape, behavior). When in doubt, surface it.
