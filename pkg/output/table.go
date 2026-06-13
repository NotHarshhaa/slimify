// Package output provides formatted terminal and JSON output for slimify reports.
package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/NotHarshhaa/slimify/pkg/analyzer"
)

// Colors and styles used for terminal output.
var (
	headerStyle  = color.New(color.FgHiCyan, color.Bold)
	successStyle = color.New(color.FgHiGreen)
	warningStyle = color.New(color.FgHiYellow)
	errorStyle   = color.New(color.FgHiRed, color.Bold)
	dimStyle     = color.New(color.FgHiBlack)
	infoStyle    = color.New(color.FgHiBlue)
)

// PrintAuditReport renders an audit report to the terminal.
func PrintAuditReport(report *analyzer.AuditReport, quiet bool) {
	if quiet {
		printQuietSummary(report)
		return
	}

	fmt.Println()
	headerStyle.Printf("  slimify audit — %s\n", report.ImageRef)
	fmt.Println("  " + strings.Repeat("─", 53))
	fmt.Println()

	// Summary
	fmt.Printf("  Image size:        %s\n", humanize.IBytes(uint64(report.TotalSize)))
	fmt.Printf("  Layers:            %d\n", report.LayerCount)
	if report.SavingsMB > 0 {
		warningStyle.Printf("  Potential savings: %.0f MB  (%.0f%%)\n", report.SavingsMB, report.SavingsPercent)
	} else {
		successStyle.Println("  Potential savings: minimal — image looks clean!")
	}

	if report.Ecosystems != nil && len(report.Ecosystems.Ecosystems) > 0 {
		fmt.Printf("  Ecosystem detected: %s\n", report.Ecosystems.String())
	}

	// Secrets warning — shown prominently above the layer table
	if len(report.SecretFiles) > 0 {
		fmt.Println()
		errorStyle.Println("  ⚠  SECRETS DETECTED IN IMAGE LAYERS:")
		for _, f := range report.SecretFiles {
			errorStyle.Printf("       %s\n", f)
		}
		errorStyle.Println("     These files should NEVER be baked into an image.")
		errorStyle.Println("     Add them to .dockerignore and rotate any exposed credentials.")
	}

	fmt.Println()

	// Layer breakdown table
	headerStyle.Println("  Layer breakdown:")
	printLayerTable(report.Layers)
	fmt.Println()

	// Top offenders per flagged layer
	for _, layer := range report.Layers {
		if len(layer.TopFiles) > 0 {
			warningStyle.Printf("  Top offenders in %s:\n", layer.Instruction)
			for _, f := range layer.TopFiles {
				sizeStr := humanize.IBytes(uint64(f.Size))
				fmt.Printf("    %-40s %s\n", f.Path, sizeStr)
			}
			fmt.Println()
		}
	}

	// Duplicate files
	if len(report.Duplicates) > 0 {
		warningStyle.Println("  Duplicate files across layers:")
		for _, d := range report.Duplicates {
			layers := make([]string, len(d.Layers))
			for i, l := range d.Layers {
				layers[i] = fmt.Sprintf("layer %d", l)
			}
			fmt.Printf("    %-40s copied in %s — consolidate\n",
				d.Path, strings.Join(layers, " and "))
		}
		fmt.Println()
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		headerStyle.Println("  Recommendations:")
		for i, rec := range report.Recommendations {
			savingsLabel := ""
			if rec.SavingsMB > 1 {
				savingsLabel = fmt.Sprintf(" → save ~%.0f MB", rec.SavingsMB)
			}
			fmt.Printf("    [%d] %s%s\n", i+1, rec.Title, savingsLabel)
			if rec.Detail != "" {
				dimStyle.Printf("        %s\n", rec.Detail)
			}
		}
		fmt.Println()
	}

	// Footer
	infoStyle.Printf("  Run `slimify fix %s --dockerfile ./Dockerfile` to apply all fixes.\n\n", report.ImageRef)
}

// printLayerTable renders the layer breakdown as an ASCII table.
func printLayerTable(layers []analyzer.LayerInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Instruction", "Size", "Files", "Delta"})
	table.SetBorder(true)
	table.SetCenterSeparator("┼")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("─")
	table.SetHeaderLine(true)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	for _, layer := range layers {
		if layer.IsEmpty {
			continue
		}
		sizeStr := humanize.IBytes(uint64(layer.Size))
		deltaStr := layer.DeltaLabel()
		if layer.Index > 0 {
			deltaStr = "+" + deltaStr
		}

		instruction := layer.Instruction
		if len(instruction) > 42 {
			instruction = instruction[:39] + "..."
		}

		fileCount := fmt.Sprintf("%d", layer.FileCount)

		row := []string{fmt.Sprintf("%d", layer.Index), instruction, sizeStr, fileCount, deltaStr}
		table.Append(row)
	}

	table.Render()
}

