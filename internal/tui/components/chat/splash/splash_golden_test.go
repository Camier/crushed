package splash_test

import (
	"path/filepath"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	splash "github.com/charmbracelet/crush/internal/tui/components/chat/splash"
	"github.com/charmbracelet/x/exp/golden"
)

func TestSplashCompact_Onboarding(t *testing.T) {
	t.Parallel()

	s := splash.New()
	s.SetOnboarding(true)
	s.SetSize(40, 20)
	golden.RequireEqual(t, []byte(s.View()))
}

func TestSplashCompact_Onboarding_Width30(t *testing.T) {
	t.Parallel()
	s := splash.New()
	s.SetOnboarding(true)
	s.SetSize(30, 20)
	golden.RequireEqual(t, []byte(s.View()))
}

func TestSplashCompact_Onboarding_Width50(t *testing.T) {
	t.Parallel()
	s := splash.New()
	s.SetOnboarding(true)
	s.SetSize(50, 20)
	golden.RequireEqual(t, []byte(s.View()))
}

func TestSplashCompact_ProjectInit(t *testing.T) {
	t.Parallel()

	// Initialize config with deterministic working dir for stable golden
	cfgSetup(t, "/proj")

	s := splash.New()
	s.SetOnboarding(false)
	s.SetProjectInit(true)
	s.SetSize(40, 20)
	golden.RequireEqual(t, []byte(s.View()))
}

// cfgSetup initializes the global config with a deterministic working dir.
func cfgSetup(t *testing.T, workingDir string) {
	t.Helper()
	tdir := t.TempDir()
	data := filepath.Join(tdir, ".crush")
	_, err := config.Init(workingDir, data, false)
	if err != nil {
		t.Fatalf("config.Init error: %v", err)
	}
}

// API key verification state is covered directly in models package tests.
