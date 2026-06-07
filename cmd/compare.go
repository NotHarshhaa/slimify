package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NotHarshhaa/slimify/pkg/analyzer"
	"github.com/NotHarshhaa/slimify/pkg/output"
)

var compareCmd = &cobra.Command{
	Use:   "compare <image-a> <image-b>",
	Short: "Diff two image versions side by side",
	Long: `Compare two Docker image versions to validate that a rebuild actually got
smaller. Shows size differences, layer counts, and shared base layers.

Examples:
  slimify compare myapp:v1.0 myapp:v2.0
  slimify compare myapp:latest myapp:slim --json`,
	Args: cobra.ExactArgs(2),
	RunE: runCompare,
}

func init() {
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, args []string) error {
	imageA := args[0]
	imageB := args[1]

	a := analyzer.NewImageAnalyzer(10, 1.0)

	report, err := a.CompareImages(imageA, imageB, false)
	if err != nil {
		return fmt.Errorf("compare failed: %w", err)
	}

	if jsonOutput {
		return output.PrintCompareJSON(report)
	}

	output.PrintCompareReport(report)
	return nil
}
