package ecosystem

import (
	"path/filepath"
	"strings"
)

// BloatPattern defines a file or directory pattern that is typically bloat.
type BloatPattern struct {
	Pattern     string // glob or path prefix
	Description string // human-readable explanation
	Ecosystem   Type   // which ecosystem this belongs to (empty = universal)
}

// UniversalBloatPatterns are patterns that are bloat in any ecosystem.
var UniversalBloatPatterns = []BloatPattern{
	{Pattern: ".git/", Description: "Git history — not needed in images"},
	{Pattern: ".git", Description: "Git directory"},
	{Pattern: ".gitignore", Description: "Git ignore file"},
	{Pattern: ".github/", Description: "GitHub metadata"},
	{Pattern: ".vscode/", Description: "VS Code settings"},
	{Pattern: ".idea/", Description: "JetBrains IDE settings"},
	{Pattern: ".editorconfig", Description: "Editor config"},
	{Pattern: "*.md", Description: "Markdown docs"},
	{Pattern: "*.txt", Description: "Text files (README, CHANGELOG, etc.)"},
	{Pattern: "LICENSE*", Description: "License files"},
	{Pattern: "CHANGELOG*", Description: "Changelog files"},
	{Pattern: "docs/", Description: "Documentation directory"},
	{Pattern: "doc/", Description: "Documentation directory"},
	{Pattern: "test/", Description: "Test directory"},
	{Pattern: "tests/", Description: "Test directory"},
	{Pattern: "spec/", Description: "Spec/test directory"},
	{Pattern: "__tests__/", Description: "Jest test directory"},
	{Pattern: "*.test.*", Description: "Test files"},
	{Pattern: "*.spec.*", Description: "Spec files"},
	{Pattern: "coverage/", Description: "Coverage reports"},
	{Pattern: ".nyc_output/", Description: "NYC coverage data"},
	{Pattern: ".env", Description: "Environment file with secrets"},
	{Pattern: ".env.*", Description: "Environment files"},
	{Pattern: "docker-compose*.yml", Description: "Docker Compose files"},
	{Pattern: "docker-compose*.yaml", Description: "Docker Compose files"},
	{Pattern: "Makefile", Description: "Build automation"},
	{Pattern: "Dockerfile*", Description: "Dockerfiles (already used for build)"},
	{Pattern: ".dockerignore", Description: "Docker ignore file"},
	{Pattern: "*.log", Description: "Log files"},
	{Pattern: "tmp/", Description: "Temporary files"},
	{Pattern: ".tmp/", Description: "Temporary files"},
}

