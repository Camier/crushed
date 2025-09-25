package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stateFile struct {
	LSP map[string]map[string]any `json:"lsp"`
}

func TestLSPList(t *testing.T) {
	t.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")
	// isolate overrides under tmp XDG_CONFIG_HOME
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

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

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	lspCmd.SetOut(&buf)

	rootCmd.SetArgs([]string{"lsp", "list", "-c", tmp})
	execErr := rootCmd.Execute()
	out := buf.String()

	if execErr != nil {
		t.Fatalf("execute lsp list: %v\noutput:%s", execErr, out)
	}
	if !strings.Contains(out, "LSP servers:") {
		t.Fatalf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "shell: enabled") {
		t.Fatalf("expected 'shell: enabled', got: %s", out)
	}
	if !strings.Contains(out, "missing: enabled") {
		t.Fatalf("expected 'missing: enabled', got: %s", out)
	}
}

func TestLSPEnableDisable(t *testing.T) {
	t.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")
	// isolate overrides under tmp XDG_CONFIG_HOME
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	cfgPath := filepath.Join(tmp, ".crush.json")
	if err := os.WriteFile(cfgPath, []byte(`{
      "lsp": {
        "shell": {"command": "bash", "args": ["--version"]}
      },
      "providers": {"noop": {"name": "noop", "type": "openai", "base_url": "http://127.0.0.1:9", "models": [{"id":"x"}]}}
    }`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// disable
	rootCmd.SetArgs([]string{"lsp", "disable", "shell", "-c", tmp})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("disable: %v", err)
	}
	// verify state file contains disabled true
	statePath := filepath.Join(tmp, "crush", "crush.state.json")
	bts, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var st stateFile
	_ = json.Unmarshal(bts, &st)
	if v := st.LSP["shell"]["disabled"]; v != true {
		t.Fatalf("expected disabled true in state, got: %#v", v)
	}

	// enable
	rootCmd.SetArgs([]string{"lsp", "enable", "shell", "-c", tmp})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("enable: %v", err)
	}
	bts, err = os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state after enable: %v", err)
	}
	st = stateFile{}
	_ = json.Unmarshal(bts, &st)
	if v := st.LSP["shell"]["disabled"]; v != false {
		t.Fatalf("expected disabled false in state, got: %#v", v)
	}
}

func TestLSPTestCommand(t *testing.T) {
	t.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")
	// isolate overrides under tmp XDG_CONFIG_HOME
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
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

	// found path
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	lspCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"lsp", "test", "shell", "-c", tmp})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("lsp test shell: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "shell: found") {
		t.Fatalf("expected 'shell: found', got: %s", out)
	}

	// missing path
	buf.Reset()
	rootCmd.SetArgs([]string{"lsp", "test", "missing", "-c", tmp})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("lsp test missing: %v", err)
	}
	out2 := buf.String()
	if !strings.Contains(out2, "missing: missing") {
		t.Fatalf("expected 'missing: missing', got: %s", out2)
	}
}
