package providerstatus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

func TestEnsureProviderReady_NoBaseURL(t *testing.T) {
	require.NoError(t, EnsureProviderReady(context.Background(), t.TempDir(), config.ProviderConfig{}))
}

func TestEnsureProviderReady_HealthyEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	prov := config.ProviderConfig{
		ID:      "test",
		Name:    "Healthy",
		BaseURL: ts.URL,
	}

	require.NoError(t, EnsureProviderReady(context.Background(), t.TempDir(), prov))
}

func TestEnsureProviderReady_Unreachable(t *testing.T) {
	prov := config.ProviderConfig{
		ID:      "offline",
		Name:    "Offline",
		BaseURL: "http://127.0.0.1:1",
	}

	err := EnsureProviderReady(context.Background(), t.TempDir(), prov)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unreachable")
}
