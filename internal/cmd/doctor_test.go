package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDoctorProvidersOutput verifies the enriched diagnostics formatting.
func TestDoctorProvidersOutput(t *testing.T) {
	t.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")

	tmp := t.TempDir()
	// Project-level config to ensure our provider is loaded by config.Init
	cfgPath := filepath.Join(tmp, ".crush.json")
	// Use a localhost URL that should fail quickly; include health path and timeout.
	if err := os.WriteFile(cfgPath, []byte(`{
      "providers": {
        "localtest": {
          "name": "LocalTest",
          "type": "openai",
          "base_url": "http://127.0.0.1:1",
          "startup_health_path": "/health",
          "startup_timeout_seconds": 1,
          "models": [{"id": "test"}]
        }
      }
    }`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	doctorCmd.SetOut(&buf)
	t.Cleanup(func() {
		rootCmd.SetOut(os.Stdout)
		doctorCmd.SetOut(os.Stdout)
	})

	rootCmd.SetArgs([]string{"doctor", "providers", "-c", tmp})
	execErr := rootCmd.Execute()
	out := buf.String()

	if execErr != nil {
		t.Fatalf("execute doctor providers: %v\noutput:%s", execErr, out)
	}
	// Basic header
	if !strings.Contains(out, "Checking providers") {
		t.Fatalf("expected 'Checking providers' in output, got: %s", out)
	}
	// Provider line and enriched fields
	if !strings.Contains(out, "localtest") {
		t.Fatalf("expected provider id in output, got: %s", out)
	}
	if !strings.Contains(out, "unreachable") {
		t.Fatalf("expected 'unreachable' status, got: %s", out)
	}
	if !strings.Contains(out, "url: http://127.0.0.1:1") {
		t.Fatalf("expected base url hint in output, got: %s", out)
	}
	if !strings.Contains(out, "health: /health (timeout 1s)") {
		t.Fatalf("expected health path + timeout hint, got: %s", out)
	}
}
