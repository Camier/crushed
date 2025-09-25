package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorMCPOutput(t *testing.T) {
	t.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")

	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, ".crush.json")
	if err := os.WriteFile(cfgPath, []byte(`{
      "mcp": {
        "context7": {"type": "http", "url": "https://api.example.com/mcp/", "headers": {"Authorization": "$(echo Bearer $CTX7)"}},
        "git": {"type": "stdio", "command": "bash", "args": ["--version"]},
        "missing": {"type": "stdio", "command": "definitely-not-a-binary"}
      },
      "providers": {"noop": {"name": "noop", "type": "openai", "base_url": "http://127.0.0.1:9", "models": [{"id":"x"}]}}
    }`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CTX7", "token")

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	doctorCmd.SetOut(&buf)
	t.Cleanup(func() {
		rootCmd.SetOut(os.Stdout)
		doctorCmd.SetOut(os.Stdout)
	})

	rootCmd.SetArgs([]string{"doctor", "mcp", "-c", tmp})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("doctor mcp: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Checking MCP entries") {
		t.Fatalf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "context7: http") {
		t.Fatalf("expected context7 http entry, got: %s", out)
	}
	if !strings.Contains(out, "auth: found") {
		t.Fatalf("expected auth: found, got: %s", out)
	}
	if !strings.Contains(out, "git: stdio") {
		t.Fatalf("expected stdio entry, got: %s", out)
	}
	if !strings.Contains(out, "missing") {
		t.Fatalf("expected missing stdio entry, got: %s", out)
	}
}
