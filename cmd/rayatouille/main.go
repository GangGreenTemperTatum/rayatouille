package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/GangGreenTemperTatum/rayatouille/internal/app"
	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// version, commit, and date are set at build time via ldflags.
// GoReleaser injects: -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cfg := &config.Config{}

	rootCmd := &cobra.Command{
		Use:           "rayatouille",
		Short:         "Terminal UI for Ray cluster monitoring",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Resolve()

			if err := cfg.Validate(); err != nil {
				return err
			}

			client := ray.NewClient(cfg.Address, cfg.Timeout)

			versionInfo, err := client.Ping(context.Background())
			if err != nil {
				return fmt.Errorf("cannot reach Ray cluster at %s: %w", cfg.Address, err)
			}

			fmt.Fprintf(os.Stderr, "Connected to Ray cluster at %s (Ray %s)\n", cfg.Address, versionInfo.RayVersion)

			model := app.New(client, cfg, versionInfo)
			p := tea.NewProgram(model)
			_, err = p.Run()
			return err
		},
	}

	rootCmd.Version = version

	rootCmd.Flags().StringVar(&cfg.Address, "address", "", "Ray Dashboard URL (env: RAY_DASHBOARD_URL)")
	rootCmd.Flags().DurationVar(&cfg.Timeout, "timeout", 10*time.Second, "API request timeout")
	rootCmd.Flags().DurationVar(&cfg.RefreshInterval, "refresh-interval", 5*time.Second, "Data refresh interval")

	// Profile subcommand group
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage cluster connection profiles",
	}

	profileCmd.AddCommand(profileListCmd())
	profileCmd.AddCommand(profileAddCmd())
	profileCmd.AddCommand(profileRemoveCmd())
	profileCmd.AddCommand(profileUseCmd())

	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(completionCmd(rootCmd))

	return rootCmd
}

func profileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all saved profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadProfileConfig()
			if err != nil {
				return err
			}

			if len(cfg.Profiles) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No profiles saved. Use 'rayatouille profile add' to create one.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-40s %s\n", "NAME", "ADDRESS", "ACTIVE")
			for _, name := range sortedKeys(cfg.Profiles) {
				p := cfg.Profiles[name]
				marker := ""
				if name == cfg.ActiveProfile {
					marker = "*"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-40s %s\n", name, p.Address, marker)
			}
			return nil
		},
	}
}

func profileAddCmd() *cobra.Command {
	var address string
	var timeout string
	var refreshInterval string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new cluster profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.LoadProfileConfig()
			if err != nil {
				return err
			}

			if _, exists := cfg.Profiles[name]; exists {
				return fmt.Errorf("profile %q already exists (remove it first or choose a different name)", name)
			}

			cfg.Profiles[name] = config.Profile{
				Address:         address,
				Timeout:         timeout,
				RefreshInterval: refreshInterval,
			}

			if err := config.SaveProfileConfig(cfg); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Profile %q added.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "Ray Dashboard URL (required)")
	_ = cmd.MarkFlagRequired("address")
	cmd.Flags().StringVar(&timeout, "timeout", "", "API request timeout (e.g., 10s)")
	cmd.Flags().StringVar(&refreshInterval, "refresh-interval", "", "Data refresh interval (e.g., 5s)")

	return cmd
}

func profileRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a saved profile",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			names, err := config.ListProfileNames()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.LoadProfileConfig()
			if err != nil {
				return err
			}

			if _, exists := cfg.Profiles[name]; !exists {
				return fmt.Errorf("profile %q not found", name)
			}

			delete(cfg.Profiles, name)
			if cfg.ActiveProfile == name {
				cfg.ActiveProfile = ""
			}

			if err := config.SaveProfileConfig(cfg); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Profile %q removed.\n", name)
			return nil
		},
	}
}

func profileUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			names, err := config.ListProfileNames()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := config.SetActiveProfile(name); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Active profile set to %q.\n", name)
			return nil
		},
	}
}

func completionCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:       "completion [bash|zsh|fish]",
		Short:     "Generate shell completion scripts",
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.ExactArgs(1),
		Long: `Generate shell completion scripts for rayatouille.

To install completions:

  Bash:
    $ rayatouille completion bash > /etc/bash_completion.d/rayatouille

  Zsh:
    $ rayatouille completion zsh > "${fpath[1]}/_rayatouille"

  Fish:
    $ rayatouille completion fish > ~/.config/fish/completions/rayatouille.fish
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletionV2(out, true)
			case "zsh":
				return rootCmd.GenZshCompletion(out)
			case "fish":
				return rootCmd.GenFishCompletion(out, true)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}

func sortedKeys(m map[string]config.Profile) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// sort alphabetically
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
