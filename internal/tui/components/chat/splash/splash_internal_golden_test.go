package splash

import (
	"testing"

	"github.com/charmbracelet/crush/internal/tui/components/dialogs/models"
	"github.com/charmbracelet/x/exp/golden"
)

func TestSplashInternal_APIKey_Error(t *testing.T) {
	s := New().(*splashCmp)
	s.needsAPIKey = true
	s.apiKeyInput.SetProviderName("Local")
	s.SetSize(40, 12)
	s.Update(models.APIKeyStateChangeMsg{State: models.APIKeyInputStateError})
	golden.RequireEqual(t, []byte(s.View()))
}

func TestSplashInternal_APIKey_Verifying(t *testing.T) {
	s := New().(*splashCmp)
	s.needsAPIKey = true
	s.apiKeyInput.SetProviderName("Local")
	s.SetSize(40, 12)
	s.Update(models.APIKeyStateChangeMsg{State: models.APIKeyInputStateVerifying})
	golden.RequireEqual(t, []byte(s.View()))
}

func TestSplashInternal_APIKey_Verified(t *testing.T) {
	s := New().(*splashCmp)
	s.needsAPIKey = true
	s.apiKeyInput.SetProviderName("Local")
	s.SetSize(40, 12)
	s.Update(models.APIKeyStateChangeMsg{State: models.APIKeyInputStateVerified})
	golden.RequireEqual(t, []byte(s.View()))
}
