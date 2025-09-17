package prompt

import "github.com/charmbracelet/crush/internal/config"

func includeEnvironmentInfoForModel(modelType config.SelectedModelType) bool {
	cfg := config.Get()
	if cfg == nil {
		return true
	}
	provider := cfg.GetProviderForModel(modelType)
	if provider == nil {
		return true
	}
	return !provider.DisableStream
}
