package header

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/fsext"
	"github.com/charmbracelet/crush/internal/lsp"
	"github.com/charmbracelet/crush/internal/lsp/protocol"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/crush/internal/tui/util"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type Header interface {
	util.Model
	SetSession(session session.Session) tea.Cmd
	SetWidth(width int) tea.Cmd
	SetDetailsOpen(open bool)
	SetProviderStatus(status app.ProviderStatus)
	ShowingDetails() bool
}

type header struct {
	width          int
	session        session.Session
	lspClients     map[string]*lsp.Client
	detailsOpen    bool
	providerStatus app.ProviderStatus
}

func New(lspClients map[string]*lsp.Client) Header {
	return &header{
		lspClients: lspClients,
		width:      0,
	}
}

func (h *header) Init() tea.Cmd {
	return nil
}

func (h *header) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pubsub.Event[session.Session]:
		if msg.Type == pubsub.UpdatedEvent {
			if h.session.ID == msg.Payload.ID {
				h.session = msg.Payload
			}
		}
	}
	return h, nil
}

func (h *header) View() string {
	if h.session.ID == "" {
		return ""
	}

	const (
		gap          = " "
		diag         = "╱"
		minDiags     = 3
		leftPadding  = 1
		rightPadding = 1
	)

	t := styles.CurrentTheme()

	var b strings.Builder

	b.WriteString(t.S().Base.Foreground(t.Secondary).Render("Charm™"))
	b.WriteString(gap)
	b.WriteString(styles.ApplyBoldForegroundGrad("CRUSH", t.Secondary, t.Primary))
	b.WriteString(gap)

	availDetailWidth := h.width - leftPadding - rightPadding - lipgloss.Width(b.String()) - minDiags
	details := h.details(availDetailWidth)

	remainingWidth := h.width -
		lipgloss.Width(b.String()) -
		lipgloss.Width(details) -
		leftPadding -
		rightPadding

	if remainingWidth > 0 {
		b.WriteString(t.S().Base.Foreground(t.Primary).Render(
			strings.Repeat(diag, max(minDiags, remainingWidth)),
		))
		b.WriteString(gap)
	}

	b.WriteString(details)

	return t.S().Base.Padding(0, rightPadding, 0, leftPadding).Render(b.String())
}

func (h *header) details(availWidth int) string {
	s := styles.CurrentTheme().S()

	var parts []string

	if providerSegment := h.providerSummary(availWidth); providerSegment != "" {
		parts = append(parts, providerSegment)
	}

	// When details are open, include a compact LSP summary if LSPs are configured.
	if h.detailsOpen {
		if lsp := h.lspSummary(); lsp != "" {
			parts = append(parts, lsp)
		}
		// And a concise per-LSP status list (found/missing/disabled).
		if lst := h.lspDetailsList(); lst != "" {
			parts = append(parts, lst)
		}
	}

	errorCount := 0
	for _, l := range h.lspClients {
		for _, diagnostics := range l.GetDiagnostics() {
			for _, diagnostic := range diagnostics {
				if diagnostic.Severity == protocol.SeverityError {
					errorCount++
				}
			}
		}
	}

	if errorCount > 0 {
		parts = append(parts, s.Error.Render(fmt.Sprintf("%s%d", styles.ErrorIcon, errorCount)))
	}

	agentCfg := config.Get().Agents["coder"]
	model := config.Get().GetModelByType(agentCfg.Model)
	percentage := (float64(h.session.CompletionTokens+h.session.PromptTokens) / float64(model.ContextWindow)) * 100
	formattedPercentage := s.Muted.Render(fmt.Sprintf("%d%%", int(percentage)))
	parts = append(parts, formattedPercentage)

	const keystroke = "ctrl+d"
	if h.detailsOpen {
		parts = append(parts, s.Muted.Render(keystroke)+s.Subtle.Render(" close"))
	} else {
		parts = append(parts, s.Muted.Render(keystroke)+s.Subtle.Render(" open "))
	}

	dot := s.Subtle.Render(" • ")
	metadata := strings.Join(parts, dot)
	metadata = dot + metadata

	// Truncate cwd if necessary, and insert it at the beginning.
	const dirTrimLimit = 4
	cwd := fsext.DirTrim(fsext.PrettyPath(config.Get().WorkingDir()), dirTrimLimit)
	cwd = ansi.Truncate(cwd, max(0, availWidth-lipgloss.Width(metadata)), "…")
	cwd = s.Muted.Render(cwd)

	return cwd + metadata
}

