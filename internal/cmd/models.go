package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/providerstatus"
	"github.com/spf13/cobra"
)

type providerRow struct {
	id   string
	prov config.ProviderConfig
}

func init() {
	modelsCmd.AddCommand(modelsListCmd)
	modelsUseCmd.Flags().StringP("type", "t", string(config.SelectedModelTypeLarge), "Model type to update: large or small")
	modelsUseCmd.Flags().Int64("max-tokens", 0, "Override max tokens for the selected model (optional)")
	modelsUseCmd.Flags().String("reasoning-effort", "", "Reasoning effort for OpenAI models (low, medium, high) (optional)")
	modelsCmd.AddCommand(modelsUseCmd)
	modelsCmd.AddCommand(modelsStatusCmd)
	rootCmd.AddCommand(modelsCmd)
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage preferred models",
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers and models",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}
		cfg, err := config.Init(cwd, "", false)
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, "Providers:")
		providers := make([]providerRow, 0)
		for id, p := range cfg.Providers.Seq2() {
			providers = append(providers, providerRow{id: id, prov: p})
		}
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].id < providers[j].id
		})
		for _, row := range providers {
			status := ""
			if row.prov.Disable {
				status = " (disabled)"
			}
			fmt.Fprintf(os.Stdout, "- %s (%s)%s\n", row.id, row.prov.Type, status)
			models := append([]catwalk.Model(nil), row.prov.Models...)
			sort.Slice(models, func(i, j int) bool {
				return models[i].ID < models[j].ID
			})
			for _, m := range models {
				fmt.Fprintf(os.Stdout, "  - %s (%s)\n", m.ID, m.Name)
			}
		}

		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Current selection:")
		if m := cfg.Models[config.SelectedModelTypeLarge]; m.Model != "" {
			fmt.Fprintf(os.Stdout, "- large: %s/%s%s\n", m.Provider, m.Model, formatReasoning(m.ReasoningEffort))
		}
		if m := cfg.Models[config.SelectedModelTypeSmall]; m.Model != "" {
			fmt.Fprintf(os.Stdout, "- small: %s/%s%s\n", m.Provider, m.Model, formatReasoning(m.ReasoningEffort))
		}
		return nil
	},
}

var modelsUseCmd = &cobra.Command{
	Use:   "use <provider> <model>",
	Short: "Select the preferred model for large/small",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerID := args[0]
		modelID := args[1]
		modelTypeStr, _ := cmd.Flags().GetString("type")
		maxTokens, _ := cmd.Flags().GetInt64("max-tokens")
		reasoning, _ := cmd.Flags().GetString("reasoning-effort")

		modelType, err := parseModelType(modelTypeStr)
		if err != nil {
			return err
		}
		if maxTokens < 0 {
			return fmt.Errorf("max-tokens must be non-negative")
		}

		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}
		cfg, err := config.Init(cwd, "", false)
		if err != nil {
			return err
		}

		prov, ok := cfg.Providers.Get(providerID)
		if !ok {
			return fmt.Errorf("provider not found: %s", providerID)
		}
		if prov.Disable {
			return fmt.Errorf("provider %s is disabled", providerID)
		}
		if cfg.GetModel(providerID, modelID) == nil {
			fmt.Fprintf(os.Stderr, "model '%s' not found in provider '%s'\n", modelID, providerID)
			fmt.Fprintln(os.Stderr, "available models:")
			models := append([]catwalk.Model(nil), prov.Models...)
			sort.Slice(models, func(i, j int) bool {
				return models[i].ID < models[j].ID
			})
			for _, m := range models {
				fmt.Fprintf(os.Stderr, "- %s (%s)\n", m.ID, m.Name)
			}
			return fmt.Errorf("unknown model")
		}

		sel := config.SelectedModel{
			Provider:        providerID,
			Model:           modelID,
			MaxTokens:       maxTokens,
			ReasoningEffort: strings.ToLower(strings.TrimSpace(reasoning)),
		}

		if err := cfg.UpdatePreferredModel(modelType, sel); err != nil {
			return err
		}

		if err := ensureProviderReady(cmd.Context(), cwd, prov); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Updated %s model to %s/%s\n", modelType, providerID, modelID)
		return nil
	},
}

var modelsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check provider readiness",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}
		cfg, err := config.Init(cwd, "", false)
		if err != nil {
			return err
		}

		providers := make([]providerRow, 0)
		for id, p := range cfg.Providers.Seq2() {
			providers = append(providers, providerRow{id: id, prov: p})
		}
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].id < providers[j].id
		})

		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		fmt.Fprintln(os.Stdout, "Provider status:")
		for _, row := range providers {
			name := row.id
			if name == "" {
				name = row.prov.Name
			}
			healthURL, err := providerstatus.BuildHealthURL(row.prov)
			if err != nil {
				fmt.Fprintf(os.Stdout, "- %s: skipped (%s)\n", name, err.Error())
				continue
			}
			ready, detail, checkErr := providerstatus.CheckHealthURL(ctx, nil, row.prov, healthURL)
			if checkErr != nil {
				fmt.Fprintf(os.Stdout, "- %s: unreachable (%s)\n", name, checkErr.Error())
				continue
			}
			if ready {
				fmt.Fprintf(os.Stdout, "- %s: ready\n", name)
			} else if detail != "" {
				fmt.Fprintf(os.Stdout, "- %s: unreachable (%s)\n", name, detail)
			} else {
				fmt.Fprintf(os.Stdout, "- %s: unreachable\n", name)
			}
		}
		return nil
	},
}

func parseModelType(value string) (config.SelectedModelType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(config.SelectedModelTypeLarge), "":
		return config.SelectedModelTypeLarge, nil
	case string(config.SelectedModelTypeSmall):
		return config.SelectedModelTypeSmall, nil
	default:
		return "", fmt.Errorf("invalid model type %q (expected large or small)", value)
	}
}

func formatReasoning(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ""
	}
	return fmt.Sprintf(" (reasoning: %s)", reason)
}

func ensureProviderReady(ctx context.Context, cwd string, prov config.ProviderConfig) error {
	if prov.StartupCommand == "" {
		return nil
	}
	if os.Getenv("CRUSH_SKIP_PROVIDER_STARTUP") == "1" {
		return nil
	}

	healthURL, err := providerstatus.BuildHealthURL(prov)
	if err != nil {
		return err
	}

	if ok, _, checkErr := providerstatus.CheckHealthURL(ctx, nil, prov, healthURL); checkErr == nil && ok {
		return nil
	}

	providerName := prov.ID
	if providerName == "" {
		providerName = prov.Name
	}
	fmt.Fprintf(os.Stdout, "Provider %s not reachable, executing startup command...\n", providerName)

	timeoutSeconds := prov.StartupTimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	deadlineCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := buildShellCommand(deadlineCtx, prov.StartupCommand)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start provider command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-deadlineCtx.Done():
			if cmd.ProcessState == nil && cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return fmt.Errorf("provider did not become ready within %d seconds", timeoutSeconds)
		case err := <-done:
			if err != nil {
				return fmt.Errorf("provider startup command exited with error: %w", err)
			}
			// Command exited successfully; continue polling in case it spawned a background process.
		case <-ticker.C:
			if ok, _, checkErr := providerstatus.CheckHealthURL(ctx, nil, prov, healthURL); checkErr == nil && ok {
				fmt.Fprintln(os.Stdout, "Provider is ready.")
				return nil
			}
		}
	}
}

func buildShellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", command)
	}
	return exec.CommandContext(ctx, "bash", "-lc", command)
}
