package cmd

import (
	"fmt"
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
)

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
  slimify ignore --ecosystem go,node`,
	Args: cobra.NoArgs,
	RunE: runIgnore,
}

func init() {
	ignoreCmd.Flags().StringVar(&ignoreWrite, "write", "", "write directly to the given file path")
	ignoreCmd.Flags().StringVar(&ignoreEcosystem, "ecosystem", "", "force specific ecosystems (comma-separated: go,node,python,rust,java,ruby)")

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
		// Scan current directory for ecosystem markers
		files, err := scanDirectory(".")
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

// scanDirectory walks a directory and returns all file paths.
func scanDirectory(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		files = append(files, path)

		// Also check one level deep for common project structures
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subEntries, err := os.ReadDir(path)
			if err != nil {
				continue
			}
			for _, sub := range subEntries {
				files = append(files, filepath.Join(path, sub.Name()))
			}
		}
	}

	return files, nil
}
