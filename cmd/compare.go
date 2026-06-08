package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NotHarshhaa/slimify/pkg/analyzer"
	"github.com/NotHarshhaa/slimify/pkg/config"
	"github.com/NotHarshhaa/slimify/pkg/output"
)

var (
	compareRemote bool
)

var compareCmd = &cobra.Command{
	Use:   "compare <image-a> <image-b>",
	Short: "Diff two image versions side by side",
	Long: `Compare two Docker image versions to validate that a rebuild actually got
smaller. Shows size differences, layer counts, and shared base layers.

Examples:
  slimify compare myapp:v1.0 myapp:v2.0
  slimify compare myapp:latest myapp:slim --json
  slimify compare myapp:v1.0 myapp:v2.0 --remote`,
	Args: cobra.ExactArgs(2),
	RunE: runCompare,
}

func init() {
	compareCmd.Flags().BoolVar(&compareRemote, "remote", false, "compare images from a remote registry without pulling")
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, args []string) error {
	imageA := args[0]
	imageB := args[1]

	// Use config values for analysis settings
	cfg, err := config.Load(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	a := analyzer.NewImageAnalyzer(cfg.Audit.TopFilesPerLayer, cfg.Audit.ThresholdMB)

	report, err := a.CompareImages(imageA, imageB, compareRemote)
	if err != nil {
		return fmt.Errorf("compare failed: %w", err)
	}

	if jsonOutput {
		return output.PrintCompareJSON(report)
	}

	output.PrintCompareReport(report)
	return nil
}
