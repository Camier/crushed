package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorLSPOutput(t *testing.T) {
	t.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")
	t.Setenv("CRUSH_LSP_VERSION_CHECK", "1")

	tmp := t.TempDir()
	// project config with two LSP entries: one resolvable, one missing
	cfgPath := filepath.Join(tmp, ".crush.json")
	if err := os.WriteFile(cfgPath, []byte(`{
      "lsp": {
        "shell": {"command": "bash", "args": ["--version"]},
        "missing": {"command": "this-lsp-does-not-exist"}
      },
      "providers": {"noop": {"name": "noop", "type": "openai", "base_url": "http://127.0.0.1:9", "models": [{"id":"x"}]}}
    }`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = oldStdout
	}()

	rootCmd.SetArgs([]string{"doctor", "lsp", "-c", tmp})
	execErr := rootCmd.Execute()
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()

	if execErr != nil {
		t.Fatalf("execute doctor lsp: %v\noutput:%s", execErr, out)
	}
	if !strings.Contains(out, "Checking LSP servers") {
		t.Fatalf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "shell: found") {
		t.Fatalf("expected 'shell: found', got: %s", out)
	}
	if !strings.Contains(out, "missing: missing") {
		t.Fatalf("expected 'missing: missing', got: %s", out)
	}
}
