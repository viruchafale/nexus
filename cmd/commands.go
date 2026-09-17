package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"nexus/internal/docker"
	"nexus/internal/git"
	"nexus/internal/network"
	"nexus/internal/process"
	"nexus/internal/system"
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
		info, err := system.Collect()
		fmt.Fprint(cmd.OutOrStdout(), system.Format(info))
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "Warning: some metrics unavailable:", err)
		}
		return nil
	},
}

// processesListFlags holds --limit/--sort values (cobra-idiomatic).
var (
	processesLimit = 15
	processesSort  = "cpu"
)

var processesCmd = &cobra.Command{
	Use:   "processes",
	Short: "Show top processes by CPU or memory (read-only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := process.ParseSortMode(processesSort)
		if err != nil {
			return err
		}
		procs, err := process.Collect()
		if err != nil {
			return fmt.Errorf("processes: %w", err)
		}
		process.Sort(procs, mode)
		top, err := process.Top(procs, processesLimit)
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), process.FormatTable(top))
		return nil
	},
}

// dockerAll backs the --all flag (cobra-idiomatic).
var dockerAll bool

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "List Docker containers (read-only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		containers, err := docker.List(dockerAll)
		if err != nil {
			var unavail *docker.UnavailableError
			if errors.As(err, &unavail) {
				fmt.Fprint(cmd.OutOrStdout(), docker.UnavailableMessage())
				fmt.Fprintln(cmd.ErrOrStderr(), "Warning:", docker.Detail(err))
				return nil
			}
			return fmt.Errorf("docker: %w", err)
		}
		fmt.Fprint(cmd.OutOrStdout(), docker.FormatTable(containers))
		return nil
	},
}

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Show Git repository overview (read-only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.Inspect("")
		if err != nil {
			var notRepo *git.NotRepoError
			if errors.As(err, &notRepo) {
				fmt.Fprint(cmd.OutOrStdout(), git.NotRepoMessage())
				return nil
			}
			return fmt.Errorf("git: %w", err)
		}
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Fprint(cmd.OutOrStdout(), git.FormatShort(repo))
		} else {
			fmt.Fprint(cmd.OutOrStdout(), git.Format(repo))
		}
		return nil
	},
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Run read-only network diagnostics",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(cmd.OutOrStdout(), network.Format(network.Collect()))
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
	processesCmd.Flags().IntVar(&processesLimit, "limit", 15, "max processes to show (>= 1)")
	processesCmd.Flags().StringVar(&processesSort, "sort", "cpu", "sort by \"cpu\" or \"memory\"")
	dockerCmd.Flags().BoolVarP(&dockerAll, "all", "a", false, "include stopped containers")
	gitCmd.Flags().Bool("short", false, "one-line summary")
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
