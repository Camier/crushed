package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

		fmt.Fprintln(os.Stdout, "Checking providers...")
		for _, row := range providers {
			name := row.id
			if name == "" {
				name = row.prov.Name
			}

			ready, detail, checkErr := providerstatus.CheckHealth(ctx, nil, row.prov)
			if checkErr != nil {
				fmt.Fprintf(os.Stdout, "- %s: health check failed (%s)\n", name, checkErr.Error())
				if row.prov.BaseURL != "" {
					fmt.Fprintf(os.Stdout, "  url: %s\n", row.prov.BaseURL)
				}
				if row.prov.StartupHealthPath != "" {
					fmt.Fprintf(os.Stdout, "  health: %s (timeout %ds)\n", row.prov.StartupHealthPath, row.prov.StartupTimeoutSeconds)
				}
				continue
			}
			if ready {
				if row.prov.BaseURL != "" {
					fmt.Fprintf(os.Stdout, "- %s: ready (url: %s)\n", name, row.prov.BaseURL)
				} else {
					fmt.Fprintf(os.Stdout, "- %s: ready\n", name)
				}
				continue
			}

			if !attemptStart || row.prov.StartupCommand == "" {
				if detail == "" {
					detail = "no response"
				}
				fmt.Fprintf(os.Stdout, "- %s: unreachable (%s)\n", name, detail)
				if row.prov.BaseURL != "" {
					fmt.Fprintf(os.Stdout, "  url: %s\n", row.prov.BaseURL)
				}
				if row.prov.StartupHealthPath != "" {
					fmt.Fprintf(os.Stdout, "  health: %s (timeout %ds)\n", row.prov.StartupHealthPath, row.prov.StartupTimeoutSeconds)
				}
				if row.prov.StartupCommand != "" {
					fmt.Fprintln(os.Stdout, "  hint: try 'crush doctor providers --start' to auto-start this provider")
				}
				continue
			}

			fmt.Fprintf(os.Stdout, "- %s: unreachable (%s), attempting startup...\n", name, detail)
			if err := providerstatus.EnsureProviderReady(ctx, cwd, row.prov); err != nil {
				fmt.Fprintf(os.Stdout, "  ✗ startup failed: %s\n", err.Error())
				continue
			}

			fmt.Fprintf(os.Stdout, "  ✓ provider is ready\n")
		}

		return nil
	},
}

var doctorLSPCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Check LSP server availability and versions",
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

		// Deterministic order
		names := make([]string, 0, len(cfg.LSP))
		for name := range cfg.LSP {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Fprintln(os.Stdout, "Checking LSP servers...")
		doVersion := os.Getenv("CRUSH_LSP_VERSION_CHECK") == "1"

		for _, name := range names {
			l := cfg.LSP[name]
			status := "missing"
			if path, err := exec.LookPath(l.Command); err == nil {
				status = "found"
				fmt.Fprintf(os.Stdout, "- %s: %s (%s)\n", name, status, path)
				if doVersion {
					// Attempt a quick version check with a short timeout
					ctx, cancel := context.WithTimeout(cmd.Context(), 500*time.Millisecond)
					defer cancel()
					vv := exec.CommandContext(ctx, l.Command, append(l.Args, "--version")...)
					vv.Env = append(os.Environ(), l.ResolvedEnv()...)
					out, _ := vv.CombinedOutput()
					line := strings.TrimSpace(string(out))
					if line != "" {
						if i := strings.IndexByte(line, '\n'); i >= 0 {
							line = line[:i]
						}
						fmt.Fprintf(os.Stdout, "  version: %s\n", line)
					}
				}
			} else {
				fmt.Fprintf(os.Stdout, "- %s: %s (command: %s)\n", name, status, l.Command)
			}
		}

		return nil
	},
}
