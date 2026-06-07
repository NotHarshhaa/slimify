// Package cmd implements the cobra CLI commands for slimify.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags.
	Version = "dev"
	// Commit is set at build time via ldflags.
	Commit = "none"

	// Global flags
	cfgFile    string
	jsonOutput bool
	quiet      bool
)

// rootCmd is the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "slimify",
	Short: "The Docker image auditor that actually tells you what to do",
	Long: `
  ███████╗██╗     ██╗███╗   ███╗██╗███████╗██╗   ██╗
  ██╔════╝██║     ██║████╗ ████║██║██╔════╝╚██╗ ██╔╝
  ███████╗██║     ██║██╔████╔██║██║█████╗   ╚████╔╝ 
  ╚════██║██║     ██║██║╚██╔╝██║██║██╔══╝    ╚██╔╝  
  ███████║███████╗██║██║ ╚═╝ ██║██║██║        ██║   
  ╚══════╝╚══════╝╚═╝╚═╝     ╚═╝╚═╝╚═╝        ╚═╝   

  Scan. Understand. Shrink.

  slimify inspects any Docker image, explains exactly where the bloat is,
  and hands you a rewritten Dockerfile + tuned .dockerignore — ready to run.`,
	Version: Version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./slimify.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "only print the summary line")

	rootCmd.SetVersionTemplate(fmt.Sprintf("slimify %s (commit: %s)\n", Version, Commit))
}
