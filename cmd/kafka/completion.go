package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/aaronjheng/kafka-cli/internal/kafka/admin"
)

func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "completion [bash|zsh|fish]",
		Short:       "Generate completion script",
		Annotations: map[string]string{"skipConfigLoad": "true"},
		Long: fmt.Sprintf(`To load completions:

Bash:

  $ source <(%[1]s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
  # macOS:
  $ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ %[1]s completion fish | source

  # To load completions for each session, execute once:
  $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish
`, "kafka"),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			genCompletion(cmd, args[0])
		},
	}

	return cmd
}

func genCompletion(cmd *cobra.Command, shell string) {
	switch shell {
	case "bash":
		_ = cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		_ = cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		_ = cmd.Root().GenFishCompletion(os.Stdout, true)
	}
}

func topicCompletionFunc(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	adminClient, closer, err := admin.NewFromConfig(cfg, cluster)
	if err != nil {
		slog.Debug("provideAdmin error", slog.Any("error", err))

		return nil, cobra.ShellCompDirectiveError
	}

	defer func() {
		err := closer(cmd.Context())
		if err != nil {
			slog.Debug("closer error", slog.Any("error", err))
		}
	}()

	topics, err := adminClient.ListTopicNames()
	if err != nil {
		slog.Debug("admin.ListTopicNames error", slog.Any("error", err))

		return nil, cobra.ShellCompDirectiveError
	}

	return topics, cobra.ShellCompDirectiveNoFileComp
}

func consumerGroupCompletionFunc(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	adminClient, closer, err := admin.NewFromConfig(cfg, cluster)
	if err != nil {
		slog.Debug("provideAdmin error", slog.Any("error", err))

		return nil, cobra.ShellCompDirectiveError
	}

	defer func() {
		err := closer(cmd.Context())
		if err != nil {
			slog.Debug("closer error", slog.Any("error", err))
		}
	}()

	groups, err := adminClient.ListConsumerGroupIDs()
	if err != nil {
		slog.Debug("admin.ListConsumerGroupIDs error", slog.Any("error", err))

		return nil, cobra.ShellCompDirectiveError
	}

	return groups, cobra.ShellCompDirectiveNoFileComp
}

func clusterCompletionFunc(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	if cfg == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	names := make([]string, 0, len(cfg.Clusters))
	for name := range cfg.Clusters {
		names = append(names, name)
	}

	slices.Sort(names)

	return names, cobra.ShellCompDirectiveNoFileComp
}
