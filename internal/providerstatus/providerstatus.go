package providerstatus

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/config"
)

const defaultHealthTimeout = 3 * time.Second

// BuildHealthURL constructs the readiness URL for a provider using its base URL and
// optional startup health path configuration.
func BuildHealthURL(prov config.ProviderConfig) (string, error) {
	baseURL := strings.TrimSpace(prov.BaseURL)
	if baseURL == "" {
		return "", fmt.Errorf("provider base_url not configured")
	}
	baseURL = strings.TrimRight(baseURL, "/")

	healthPath := strings.TrimSpace(prov.StartupHealthPath)
	if healthPath == "" {
		healthPath = "/models"
	}
	if !strings.HasPrefix(healthPath, "/") {
		healthPath = "/" + healthPath
	}

	return baseURL + healthPath, nil
}

// CheckHealth verifies whether a provider responds successfully on its readiness endpoint.
// It returns a boolean indicating readiness, an optional detail string for failures, and
// an error when the request could not be constructed.
func CheckHealth(ctx context.Context, client *http.Client, prov config.ProviderConfig) (bool, string, error) {
	resolved := resolveProviderConfig(prov)
	healthURL, err := BuildHealthURL(resolved)
	if err != nil {
		return false, "", err
	}
	return CheckHealthURL(ctx, client, resolved, healthURL)
}

// CheckHealthURL verifies the readiness of a provider against the supplied URL and configuration.
func CheckHealthURL(ctx context.Context, client *http.Client, prov config.ProviderConfig, healthURL string) (bool, string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, defaultHealthTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to create health request: %w", err)
	}
	applyHealthHeaders(req, prov)

	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return true, "", nil
	}
	return false, fmt.Sprintf("status %d", resp.StatusCode), nil
}

func applyHealthHeaders(req *http.Request, prov config.ProviderConfig) {
	switch prov.Type {
	case catwalk.TypeOpenAI, catwalk.TypeAzure:
		if prov.APIKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", prov.APIKey))
		}
	case catwalk.TypeAnthropic:
		if prov.APIKey != "" {
			req.Header.Set("x-api-key", prov.APIKey)
		}
	default:
		if prov.APIKey != "" {
			req.Header.Set("Authorization", prov.APIKey)
		}
	}

	for key, value := range prov.ExtraHeaders {
		if strings.EqualFold(key, "authorization") && prov.APIKey != "" {
			continue
		}
		req.Header.Set(key, value)
	}
}

func resolveProviderConfig(prov config.ProviderConfig) config.ProviderConfig {
	cfg := config.Get()
	if cfg == nil {
		return prov
	}

	resolved := prov

	if resolved.BaseURL != "" {
		if v, err := cfg.Resolve(resolved.BaseURL); err == nil && v != "" {
			resolved.BaseURL = v
		}
	}
	if resolved.StartupHealthPath != "" {
		if v, err := cfg.Resolve(resolved.StartupHealthPath); err == nil && v != "" {
			resolved.StartupHealthPath = v
		}
	}
	if resolved.APIKey != "" {
		if v, err := cfg.Resolve(resolved.APIKey); err == nil {
			resolved.APIKey = v
		}
	}

	if len(prov.ExtraHeaders) > 0 {
		resolvedHeaders := make(map[string]string, len(prov.ExtraHeaders))
		for key, value := range prov.ExtraHeaders {
			if v, err := cfg.Resolve(value); err == nil {
				resolvedHeaders[key] = v
			} else {
				resolvedHeaders[key] = value
			}
		}
		resolved.ExtraHeaders = resolvedHeaders
	}

	return resolved
}
