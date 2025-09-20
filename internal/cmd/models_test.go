package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/providerstatus"
	"github.com/stretchr/testify/require"
)

func TestBuildHealthURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		prov   config.ProviderConfig
		expect string
	}{
		{
			name:   "default path",
			prov:   config.ProviderConfig{BaseURL: "https://api.example.com/"},
			expect: "https://api.example.com/models",
		},
		{
			name:   "custom path",
			prov:   config.ProviderConfig{BaseURL: "https://api.example.com", StartupHealthPath: "healthz"},
			expect: "https://api.example.com/healthz",
		},
		{
			name:   "custom path with slash",
			prov:   config.ProviderConfig{BaseURL: "https://api.example.com/v1/", StartupHealthPath: "/ready"},
			expect: "https://api.example.com/v1/ready",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := providerstatus.BuildHealthURL(tc.prov)
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestBuildHealthURLMissingBase(t *testing.T) {
	t.Parallel()

	_, err := providerstatus.BuildHealthURL(config.ProviderConfig{})
	require.Error(t, err)
}

func TestProviderHealthCheck(t *testing.T) {
	t.Parallel()

	headerCh := make(chan http.Header, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerCh <- r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	prov := config.ProviderConfig{
		Type:         catwalk.TypeOpenAI,
		APIKey:       "token",
		ExtraHeaders: map[string]string{"X-Extra": "yay"},
	}

	ready, detail, err := providerstatus.CheckHealthURL(context.Background(), nil, prov, srv.URL)
	require.NoError(t, err)
	require.True(t, ready)
	require.Empty(t, detail)

	headers := <-headerCh
	require.Equal(t, "Bearer token", headers.Get("Authorization"))
	require.Equal(t, "yay", headers.Get("X-Extra"))
}

func TestProviderHealthCheckFailure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	ready, detail, err := providerstatus.CheckHealthURL(context.Background(), nil, config.ProviderConfig{}, srv.URL)
	require.NoError(t, err)
	require.False(t, ready)
	require.Equal(t, "status 502", detail)
}

func TestProviderHealthCheckUnreachable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.NewServeMux())
	srv.Close()

	ready, detail, err := providerstatus.CheckHealthURL(context.Background(), nil, config.ProviderConfig{}, srv.URL)
	require.NoError(t, err)
	require.False(t, ready)
	require.NotEmpty(t, detail)
}
