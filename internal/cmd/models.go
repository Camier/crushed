package cmd

import (
    "fmt"
    "os"

    "github.com/charmbracelet/crush/internal/config"
    "github.com/spf13/cobra"
)

func init() {
    modelsCmd.AddCommand(modelsListCmd)
    modelsUseCmd.Flags().StringP("type", "t", "large", "Model type to update: large or small")
    modelsUseCmd.Flags().Int64("max-tokens", 0, "Override max tokens for the selected model (optional)")
    modelsUseCmd.Flags().String("reasoning-effort", "", "Reasoning effort for OpenAI models (low, medium, high) (optional)")
    modelsCmd.AddCommand(modelsUseCmd)
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
        for id, p := range cfg.Providers.Seq2() {
            fmt.Fprintf(os.Stdout, "- %s (%s)\n", id, p.Type)
            for _, m := range p.Models {
                fmt.Fprintf(os.Stdout, "  • %s (%s)\n", m.ID, m.Name)
            }
        }

        fmt.Fprintln(os.Stdout)
        fmt.Fprintln(os.Stdout, "Current selection:")
        if m := cfg.Models[config.SelectedModelTypeLarge]; m.Model != "" {
            fmt.Fprintf(os.Stdout, "- large:   %s/%s\n", m.Provider, m.Model)
        }
        if m := cfg.Models[config.SelectedModelTypeSmall]; m.Model != "" {
            fmt.Fprintf(os.Stdout, "- small:   %s/%s\n", m.Provider, m.Model)
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

        var modelType config.SelectedModelType
        switch modelTypeStr {
        case string(config.SelectedModelTypeSmall):
            modelType = config.SelectedModelTypeSmall
        default:
            modelType = config.SelectedModelTypeLarge
        }

        cwd, err := ResolveCwd(cmd)
        if err != nil {
            return err
        }
        cfg, err := config.Init(cwd, "", false)
        if err != nil {
            return err
        }

        // Validate provider and model exist
        prov, ok := cfg.Providers.Get(providerID)
        if !ok {
            return fmt.Errorf("provider not found: %s", providerID)
        }
        if cfg.GetModel(providerID, modelID) == nil {
            // if not found, but provider is known: list available
            fmt.Fprintf(os.Stderr, "model '%s' not found in provider '%s'\n", modelID, providerID)
            fmt.Fprintln(os.Stderr, "available models:")
            for _, m := range prov.Models {
                fmt.Fprintf(os.Stderr, "- %s (%s)\n", m.ID, m.Name)
            }
            return fmt.Errorf("unknown model")
        }

        sel := config.SelectedModel{
            Provider:        providerID,
            Model:           modelID,
            MaxTokens:       maxTokens,
            ReasoningEffort: reasoning,
        }

        if err := cfg.UpdatePreferredModel(modelType, sel); err != nil {
            return err
        }

        fmt.Fprintf(os.Stdout, "Updated %s model to %s/%s\n", modelType, providerID, modelID)
        return nil
    },
}

