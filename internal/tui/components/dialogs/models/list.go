package models

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/tui/exp/list"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/crush/internal/tui/util"
	"github.com/charmbracelet/x/ansi"
)

type listModel = list.FilterableGroupList[list.CompletionItem[ModelOption]]

type ModelListComponent struct {
	list             listModel
	modelType        int
	providers        []catwalk.Provider
	providerStatuses map[string]providerHealth
}

func NewModelListComponent(keyMap list.KeyMap, inputPlaceholder string, shouldResize bool) *ModelListComponent {
	t := styles.CurrentTheme()
	inputStyle := t.S().Base.PaddingLeft(1).PaddingBottom(1)
	options := []list.ListOption{
		list.WithKeyMap(keyMap),
		list.WithWrapNavigation(),
	}
	if shouldResize {
		options = append(options, list.WithResizeByList())
	}
	modelList := list.NewFilterableGroupedList(
		[]list.Group[list.CompletionItem[ModelOption]]{},
		list.WithFilterInputStyle(inputStyle),
		list.WithFilterPlaceholder(inputPlaceholder),
		list.WithFilterListOptions(
			options...,
		),
	)

	return &ModelListComponent{
		list:             modelList,
		modelType:        LargeModelType,
		providerStatuses: map[string]providerHealth{},
	}
}

func (m *ModelListComponent) Init() tea.Cmd {
	var cmds []tea.Cmd
	if len(m.providers) == 0 {
		cfg := config.Get()
		providers, err := config.Providers(cfg)
		filteredProviders := []catwalk.Provider{}
		for _, p := range providers {
			hasAPIKeyEnv := strings.HasPrefix(p.APIKey, "$")
			if hasAPIKeyEnv && p.ID != catwalk.InferenceProviderAzure {
				filteredProviders = append(filteredProviders, p)
			}
		}

		presets := localProviderPresets()
		for i := len(presets) - 1; i >= 0; i-- {
			preset := presets[i]
			if _, exists := cfg.Providers.Get(string(preset.ID)); exists {
				continue
			}
			if slices.ContainsFunc(filteredProviders, func(existing catwalk.Provider) bool {
				return existing.ID == preset.ID
			}) {
				continue
			}
			filteredProviders = append([]catwalk.Provider{preset}, filteredProviders...)
		}

		m.providers = filteredProviders
		if err != nil {
			cmds = append(cmds, util.ReportError(err))
		}
	}
	cmds = append(cmds, m.list.Init(), m.SetModelType(m.modelType))
	return tea.Batch(cmds...)
}

func (m *ModelListComponent) Update(msg tea.Msg) (*ModelListComponent, tea.Cmd) {
	u, cmd := m.list.Update(msg)
	m.list = u.(listModel)
	return m, cmd
}

func (m *ModelListComponent) View() string {
	return m.list.View()
}

func (m *ModelListComponent) Cursor() *tea.Cursor {
	return m.list.Cursor()
}

func (m *ModelListComponent) SetSize(width, height int) tea.Cmd {
	return m.list.SetSize(width, height)
}

func (m *ModelListComponent) SelectedModel() *ModelOption {
	s := m.list.SelectedItem()
	if s == nil {
		return nil
	}
	sv := *s
	model := sv.Value()
	return &model
}

func (m *ModelListComponent) SetProviderStatuses(statuses map[string]providerHealth) tea.Cmd {
	if statuses == nil {
		m.providerStatuses = map[string]providerHealth{}
	} else {
		m.providerStatuses = statuses
	}
	return m.SetModelType(m.modelType)
}

