package header_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/csync"
	"github.com/charmbracelet/crush/internal/lsp"
	"github.com/charmbracelet/crush/internal/session"
	hdr "github.com/charmbracelet/crush/internal/tui/components/chat/header"
	"github.com/charmbracelet/x/exp/golden"
)

func setupTestConfig(t *testing.T) *config.Config {
	t.Helper()
	os.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")
	tdir := t.TempDir()
	work := tdir
	data := filepath.Join(tdir, ".crush")
	cfg, err := config.Init(work, data, false)
	if err != nil {
		t.Fatalf("config.Init error: %v", err)
	}
	cfg.Providers = csync.NewMap[string, config.ProviderConfig]()
	cfg.Providers.Set("local", config.ProviderConfig{
		ID:   "local",
		Name: "Local",
		Type: catwalk.TypeOpenAI,
		Models: []catwalk.Model{{
			ID: "test", Name: "Test", ContextWindow: 1000, DefaultMaxTokens: 200,
		}},
	})
	cfg.Models[config.SelectedModelTypeLarge] = config.SelectedModel{Provider: "local", Model: "test", MaxTokens: 200}
	cfg.SetupAgents()
	return cfg
}

func TestHeaderCompact(t *testing.T) {
	_ = setupTestConfig(t)

	t.Run("Closed", func(t *testing.T) {
		h := hdr.New(nil)
		h.SetWidth(40)
		h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
		golden.RequireEqual(t, []byte(h.View()))
	})

	t.Run("Open", func(t *testing.T) {
		h := hdr.New(nil)
		h.SetWidth(40)
		h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
		h.SetDetailsOpen(true)
		golden.RequireEqual(t, []byte(h.View()))
	})
}

func TestHeaderCompact_WithProvider(t *testing.T) {
	_ = setupTestConfig(t)

	h := hdr.New(nil)
	h.SetWidth(40)
	h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
	// Simulate provider status populated
	h.SetProviderStatus(app.ProviderStatus{
		ProviderName:  "Local",
		ProviderID:    "local",
		ModelID:       "test",
		Ready:         true,
		StreamEnabled: true,
	})

	golden.RequireEqual(t, []byte(h.View()))
}

func TestHeaderCompact_WithProviderStreamOff(t *testing.T) {
	_ = setupTestConfig(t)

	h := hdr.New(nil)
	h.SetWidth(40)
	h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
	// Ready but stream disabled should show a "stream off" badge
	h.SetProviderStatus(app.ProviderStatus{
		ProviderName:  "Local",
		ProviderID:    "local",
		ModelID:       "test",
		Ready:         true,
		StreamEnabled: false,
	})

	golden.RequireEqual(t, []byte(h.View()))
}

func TestHeaderCompact_WithProviderTruncation(t *testing.T) {
	_ = setupTestConfig(t)

	h := hdr.New(nil)
	h.SetWidth(40)
	h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
	// Long model id to force truncation in provider summary
	h.SetProviderStatus(app.ProviderStatus{
		ProviderName:  "Local",
		ProviderID:    "local",
		ModelID:       "this-is-a-very-long-model-identifier-that-should-be-truncated",
		Ready:         true,
		StreamEnabled: true,
	})

	golden.RequireEqual(t, []byte(h.View()))
}

func TestHeaderCompact_NarrowWidth(t *testing.T) {
	_ = setupTestConfig(t)

	h := hdr.New(nil)
	h.SetWidth(20)
	h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
	golden.RequireEqual(t, []byte(h.View()))
}

func TestHeaderCompact_WithProviderWarning(t *testing.T) {
	_ = setupTestConfig(t)

	h := hdr.New(nil)
	h.SetWidth(40)
	h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
	// Not ready with detail and stream off to exercise warning summary
	h.SetProviderStatus(app.ProviderStatus{
		ProviderName:  "Local",
		ProviderID:    "local",
		ModelID:       "test",
		Ready:         false,
		Detail:        "auth required",
		StreamEnabled: false,
	})

	golden.RequireEqual(t, []byte(h.View()))
}

func TestHeaderCompact_NoProvider(t *testing.T) {
	_ = setupTestConfig(t)

	h := hdr.New(nil)
	h.SetWidth(40)
	h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
	h.SetProviderStatus(app.ProviderStatus{})
	golden.RequireEqual(t, []byte(h.View()))
}

func TestHeader_WithLSPSummaryAndDetails(t *testing.T) {
	cfg := setupTestConfig(t)
	// Configure three LSPs: one active (shell), one missing, one disabled
	cfg.LSP = config.LSPs{
		"shell":   {Command: "bash", Args: []string{"--version"}},
		"missing": {Command: "this-lsp-does-not-exist"},
		"off":     {Command: "bash", Disabled: true},
	}

	clients := map[string]*lsp.Client{
		"shell": {}, // zero-value client OK for header checks
	}

	h := hdr.New(clients)
	h.SetWidth(80)
	h.SetSession(session.Session{ID: "s1", Title: "My Session", PromptTokens: 10, CompletionTokens: 20})
	h.SetDetailsOpen(true)

	golden.RequireEqual(t, []byte(h.View()))
}
