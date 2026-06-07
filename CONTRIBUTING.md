# Contributing to slimify

Thanks for your interest in contributing to **slimify**! This document covers everything you need to get started.

## Quick Start

```bash
git clone https://github.com/NotHarshhaa/slimify
cd slimify
go mod tidy
go run . --help
```

## Prerequisites

- **Go 1.22+** — [install](https://go.dev/dl/)
- **Docker** (optional) — only needed for testing `audit` and `fix` commands against real images
- **golangci-lint** (optional) — for running linters locally

## Development

### Building

```bash
# Build the binary
go build -o slimify .

# Run directly without building
go run . audit --help
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
# Install golangci-lint (if not already installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run ./...
```

## Project Structure

```
slimify/
├── main.go                  # Entry point
├── cmd/                     # CLI commands (cobra)
│   ├── root.go              # Root command + global flags
│   ├── audit.go             # `slimify audit` command
│   ├── fix.go               # `slimify fix` command
│   ├── compare.go           # `slimify compare` command
│   └── ignore.go            # `slimify ignore` command
├── pkg/
│   ├── analyzer/            # Image analysis engine
│   │   ├── image.go         # Image loading + analysis
│   │   ├── layer.go         # Per-layer file extraction
│   │   ├── duplicates.go    # Cross-layer duplicate detection
│   │   └── report.go        # Report structs + comparison
│   ├── config/              # Configuration management
│   │   └── config.go        # Config loading (viper)
│   ├── dockerfile/          # Dockerfile parsing + rewriting
│   │   ├── parser.go        # Instruction parser
│   │   └── rewriter.go      # Multi-stage rewrite engine
│   ├── ecosystem/           # Language ecosystem detection
│   │   ├── detect.go        # File-based ecosystem detection
│   │   └── patterns.go      # Bloat patterns per ecosystem
│   ├── ignore/              # .dockerignore generation
│   │   └── generator.go     # Pattern-based generator
│   └── output/              # Output formatting
│       ├── table.go         # Terminal table output
│       └── json.go          # JSON output
├── Dockerfile               # Multi-stage build for slimify itself
├── .goreleaser.yml          # Release configuration
├── install.sh               # Curl-pipe installer
└── slimify.yaml             # Example config
```

## Making Changes

### Adding a New Command

1. Create a new file in `cmd/` (e.g., `cmd/newcmd.go`)
2. Define a cobra command and register it with `rootCmd` in `init()`
3. Add tests in `cmd/newcmd_test.go`

### Adding a New Ecosystem

1. Add markers to `pkg/ecosystem/detect.go` → `AllMarkers`
2. Add bloat patterns to `pkg/ecosystem/patterns.go` → `EcosystemBloatPatterns`
3. Add ignore patterns to `pkg/ignore/generator.go` → `getEcosystemIgnorePatterns()`
4. Add build/prod stage templates to `pkg/dockerfile/rewriter.go`

### Adding a New Recommendation

1. Add detection logic in `pkg/analyzer/image.go` → `generateRecommendations()`
2. Make sure recommendations have a meaningful `SavingsMB` estimate

## Pull Request Guidelines

1. **One feature/fix per PR** — keep PRs focused and reviewable
2. **Tests required** — add or update tests for any behavior changes
3. **Docs** — update README.md if you're adding user-facing features
4. **Commits** — use conventional commits (`feat:`, `fix:`, `docs:`, `chore:`)
5. **CI must pass** — all checks must be green before merge

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use descriptive variable names
- Add comments for exported functions and types
- Keep functions focused and under ~50 lines where possible

## Releasing

Releases are automated via GoReleaser when a new tag is pushed:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This triggers the release workflow which:
- Builds binaries for Linux, macOS, and Windows
- Creates a GitHub Release with changelogs
- Publishes to Homebrew tap
- Pushes Docker image to GHCR

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](./LICENSE).
