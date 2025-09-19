package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/config"
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

			got, err := buildHealthURL(tc.prov)
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}

func TestBuildHealthURLMissingBase(t *testing.T) {
	t.Parallel()

	_, err := buildHealthURL(config.ProviderConfig{})
	require.Error(t, err)
}

func TestApplyHealthHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		prov   config.ProviderConfig
		header string
		value  string
	}{
		{
			name:   "openai",
			prov:   config.ProviderConfig{Type: catwalk.TypeOpenAI, APIKey: "secret"},
			header: "Authorization",
			value:  "Bearer secret",
		},
		{
			name:   "anthropic",
			prov:   config.ProviderConfig{Type: catwalk.TypeAnthropic, APIKey: "anthro"},
			header: "x-api-key",
			value:  "anthro",
		},
		{
			name:   "azure",
			prov:   config.ProviderConfig{Type: catwalk.TypeAzure, APIKey: "azkey"},
			header: "api-key",
			value:  "azkey",
		},
	}

	for _, tc := range tests {
		testCase := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			require.NoError(t, err)

			testCase.prov.ExtraHeaders = map[string]string{"X-Test": "ok"}

			applyHealthHeaders(req, testCase.prov)

			require.Equal(t, testCase.value, req.Header.Get(testCase.header))
			require.Equal(t, "ok", req.Header.Get("X-Test"))
		})
	}
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

	ready, detail := providerHealthCheck(context.Background(), prov, srv.URL)
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

	ready, detail := providerHealthCheck(context.Background(), config.ProviderConfig{}, srv.URL)
	require.False(t, ready)
	require.Equal(t, "status 502", detail)
}

func TestProviderHealthCheckUnreachable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.NewServeMux())
	srv.Close()

	ready, detail := providerHealthCheck(context.Background(), config.ProviderConfig{}, srv.URL)
	require.False(t, ready)
	require.NotEmpty(t, detail)
}