// EcosystemBloatPatterns maps each ecosystem to its specific bloat patterns.
var EcosystemBloatPatterns = map[Type][]BloatPattern{
	NodeJS: {
		{Pattern: "node_modules/", Description: "Node dependencies — use multi-stage build", Ecosystem: NodeJS},
		{Pattern: ".npm/", Description: "npm cache", Ecosystem: NodeJS},
		{Pattern: ".yarn/", Description: "Yarn cache", Ecosystem: NodeJS},
		{Pattern: ".pnpm-store/", Description: "pnpm store", Ecosystem: NodeJS},
		{Pattern: "*.map", Description: "Source maps — strip in production", Ecosystem: NodeJS},
		{Pattern: "*.d.ts", Description: "TypeScript declarations", Ecosystem: NodeJS},
		{Pattern: ".eslintrc*", Description: "ESLint config", Ecosystem: NodeJS},
		{Pattern: ".prettierrc*", Description: "Prettier config", Ecosystem: NodeJS},
		{Pattern: "tsconfig*.json", Description: "TypeScript config", Ecosystem: NodeJS},
		{Pattern: "jest.config*", Description: "Jest config", Ecosystem: NodeJS},
		{Pattern: ".babelrc", Description: "Babel config", Ecosystem: NodeJS},
		{Pattern: "webpack.config*", Description: "Webpack config (if pre-built)", Ecosystem: NodeJS},
		{Pattern: "storybook/", Description: "Storybook", Ecosystem: NodeJS},
		{Pattern: ".storybook/", Description: "Storybook config", Ecosystem: NodeJS},
	},
	Go: {
		{Pattern: "vendor/", Description: "Go vendor directory — use go mod download", Ecosystem: Go},
		{Pattern: "*_test.go", Description: "Go test files", Ecosystem: Go},
		{Pattern: "testdata/", Description: "Go test data", Ecosystem: Go},
		{Pattern: ".golangci.yml", Description: "Linter config", Ecosystem: Go},
		{Pattern: ".golangci.yaml", Description: "Linter config", Ecosystem: Go},
	},
	Python: {
		{Pattern: "__pycache__/", Description: "Python bytecode cache", Ecosystem: Python},
		{Pattern: "*.pyc", Description: "Compiled Python files", Ecosystem: Python},
		{Pattern: "*.pyo", Description: "Optimized Python files", Ecosystem: Python},
		{Pattern: ".pytest_cache/", Description: "Pytest cache", Ecosystem: Python},
		{Pattern: ".mypy_cache/", Description: "Mypy cache", Ecosystem: Python},
		{Pattern: ".tox/", Description: "Tox environments", Ecosystem: Python},
		{Pattern: "*.egg-info/", Description: "Python egg info", Ecosystem: Python},
		{Pattern: "dist/", Description: "Python distribution", Ecosystem: Python},
		{Pattern: "build/", Description: "Python build directory", Ecosystem: Python},
		{Pattern: ".venv/", Description: "Virtual environment", Ecosystem: Python},
		{Pattern: "venv/", Description: "Virtual environment", Ecosystem: Python},
		{Pattern: "env/", Description: "Virtual environment", Ecosystem: Python},
	},
	Rust: {
		{Pattern: "target/", Description: "Rust build artifacts", Ecosystem: Rust},
		{Pattern: "target/debug/", Description: "Debug build artifacts", Ecosystem: Rust},
		{Pattern: "target/release/deps/", Description: "Release dependencies", Ecosystem: Rust},
		{Pattern: "*.rs.bk", Description: "Rust backup files", Ecosystem: Rust},
	},
	Java: {
		{Pattern: "target/", Description: "Maven build output", Ecosystem: Java},
		{Pattern: "build/", Description: "Gradle build output", Ecosystem: Java},
		{Pattern: ".gradle/", Description: "Gradle cache", Ecosystem: Java},
		{Pattern: "*.class", Description: "Compiled class files", Ecosystem: Java},
		{Pattern: "*.jar", Description: "JAR files (use multi-stage)", Ecosystem: Java},
		{Pattern: "*.war", Description: "WAR files", Ecosystem: Java},
		{Pattern: ".mvn/", Description: "Maven wrapper", Ecosystem: Java},
	},
	Ruby: {
		{Pattern: ".bundle/", Description: "Bundler metadata", Ecosystem: Ruby},
		{Pattern: "vendor/bundle/", Description: "Vendored gems", Ecosystem: Ruby},
		{Pattern: ".rubocop*", Description: "Rubocop config", Ecosystem: Ruby},
		{Pattern: "*.gem", Description: "Gem files", Ecosystem: Ruby},
		{Pattern: "log/", Description: "Rails log directory", Ecosystem: Ruby},
	},
}

// GetBloatPatterns returns all bloat patterns for the given ecosystems,
// combined with universal patterns.
func GetBloatPatterns(ecosystems *DetectResult) []BloatPattern {
	patterns := make([]BloatPattern, 0, len(UniversalBloatPatterns)+50)
	patterns = append(patterns, UniversalBloatPatterns...)

	for eco := range ecosystems.Ecosystems {
		if ecoPatterns, ok := EcosystemBloatPatterns[eco]; ok {
			patterns = append(patterns, ecoPatterns...)
		}
	}

	return patterns
}

// MatchesBloatPattern checks if a file path matches any of the given bloat patterns.
func MatchesBloatPattern(path string, patterns []BloatPattern) (bool, *BloatPattern) {
	base := filepath.Base(path)
	dir := filepath.Dir(path)

	for i := range patterns {
		p := &patterns[i]

		// Check directory prefix matches (patterns ending with /)
		if strings.HasSuffix(p.Pattern, "/") {
			dirPattern := strings.TrimSuffix(p.Pattern, "/")
			// Check if path starts with or contains this directory
			if strings.HasPrefix(path, dirPattern+"/") || strings.Contains(path, "/"+dirPattern+"/") {
				return true, p
			}
			// Also check the base directory name
			if base == dirPattern || filepath.Base(dir) == dirPattern {
				return true, p
			}
			continue
		}

		// Check glob patterns
		if strings.Contains(p.Pattern, "*") {
			if matched, _ := filepath.Match(p.Pattern, base); matched {
				return true, p
			}
			continue
		}

		// Exact name match
		if base == p.Pattern {
			return true, p
		}
	}

	return false, nil
}
