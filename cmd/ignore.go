package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NotHarshhaa/slimify/pkg/config"
	"github.com/NotHarshhaa/slimify/pkg/ecosystem"
	"github.com/NotHarshhaa/slimify/pkg/ignore"
	"github.com/NotHarshhaa/slimify/pkg/output"
)

var (
	ignoreWrite     string
	ignoreEcosystem string
	ignoreDir       string
)

// maxScanDepth is the maximum directory depth for ecosystem detection.
const maxScanDepth = 4

var ignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Generate a .dockerignore file for the current project",
	Long: `Standalone .dockerignore generator — run it in any project directory to
generate a .dockerignore without auditing an image first.

Auto-detects your ecosystem from lock files and project structure.
Multiple ecosystems in the same project are supported.

Examples:
  slimify ignore
  slimify ignore > .dockerignore
  slimify ignore --write .dockerignore
  slimify ignore --ecosystem go,node,php
  slimify ignore --dir ./services/api`,
	Args: cobra.NoArgs,
	RunE: runIgnore,
}

func init() {
	ignoreCmd.Flags().StringVar(&ignoreWrite, "write", "", "write directly to the given file path")
	ignoreCmd.Flags().StringVar(&ignoreEcosystem, "ecosystem", "", "force specific ecosystems (comma-separated: go,node,python,rust,java,ruby,php,elixir,dotnet)")
	ignoreCmd.Flags().StringVar(&ignoreDir, "dir", ".", "directory to scan for ecosystem detection (useful for monorepos)")

	rootCmd.AddCommand(ignoreCmd)
}

func runIgnore(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Detect ecosystems
	var eco *ecosystem.DetectResult
	if ignoreEcosystem != "" {
		eco = ecosystem.DetectFromEcosystemFlag(ignoreEcosystem)
	} else {
		// Scan the target directory for ecosystem markers
		scanDir := ignoreDir
		if scanDir == "" {
			scanDir = "."
		}

		// Resolve to absolute path for clarity
		if abs, err := filepath.Abs(scanDir); err == nil {
			scanDir = abs
		}

		if _, err := os.Stat(scanDir); err != nil {
			return fmt.Errorf("scan directory %q not found: %w", scanDir, err)
		}

		files, err := scanDirectory(scanDir)
		if err != nil {
			files = []string{}
		}
		eco = ecosystem.DetectFromFiles(files)
	}

	// Generate ignore file
	gen := ignore.NewGenerator(cfg, eco)
	content := gen.Generate()

	// Count patterns (non-comment, non-empty lines)
	patternCount := 0
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patternCount++
		}
	}

	if ignoreWrite != "" {
		// Write to file
		if err := gen.WriteToFile(ignoreWrite, true); err != nil {
			return fmt.Errorf("failed to write ignore file: %w", err)
		}

		if !quiet {
			output.PrintIgnoreSummary(eco.String(), patternCount, true, ignoreWrite)
		}
	} else {
		// Print to stdout
		fmt.Print(content)

		if !quiet && !jsonOutput {
			output.PrintIgnoreSummary(eco.String(), patternCount, false, "")
		}
	}

	return nil
}

// scanDirectory walks a directory up to maxScanDepth levels and returns all file paths.
func scanDirectory(dir string) ([]string, error) {
	var files []string

	baseDepth := strings.Count(filepath.ToSlash(dir), "/")

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		// Skip hidden directories (other than the root itself)
		if d.IsDir() && path != dir && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		// Enforce depth limit
		currentDepth := strings.Count(filepath.ToSlash(path), "/")
		if d.IsDir() && currentDepth-baseDepth >= maxScanDepth {
			return filepath.SkipDir
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
