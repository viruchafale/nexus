package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Phase 1: command skeletons only. Each prints a clear
// "not implemented yet" message and returns nil so the
// binary stays runnable and scripts don't break.

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open interactive dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: dashboard (interactive Bubble Tea UI comes in a later phase)")
		return nil
	},
}

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Show system information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: system")
		return nil
	},
}

var processesCmd = &cobra.Command{
	Use:   "processes",
	Short: "Show running processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: processes")
		return nil
	},
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Inspect Docker containers and images",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: docker")
		return nil
	},
}

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Inspect current Git repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: git")
		return nil
	},
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Run network diagnostics",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: network")
		return nil
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose common development problems",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: doctor")
		return nil
	},
}

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask AI for troubleshooting help (optional)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "not implemented yet: ask (optional AI troubleshooting comes in a later phase)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(
		dashboardCmd,
		systemCmd,
		processesCmd,
		dockerCmd,
		gitCmd,
		networkCmd,
		doctorCmd,
		askCmd,
	)
}
