package providerstatus

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/config"
)

const defaultStartupPollInterval = 2 * time.Second

// EnsureProviderReady verifies that the given provider is reachable. If the provider is
// unreachable and a startup command is configured, the command is executed and the
// health check is retried until it becomes ready or the timeout expires.
func EnsureProviderReady(ctx context.Context, cwd string, prov config.ProviderConfig) error {
	resolved := resolveProviderConfig(prov)

	if resolved.BaseURL == "" {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if ready, detail, err := CheckHealth(ctx, nil, resolved); err != nil {
		return fmt.Errorf("provider %s health check failed: %w", providerIdentifier(resolved), err)
	} else if ready {
		return nil
	} else if resolved.StartupCommand == "" || os.Getenv("CRUSH_SKIP_PROVIDER_STARTUP") == "1" {
		if detail == "" {
			detail = "no response"
		}
		return fmt.Errorf("provider %s is unreachable (%s); try selecting a different provider with `crush models use` or configure a startup_command", providerIdentifier(resolved), detail)
	}

	providerName := providerIdentifier(resolved)
	timeoutSeconds := resolved.StartupTimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}

	slog.Info("Attempting to start provider", "provider", providerName, "command", resolved.StartupCommand)
	deadlineCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd := buildShellCommand(deadlineCtx, resolved.StartupCommand)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start provider %s: %w", providerName, err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	ticker := time.NewTicker(defaultStartupPollInterval)
	defer ticker.Stop()

	var lastDetail string
	for {
		select {
		case <-deadlineCtx.Done():
			if cmd.ProcessState == nil && cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			if lastDetail == "" {
				lastDetail = deadlineCtx.Err().Error()
			}
			return fmt.Errorf("provider %s did not become ready within %d seconds (%s)", providerName, timeoutSeconds, lastDetail)
		case err := <-done:
			if err != nil {
				return fmt.Errorf("provider %s startup command failed: %w", providerName, err)
			}
			// Command exited cleanly; continue polling in case it spawned a daemon.
		case <-ticker.C:
			ready, detail, checkErr := CheckHealth(ctx, nil, resolved)
			if checkErr != nil {
				lastDetail = checkErr.Error()
				continue
			}
			if ready {
				slog.Info("Provider is ready", "provider", providerName)
				return nil
			}
			lastDetail = detail
		}
	}
}

func buildShellCommand(ctx context.Context, command string) *exec.Cmd {
	// Use POSIX-compatible shell invocation consistently
	return exec.CommandContext(ctx, "bash", "-lc", command)
}

func providerIdentifier(prov config.ProviderConfig) string {
	if strings.TrimSpace(prov.Name) != "" {
		return prov.Name
	}
	if strings.TrimSpace(prov.ID) != "" {
		return prov.ID
	}
	if strings.TrimSpace(prov.BaseURL) != "" {
		return prov.BaseURL
	}
	return "unknown"
}
