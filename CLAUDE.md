# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Language Preference

**IMPORTANT: All instructions and responses should be in Japanese (日本語).**
When working in this repository, Claude should communicate in Japanese for all explanations, suggestions, and discussions.

## Project Overview

This is **omusubi-platform-codegen**, a code generation tool for the Omusubi embedded framework. It parses C++ abstract base classes (interfaces) using tree-sitter and generates implementation skeletons (.hpp/.cpp files) with empty method bodies ready for manual implementation.

**Key technologies:**
- Go 1.23+ with CGO (required for tree-sitter)
- tree-sitter C++ parser for AST analysis
- Cobra CLI framework
- AlecAivazis/survey for interactive prompts
- Go templates for code generation

## Common Commands

### Building
```bash
# Build the binary (creates ./omusubi-codegen)
make build

# Clean build artifacts
make clean

# Install dependencies
make deps
```

### Testing
```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage
```

### Code Quality
```bash
# Format code
make fmt

# Run linters (requires golangci-lint)
make lint
```

### Running the Tool

#### Basic Usage (Implementation Generation Only)
```bash
# Parse a repository and list abstract classes
./omusubi-codegen parse --repo /path/to/omusubi/include
./omusubi-codegen parse --repo /path/to/omusubi/include --verbose  # Show method details

# Generate implementations (interactive mode with multi-select)
./omusubi-codegen generate --repo /path/to/omusubi/include

# Generate with CLI arguments
./omusubi-codegen generate \
  --repo /path/to/omusubi/include \
  --base IDevice \
  --class MyDevice \
  --output ./output
```

#### Multi-Repository Workspace Usage (Recommended)
The tool automatically detects the omusubi workspace structure when run from a project directory:

**Workspace Structure:**
```
workspace/
├── omusubi/                  # Core library
│   ├── include/omusubi/
│   └── ...
├── omusubi-m5stack/          # Platform implementation
│   ├── include/
│   └── src/
└── my-project/               # Your project (generated)
    ├── platformio.ini
    ├── src/
    │   ├── main.cpp
    │   └── my_device.hpp
    └── ...
```

**Auto-detection Example:**
```bash
# From workspace directory - auto-detects omusubi/ and omusubi-m5stack/
./omusubi-codegen generate --project --project-name my-m5stack-project

# For alpha version (pre-omusubi) - will be removed after release
./omusubi-codegen generate --legacy-name --project --project-name my-m5stack-project

# Specify paths explicitly
./omusubi-codegen generate \
  --project \
  --project-name my-m5stack-project \
  --core-lib ./omusubi \
  --platform-lib ./omusubi-m5stack \
  --board m5stack-core-esp32

# Generate only implementations (no project structure)
./omusubi-codegen generate \
  --core-lib ./omusubi \
  --output ./my-project/src
```

**What Gets Generated:**
- `--project` flag creates:
  - `platformio.ini` with correct relative paths to libraries
  - `src/main.cpp` with basic Arduino setup
  - `.gitignore` for PlatformIO
  - Implementation files (.hpp/.cpp) in `src/` directory
- Without `--project`:
  - Only implementation files (.hpp/.cpp) in specified output directory

### Development Container
```bash
# Build devcontainer (uses PROJECT_NAME from .env)
make devcontainer-build

# Clean all devcontainer resources (containers, volumes, networks)
make devcontainer-clean

# Direct Docker Compose usage
docker compose -f .devcontainer/compose.yaml up -d
docker exec -it ${PROJECT_NAME}-devcontainer /bin/bash
```

## Architecture

### Core Data Flow
```
C++ Header Files (.hpp/.h)
    ↓
Parser (tree-sitter) → AST traversal
    ↓
Model (ClassInfo, MethodInfo) → Extract pure virtual methods
    ↓
Generator + Templates → Execute Go templates
    ↓
Generated Files (.hpp + .cpp/.c)
```

### Package Structure

**cmd/codegen/main.go**
- CLI entry point using Cobra framework
- Two main commands: `parse` and `generate`
- Interactive user prompts with survey/v2 (multi-select for class selection)
- Handles "select all" vs "choose individually" workflow
- **NEW:** Workspace auto-detection via `parser.DetectWorkspace()`
- **NEW:** PlatformIO project generation with `--project` flag

