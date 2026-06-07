<div align="center">
  <br/>

  ```
  ███████╗██╗     ██╗███╗   ███╗██╗███████╗██╗   ██╗
  ██╔════╝██║     ██║████╗ ████║██║██╔════╝╚██╗ ██╔╝
  ███████╗██║     ██║██╔████╔██║██║█████╗   ╚████╔╝ 
  ╚════██║██║     ██║██║╚██╔╝██║██║██╔══╝    ╚██╔╝  
  ███████║███████╗██║██║ ╚═╝ ██║██║██║        ██║   
  ╚══════╝╚══════╝╚═╝╚═╝     ╚═╝╚═╝╚═╝        ╚═╝   
  ```

  <h3>The Docker image auditor that actually tells you what to do.</h3>

  <p>
    <a href="https://github.com/NotHarshhaa/slimify/releases"><img src="https://img.shields.io/github/v/release/NotHarshhaa/slimify?style=flat-square&color=0F6E56" alt="Latest Release"/></a>
    <a href="https://goreportcard.com/report/github.com/NotHarshhaa/slimify"><img src="https://goreportcard.com/badge/github.com/NotHarshhaa/slimify?style=flat-square" alt="Go Report"/></a>
    <a href="https://github.com/NotHarshhaa/slimify/blob/main/LICENSE"><img src="https://img.shields.io/github/license/NotHarshhaa/slimify?style=flat-square" alt="License"/></a>
    <a href="https://github.com/NotHarshhaa/slimify/stargazers"><img src="https://img.shields.io/github/stars/NotHarshhaa/slimify?style=flat-square" alt="Stars"/></a>
    <img src="https://img.shields.io/badge/platform-linux%20%7C%20mac%20%7C%20windows-informational?style=flat-square" alt="Platform"/>
  </p>

  <br/>

  > **Scan. Understand. Shrink.**
  > `slimify` inspects any Docker image, explains exactly where the bloat is,
  > and hands you a rewritten Dockerfile + tuned `.dockerignore` — ready to run.

  <br/>
</div>

---

## The problem with existing tools

| Tool | What it does | What's missing |
|---|---|---|
| `untracked` | Generates `.dockerignore` from npm deps | Node/npm only. No image analysis. No layer insight. |
| `docker-repack` | Reorders/squashes layers post-build | Doesn't remove actual bloat. No file-level visibility. |
| `dive` | Visual layer explorer (TUI) | Read-only. No ecosystem detection. No fix output. |
| `docker-slim` | Strips images by tracing syscalls at runtime | Complex setup. Requires running the container. Can break apps. |

**`slimify` does all of it in a single binary** — without running your container, without Node lock-in, and without guesswork. It reads the image, finds the bloat, and tells you exactly how to fix it.

---

## Features

- 🔍 **Layer-by-layer audit** — exact MB per instruction, top 10 largest files per layer
- 🌐 **Multi-ecosystem ignore generation** — Node, Go, Python, Rust, Java, Ruby auto-detected
- 📋 **Duplicate file detection** — catches files silently copied across layers (common in `RUN apt-get` chains)
- 🔧 **Dockerfile rewriter** — outputs a multi-stage optimized Dockerfile
- 📦 **`.dockerignore` generator** — tuned to your actual dependency graph, not generic patterns
- 💰 **Savings estimate** — tells you how much space you'll save *before* you rebuild
- 🚀 **Single Go binary** — no runtime, no daemon, no Docker socket required for `audit`
- 🖥️ **CI-friendly** — JSON output mode, exit codes, quiet flag

---

## Installation

### Homebrew (macOS/Linux)

```bash
brew install NotHarshhaa/tap/slimify
```

### Go install

```bash
go install github.com/NotHarshhaa/slimify@latest
```

### Binary releases

```bash
curl -sSfL https://raw.githubusercontent.com/NotHarshhaa/slimify/main/install.sh | sh
```

