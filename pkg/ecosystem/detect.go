// Package ecosystem detects the language/framework ecosystems present in a
// Docker image or project directory by looking for well-known lock files
// and project structure markers.
package ecosystem

import (
	"path/filepath"
	"strings"
)

// Type represents a detected ecosystem.
type Type string

const (
	NodeJS Type = "Node.js"
	Go     Type = "Go"
	Python Type = "Python"
	Rust   Type = "Rust"
	Java   Type = "Java"
	Ruby   Type = "Ruby"
)

// Marker maps a file or directory name to its ecosystem.
type Marker struct {
	Path      string
	Ecosystem Type
	Label     string // human-readable label like "npm", "pip", etc.
}

// AllMarkers is the full list of ecosystem markers we look for.
var AllMarkers = []Marker{
	// Node.js
	{Path: "package.json", Ecosystem: NodeJS, Label: "npm"},
	{Path: "package-lock.json", Ecosystem: NodeJS, Label: "npm"},
	{Path: "yarn.lock", Ecosystem: NodeJS, Label: "yarn"},
	{Path: "pnpm-lock.yaml", Ecosystem: NodeJS, Label: "pnpm"},
	{Path: "bun.lockb", Ecosystem: NodeJS, Label: "bun"},

	// Go
	{Path: "go.mod", Ecosystem: Go, Label: "go modules"},
	{Path: "go.sum", Ecosystem: Go, Label: "go modules"},

	// Python
	{Path: "requirements.txt", Ecosystem: Python, Label: "pip"},
	{Path: "Pipfile", Ecosystem: Python, Label: "pipenv"},
	{Path: "Pipfile.lock", Ecosystem: Python, Label: "pipenv"},
	{Path: "pyproject.toml", Ecosystem: Python, Label: "poetry/pip"},
	{Path: "setup.py", Ecosystem: Python, Label: "setuptools"},
	{Path: "setup.cfg", Ecosystem: Python, Label: "setuptools"},
	{Path: "uv.lock", Ecosystem: Python, Label: "uv"},

	// Rust
	{Path: "Cargo.toml", Ecosystem: Rust, Label: "cargo"},
	{Path: "Cargo.lock", Ecosystem: Rust, Label: "cargo"},

	// Java
	{Path: "pom.xml", Ecosystem: Java, Label: "maven"},
	{Path: "build.gradle", Ecosystem: Java, Label: "gradle"},
	{Path: "build.gradle.kts", Ecosystem: Java, Label: "gradle"},
	{Path: "gradlew", Ecosystem: Java, Label: "gradle"},

	// Ruby
	{Path: "Gemfile", Ecosystem: Ruby, Label: "bundler"},
	{Path: "Gemfile.lock", Ecosystem: Ruby, Label: "bundler"},
}

// DetectResult holds the detected ecosystems and their labels.
type DetectResult struct {
	Ecosystems map[Type][]string // ecosystem -> labels (e.g., NodeJS -> ["npm", "yarn"])
}

// String returns a human-readable summary like "Node.js (npm), Python (pip)".
func (d *DetectResult) String() string {
	if len(d.Ecosystems) == 0 {
		return "none detected"
	}
	var parts []string
	for eco, labels := range d.Ecosystems {
		unique := uniqueStrings(labels)
		parts = append(parts, string(eco)+" ("+strings.Join(unique, ", ")+")")
	}
	return strings.Join(parts, ", ")
}

// Types returns just the ecosystem types detected.
func (d *DetectResult) Types() []Type {
	var types []Type
	for t := range d.Ecosystems {
		types = append(types, t)
	}
	return types
}

// HasEcosystem checks if a specific ecosystem was detected.
func (d *DetectResult) HasEcosystem(t Type) bool {
	_, ok := d.Ecosystems[t]
	return ok
}

// DetectFromFiles takes a list of file paths (relative to image root) and
// returns which ecosystems are present.
func DetectFromFiles(files []string) *DetectResult {
	result := &DetectResult{
		Ecosystems: make(map[Type][]string),
	}

	// Build a set of basenames for fast lookup
	baseNames := make(map[string]bool)
	for _, f := range files {
		base := filepath.Base(f)
		baseNames[base] = true
	}

	for _, marker := range AllMarkers {
		if baseNames[marker.Path] {
			result.Ecosystems[marker.Ecosystem] = append(
				result.Ecosystems[marker.Ecosystem],
				marker.Label,
			)
		}
	}

	return result
}

// DetectFromEcosystemFlag parses a comma-separated ecosystem flag value
// like "go,node" into DetectResult.
func DetectFromEcosystemFlag(flag string) *DetectResult {
	result := &DetectResult{
		Ecosystems: make(map[Type][]string),
	}

	if flag == "" {
		return result
	}

	for _, part := range strings.Split(flag, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		switch part {
		case "node", "nodejs", "npm":
			result.Ecosystems[NodeJS] = []string{"npm"}
		case "go", "golang":
			result.Ecosystems[Go] = []string{"go modules"}
		case "python", "pip":
			result.Ecosystems[Python] = []string{"pip"}
		case "rust", "cargo":
			result.Ecosystems[Rust] = []string{"cargo"}
		case "java", "maven", "gradle":
			result.Ecosystems[Java] = []string{part}
		case "ruby", "bundler":
			result.Ecosystems[Ruby] = []string{"bundler"}
		}
	}

	return result
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
