package doctor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/providerstatus"
	"github.com/charmbracelet/crush/internal/tui/components/core"
	"github.com/charmbracelet/crush/internal/tui/components/dialogs"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/crush/internal/tui/util"
	"github.com/charmbracelet/lipgloss/v2"
)

const ProvidersDialogID dialogs.DialogID = "doctor-providers"

type providerDiag struct {
	ID             string
	Name           string
	Ready          bool
	Detail         string
	StartupCommand bool
	AttemptedStart bool
}

type providersDialog struct {
	spinner       spinner.Model
	entries       []providerDiag
	loading       bool
	width         int
	height        int
	screenWidth   int
	screenHeight  int
	keyMap        keyMap
	help          help.Model
	lastUpdatedAt time.Time
}

type providerDiagMsg struct {
	entries        []providerDiag
	loading        bool
	timestamp      time.Time
	attemptedStart bool
}

func NewProvidersDialog() dialogs.DialogModel {
	theme := styles.CurrentTheme()
	sp := spinner.New()
	sp.Style = theme.S().Subtle
	sp.Spinner = spinner.Dot

	helpModel := help.New()
	helpModel.Styles = theme.S().Help

	d := &providersDialog{
		spinner: sp,
		loading: true,
		width:   72,
		height:  18,
		keyMap:  defaultKeyMap(),
		help:    helpModel,
	}

	return d
}

func (d *providersDialog) Init() tea.Cmd {
	return tea.Batch(d.spinner.Tick, d.refresh(false))
}

func (d *providersDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.screenWidth = msg.Width
		d.screenHeight = msg.Height
		d.width = min(msg.Width-4, 90)
		d.height = min(msg.Height-4, 24)
		d.help.Width = d.width - 4
		return d, nil
	case providerDiagMsg:
		d.entries = msg.entries
		d.loading = false
		d.lastUpdatedAt = msg.timestamp
		return d, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, d.keyMap.Refresh):
			d.loading = true
			return d, tea.Batch(d.spinner.Tick, d.refresh(false))
		case key.Matches(msg, d.keyMap.Start):
			d.loading = true
			return d, tea.Batch(d.spinner.Tick, d.refresh(true))
		case key.Matches(msg, d.keyMap.Close):
			return d, util.CmdHandler(dialogs.CloseDialogMsg{})
		}
	}

	if d.loading {
		sp, cmd := d.spinner.Update(msg)
		d.spinner = sp
		return d, cmd
	}

	return d, nil
}

func (d *providersDialog) View() string {
	theme := styles.CurrentTheme()

	var body string
	if d.loading {
		body = theme.S().Base.Padding(2, 2, 2, 2).Render(fmt.Sprintf("%s Checking providers...", d.spinner.View()))
	} else if len(d.entries) == 0 {
		body = theme.S().Base.Padding(2, 2, 2, 2).Render("No providers configured.")
	} else {
		var lines []string
		for _, entry := range d.entries {
			icon := styles.CheckIcon
			statusStyle := theme.S().Success
			if !entry.Ready {
				icon = styles.ErrorIcon
				statusStyle = theme.S().Error
			}

			name := theme.S().Base.Render(entry.Name)
			status := statusStyle.Render(icon + " " + statusLabel(entry))
			detail := ""
			if entry.Detail != "" {
				detail = theme.S().Muted.Render(" · " + entry.Detail)
			}
			if entry.AttemptedStart && entry.StartupCommand {
				detail += theme.S().Subtle.Render(" · startup attempted")
			}
			lines = append(lines, fmt.Sprintf("%s%s%s", status, "  ", name+detail))
		}

		if !d.lastUpdatedAt.IsZero() {
			lines = append(lines, "")
			lines = append(lines, theme.S().Subtle.Render("Last checked "+time.Since(d.lastUpdatedAt).Truncate(time.Second).String()+" ago"))
		}

		body = theme.S().Base.Padding(1, 2, 1, 2).Render(strings.Join(lines, "\n"))
	}

	header := theme.S().Base.Padding(0, 1, 1, 1).Render(core.Title("Provider Doctor", d.width-4))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		theme.S().Base.Padding(1, 1, 0, 1).Render(d.help.View(d.keyMap)),
	)

	return theme.S().Base.
		Width(d.width).
		Height(d.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderFocus).
		Render(content)
}

func statusLabel(entry providerDiag) string {
	if entry.Ready {
		return "ready"
	}
	if entry.Detail != "" {
		return entry.Detail
	}
	if entry.StartupCommand {
		return "unreachable"
	}
	return "unreachable (no startup command)"
}

func (d *providersDialog) Position() (int, int) {
	row := max(1, d.screenHeight/2-d.height/2)
	col := max(2, d.screenWidth/2-d.width/2)
	return row, col
}

func (d *providersDialog) ID() dialogs.DialogID {
	return ProvidersDialogID
}

func (d *providersDialog) refresh(attemptStart bool) tea.Cmd {
	return func() tea.Msg {
		cfg := config.Get()
		entries := make([]providerDiag, 0, cfg.Providers.Len())
		for providerID, providerCfg := range cfg.Providers.Seq2() {
			if providerCfg.Disable {
				continue
			}

			entry := providerDiag{
				ID:             providerID,
				Name:           providerCfg.Name,
				StartupCommand: providerCfg.StartupCommand != "",
			}
			if entry.Name == "" {
				entry.Name = providerID
			}

			ready, detail, err := providerstatus.CheckHealth(context.Background(), nil, providerCfg)
			if err != nil {
				entry.Detail = err.Error()
			} else {
				entry.Detail = detail
			}
			entry.Ready = ready && err == nil

			if attemptStart && !entry.Ready && providerCfg.StartupCommand != "" {
				entry.AttemptedStart = true
				if err := providerstatus.EnsureProviderReady(context.Background(), cfg.WorkingDir(), providerCfg); err != nil {
					entry.Detail = err.Error()
				} else {
					ready, detail, err = providerstatus.CheckHealth(context.Background(), nil, providerCfg)
					if err != nil {
						entry.Detail = err.Error()
					} else {
						entry.Detail = detail
					}
					entry.Ready = ready && err == nil
				}
			}

			if entry.Detail == "" && !entry.Ready {
				entry.Detail = "unreachable"
			}
			entries = append(entries, entry)
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
		return providerDiagMsg{entries: entries, loading: false, timestamp: time.Now(), attemptedStart: attemptStart}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type keyMap struct {
	Refresh key.Binding
	Start   key.Binding
	Close   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "retry startup"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Refresh, k.Start, k.Close}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Refresh, k.Start, k.Close}}
}
