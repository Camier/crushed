package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/providerstatus"
	"github.com/spf13/cobra"
)

func init() {
	doctorProvidersCmd.Flags().Bool("start", false, "Attempt to start providers using their startup_command when unreachable")
	doctorCmd.AddCommand(doctorProvidersCmd)
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
				continue
			}
			if ready {
				fmt.Fprintf(os.Stdout, "- %s: ready\n", name)
				continue
			}

			if !attemptStart || row.prov.StartupCommand == "" {
				if detail == "" {
					detail = "no response"
				}
				fmt.Fprintf(os.Stdout, "- %s: unreachable (%s)\n", name, detail)
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
