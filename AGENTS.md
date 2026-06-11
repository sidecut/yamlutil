# AGENTS.md - yamlutil Codebase Guide

## Project Overview

`yamlutil` is a Go CLI tool for YAML file manipulation. It provides utilities to list keys and sort YAML files using the Cobra CLI framework. The tool can process files from arguments or stdin, making it suitable for pipeline usage.

## Essential Commands

### Build and Run
```bash
# Build the project
go build -v

# Run directly with go
go run main.go [command]

# Build and run binary
./yamlutil [command]

# Get help
./yamlutil --help
./yamlutil [command] --help
```

### Available Commands
- `yamlutil list [files...]` - Lists all keys in YAML files (dot notation)
- `yamlutil sort [files...]` - Sorts YAML keys alphabetically
  - `-r, --replace` - In-place sort (replaces original files)
  - `-a, --auto` - Automatic *.sorted.yaml filename generation
- `yamlutil completion` - Shell completion scripts

### Dependencies
```bash
# Install/update dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Architecture

### Project Structure
```
├── main.go                 # Entry point, calls cmd.Execute()
├── cmd/                   # Cobra CLI commands
│   ├── root.go           # Root command and configuration
│   ├── list.go           # List command implementation  
│   └── sort.go           # Sort command implementation
├── argsprocessor/        # Shared argument processing logic
│   └── args.go          # File/stdin input handling
├── sample.yaml          # Test files
└── sample2.yaml
```

### Key Components

**Main Entry Point (`main.go`)**
- Simple wrapper that calls `cmd.Execute()`

**Root Command (`cmd/root.go`)**
- Cobra setup with Viper configuration
- Config file support (`$HOME/.yamlutil.yaml`)
- Global `--config` flag

**Command Pattern (`cmd/list.go`, `cmd/sort.go`)**
- Each command follows Cobra pattern with `Use`, `Short`, `Run`
- Commands use `argsprocessor.ProcessArgs()` for consistent file/stdin handling
- Both support multiple files and stdin input

**Args Processor (`argsprocessor/args.go`)**
- Centralizes file vs stdin logic
- Uses `-` constant to represent stdin
- Provides `ProcessArgs(args, stdin, callback)` interface

## Code Patterns and Conventions

### Type Definitions
```go
type genericMap map[interface{}]interface{}  // For flexible YAML parsing
type stringMap map[string]interface{}        // For string-keyed maps
```

### Key Handling Pattern
- Non-string keys are quoted in output: `."0"`, `."false"`, `."true"`
- String keys use dot notation: `.key`, `.parent.child`
- Array indices shown as: `.key[0]`, `.key[1]`

### Error Handling
- Uses `cobra.CheckErr(err)` throughout for consistent CLI error handling
- Commands return errors that bubble up to Cobra's error handling

### YAML Processing
- Uses `gopkg.in/yaml.v2` for parsing/marshaling
- Handles various YAML data types (string, bool, int, arrays, maps, nil)
- The `list` command recursively traverses nested structures
- The `sort` command uses Go's map iteration (which is naturally sorted when marshaled)

## Testing and Development

### No Test Suite
- **Important**: This project currently has no automated tests
- Test manually using the provided sample files:
  ```bash
  ./yamlutil list sample.yaml
  ./yamlutil sort sample.yaml
  ./yamlutil sort -a sample.yaml  # Creates sample.sorted.yaml
  ```

### Sample Files
- `sample.yaml` - Basic test file with various YAML data types
- `sample2.yaml` - Multi-document YAML (starts with `---`)

### Manual Testing Commands
```bash
# Test stdin processing
cat sample.yaml | ./yamlutil list
cat sample.yaml | ./yamlutil sort

# Test multiple files
./yamlutil list sample.yaml sample2.yaml

# Test sort options
./yamlutil sort --replace sample.yaml     # In-place replacement
./yamlutil sort --auto sample.yaml       # Creates sample.sorted.yaml
```

## Gotchas and Important Details

### Key Output Format
- String keys: `.key`
- Non-string keys: `."key"` (quoted)
- This affects keys like `0`, `false`, `true`, which are non-string in YAML

### Sort Command Behavior
- Mutually exclusive flags: can't use `--auto` and `--replace` together
- Can't use `--auto` or `--replace` with stdin (no filename to work with)
- When no output specified, writes to stdout

### File Processing
- Both commands support multiple file processing
- When processing multiple files, each filename is printed as a header
- The `argsprocessor` package handles the file vs stdin logic consistently

### Dependencies & Modernization Needed
- **Code Quality**: Multiple deprecation warnings for `io/ioutil` (should use `io` or `os` instead)
- **Go Version**: Uses `interface{}` instead of modern `any` type (Go 1.18+ feature)
- **Dependencies**: Uses older yaml.v2 instead of yaml.v3

### Configuration
- Supports config file at `$HOME/.yamlutil.yaml`
- Uses Viper for configuration management
- Environment variables automatically loaded with `viper.AutomaticEnv()`

### Build Details
- Requires Go 1.18+ (specified in go.mod)
- No external build tools (Makefile, CI/CD) currently configured
- Simple `go build` creates the binary

## Development Tasks

If extending this project:

1. **Add Tests**: Create `*_test.go` files with table-driven tests
2. **Modernize Code**: Replace `io/ioutil` with `io`/`os`, use `any` instead of `interface{}`
3. **Add CI/CD**: Consider GitHub Actions for testing and building
4. **Error Handling**: The sort command has complex file handling that could benefit from better error messages
5. **YAML v3**: Consider upgrading to `gopkg.in/yaml.v3` for better YAML 1.2 support

## Configuration Examples

Example `$HOME/.yamlutil.yaml`:
```yaml
# Currently no documented configuration options
# Config file support is scaffolded but not used
```

This codebase is straightforward but uses solid patterns for CLI tools. The separation between command logic and argument processing makes it easy to add new commands.