**internal/parser/**
- `Parser` wraps tree-sitter C++ parser
- `ParseDirectory()`: Recursively walks directories for .hpp/.h files
- `ParseFile()`: Parses individual files, detects file extension (.h vs .hpp)
- AST traversal via `traverseNode()`: Handles namespaces, classes, methods, fields
- Detects pure virtual methods (`= 0`) to mark classes as abstract
- Extracts: access levels (public/protected/private), const/static/virtual modifiers, parameters with default values
- **NEW:** `DetectWorkspace()`: Automatically finds omusubi core and platform libraries in parent directories

**internal/generator/**
- `Generator`: Manages template rendering and file output
- `GenerateImplementation()`: Creates both header and source files
- Template functions: `formatParameters`, `formatMethodSignature`, `toSnakeCase`
- File extension handling: .h→.c, .hpp→.cpp
- Output filenames are automatically snake_cased (e.g., MyDevice → my_device.hpp)
- **NEW:** `GenerateProject()`: Creates complete PlatformIO project structure
- **NEW:** Generates `platformio.ini`, `main.cpp`, and `.gitignore`

**internal/model/**
- Data structures: `ClassInfo`, `MethodInfo`, `ParameterInfo`, `FieldInfo`, `FileInfo`
- `AccessLevel` enum: Public, Protected, Private
- `SourceFileExt` field tracks original file extension for correct generation
- **NEW:** `ProjectConfig`: Holds PlatformIO project configuration with relative library paths

**internal/template/templates/**
- `class_header.tmpl`: Generates .hpp/.h with class declaration, override methods
- `class_source.tmpl`: Generates .cpp/.c with empty method implementations
- **NEW:** `platformio.ini.tmpl`: Generates PlatformIO configuration with library paths
- **NEW:** `main.cpp.tmpl`: Generates basic Arduino setup code
- **NEW:** `gitignore.tmpl`: Generates .gitignore for PlatformIO projects
- Templates use Go text/template with custom functions
- All method signatures include `override` keyword

### Important Implementation Details

**tree-sitter AST Nodes:**
- `namespace_definition`: C++ namespace blocks
- `class_specifier`: Class definitions
- `function_definition`/`declaration`: Method declarations
- `field_declaration`: Member variables
- `access_specifier`: public/protected/private sections

**Pure Virtual Detection:**
- Parser searches for `= 0` in method declaration text
- Only classes with at least one pure virtual method are considered "abstract"
- Only pure virtual methods are included in generated implementations

**File Extension Logic:**
- Original file extension (.h or .hpp) is stored in `ClassInfo.SourceFileExt`
- Generated files match the original convention:
  - `.h` source → `.h` + `.c` output
  - `.hpp` source → `.hpp` + `.cpp` output
- Include statements in generated files use the correct extension

**Interactive CLI Flow:**
1. User runs `generate` command with `--repo`
2. Parser scans repository for abstract classes
3. If classes found: "Select all?" prompt (Yes/No)
4. If "No": Multi-select prompt (arrow keys + space to toggle)
5. Prompt for class name prefix (default: "My")
6. Generate `<Prefix><BaseClassName>` for each selected class

## Dev Container Notes

**Environment Variables (`.env` file):**
- `PROJECT_NAME`: Prefix for container/volume/network names (prevents conflicts)
- `USER`, `UID`, `GID`: Host user mapping (important for file permissions)
- On macOS, UID/GID may need manual configuration

**Resource Naming:**
- Containers: `${PROJECT_NAME}-devcontainer`
- Volumes: `${PROJECT_NAME}-go-modules`, `${PROJECT_NAME}-vscode-extensions`
- Networks: `${PROJECT_NAME}-network`

**CGO Requirement:**
- tree-sitter requires CGO, so `CGO_ENABLED=1` is mandatory
- Cross-compilation is complex; build on target platform when possible

## Testing

Tests use Go's standard testing package. Example pattern:

```go
func TestToSnakeCase(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"MyDevice", "my_device"},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := toSnakeCase(tt.input)
            if result != tt.expected {
                t.Errorf("toSnakeCase(%q) = %q; want %q", tt.input, result, tt.expected)
            }
        })
    }
}
```

Test data is in `testdata/` directory.

## Template Customization

Templates are in `internal/template/templates/` and use Go's `text/template`:

**Available template variables:**
- `.ClassName`: Derived class name
- `.BaseClass`: Base class name
- `.BaseClassExt`: File extension (h/hpp)
- `.Namespace`: C++ namespace
- `.Methods`: Array of MethodInfo (pure virtual methods only)
- `.HasNamespace`: Boolean

**Template functions:**
- `formatParameters`: Formats parameter list
- `formatMethodSignature`: Builds complete method signature with override
- `toSnakeCase`: Converts PascalCase to snake_case
- `toLower`, `toUpper`: String case conversion

**To modify templates:**
1. Edit files in `internal/template/templates/`
2. Templates are embedded via `go:embed`, so rebuild is required: `make clean && make build`

## Key Conventions

- **File naming**: Generated files use snake_case (e.g., `my_device.hpp`)
- **Include guards**: Use `#pragma once` (modern C++ standard)
- **Copy prevention**: Generated classes delete copy constructor and assignment operator
- **Namespace handling**: Parser correctly extracts nested namespaces and preserves them in generated code
- **Method signatures**: Include `override` keyword for all pure virtual method implementations
- **Default constructors**: Use `= default` for constructor/destructor

## Binary Name

The binary is named `omusubi-codegen` (not `codegen`) to avoid conflicts with other tools. The build process uses:
- Makefile variable: `BINARY_NAME=omusubi-codegen`
- Install path via `go install` creates `codegen` in `$GOPATH/bin`
- Symbolic link can be created if needed: `ln -s $GOPATH/bin/codegen $GOPATH/bin/omusubi-codegen`

## Troubleshooting

**Templates not updating:**
Templates are embedded at compile time. After editing templates, run `make clean && make build`.

**tree-sitter build errors:**
Use the devcontainer which has the correct C++ compiler and tree-sitter setup. CGO is required.

**macOS security warnings:**
If the binary is blocked by macOS Gatekeeper, run: `xattr -d com.apple.quarantine ./omusubi-codegen`
