package app

import (
	"context"
	"time"

	"github.com/charmbracelet/crush/internal/providerstatus"
	"github.com/charmbracelet/crush/internal/pubsub"
)

const providerStatusInterval = 30 * time.Second

// ProviderStatus captures runtime readiness information for the active agent provider.
type ProviderStatus struct {
	ProviderID    string
	ProviderName  string
	ModelID       string
	ModelName     string
	BaseURL       string
	Ready         bool
	Detail        string
	StreamEnabled bool
	LastChecked   time.Time
}

// ProviderStatus returns the latest cached provider status snapshot.
func (app *App) ProviderStatus() ProviderStatus {
	app.providerStatusMu.RLock()
	defer app.providerStatusMu.RUnlock()
	return app.providerStatus
}

// SubscribeProviderStatus exposes the provider status broker for UI consumers.
func (app *App) SubscribeProviderStatus(ctx context.Context) <-chan pubsub.Event[ProviderStatus] {
	return app.providerStatusBroker.Subscribe(ctx)
}

func (app *App) startProviderStatusMonitor() {
	ctx, cancel := context.WithCancel(app.globalCtx)
	app.cleanupFuncs = append(app.cleanupFuncs, func() error {
		cancel()
		return nil
	})

	app.serviceEventsWG.Go(func() {
		ticker := time.NewTicker(providerStatusInterval)
		defer ticker.Stop()

		app.refreshProviderStatus()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				app.refreshProviderStatus()
			}
		}
	})
}

func (app *App) refreshProviderStatus() {
	status := app.computeProviderStatus()
	app.providerStatusMu.Lock()
	app.providerStatus = status
	app.providerStatusMu.Unlock()
	app.providerStatusBroker.Publish(pubsub.UpdatedEvent, status)
}

func (app *App) computeProviderStatus() ProviderStatus {
	cfg := app.config
	status := ProviderStatus{
		LastChecked: time.Now(),
	}
	if cfg == nil {
		status.Detail = "configuration unavailable"
		return status
	}

	agentCfg, ok := cfg.Agents["coder"]
	if !ok || agentCfg.Model == "" {
		status.Detail = "coder agent not configured"
		return status
	}

	modelCfg, ok := cfg.Models[agentCfg.Model]
	if !ok {
		status.Detail = "model not selected"
		return status
	}
	status.ModelID = modelCfg.Model

	model := cfg.GetModelByType(agentCfg.Model)
	if model != nil {
		status.ModelName = model.Name
	}

	providerCfg := cfg.GetProviderForModel(agentCfg.Model)
	if providerCfg == nil {
		status.Detail = "provider not found"
		return status
	}

	status.ProviderID = providerCfg.ID
	if providerCfg.Name != "" {
		status.ProviderName = providerCfg.Name
	} else {
		status.ProviderName = providerCfg.ID
	}
	status.BaseURL = providerCfg.BaseURL
	status.StreamEnabled = !providerCfg.DisableStream

	if providerCfg.BaseURL == "" {
		status.Ready = true
		status.Detail = "no base URL configured"
		return status
	}

	ready, detail, err := providerstatus.CheckHealth(app.globalCtx, nil, *providerCfg)
	if err != nil {
		status.Detail = err.Error()
		return status
	}
	status.Ready = ready
	status.Detail = detail
	return status
}
