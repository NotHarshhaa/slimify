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
	auditNoSecrets bool
	auditExitCode  bool
)

var auditCmd = &cobra.Command{
	Use:   "audit <image>",
	Short: "Inspect a Docker image and produce a full bloat report",
	Long: `Inspects a local or remote Docker image and produces a full bloat report.

Analyzes each layer to find the largest files, detects ecosystems (Node, Go,
Python, Rust, Java, Ruby, PHP, Elixir, .NET), identifies duplicate files across
layers, scans for secret files, and provides actionable recommendations to
reduce image size.

Examples:
  slimify audit myapp:latest
  slimify audit node:20-alpine --remote
  slimify audit myapp:latest --json --top 20
  slimify audit myapp:latest --threshold 5
  slimify audit myapp:latest --exit-code   # exits 1 when savings > threshold
  slimify audit myapp:latest --no-secrets  # skip secret file scanning`,
	Args: cobra.ExactArgs(1),
	RunE: runAudit,
}

func init() {
	auditCmd.Flags().BoolVar(&auditRemote, "remote", false, "audit a remote image from a registry without pulling")
	auditCmd.Flags().IntVar(&auditTop, "top", 10, "show top N largest files per layer")
	auditCmd.Flags().Float64Var(&auditThreshold, "threshold", 1.0, "only flag files larger than N MB")
	auditCmd.Flags().BoolVar(&auditNoSecrets, "no-secrets", false, "skip scanning for secret files in layers")
	auditCmd.Flags().BoolVar(&auditExitCode, "exit-code", false, "exit with code 1 if potential savings exceed threshold (useful for CI)")

	rootCmd.AddCommand(auditCmd)
}

func runAudit(cmd *cobra.Command, args []string) error {
	imageRef := args[0]

	a := analyzer.NewImageAnalyzer(auditTop, auditThreshold)
	a.ScanSecrets = !auditNoSecrets

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

	// --exit-code: exit 1 when there are significant savings (without corrupting JSON output)
	if auditExitCode && report.SavingsMB > auditThreshold {
		os.Exit(1)
	}

	// Legacy quiet gate (kept for backwards compat, but only when not in JSON mode)
	if quiet && !jsonOutput && report.SavingsMB > 100 {
		os.Exit(1)
	}

	return nil
}
