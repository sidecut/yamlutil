# AGENTS.md - yamlutil Codebase Guide

## Project Overview

`yamlutil` is a Rust CLI tool for YAML file manipulation. It provides utilities to list keys and sort YAML files using the clap CLI framework. The tool can process files from arguments or stdin, making it suitable for pipeline usage.

## Essential Commands

### Build, Test, and Run
```bash
# Build the project
cargo build

# Release build
cargo build --release

# Run tests
cargo test

# Lint
cargo clippy

# Format
cargo fmt

# Run directly with cargo
cargo run -- [command]

# Run the built binary
./target/debug/yamlutil [command]

# Get help
./target/debug/yamlutil --help
./target/debug/yamlutil [command] --help
```

### Available Commands
- `yamlutil list [files...]` - Lists all keys in YAML files (dot notation)
- `yamlutil sort [files...]` - Sorts YAML keys alphabetically
  - `-r, --replace` - In-place sort (replaces original files)
  - `-a, --auto` - Automatic *.sorted.yaml filename generation
- `yamlutil completion <shell>` - Shell completion scripts (bash, zsh, fish, etc.)

## Architecture

### Project Structure
```
├── Cargo.toml             # Package manifest and dependencies
├── src/
│   ├── main.rs           # Entry point, clap CLI definition, dispatch
│   ├── list.rs           # List command implementation
│   ├── sort.rs           # Sort command implementation
│   └── argsprocessor.rs  # Shared file/stdin input handling
├── sample.yaml           # Test files
└── sample2.yaml
```

### Key Components

**Main Entry Point (`src/main.rs`)**
- clap derive-based CLI (`Cli` struct, `Commands` enum)
- Global `--config` flag (accepted for compatibility, currently unused)
- `completion` subcommand via `clap_complete`
- Errors print to stderr and exit with code 1

**Commands (`src/list.rs`, `src/sort.rs`)**
- Each command exposes a `run()` function called from `main.rs`
- Commands use `argsprocessor::process_args()` for consistent file/stdin handling
- Both support multiple files and stdin input

**Args Processor (`src/argsprocessor.rs`)**
- Centralizes file vs stdin logic
- Uses `STDIN` constant (`-`) to represent stdin
- `process_args(args, callback)` passes filename and file contents to the callback
- `first_document()` parses only the first YAML document (multi-doc files are truncated to the first document, matching the original Go behavior)

## Code Patterns and Conventions

### YAML Processing
- Uses `serde_yaml` (`serde_yaml::Value`, `serde_yaml::Mapping`)
- Handles all YAML value types (string, bool, number, sequence, mapping, null)
- The `list` command recursively traverses nested mappings; sequences list indices only (no recursion into elements)
- The `sort` command recursively sorts mapping keys alphabetically before serializing

### Key Handling Pattern
- Non-string keys are quoted in output: `."0"`, `."false"`, `."null"`
- String keys use dot notation: `.key`, `.parent.child`
- Array indices shown as: `.key[0]`, `.key[1]`

### Error Handling
- Uses `anyhow::Result` throughout; `bail!` for validation errors
- `main()` prints `Error: <message>` to stderr and exits 1

## Testing

### Unit Tests
- Inline `#[cfg(test)]` modules in `src/list.rs` and `src/sort.rs`
- Run with `cargo test`
- Cover key formatting, output filename generation, and recursive sorting

### Manual Testing Commands
```bash
B=./target/debug/yamlutil

# Test stdin processing
cat sample.yaml | $B list
cat sample.yaml | $B sort

# Test multiple files
$B list sample.yaml sample2.yaml

# Test sort options
$B sort --replace sample.yaml     # In-place replacement
$B sort --auto sample.yaml        # Creates sample.sorted.yaml
```

## Gotchas and Important Details

### Key Output Format
- String keys: `.key`
- Non-string keys: `."key"` (quoted)
- serde_yaml is YAML 1.2: bare `yes`/`no` are strings, unlike the original Go yaml.v2 (YAML 1.1) which treated them as booleans. So a `yes:` key lists as `.yes`, not `."true"`.

### Sort Command Behavior
- Mutually exclusive flags: can't use `--auto` and `--replace` together
- Can't use `--auto` or `--replace` with stdin (no filename to work with)
- When no output specified, writes to stdout
- Sorting is recursive (nested mappings are sorted too)

### File Processing
- Both commands support multiple file processing
- When processing multiple files, each filename is printed as a header (`filename:`)
- Multi-document YAML files: only the first document is processed

### Configuration
- `--config` flag is accepted but currently unused (carried over from the Go version's scaffolding)

### Build Details
- Standard cargo project; no external build tools
- Binary output: `target/debug/yamlutil` (or `target/release/yamlutil`)
- `serde_yaml` 0.9 is in maintenance mode (deprecated upstream) but stable; consider `serde_yaml_ng` or `serde_yml` if migration is needed

## History

This project was originally written in Go (Cobra + yaml.v2) and was converted to Rust (clap + serde_yaml). The CLI surface and output format were preserved, except for YAML 1.2 scalar interpretation noted above.
