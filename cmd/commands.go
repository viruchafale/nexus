package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"nexus/config"
	"nexus/internal/ai"
	"nexus/internal/docker"
	"nexus/internal/doctor"
	"nexus/internal/git"
	"nexus/internal/network"
	"nexus/internal/process"
	"nexus/internal/system"
	"nexus/ui"
)

// Phase 1: command skeletons only. Each prints a clear
// "not implemented yet" message and returns nil so the
// binary stays runnable and scripts don't break.

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open interactive dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.Run(config.Load().NoColor)
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
	Short: "Run read-only diagnostics (never fixes anything)",
	RunE: func(cmd *cobra.Command, args []string) error {
		portsFlag, _ := cmd.Flags().GetString("ports")
		ports, err := doctor.ParsePorts(portsFlag)
		if err != nil {
			return err
		}
		opts := doctor.DefaultOptions()
		opts.Ports = ports
		fmt.Fprint(cmd.OutOrStdout(), doctor.Format(doctor.RunWith(opts)))
		return nil
	},
}

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask AI for troubleshooting help (optional, needs API key)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		if !cfg.HasAIKey() {
			fmt.Fprintln(cmd.OutOrStdout(), "AI functionality is unavailable.")
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "Set the required environment variable to enable it:")
			fmt.Fprintln(cmd.OutOrStdout(), "  export NEXUS_OPENAI_API_KEY=<your-key>")
			return nil
		}
		if cfg.AIProvider != "openai" {
			return fmt.Errorf("unsupported NEXUS_AI_PROVIDER %q: only \"openai\" is supported", cfg.AIProvider)
		}
		model := cfg.AIModel
		if model == "" {
			model = config.DefaultAIModel
		}
		question := strings.Join(args, " ")
		snapshot := ai.CollectContext()

		showCtx, _ := cmd.Flags().GetBool("show-context")
		if showCtx {
			fmt.Fprintln(cmd.OutOrStdout(), "Context sent to the model:")
			fmt.Fprintln(cmd.OutOrStdout(), "────────────────────────")
			fmt.Fprintln(cmd.OutOrStdout(), snapshot)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "(Tip: re-run with --show-context to see the environment snapshot sent to the model.)")
		}

		provider := &ai.OpenAIProvider{APIKey: cfg.OpenAIKey, BaseURL: cfg.AIBaseURL}
		ctx, cancel := context.WithTimeout(cmd.Context(), cfg.AITimeout)
		defer cancel()
		answer, err := provider.Complete(ctx, ai.Request{
			Model: model, System: ai.SystemPrompt, Question: question, Context: snapshot,
		})
		if err != nil {
			return fmt.Errorf("ask: %w", err)
		}
		// Advisory text only: NEXUS never executes anything the AI suggests.
		fmt.Fprintln(cmd.OutOrStdout(), answer)
		return nil
	},
}

func init() {
	processesCmd.Flags().IntVar(&processesLimit, "limit", 15, "max processes to show (>= 1)")
	processesCmd.Flags().StringVar(&processesSort, "sort", "cpu", "sort by \"cpu\" or \"memory\"")
	dockerCmd.Flags().BoolVarP(&dockerAll, "all", "a", false, "include stopped containers")
	gitCmd.Flags().Bool("short", false, "one-line summary")
	doctorCmd.Flags().String("ports", "", "comma-separated ports to probe (default: 3000,5173,8000,8080,5432,6379)")
	askCmd.Flags().Bool("show-context", false, "print the environment snapshot sent to the model")
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
