package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"nexus/config"
	"nexus/internal/logger"
)

// verbose is bound to the --verbose persistent flag (cobra-idiomatic
// package state; everything else is passed explicitly, no globals).
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "nexus",
	Short: "NEXUS — terminal-based Developer Command Center",
	Long: `NEXUS is a terminal-based Developer Command Center.

Inspect your machine, diagnose problems, check Docker,
Git, networking, and optionally ask an LLM for help.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default `nexus` with no args shows help.
		// dashboard is the interactive entrypoint (later phase).
		_ = cmd.Help()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging to stderr")
}

// Execute runs the root command. Errors are printed to stderr
// by the caller so output stays clean for piping.
func Execute() error {
	cfg := config.Load()
	log := logger.New(verbose || cfg.Debug)
	if err := rootCmd.Execute(); err != nil {
		log.Debug("command failed", "error", err)
		fmt.Fprintln(os.Stderr, "Error:", err)
		return err
	}
	return nil
}
