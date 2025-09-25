package models_test

import (
	"testing"

	"github.com/charmbracelet/crush/internal/tui/components/dialogs/models"
	"github.com/charmbracelet/x/exp/golden"
)

func TestAPIKeyInput_Verifying(t *testing.T) {
	in := models.NewAPIKeyInput()
	in.SetProviderName("Local")
	in.SetWidth(40)
	// trigger verifying
	in.Update(models.APIKeyStateChangeMsg{State: models.APIKeyInputStateVerifying})
	golden.RequireEqual(t, []byte(in.View()))
}

func TestAPIKeyInput_Verified(t *testing.T) {
	in := models.NewAPIKeyInput()
	in.SetProviderName("Local")
	in.SetWidth(40)
	// trigger verified
	in.Update(models.APIKeyStateChangeMsg{State: models.APIKeyInputStateVerified})
	golden.RequireEqual(t, []byte(in.View()))
}