func (m *ModelListComponent) SetModelType(modelType int) tea.Cmd {
	t := styles.CurrentTheme()
	m.modelType = modelType

	var groups []list.Group[list.CompletionItem[ModelOption]]
	// first none section
	selectedItemID := ""

	cfg := config.Get()
	var currentModel config.SelectedModel
	if m.modelType == LargeModelType {
		currentModel = cfg.Models[config.SelectedModelTypeLarge]
	} else {
		currentModel = cfg.Models[config.SelectedModelTypeSmall]
	}

	configuredIcon := t.S().Base.Foreground(t.Success).Render(styles.CheckIcon)
	configured := fmt.Sprintf("%s %s", configuredIcon, t.S().Subtle.Render("Configured"))

	// Create a map to track which providers we've already added
	addedProviders := make(map[string]bool)

	// First, add any configured providers that are not in the known providers list
	// These should appear at the top of the list
	knownProviders, err := config.Providers(cfg)
	if err != nil {
		return util.ReportError(err)
	}
	for providerID, providerConfig := range cfg.Providers.Seq2() {
		if providerConfig.Disable {
			continue
		}

		// Check if this provider is not in the known providers list
		if !slices.ContainsFunc(knownProviders, func(p catwalk.Provider) bool { return p.ID == catwalk.InferenceProvider(providerID) }) ||
			!slices.ContainsFunc(m.providers, func(p catwalk.Provider) bool { return p.ID == catwalk.InferenceProvider(providerID) }) {
			// Convert config provider to provider.Provider format
			configProvider := catwalk.Provider{
				Name:   providerConfig.Name,
				ID:     catwalk.InferenceProvider(providerID),
				Models: make([]catwalk.Model, len(providerConfig.Models)),
			}

			// Convert models
			for i, model := range providerConfig.Models {
				configProvider.Models[i] = catwalk.Model{
					ID:                     model.ID,
					Name:                   model.Name,
					CostPer1MIn:            model.CostPer1MIn,
					CostPer1MOut:           model.CostPer1MOut,
					CostPer1MInCached:      model.CostPer1MInCached,
					CostPer1MOutCached:     model.CostPer1MOutCached,
					ContextWindow:          model.ContextWindow,
					DefaultMaxTokens:       model.DefaultMaxTokens,
					CanReason:              model.CanReason,
					HasReasoningEffort:     model.HasReasoningEffort,
					DefaultReasoningEffort: model.DefaultReasoningEffort,
					SupportsImages:         model.SupportsImages,
				}
			}

			// Add this unknown provider to the list
			name := configProvider.Name
			if name == "" {
				name = string(configProvider.ID)
			}
			section := list.NewItemSection(name)
			if info := m.providerInfo(t, providerConfig.ID, true, configured); info != "" {
				section.SetInfo(info)
			}
			group := list.Group[list.CompletionItem[ModelOption]]{
				Section: section,
			}
			for _, model := range configProvider.Models {
				item := list.NewCompletionItem(model.Name, ModelOption{
					Provider: configProvider,
					Model:    model,
				},
					list.WithCompletionID(
						fmt.Sprintf("%s:%s", providerConfig.ID, model.ID),
					),
				)

				group.Items = append(group.Items, item)
				if model.ID == currentModel.Model && string(configProvider.ID) == currentModel.Provider {
					selectedItemID = item.ID()
				}
			}
			groups = append(groups, group)

			addedProviders[providerID] = true
		}
	}

	// Then add the known providers from the predefined list
	for _, provider := range m.providers {
		// Skip if we already added this provider as an unknown provider
		if addedProviders[string(provider.ID)] {
			continue
		}

		// Check if this provider is configured and not disabled
		if providerConfig, exists := cfg.Providers.Get(string(provider.ID)); exists && providerConfig.Disable {
			continue
		}

		name := provider.Name
		if name == "" {
			name = string(provider.ID)
		}

		section := list.NewItemSection(name)
		if info := m.providerInfo(t, string(provider.ID), m.isProviderConfigured(cfg, string(provider.ID)), configured); info != "" {
			section.SetInfo(info)
		}
		group := list.Group[list.CompletionItem[ModelOption]]{
			Section: section,
		}
		for _, model := range provider.Models {
			item := list.NewCompletionItem(model.Name, ModelOption{
				Provider: provider,
				Model:    model,
			},
				list.WithCompletionID(
					fmt.Sprintf("%s:%s", provider.ID, model.ID),
				),
			)
			group.Items = append(group.Items, item)
			if model.ID == currentModel.Model && string(provider.ID) == currentModel.Provider {
				selectedItemID = item.ID()
			}
		}
		groups = append(groups, group)
	}

	var cmds []tea.Cmd

	cmd := m.list.SetGroups(groups)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	cmd = m.list.SetSelected(selectedItemID)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return tea.Sequence(cmds...)
}