Pre-built binaries for Linux (`amd64`, `arm64`), macOS (`arm64`), and Windows are available on the [Releases](https://github.com/NotHarshhaa/slimify/releases) page.

### Docker

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/NotHarshhaa/slimify audit myapp:latest
```

---

## Quick Start

```bash
# Audit any image — no config needed
slimify audit myapp:latest

# Get a full fix: rewritten Dockerfile + .dockerignore
slimify fix myapp:latest --dockerfile ./Dockerfile

# Audit a remote image without pulling first
slimify audit node:20-alpine --remote
```

---

## Commands

### `slimify audit <image>`

Inspects a local or remote Docker image and produces a full bloat report.

```
$ slimify audit myapp:latest

  slimify audit — myapp:latest
  ─────────────────────────────────────────────────────

  Image size:       847 MB
  Potential savings: 312 MB  (36%)
  Ecosystem detected: Node.js (npm), Python (pip)

  Layer breakdown:
  ┌──────────────────────────────────────┬──────────┬──────────┐
  │ Instruction                          │ Size     │ Delta    │
  ├──────────────────────────────────────┼──────────┼──────────┤
  │ FROM node:20                         │  342 MB  │ baseline │
  │ RUN apt-get install -y build-essential│  61 MB  │ +61 MB   │
  │ COPY . .                             │  218 MB  │ +218 MB  │  ← bloat
  │ RUN npm install                      │  187 MB  │ +187 MB  │  ← bloat
  │ RUN npm run build                    │  39 MB   │ +39 MB   │
  └──────────────────────────────────────┴──────────┴──────────┘

  Top offenders in COPY . . :
    node_modules/       182 MB   ← not needed, add to .dockerignore
    .git/               28 MB    ← not needed, add to .dockerignore
    coverage/           4.2 MB   ← not needed
    *.map files         2.1 MB   ← strip source maps

  Duplicate files across layers:
    package-lock.json   copied in layer 3 and layer 5 — consolidate

  Recommendations:
    [1] Switch to multi-stage build                → save ~187 MB (node_modules)
    [2] Use node:20-alpine as base                 → save ~95 MB
    [3] Generate .dockerignore (run slimify fix)   → save ~34 MB from COPY context
    [4] Merge RUN apt-get + cleanup in one layer   → save ~18 MB

  Run `slimify fix myapp:latest --dockerfile ./Dockerfile` to apply all fixes.
```

**Flags:**

| Flag | Description |
|---|---|
| `--remote` | Audit a remote image from a registry without pulling |
| `--json` | Output report as JSON (for CI/CD pipelines) |
| `--quiet` | Only print the savings summary line |
| `--top N` | Show top N largest files per layer (default: 10) |
| `--threshold MB` | Only flag files larger than N MB (default: 1) |

---

### `slimify fix <image>`

Generates a `.dockerignore`, an optimized multi-stage `Dockerfile`, and a `slimify.yaml` config.

```bash
slimify fix myapp:latest --dockerfile ./Dockerfile --out ./slimify-out/
```

```
  slimify fix — myapp:latest
  ─────────────────────────────────────────────────────

  ✓ Generated .dockerignore         (removes 34 MB from build context)
  ✓ Rewritten Dockerfile            (multi-stage, alpine base)
  ✓ Estimated new image size: 198 MB  (was 847 MB — 76% smaller)

  Output written to ./slimify-out/
    ├── Dockerfile.slimified
    ├── .dockerignore
    └── slimify.yaml
```

**Flags:**

| Flag | Description |
|---|---|
| `--dockerfile PATH` | Path to your existing Dockerfile (required for rewrite) |
| `--out DIR` | Output directory for generated files (default: `.`) |
| `--write` | Write files in-place, overwriting existing ones |
| `--no-rewrite` | Only generate `.dockerignore`, skip Dockerfile rewrite |
| `--dry-run` | Print the generated output to stdout, don't write files |

---

### `slimify compare <image-a> <image-b>`

Diff two image versions side by side — useful for validating that a rebuild actually got smaller.

```bash
slimify compare myapp:v1.0 myapp:v2.0
```

```
  Image A (myapp:v1.0):   847 MB
  Image B (myapp:v2.0):   198 MB
  Reduction:              649 MB (76%)

  New layers in B:        3
  Removed layers in B:    7
  Shared base layers:     2
```

---

### `slimify ignore`

Standalone ignore file generator — run it in any project directory to generate a `.dockerignore` without auditing an image first.

```bash
# Auto-detect ecosystem and generate
slimify ignore

# Write directly to file
slimify ignore > .dockerignore

# Update existing file, preserving custom rules
slimify ignore --write .dockerignore

# Force a specific ecosystem
slimify ignore --ecosystem go,node
```

The `ignore` command detects your ecosystem from lock files and project structure:

| Detected file | Ecosystem |
|---|---|
| `package.json`, `package-lock.json`, `yarn.lock` | Node.js |
| `go.mod`, `go.sum` | Go |
| `requirements.txt`, `Pipfile`, `pyproject.toml` | Python |
| `Cargo.toml`, `Cargo.lock` | Rust |
| `pom.xml`, `build.gradle` | Java |
| `Gemfile`, `Gemfile.lock` | Ruby |

Multiple ecosystems in the same project are supported.

---

## Configuration

Customize `slimify` behavior via `slimify.yaml` in your project root (or `--config` flag):

```yaml
# slimify.yaml

ignore:
  whitelist:
    - bin/            # don't ignore this even if detected as bloat
    - config/prod/
  blacklist:
    - node_modules/.cache/puppeteer  # always ignore, even if not auto-detected

audit:
  threshold_mb: 2             # flag files larger than 2 MB
  top_files_per_layer: 20

fix:
  base_image: node:20-alpine  # override auto-selected base
  multi_stage: true           # always rewrite as multi-stage
  output_dir: ./docker/
```

You can also use any [cosmiconfig](https://github.com/davidtheclark/cosmiconfig)-compatible format: `.slimifyrc`, `.slimifyrc.json`, `.slimifyrc.yaml`, or `slimify.config.js`.

---

## CI / CD Integration

### GitHub Actions

```yaml
- name: Audit Docker image
  uses: NotHarshhaa/slimify-action@v1
  with:
    image: myapp:${{ github.sha }}
    fail-if-savings-above: 100  # fail PR if >100 MB of bloat detected

- name: Comment savings report on PR
  uses: NotHarshhaa/slimify-action@v1
  with:
    image: myapp:${{ github.sha }}
    comment-on-pr: true
```

### Shell (any CI)

```bash
# Exit 1 if image has >100 MB of potential savings
slimify audit myapp:latest --json | jq -e '.savings_mb < 100'
```

---

## How it works

`slimify` reads your Docker image as a sequence of OCI-compatible layers — no Docker daemon or `docker run` required for the audit step. For each layer:

1. Extracts the layer tarball and builds a file tree with sizes
2. Computes per-file deltas against the previous layer
3. Cross-references against known ecosystem bloat patterns (test dirs, source maps, lock files, `.git`, docs, type definitions, etc.)
4. Detects duplicate inodes across layers (files added and overwritten without `--squash`)

The `fix` command feeds the audit output into a Dockerfile parser, identifies stage boundaries, and rewrites using:
- Multi-stage builds that discard the build stage's `node_modules` / build cache
- `--no-cache` flags on package manager installs
- Merged `RUN` instructions to eliminate dead layer space
- `alpine` or `distroless` base image suggestions where viable

Everything runs locally. `slimify` never uploads your image or Dockerfile anywhere.

---

## Comparison

| | `slimify` | `untracked` | `docker-repack` | `dive` | `docker-slim` |
|---|:---:|:---:|:---:|:---:|:---:|
| Layer-level analysis | ✅ | ❌ | ✅ | ✅ | ✅ |
| File-level breakdown | ✅ | ❌ | ❌ | ✅ | ❌ |
| Multi-ecosystem support | ✅ | ❌ (Node only) | N/A | N/A | N/A |
| `.dockerignore` generation | ✅ | ✅ | ❌ | ❌ | ❌ |
| Dockerfile rewriter | ✅ | ❌ | ❌ | ❌ | ❌ |
| Savings estimate upfront | ✅ | ❌ | ❌ | ❌ | ❌ |
| No container runtime needed | ✅ | ✅ | ❌ | ❌ | ❌ |
| CI-friendly JSON output | ✅ | ❌ | ❌ | ❌ | ✅ |
| Single binary | ✅ | ❌ (npm) | ❌ (Rust) | ✅ | ✅ |

---

## Roadmap

- [ ] `slimify audit` — remote registry support (ECR, GCR, ACR) without local pull
- [ ] `slimify fix --base distroless` — distroless base image rewrite support
- [ ] VS Code extension — inline `.dockerignore` suggestions in the editor
- [ ] GitHub Actions Marketplace release
- [ ] GitLab CI/CD Catalog release
- [ ] SBOM output alongside audit report
- [ ] `slimify watch` — re-audit on every `docker build` in a dev loop

---

## Contributing

PRs are welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for local setup and guidelines.

```bash
git clone https://github.com/NotHarshhaa/slimify
cd slimify
go mod tidy
go run . audit --help
```

---

## Related projects

- [untracked](https://github.com/Kikobeats/untracked) — npm-specific `.dockerignore` generator
- [docker-repack](https://github.com/orf/docker-repack) — layer squashing and reordering
- [dive](https://github.com/wagoodman/dive) — interactive layer explorer TUI
- [docker-slim](https://github.com/slimtoolkit/slim) — runtime-based image minifier

---

## License

**slimify** © [NotHarshhaa](https://github.com/NotHarshhaa), released under the [MIT](./LICENSE) License.
