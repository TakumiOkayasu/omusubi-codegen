# CLAUDE.md

## Language Preference

**IMPORTANT: All instructions and responses should be in Japanese (日本語).**

## Project Overview

**omusubi-codegen**: Parses C++ abstract base classes using tree-sitter and generates implementation skeletons (.hpp/.cpp).

**Tech Stack:** Go 1.23+ (CGO required), tree-sitter, Cobra CLI, survey/v2, Go templates

## Common Commands

```bash
# Build / Test / Lint
make build          # Build binary (./omusubi-codegen)
make test           # Run tests
make fmt && make lint

# Basic usage
./omusubi-codegen parse --repo /path/to/include [--verbose]
./omusubi-codegen generate --repo /path/to/include
./omusubi-codegen generate  # Auto-detect workspace

# Alpha version (pre-omusubi)
./omusubi-codegen generate --legacy-name

# DevContainer
make devcontainer-build
make devcontainer-clean
```

**Workspace Structure (auto-detected):**

```text
workspace/
├── omusubi/              # Core library (auto-detected)
└── omusubi-m5stack/      # Platform implementation (optional, auto-detected)
```

## Architecture

```text
C++ Headers → Parser (tree-sitter) → Model → Generator + Templates → Generated Files
```

### Package Structure

- **cmd/codegen/main.go**: CLI entry point (Cobra), `parse`/`generate` commands, interactive prompts
- **internal/parser/**: tree-sitter wrapper, AST traversal, pure virtual detection (`= 0`), workspace auto-detection
- **internal/generator/**: Template execution, file output
- **internal/model/**: Data structures (`ClassInfo`, `MethodInfo`, `FileInfo`, etc.)
- **internal/generator/templates/**: `class_header.tmpl`, `class_source.tmpl`

### Key Implementation Details

- **Pure Virtual Detection**: Only methods containing `= 0` are treated as abstract
- **File Extensions**: Preserves original extension (.h→.h+.cpp, .hpp→.hpp+.cpp)
- **Output Filenames**: Auto-converted to snake_case (MyDevice → my_device.hpp)
- **Output Structure**: PlatformIO-compatible (headers in `include/`, sources in `src/`)

## Template Customization

Templates: `internal/generator/templates/` (uses `go:embed`, requires `make clean && make build` after changes)

**Variables:** `.ClassName`, `.BaseClass`, `.BaseClassExt`, `.HeaderExt`, `.Namespace`, `.Methods`, `.HasNamespace`

**Functions:** `formatParameters`, `formatMethodSignature`, `toSnakeCase`, `toLower`, `toUpper`

## Key Conventions

- snake_case filenames, `#pragma once`, copy/move prevention (deleted), `override` keyword, `= default` constructors, `noexcept` destructors

## Troubleshooting

- **Templates not updating**: `make clean && make build`
- **tree-sitter build errors**: Use devcontainer (CGO required)
- **macOS security warning**: `xattr -d com.apple.quarantine ./omusubi-codegen`