// GetModelType returns the current model type
func (m *ModelListComponent) GetModelType() int {
	return m.modelType
}

func (m *ModelListComponent) SetInputPlaceholder(placeholder string) {
	m.list.SetInputPlaceholder(placeholder)
}

func (m *ModelListComponent) providerInfo(t *styles.Theme, providerID string, configured bool, configuredLabel string) string {
	parts := make([]string, 0, 2)
	if configured {
		parts = append(parts, configuredLabel)
	}
	if status, ok := m.providerStatuses[providerID]; ok && status.Checked {
		if status.Ready {
			parts = append(parts, t.S().Success.Render("Ready"))
		} else {
			detail := status.Detail
			if detail == "" {
				detail = "Offline"
			}
			detail = ansi.Truncate(detail, 32, "…")
			parts = append(parts, t.S().Error.Render(detail))
		}
	}
	return strings.Join(parts, " · ")
}

func (m *ModelListComponent) isProviderConfigured(cfg *config.Config, providerID string) bool {
	if cfg == nil {
		return false
	}
	if providerConfig, ok := cfg.Providers.Get(providerID); ok {
		return !providerConfig.Disable
	}
	return false
}

func localProviderPresets() []catwalk.Provider {
	return []catwalk.Provider{
		{
			Name:        "LM Studio",
			ID:          catwalk.InferenceProvider("lmstudio"),
			Type:        catwalk.TypeOpenAI,
			APIEndpoint: "http://127.0.0.1:1234/v1/",
			Models: []catwalk.Model{
				{
					ID:               "llama.cpp/models/meta-llama-3-8b-instruct-q5_k_m.gguf",
					Name:             "Llama 3 8B Instruct",
					ContextWindow:    8192,
					DefaultMaxTokens: 4096,
				},
				{
					ID:               "llama.cpp/models/mistral-nemo-instruct-2407-q5_k_m.gguf",
					Name:             "Mistral Nemo Instruct",
					ContextWindow:    8192,
					DefaultMaxTokens: 4096,
				},
			},
		},
		{
			Name:        "llama.cpp",
			ID:          catwalk.InferenceProvider("llamacpp"),
			Type:        catwalk.TypeOpenAI,
			APIEndpoint: "http://127.0.0.1:8080/v1/",
			Models: []catwalk.Model{
				{
					ID:               "llama.cpp/models/Qwen2.5-14B-Instruct-Q5_K_M.gguf",
					Name:             "Qwen2.5 14B Instruct",
					ContextWindow:    32768,
					DefaultMaxTokens: 4096,
				},
				{
					ID:               "llama.cpp/models/qwen2.5-7b-instruct-q4_k_m.gguf",
					Name:             "Qwen2.5 7B Instruct",
					ContextWindow:    32768,
					DefaultMaxTokens: 2048,
				},
			},
		},
		{
			Name:        "vLLM",
			ID:          catwalk.InferenceProvider("vllm"),
			Type:        catwalk.TypeOpenAI,
			APIEndpoint: "http://127.0.0.1:8000/v1/",
			Models: []catwalk.Model{
				{
					ID:               "nous-hermes-7b",
					Name:             "Nous Hermes 7B",
					ContextWindow:    32768,
					DefaultMaxTokens: 4096,
				},
				{
					ID:               "openorca-7b",
					Name:             "OpenOrca 7B",
					ContextWindow:    32768,
					DefaultMaxTokens: 4096,
				},
			},
		},
	}
}
