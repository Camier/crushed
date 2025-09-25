package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/providerstatus"
	"github.com/spf13/cobra"
)

func init() {
	doctorProvidersCmd.Flags().Bool("start", false, "Attempt to start providers using their startup_command when unreachable")
	doctorCmd.AddCommand(doctorProvidersCmd)
	doctorCmd.AddCommand(doctorLSPCmd)
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose common Crush configuration issues",
}

var doctorProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Check provider connectivity and optionally attempt repairs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()

		debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
		dataDir, _ := cmd.Root().PersistentFlags().GetString("data-dir")

		cfg, err := config.Init(cwd, dataDir, debug)
		if err != nil {
			return err
		}

		providers := make([]providerRow, 0)
		for id, p := range cfg.Providers.Seq2() {
			providers = append(providers, providerRow{id: id, prov: p})
		}
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].id < providers[j].id
		})

		attemptStart, _ := cmd.Flags().GetBool("start")

		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		fmt.Fprintln(out, "Checking providers...")
		for _, row := range providers {
			name := row.id
			if name == "" {
				name = row.prov.Name
			}

			ready, detail, checkErr := providerstatus.CheckHealth(ctx, nil, row.prov)
			if checkErr != nil {
				fmt.Fprintf(out, "- %s: health check failed (%s)\n", name, checkErr.Error())
				if row.prov.BaseURL != "" {
					fmt.Fprintf(out, "  url: %s\n", row.prov.BaseURL)
				}
				if row.prov.StartupHealthPath != "" {
					fmt.Fprintf(out, "  health: %s (timeout %ds)\n", row.prov.StartupHealthPath, row.prov.StartupTimeoutSeconds)
				}
				continue
			}
			if ready {
				if row.prov.BaseURL != "" {
					fmt.Fprintf(out, "- %s: ready (url: %s)\n", name, row.prov.BaseURL)
				} else {
					fmt.Fprintf(out, "- %s: ready\n", name)
				}
				continue
			}

			if !attemptStart || row.prov.StartupCommand == "" {
				if detail == "" {
					detail = "no response"
				}
				fmt.Fprintf(out, "- %s: unreachable (%s)\n", name, detail)
				if row.prov.BaseURL != "" {
					fmt.Fprintf(out, "  url: %s\n", row.prov.BaseURL)
				}
				if row.prov.StartupHealthPath != "" {
					fmt.Fprintf(out, "  health: %s (timeout %ds)\n", row.prov.StartupHealthPath, row.prov.StartupTimeoutSeconds)
				}
				if row.prov.StartupCommand != "" {
					fmt.Fprintln(out, "  hint: try 'crush doctor providers --start' to auto-start this provider")
				}
				continue
			}

			fmt.Fprintf(out, "- %s: unreachable (%s), attempting startup...\n", name, detail)
			if err := providerstatus.EnsureProviderReady(ctx, cwd, row.prov); err != nil {
				fmt.Fprintf(out, "  ✗ startup failed: %s\n", err.Error())
				continue
			}

			fmt.Fprintf(out, "  ✓ provider is ready\n")
		}

		return nil
	},
}

var lspInstallHints = map[string]string{
	"gopls":                      "go install golang.org/x/tools/gopls@latest",
	"typescript-language-server": "npm install -g typescript-language-server typescript",
	"pylsp":                      "pip install 'python-lsp-server[all]'",
	"pyright":                    "npm install -g pyright",
	"rust-analyzer":              "rustup component add rust-analyzer",
	"bash-language-server":       "npm install -g bash-language-server",
}

var doctorLSPCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Check LSP server availability and versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := ResolveCwd(cmd)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()

		debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
		dataDir, _ := cmd.Root().PersistentFlags().GetString("data-dir")

		cfg, err := config.Init(cwd, dataDir, debug)
		if err != nil {
			return err
		}

		// Deterministic order
		names := make([]string, 0, len(cfg.LSP))
		for name := range cfg.LSP {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Fprintln(out, "Checking LSP servers...")
		doVersion := os.Getenv("CRUSH_LSP_VERSION_CHECK") == "1"

		for _, name := range names {
			l := cfg.LSP[name]
			status := "missing"
			if path, err := exec.LookPath(l.Command); err == nil {
				status = "found"
				fmt.Fprintf(out, "- %s: %s (%s)\n", name, status, path)
				if doVersion {
					// Attempt a quick version check with a short timeout
					ctx, cancel := context.WithTimeout(cmd.Context(), 500*time.Millisecond)
					defer cancel()
					vv := exec.CommandContext(ctx, l.Command, append(l.Args, "--version")...)
					vv.Env = append(os.Environ(), l.ResolvedEnv()...)
					outBytes, _ := vv.CombinedOutput()
					line := strings.TrimSpace(string(outBytes))
					if line != "" {
						if i := strings.IndexByte(line, '\n'); i >= 0 {
							line = line[:i]
						}
						fmt.Fprintf(out, "  version: %s\n", line)
					}
				}
			} else {
				fmt.Fprintf(out, "- %s: %s (command: %s)\n", name, status, l.Command)
				cmdName := strings.ToLower(filepath.Base(l.Command))
				if hint, ok := lspInstallHints[cmdName]; ok {
					fmt.Fprintf(out, "  hint: install via `%s`\n", hint)
				} else {
					fmt.Fprintf(out, "  hint: ensure %s is installed and on PATH or set an absolute command\n", l.Command)
				}
			}
		}

		return nil
	},
}
