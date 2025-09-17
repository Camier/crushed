package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/csync"
	"github.com/stretchr/testify/require"
)

func newTestConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()
	cfg := &Config{
		Models:        make(map[SelectedModelType]SelectedModel),
		Providers:     csync.NewMap[string, ProviderConfig](),
		dataConfigDir: filepath.Join(dir, "config.json"),
	}
	return cfg
}

func TestUpdatePreferredModelValidatesInputs(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.Providers.Set("local", ProviderConfig{
		ID:     "local",
		Type:   catwalk.TypeOpenAI,
		Models: []catwalk.Model{{ID: "m1", Name: "Model One"}},
	})

	t.Run("unknown provider", func(t *testing.T) {
		err := cfg.UpdatePreferredModel(SelectedModelTypeLarge, SelectedModel{Provider: "missing", Model: "m1"})
		require.Error(t, err)
	})

	t.Run("unknown model", func(t *testing.T) {
		err := cfg.UpdatePreferredModel(SelectedModelTypeLarge, SelectedModel{Provider: "local", Model: "nope"})
		require.Error(t, err)
	})

	t.Run("disabled provider", func(t *testing.T) {
		cfg := newTestConfig(t)
		cfg.Providers.Set("local", ProviderConfig{
			ID:      "local",
			Disable: true,
			Models:  []catwalk.Model{{ID: "m1", Name: "Model One"}},
		})
		err := cfg.UpdatePreferredModel(SelectedModelTypeLarge, SelectedModel{Provider: "local", Model: "m1"})
		require.Error(t, err)
	})

	t.Run("negative tokens", func(t *testing.T) {
		err := cfg.UpdatePreferredModel(SelectedModelTypeLarge, SelectedModel{Provider: "local", Model: "m1", MaxTokens: -1})
		require.Error(t, err)
	})

	t.Run("invalid reasoning", func(t *testing.T) {
		err := cfg.UpdatePreferredModel(SelectedModelTypeLarge, SelectedModel{Provider: "local", Model: "m1", ReasoningEffort: "extreme"})
		require.Error(t, err)
	})

	t.Run("valid selection", func(t *testing.T) {
		err := cfg.UpdatePreferredModel(SelectedModelTypeLarge, SelectedModel{Provider: "local", Model: "m1", ReasoningEffort: "Medium"})
		require.NoError(t, err)
		stored := cfg.Models[SelectedModelTypeLarge]
		require.Equal(t, "medium", stored.ReasoningEffort)
		data, readErr := os.ReadFile(cfg.dataConfigDir)
		require.NoError(t, readErr)
		require.Contains(t, string(data), "m1")
	})

	t.Run("non-openai provider clears reasoning", func(t *testing.T) {
		cfg := newTestConfig(t)
		cfg.Providers.Set("anthropic", ProviderConfig{
			ID:     "anthropic",
			Type:   catwalk.TypeAnthropic,
			Models: []catwalk.Model{{ID: "claude", Name: "Claude"}},
		})
		err := cfg.UpdatePreferredModel(SelectedModelTypeLarge, SelectedModel{Provider: "anthropic", Model: "claude", ReasoningEffort: "low"})
		require.NoError(t, err)
		stored := cfg.Models[SelectedModelTypeLarge]
		require.Equal(t, "", stored.ReasoningEffort)
	})
}