// PrintCompareReport renders a comparison report to the terminal.
func PrintCompareReport(report *analyzer.CompareReport) {
	fmt.Println()
	headerStyle.Printf("  slimify compare\n")
	fmt.Println("  " + strings.Repeat("─", 53))
	fmt.Println()

	fmt.Printf("  Image A:  %-30s %s\n", report.ImageA, humanize.IBytes(uint64(report.SizeA)))
	fmt.Printf("  Image B:  %-30s %s\n", report.ImageB, humanize.IBytes(uint64(report.SizeB)))
	fmt.Println()

	if report.Reduction > 0 {
		successStyle.Printf("  ▼ Reduction:  %s (%.1f%% smaller)\n",
			humanize.IBytes(uint64(report.Reduction)), report.ReductionPercent)
	} else if report.Reduction < 0 {
		warningStyle.Printf("  ▲ Increase:   %s (%.1f%% larger)\n",
			humanize.IBytes(uint64(-report.Reduction)), -report.ReductionPercent)
	} else {
		dimStyle.Println("  ↔ No change:  same compressed size")
	}

	fmt.Println()
	fmt.Printf("  Layers A → B:    %d → %d\n", report.LayersA, report.LayersB)
	fmt.Printf("  New layers in B:       %d\n", report.NewLayersInB)
	fmt.Printf("  Removed layers in B:   %d\n", report.RemovedLayersInB)
	fmt.Printf("  Shared base layers:    %d\n", report.SharedBaseLayers)
	fmt.Println()
}

// PrintFixSummary renders the output of a fix command.
func PrintFixSummary(outputDir string, ignoreSaved int64, originalSize int64, hasDockerfile bool, savingsMB float64) {
	fmt.Println()
	headerStyle.Println("  slimify fix")
	fmt.Println("  " + strings.Repeat("─", 53))
	fmt.Println()

	successStyle.Printf("  ✓ Generated .dockerignore")
	if ignoreSaved > 0 {
		dimStyle.Printf("         (removes %s from build context)", humanize.IBytes(uint64(ignoreSaved)))
	}
	fmt.Println()

	if hasDockerfile {
		successStyle.Println("  ✓ Rewritten Dockerfile            (multi-stage, optimized base)")

		if originalSize > 0 {
			var estimated float64
			if savingsMB > 0 {
				// Use the real analysis-derived savings
				estimated = float64(originalSize) - savingsMB*1024*1024
				if estimated < 0 {
					estimated = float64(originalSize) * 0.3
				}
			} else {
				// Fallback rough estimate
				estimated = float64(originalSize) * 0.3
			}
			savingsPct := (1 - estimated/float64(originalSize)) * 100
			successStyle.Printf("  ✓ Estimated new image size: %s",
				humanize.IBytes(uint64(estimated)))
			dimStyle.Printf("  (was %s — %.0f%% smaller)\n",
				humanize.IBytes(uint64(originalSize)),
				savingsPct)
		}
	}

	fmt.Println()
	fmt.Printf("  Output written to %s/\n", outputDir)
	if hasDockerfile {
		fmt.Println("    ├── Dockerfile.slimified")
		fmt.Println("    ├── .dockerignore")
		fmt.Println("    └── slimify.yaml")
	} else {
		fmt.Println("    ├── .dockerignore")
		fmt.Println("    └── slimify.yaml")
	}
	fmt.Println()
}

// PrintIgnoreSummary renders the output of the ignore command.
func PrintIgnoreSummary(ecosystems string, patternCount int, written bool, path string) {
	fmt.Println()
	headerStyle.Println("  slimify ignore")
	fmt.Println("  " + strings.Repeat("─", 53))
	fmt.Println()

	if ecosystems != "" {
		fmt.Printf("  Detected: %s\n", ecosystems)
	}
	fmt.Printf("  Patterns: %d rules generated\n", patternCount)

	if written {
		successStyle.Printf("  ✓ Written to %s\n", path)
	}
	fmt.Println()
}

// printQuietSummary prints only the one-line savings summary.
func printQuietSummary(report *analyzer.AuditReport) {
	secretsNote := ""
	if len(report.SecretFiles) > 0 {
		secretsNote = fmt.Sprintf(" | ⚠ %d secret file(s) detected", len(report.SecretFiles))
	}
	fmt.Printf("%s: %s → savings: %.0f MB (%.0f%%)%s\n",
		report.ImageRef,
		humanize.IBytes(uint64(report.TotalSize)),
		report.SavingsMB,
		report.SavingsPercent,
		secretsNote,
	)
}
