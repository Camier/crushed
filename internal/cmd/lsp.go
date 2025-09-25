package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	lspCmd.AddCommand(lspListCmd)
	lspCmd.AddCommand(lspEnableCmd)
	lspCmd.AddCommand(lspDisableCmd)
	lspCmd.AddCommand(lspTestCmd)
	rootCmd.AddCommand(lspCmd)
}

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Manage LSP (Language Server Protocol) integrations",
}

var lspListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured LSP servers and status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}
		debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
		dataDir, _ := cmd.Root().PersistentFlags().GetString("data-dir")
		cfg, err := config.Init(cwd, dataDir, debug)
		if err != nil {
			return err
		}

		names := make([]string, 0, len(cfg.LSP))
		for name := range cfg.LSP {
			names = append(names, name)
		}
		sort.Strings(names)

		var b bytes.Buffer
		b.WriteString("LSP servers:\n")
		for _, name := range names {
			l := cfg.LSP[name]
			status := "enabled"
			if l.Disabled {
				status = "disabled"
			}
			path := "missing"
			if p, err := exec.LookPath(l.Command); err == nil {
				path = p
			}
			fmt.Fprintf(&b, "- %s: %s (command: %s, path: %s)\n", name, status, l.Command, path)
		}
		cmd.Print(b.String())
		return nil
	},
}

var lspEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a configured LSP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setLSPDisabled(cmd, args[0], false)
	},
}

var lspDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a configured LSP server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setLSPDisabled(cmd, args[0], true)
	},
}

var lspTestCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test a configured LSP server (path and quick version)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}
		debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
		dataDir, _ := cmd.Root().PersistentFlags().GetString("data-dir")
		cfg, err := config.Init(cwd, dataDir, debug)
		if err != nil {
			return err
		}
		l, ok := cfg.LSP[name]
		if !ok {
			return fmt.Errorf("lsp %q not found", name)
		}
		path, err := exec.LookPath(l.Command)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s: missing (command: %s)\n", name, l.Command)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "- %s: found (%s)\n", name, path)
		vv := exec.Command(l.Command, append(l.Args, "--version")...)
		vv.Env = append(os.Environ(), l.ResolvedEnv()...)
		out, _ := vv.CombinedOutput()
		line := strings.TrimSpace(string(out))
		if i := strings.IndexByte(line, '\n'); i >= 0 {
			line = line[:i]
		}
		if line != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  version: %s\n", line)
		}
		return nil
	},
}

func setLSPDisabled(cmd *cobra.Command, name string, disabled bool) error {
	cwd, err := ResolveCwd(cmd)
	if err != nil {
		return err
	}
	debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
	dataDir, _ := cmd.Root().PersistentFlags().GetString("data-dir")
	cfg, err := config.Init(cwd, dataDir, debug)
	if err != nil {
		return err
	}
	// Update in-memory config so the change is immediate
	if l, ok := cfg.LSP[name]; ok {
		l.Disabled = disabled
		cfg.LSP[name] = l
	} else {
		return fmt.Errorf("lsp %q not found", name)
	}

	// Persist to overrides file
	overrides := config.GlobalConfigData()
	_ = os.MkdirAll(filepath.Dir(overrides), 0o755)
	data := map[string]any{}
	if b, err := os.ReadFile(overrides); err == nil && len(b) > 0 {
		_ = json.Unmarshal(b, &data)
	}
	lspObj := map[string]any{}
	if v, ok := data["lsp"].(map[string]any); ok {
		lspObj = v
	}
	entry := map[string]any{"disabled": disabled}
	if existing, ok := lspObj[name].(map[string]any); ok {
		existing["disabled"] = disabled
		lspObj[name] = existing
	} else {
		lspObj[name] = entry
	}
	data["lsp"] = lspObj
	bts, _ := json.MarshalIndent(data, "", "  ")
	if err := os.WriteFile(overrides, bts, 0o600); err != nil {
		return fmt.Errorf("failed to write overrides: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "updated lsp %s: disabled=%v\n", name, disabled)
	return nil
}