// lspSummary returns a compact status for configured LSP servers when details are open.
// Format: "✓ LSP a/b" when all configured servers are found; otherwise "⚠ LSP a/b".
// It returns an empty string when no LSPs are configured or none are enabled.
func (h *header) lspSummary() string {
	cfg := config.Get()
	if cfg == nil {
		return ""
	}
	total := 0
	for _, l := range cfg.LSP {
		if !l.Disabled {
			total++
		}
	}
	if total == 0 {
		return ""
	}
	// Compute active by name match; fall back to clamping by total.
	active := 0
	for name, l := range cfg.LSP {
		if l.Disabled {
			continue
		}
		if _, ok := h.lspClients[name]; ok {
			active++
		}
	}
	if active == 0 && len(h.lspClients) > 0 {
		if len(h.lspClients) < total {
			active = len(h.lspClients)
		} else {
			active = total
		}
	}
	missing := total - active
	// Show spinner when any client reports recent work progress
	busy := false
	for _, c := range h.lspClients {
		if c != nil && c.WorkInProgress() {
			busy = true
			break
		}
	}

	t := styles.CurrentTheme().S()
	label := fmt.Sprintf("LSP %d/%d", active, total)
	if missing > 0 {
		return t.Warning.Render(fmt.Sprintf("%s %s", styles.WarningIcon, label))
	}
	if busy {
		return t.Info.Render(fmt.Sprintf("%s %s", styles.LoadingIcon, label))
	}
	return t.Success.Render(fmt.Sprintf("%s %s", styles.ToolSuccess, label))
}

// lspDetailsList returns a compact, per-LSP status list when details are open.
// Example: "gopls ✓, pyright ⚠, rust-analyzer off"
func (h *header) lspDetailsList() string {
	cfg := config.Get()
	if cfg == nil || len(cfg.LSP) == 0 {
		return ""
	}
	entries := cfg.LSP.Sorted()
	if len(entries) == 0 {
		return ""
	}
	s := styles.CurrentTheme().S()
	var items []string
	for _, entry := range entries {
		name := entry.Name
		l := entry.LSP
		if l.Disabled {
			items = append(items, s.Muted.Render(name+" off"))
			continue
		}
		c, ok := h.lspClients[name]
		if ok && c != nil {
			if c.WorkInProgress() {
				items = append(items, s.Info.Render(name+" "+styles.LoadingIcon))
			} else {
				items = append(items, s.Success.Render(name+" "+styles.ToolSuccess))
			}
			continue
		}
		// Configured but not active
		items = append(items, s.Warning.Render(name+" "+styles.WarningIcon))
	}
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, ", ")
}

func (h *header) providerSummary(availWidth int) string {
	status := h.providerStatus
	if status.ProviderID == "" && status.ProviderName == "" && status.ModelID == "" {
		return ""
	}

	t := styles.CurrentTheme()
	name := status.ProviderName
	if name == "" {
		name = status.ProviderID
	}

	model := status.ModelID
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		model = model[idx+1:]
	}
	if model != "" {
		name = fmt.Sprintf("%s/%s", name, model)
	}

	icon := styles.ToolSuccess
	style := t.S().Success
	if !status.Ready {
		if status.Detail != "" {
			icon = styles.WarningIcon
			style = t.S().Warning
		} else {
			icon = styles.ErrorIcon
			style = t.S().Error
		}
	}

	summary := fmt.Sprintf("%s %s", icon, name)
	if !status.StreamEnabled {
		summary += " · stream off"
	}
	if !status.Ready && status.Detail != "" {
		summary += " · " + status.Detail
	}

	if availWidth > 0 {
		summary = ansi.Truncate(summary, availWidth, "…")
	}
	return style.Render(summary)
}

func (h *header) SetDetailsOpen(open bool) {
	h.detailsOpen = open
}

func (h *header) SetProviderStatus(status app.ProviderStatus) {
	h.providerStatus = status
}

// SetSession implements Header.
func (h *header) SetSession(session session.Session) tea.Cmd {
	h.session = session
	return nil
}

// SetWidth implements Header.
func (h *header) SetWidth(width int) tea.Cmd {
	h.width = width
	return nil
}

// ShowingDetails implements Header.
func (h *header) ShowingDetails() bool {
	return h.detailsOpen
}
