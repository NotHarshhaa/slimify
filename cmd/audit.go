package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NotHarshhaa/slimify/pkg/analyzer"
	"github.com/NotHarshhaa/slimify/pkg/output"
)

var (
	auditRemote    bool
	auditTop       int
	auditThreshold float64
)

var auditCmd = &cobra.Command{
	Use:   "audit <image>",
	Short: "Inspect a Docker image and produce a full bloat report",
	Long: `Inspects a local or remote Docker image and produces a full bloat report.

Analyzes each layer to find the largest files, detects ecosystems (Node, Go,
Python, Rust, Java, Ruby), identifies duplicate files across layers, and
provides actionable recommendations to reduce image size.

Examples:
  slimify audit myapp:latest
  slimify audit node:20-alpine --remote
  slimify audit myapp:latest --json --top 20
  slimify audit myapp:latest --threshold 5`,
	Args: cobra.ExactArgs(1),
	RunE: runAudit,
}

func init() {
	auditCmd.Flags().BoolVar(&auditRemote, "remote", false, "audit a remote image from a registry without pulling")
	auditCmd.Flags().IntVar(&auditTop, "top", 10, "show top N largest files per layer")
	auditCmd.Flags().Float64Var(&auditThreshold, "threshold", 1.0, "only flag files larger than N MB")

	rootCmd.AddCommand(auditCmd)
}

func runAudit(cmd *cobra.Command, args []string) error {
	imageRef := args[0]

	a := analyzer.NewImageAnalyzer(auditTop, auditThreshold)

	report, err := a.AnalyzeImage(imageRef, auditRemote)
	if err != nil {
		return fmt.Errorf("audit failed: %w", err)
	}

	// Update computed fields
	report.TotalSizeMB = float64(report.TotalSize) / (1024 * 1024)

	if jsonOutput {
		return output.PrintAuditJSON(report)
	}

	output.PrintAuditReport(report, quiet)

	// Exit with code 1 if there are significant savings to be made
	if quiet && report.SavingsMB > 100 {
		os.Exit(1)
	}

	return nil
}
