package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	doctorCmd.AddCommand(doctorMCPCmd)
}

var doctorMCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Check MCP (Model Context Protocol) configuration and basic connectivity",
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

		fmt.Fprintln(cmd.OutOrStdout(), "Checking MCP entries...")
		// deterministic order
		names := make([]string, 0, len(cfg.MCP))
		for name := range cfg.MCP {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			m := cfg.MCP[name]
			typeStr := string(m.Type)
			if typeStr == "" {
				typeStr = "stdio"
			}
			summary := fmt.Sprintf("- %s: %s", name, typeStr)
			fmt.Fprintln(cmd.OutOrStdout(), summary)
			switch strings.ToLower(typeStr) {
			case "stdio":
				status := "missing"
				if m.Command != "" {
					if _, err := execLookPath(m.Command); err == nil {
						status = "found"
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  command: %s (%s)\n", m.Command, status)
			case "http", "sse":
				url := m.URL
				fmt.Fprintf(cmd.OutOrStdout(), "  url: %s\n", url)
				// auth header presence (don’t print value)
				auth := "missing"
				if h := m.ResolvedHeaders(); strings.TrimSpace(h["Authorization"]) != "" {
					auth = "found"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  auth: %s\n", auth)
				// quick connectivity check (best effort)
				if url != "" {
					ctx, cancel := context.WithTimeout(cmd.Context(), 800*time.Millisecond)
					defer cancel()
					req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
					for k, v := range m.ResolvedHeaders() {
						req.Header.Set(k, v)
					}
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "  check: unreachable (%s)\n", err.Error())
					} else {
						_ = resp.Body.Close()
						fmt.Fprintf(cmd.OutOrStdout(), "  check: %s\n", resp.Status)
					}
				}
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "  note: unknown MCP type %q\n", typeStr)
			}
		}
		return nil
	},
}

// execLookPath is a small wrapper so tests can patch it if needed.
var execLookPath = func(file string) (string, error) { return execLookPathImpl(file) }

func execLookPathImpl(file string) (string, error) {
	return execLookPathStd(file)
}

// split out to avoid import cycles in tests
func execLookPathStd(file string) (string, error) {
	return exec.LookPath(file)
}
