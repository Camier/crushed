package status

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/help"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/lsp"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/crush/internal/tui/util"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

type StatusCmp interface {
	util.Model
	ToggleFullHelp()
	SetKeyMap(keyMap help.KeyMap)
}

type statusCmp struct {
	info       util.InfoMsg
	width      int
	messageTTL time.Duration
	help       help.Model
	keyMap     help.KeyMap

	lspStatuses   map[string]app.LSPEvent
	notifications []util.InfoMsg
}

// clearMessageCmd is a command that clears status messages after a timeout
func (m *statusCmp) clearMessageCmd(ttl time.Duration) tea.Cmd {
	return tea.Tick(ttl, func(time.Time) tea.Msg {
		return util.ClearStatusMsg{}
	})
}

func (m *statusCmp) Init() tea.Cmd {
	return nil
}

func (m *statusCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.help.Width = msg.Width - 2
		return m, nil

	// Handle status info
	case util.InfoMsg:
		m.info = msg
		m.notifications = append([]util.InfoMsg{msg}, m.notifications...)
		if len(m.notifications) > 5 {
			m.notifications = m.notifications[:5]
		}
		ttl := msg.TTL
		if ttl == 0 {
			ttl = m.messageTTL
		}
		return m, m.clearMessageCmd(ttl)
	case util.ClearStatusMsg:
		m.info = util.InfoMsg{}
	case pubsub.Event[app.LSPEvent]:
		if m.lspStatuses == nil {
			m.lspStatuses = make(map[string]app.LSPEvent)
		}
		m.lspStatuses[msg.Payload.Name] = msg.Payload
	}
	return m, nil
}

func (m *statusCmp) View() string {
	t := styles.CurrentTheme()
	output := t.S().Base.Padding(0, 1, 1, 1).Render(m.help.View(m.keyMap))
	if m.info.Msg != "" {
		output = m.infoMsg()
	}

	if lspStatus := m.renderLSPStatuses(); lspStatus != "" {
		if output != "" {
			output += "\n"
		}
		output += lspStatus
	}

	if notices := m.renderNotifications(); notices != "" {
		if output != "" {
			output += "\n"
		}
		output += notices
	}

	return output
}

func (m *statusCmp) infoMsg() string {
	t := styles.CurrentTheme()
	message := ""
	infoType := ""
	switch m.info.Type {
	case util.InfoTypeError:
		infoType = t.S().Base.Background(t.Red).Padding(0, 1).Render("ERROR")
		widthLeft := m.width - (lipgloss.Width(infoType) + 2)
		info := ansi.Truncate(m.info.Msg, widthLeft, "…")
		message = t.S().Base.Background(t.Error).Width(widthLeft+2).Foreground(t.White).Padding(0, 1).Render(info)
	case util.InfoTypeWarn:
		infoType = t.S().Base.Foreground(t.BgOverlay).Background(t.Yellow).Padding(0, 1).Render("WARNING")
		widthLeft := m.width - (lipgloss.Width(infoType) + 2)
		info := ansi.Truncate(m.info.Msg, widthLeft, "…")
		message = t.S().Base.Foreground(t.BgOverlay).Width(widthLeft+2).Background(t.Warning).Padding(0, 1).Render(info)
	default:
		infoType = t.S().Base.Foreground(t.BgOverlay).Background(t.Green).Padding(0, 1).Render("OKAY!")
		widthLeft := m.width - (lipgloss.Width(infoType) + 2)
		info := ansi.Truncate(m.info.Msg, widthLeft, "…")
		message = t.S().Base.Background(t.Success).Width(widthLeft+2).Foreground(t.White).Padding(0, 1).Render(info)
	}
	return ansi.Truncate(infoType+message, m.width, "…")
}

func (m *statusCmp) renderLSPStatuses() string {
	if len(m.lspStatuses) == 0 || m.width <= 0 {
		return ""
	}

	names := make([]string, 0, len(m.lspStatuses))
	for name := range m.lspStatuses {
		names = append(names, name)
	}
	sort.Strings(names)

	theme := styles.CurrentTheme()
	segments := make([]string, 0, len(names))

	for _, name := range names {
		event := m.lspStatuses[name]
		icon := styles.ToolPending
		style := theme.S().Muted

		switch event.State {
		case lsp.StateReady:
			icon = styles.ToolSuccess
			style = theme.S().Success
		case lsp.StateError:
			icon = styles.ErrorIcon
			style = theme.S().Error
		case lsp.StateStarting:
			icon = styles.WarningIcon
			style = theme.S().Warning
		}

		label := name
		if event.DiagnosticCount > 0 {
			label = fmt.Sprintf("%s (%d)", label, event.DiagnosticCount)
		}
		if event.State == lsp.StateError && event.Error != nil {
			label = fmt.Sprintf("%s !", label)
		}

		segments = append(segments, style.Render(fmt.Sprintf("%s %s", icon, label)))
	}

	line := theme.S().Muted.Render("LSP") + " " + strings.Join(segments, theme.S().Muted.Render("  "))
	return ansi.Truncate(line, m.width, "…")
}

func (m *statusCmp) renderNotifications() string {
	notes := m.notifications
	if m.info.Msg != "" && len(notes) > 0 && notes[0] == m.info {
		notes = notes[1:]
	}
	if len(notes) == 0 || m.width <= 0 {
		return ""
	}

	theme := styles.CurrentTheme()
	segments := make([]string, 0, len(notes))
	for _, note := range notes {
		style := theme.S().Info
		icon := styles.InfoIcon
		switch note.Type {
		case util.InfoTypeWarn:
			style = theme.S().Warning
			icon = styles.WarningIcon
		case util.InfoTypeError:
			style = theme.S().Error
			icon = styles.ErrorIcon
		}
		segments = append(segments, style.Render(fmt.Sprintf("%s %s", icon, ansi.Truncate(note.Msg, max(0, m.width-10), "…"))))
	}

	line := theme.S().Muted.Render("Alerts") + " " + strings.Join(segments, theme.S().Muted.Render("  "))
	return ansi.Truncate(line, m.width, "…")
}

func (m *statusCmp) ToggleFullHelp() {
	m.help.ShowAll = !m.help.ShowAll
}

func (m *statusCmp) SetKeyMap(keyMap help.KeyMap) {
	m.keyMap = keyMap
}

func NewStatusCmp() StatusCmp {
	t := styles.CurrentTheme()
	help := help.New()
	help.Styles = t.S().Help
	return &statusCmp{
		messageTTL:  5 * time.Second,
		help:        help,
		lspStatuses: make(map[string]app.LSPEvent),
	}
}
