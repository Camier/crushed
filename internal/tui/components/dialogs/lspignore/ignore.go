package lspignore

import (
	"strings"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textarea"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/lsp/watcher"
	"github.com/charmbracelet/crush/internal/tui/components/core"
	"github.com/charmbracelet/crush/internal/tui/components/dialogs"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/crush/internal/tui/util"
	"github.com/charmbracelet/lipgloss/v2"
)

const DialogID dialogs.DialogID = "lsp-ignore"

type Dialog struct {
	textarea     *textarea.Model
	keyMap       keyMap
	help         help.Model
	width        int
	height       int
	screenWidth  int
	screenHeight int
}

func New() *Dialog {
	ta := textarea.New()
	ta.SetWidth(60)
	ta.SetHeight(12)
	ta.ShowLineNumbers = false
	ta.CharLimit = -1
	ta.Placeholder = "One glob or path fragment per line"

	theme := styles.CurrentTheme()
	ta.SetStyles(theme.S().TextArea)

	helpModel := help.New()
	helpModel.Styles = theme.S().Help

	return &Dialog{
		textarea: ta,
		keyMap:   defaultKeyMap(),
		help:     helpModel,
		width:    64,
		height:   18,
	}
}

func (d *Dialog) Init() tea.Cmd {
	cfg := config.Get()
	lines := strings.Join(cfg.CustomLSPIgnorePaths(), "\n")
	d.textarea.SetValue(lines)
	d.help.Width = d.width - 4
	return d.textarea.Focus()
}

func (d *Dialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.screenWidth = msg.Width
		d.screenHeight = msg.Height
		d.width = min(msg.Width-4, 80)
		d.height = min(msg.Height-4, 20)
		d.textarea.SetWidth(d.width - 4)
		d.textarea.SetHeight(d.height - 6)
		d.help.Width = d.width - 4
		return d, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, d.keyMap.Save):
			return d, d.save()
		case key.Matches(msg, d.keyMap.Close):
			return d, util.CmdHandler(dialogs.CloseDialogMsg{})
		}
	}

	ta, cmd := d.textarea.Update(msg)
	d.textarea = ta
	return d, cmd
}

func (d *Dialog) View() string {
	theme := styles.CurrentTheme()
	title := core.Title("LSP Ignore Paths", d.width-4)
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		theme.S().Base.Padding(0, 1, 1, 1).Render(title),
		theme.S().Subtle.Padding(0, 1, 1, 1).Render("Patterns apply to the workspace root. Leave blank lines to remove entries."),
		theme.S().Base.Padding(0, 1, 1, 1).Render(d.textarea.View()),
		theme.S().Base.Padding(1, 1, 0, 1).Render(d.help.View(d.keyMap)),
	)
	return theme.S().Base.
		Width(d.width).
		Height(d.height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderFocus).
		Render(body)
}

func (d *Dialog) Position() (int, int) {
	row := max(1, d.screenHeight/2-d.height/2)
	col := max(2, d.screenWidth/2-d.width/2)
	return row, col
}

func (d *Dialog) ID() dialogs.DialogID {
	return DialogID
}

func (d *Dialog) save() tea.Cmd {
	raw := d.textarea.Value()
	lines := strings.Split(raw, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	if err := config.Get().SetLSPIgnorePaths(cleaned); err != nil {
		return util.ReportError(err)
	}

	watcher.ReloadIgnoreSystem()

	return tea.Sequence(
		util.ReportInfo("Updated LSP ignore paths"),
		util.CmdHandler(dialogs.CloseDialogMsg{}),
	)
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
	Save  key.Binding
	Close key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Save, k.Close}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Save, k.Close}}
}

func (k keyMap) KeyBindings() []key.Binding {
	return []key.Binding{k.Save, k.Close}
}
