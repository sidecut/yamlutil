# Copilot Instructions

## Project overview

`yamlutil` is a small Rust CLI (clap + serde_yml) for working with YAML files.
Entry point is `src/main.rs`, which defines the clap root command and
dispatches to the `list` and `sort` subcommands. The Go sources were removed
in the Rust port; the crate layout is the standard `cargo new --bin` output
(`src/main.rs` plus subcommand modules in `src/`):

- `src/main.rs` — root command, config-file probe (`--config` or
  `$HOME/.yamlutil.yaml`), dispatches to subcommands
- `src/args.rs` — `process_args` / `Source` / `STDIN` constant (`"-"`).
  Mirrors the old `argsprocessor` package. **Always** route file-or-stdin
  iteration through this helper so the per-file header rule (printed when
  more than one file is supplied, suppressed for stdin) stays consistent.
- `src/list.rs` — `list` subcommand. Recursively prints dotted paths
  (`parent.child`) and array indices (`key[i]`). Uses `serde_yml::Value`
  via an `IndexMap` to preserve document insertion order and the original
  scalar styles of keys like `0`, `false`, `null`.
- `src/sort.rs` — `sort` subcommand. Sorts the **top level** of a mapping
  lexically and re-serializes. Flags: `-r/--replace` (in-place) and
  `-a/--auto` (writes `name.sorted.yaml` alongside the input).

## Build, test, lint

Standard cargo toolchain (Rust 1.96 / edition 2024):

- Build: `cargo build` (run from the repo root)
- Run: `cargo run -- <subcommand> [args...]` (e.g. `cargo run -- list sample.yaml`)
- Release build: `cargo build --release`
- Format: `cargo fmt` (no `rustfmt.toml` is checked in — defaults apply)
- Lint: `cargo clippy --all-targets -- -D warnings`
- Tests: there are none today. If you add a test file under `src/`, run
  a single test with `cargo test name_substring` and the full suite with
  `cargo test`. Module path is `yamlutil` (see `Cargo.toml`).

## Conventions and gotchas

- **Subcommand pattern**: each subcommand module exposes a `*Args` struct
  deriving `clap::Args` and a `pub fn run(args: *Args) -> anyhow::Result<()>`.
  Follow this pattern for new subcommands; the root enum is the only place
  the variant list lives.
- **Flag validation lives in `run`**, not in clap validators. Build
  `anyhow::bail!("…")` (or `bail!` from `anyhow`) for argument errors and
  return the `Result` from `run`. See `sort::run` for the style: stdin
  vs `--auto`/`--replace` checks happen up front before opening files.
- **YAML library is `serde_yml`**, the maintained fork of `serde_yaml`
  (the latter is deprecated upstream). serde_yml is built on `noyalib`,
  whose `Mapping` only accepts `String` keys — non-string YAML keys must
  be normalized to strings before insertion. See `list::key_to_string`
  and `sort::value_key_to_string` for the conversion.
- **Reading input**: always go through `args::process_args` for the
  file-or-stdin branching. The callback receives a `Source`; both
  `list` and `sort` decide whether to print the `filename:` header
  based on `args.files.len() > 1 && source.label != STDIN`. Keep this
  duplicated check in sync across subcommands — there is no shared
  helper for it on purpose (the two subcommands have slightly
  different output semantics).
- **Key formatting in `list`**: `list_keys` walks the document in
  insertion order using an `IndexMap<String, Value>`. Scalar-style
  artifacts (e.g. `Yes` vs `true`, `null` vs `~`) are preserved by
  round-tripping through `serde_yml::Value`. Array elements are
  addressed as `key[i]`. Keep this format stable — `sort` does not
  currently produce these paths, but any future cross-command tooling
  will depend on this exact syntax.
- **Sort output**: `sort_top_level` re-parses with `serde_yml`, copies
  into an `IndexMap`, sorts the keys, and re-serializes. The sort is
  only at the **top level** because nested mappings round-trip
  positionally through `serde_yml::Value`. If a recursive sort is
  added, it must be done in-memory before re-marshaling.
- **Sample fixtures** (`sample.yaml`, `sample2.yaml`) intentionally
  contain weird keys (`0`, `false`, `null`, YAML booleans like `Yes`,
  comments). They are the regression suite for this project — use them
  to verify behavior rather than discarding them.
- **`*.sorted.yaml` is in `.gitignore`**: the `sort --auto` command
  creates sibling files (e.g. `sample.sorted.yaml`) that should not
  be committed. If you add a new sample that produces a sorted
  companion, the existing ignore rule covers it.
- **No Go remains in the tree.** `main.go`, `go.mod`, `go.sum`, `cmd/`,
  and `argsprocessor/` were all removed. If you see any of those
  referenced, it is stale.
- **No external CI config is checked in.** Verify with
  `gh workflow list` before adding CI-specific instructions to PRs.

