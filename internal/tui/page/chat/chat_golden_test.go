package chat

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/csync"
	"github.com/charmbracelet/crush/internal/history"
	lspclient "github.com/charmbracelet/crush/internal/lsp"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/session"
	chatcmp "github.com/charmbracelet/crush/internal/tui/components/chat"
	"github.com/charmbracelet/crush/internal/tui/components/dialogs/commands"
	"github.com/charmbracelet/x/exp/golden"
)

// Placeholders are deterministic by default in editor.New()

func setupChatTestConfig(t *testing.T) *config.Config {
	t.Helper()
	os.Setenv("CRUSH_DISABLE_PROVIDER_AUTO_UPDATE", "1")
	tdir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tdir, "config"))
	// Use a deterministic working directory to stabilize snapshots
	work := "/proj"
	data := filepath.Join(tdir, ".crush")
	cfg, err := config.Init(work, data, false)
	if err != nil {
		t.Fatalf("config.Init error: %v", err)
	}
	cfg.Providers = csync.NewMap[string, config.ProviderConfig]()
	cfg.Providers.Set("local", config.ProviderConfig{
		ID:   "local",
		Name: "Local",
		Type: catwalk.TypeOpenAI,
		Models: []catwalk.Model{{
			ID: "test", Name: "Test", ContextWindow: 1000, DefaultMaxTokens: 200,
		}},
	})
	cfg.Models[config.SelectedModelTypeLarge] = config.SelectedModel{Provider: "local", Model: "test", MaxTokens: 200}
	cfg.SetupAgents()
	// Ensure HasInitialDataConfig() is true
	overrides := config.GlobalConfigData()
	if err := os.MkdirAll(filepath.Dir(overrides), 0o755); err != nil {
		t.Fatalf("mkdir overrides: %v", err)
	}
	if err := os.WriteFile(overrides, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write overrides: %v", err)
	}
	return cfg
}

// minimalApp returns a stub app with only the fields used by ChatPage.
type fakeMessages struct{}

func (fakeMessages) Subscribe(ctx context.Context) <-chan pubsub.Event[message.Message] {
	ch := make(chan pubsub.Event[message.Message])
	close(ch)
	return ch
}

func (fakeMessages) Create(context.Context, string, message.CreateMessageParams) (message.Message, error) {
	return message.Message{}, nil
}
func (fakeMessages) Update(context.Context, message.Message) error { return nil }
func (fakeMessages) Get(context.Context, string) (message.Message, error) {
	return message.Message{}, nil
}

func (fakeMessages) List(context.Context, string) ([]message.Message, error) {
	return []message.Message{}, nil
}
func (fakeMessages) Delete(context.Context, string) error                { return nil }
func (fakeMessages) DeleteSessionMessages(context.Context, string) error { return nil }

type fakeHistory struct{}

func (fakeHistory) Subscribe(ctx context.Context) <-chan pubsub.Event[history.File] {
	ch := make(chan pubsub.Event[history.File])
	close(ch)
	return ch
}

func (fakeHistory) Create(context.Context, string, string, string) (history.File, error) {
	return history.File{}, nil
}

func (fakeHistory) CreateVersion(context.Context, string, string, string) (history.File, error) {
	return history.File{}, nil
}
func (fakeHistory) Get(context.Context, string) (history.File, error) { return history.File{}, nil }
func (fakeHistory) GetByPathAndSession(context.Context, string, string) (history.File, error) {
	return history.File{}, nil
}

func (fakeHistory) ListBySession(context.Context, string) ([]history.File, error) {
	return []history.File{}, nil
}

func (fakeHistory) ListLatestSessionFiles(context.Context, string) ([]history.File, error) {
	return []history.File{}, nil
}
func (fakeHistory) Delete(context.Context, string) error             { return nil }
func (fakeHistory) DeleteSessionFiles(context.Context, string) error { return nil }

func minimalApp(cfg *config.Config) *app.App {
	return &app.App{
		LSPClients: make(map[string]*lspclient.Client),
		Messages:   fakeMessages{},
		History:    fakeHistory{},
	}
}

func setupFullLayoutPage(t *testing.T) *chatPage {
	cfg := setupChatTestConfig(t)
	// Ensure compact mode is off by default
	cfg.Options.TUI.CompactMode = false
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	// Large size to guarantee non-compact
	p.SetSize(130, 40)
	return p
}

func TestChatPage_NoSession_Compact(t *testing.T) {
	// Run sequentially to avoid cross-test config interference
	cfg := setupChatTestConfig(t)
	_ = cfg

	cp := New(minimalApp(cfg)).(*chatPage)
	// Initialize and size for compact
	_ = cp.Init()
	cp.SetSize(80, 20)
	golden.RequireEqual(t, []byte(cp.View()))
}

func TestChatPage_Session_Compact(t *testing.T) {
	// Run sequentially to avoid cross-test config interference
	cfg := setupChatTestConfig(t)
	_ = cfg

	cp := New(minimalApp(cfg)).(*chatPage)
	_ = cp.Init()
	cp.SetSize(80, 20)
	// Select a session
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, cmd := cp.Update(chatcmp.SessionSelectedMsg(sess))
	if cmd != nil {
		_ = cmd()
	}
	golden.RequireEqual(t, []byte(cp.View()))
}

func TestChatPage_Session_Compact_Details(t *testing.T) {
	// Run sequentially to avoid cross-test config interference
	cfg := setupChatTestConfig(t)
	_ = cfg

	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(80, 20)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Open details overlay
	p.showingDetails = true
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout(t *testing.T) {
	p := setupFullLayoutPage(t)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_Details(t *testing.T) {
	p := setupFullLayoutPage(t)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	p.showingDetails = true
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_Width120(t *testing.T) {
	p := New(minimalApp(setupChatTestConfig(t))).(*chatPage)
	_ = p.Init()
	// Large enough height, width at breakpoint
	p.SetSize(120, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithFiles(t *testing.T) {
	p := setupFullLayoutPage(t)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Seed file history via events
	// src/main.go
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.CreatedEvent, Payload: history.File{Path: filepath.Join("/proj", "src", "main.go"), Version: 0, Content: "package main\nfunc main() {}\n"}})
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.UpdatedEvent, Payload: history.File{Path: filepath.Join("/proj", "src", "main.go"), Version: 1, Content: "package main\nfunc main() { println(\"hi\") }\n"}})
	// README.md
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.CreatedEvent, Payload: history.File{Path: filepath.Join("/proj", "README.md"), Version: 0, Content: "hello\n"}})
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.UpdatedEvent, Payload: history.File{Path: filepath.Join("/proj", "README.md"), Version: 2, Content: "hello world\n"}})
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithServices(t *testing.T) {
	cfg := setupChatTestConfig(t)
	// Add several LSPs and MCPs to exercise limits
	cfg.LSP = config.LSPs{
		"go":         {Command: "gopls"},
		"typescript": {Command: "typescript-language-server", Args: []string{"--stdio"}},
		"python":     {Command: "pyright-langserver"},
		"nix":        {Command: "nil"},
		"rust":       {Command: "rust-analyzer"},
		"bash":       {Command: "bash-language-server"},
		"json":       {Command: "vscode-json-languageserver"},
		"yaml":       {Command: "yaml-language-server"},
	}
	cfg.MCP = config.MCPs{
		"filesystem": {Type: config.MCPStdio, Command: "node", Args: []string{"fs.js"}},
		"http":       {Type: config.MCPHttp, URL: "https://example.com/mcp/"},
		"sse":        {Type: config.MCPSse, URL: "https://example.com/mcp/sse"},
		"build":      {Type: config.MCPStdio, Command: "make"},
		"grep":       {Type: config.MCPStdio, Command: "grep"},
		"tools":      {Type: config.MCPStdio, Command: "tools"},
		"docker":     {Type: config.MCPStdio, Command: "docker"},
		"kubectl":    {Type: config.MCPStdio, Command: "kubectl"},
	}
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(130, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithServices_Width200(t *testing.T) {
	p := setupFullLayoutPage(t)
	p.SetSize(200, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Use configured services; size change exercises limits
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithServices_Width160(t *testing.T) {
	p := setupFullLayoutPage(t)
	p.SetSize(160, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Use configured services; size change exercises limits
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithFiles_Width200(t *testing.T) {
	p := setupFullLayoutPage(t)
	p.SetSize(200, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Seed files via events for variety
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.CreatedEvent, Payload: history.File{Path: filepath.Join("/proj", "A.go"), Version: 0, Content: "package a\n"}})
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.UpdatedEvent, Payload: history.File{Path: filepath.Join("/proj", "A.go"), Version: 1, Content: "package a\nfunc A(){}\n"}})
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithFiles_Width160(t *testing.T) {
	p := setupFullLayoutPage(t)
	p.SetSize(160, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Seed files via events for variety
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.CreatedEvent, Payload: history.File{Path: filepath.Join("/proj", "A.go"), Version: 0, Content: "package a\n"}})
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.UpdatedEvent, Payload: history.File{Path: filepath.Join("/proj", "A.go"), Version: 1, Content: "package a\nfunc A(){}\n"}})
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_Details_Width160(t *testing.T) {
	p := setupFullLayoutPage(t)
	p.SetSize(160, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	p.showingDetails = true
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithServices_Width100(t *testing.T) {
	p := setupFullLayoutPage(t)
	p.SetSize(100, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_FullLayout_WithFiles_Width100(t *testing.T) {
	p := setupFullLayoutPage(t)
	p.SetSize(100, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Seed single file to ensure section draws
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.CreatedEvent, Payload: history.File{Path: filepath.Join("/proj", "B.go"), Version: 0, Content: "package b\n"}})
	_, _ = p.Update(pubsub.Event[history.File]{Type: pubsub.UpdatedEvent, Payload: history.File{Path: filepath.Join("/proj", "B.go"), Version: 1, Content: "package b\nfunc B(){}\n"}})
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_FocusSwitch_Compact(t *testing.T) {
	cfg := setupChatTestConfig(t)
	cfg.Options.TUI.CompactMode = true
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(80, 20)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Switch focus to chat panel
	p.changeFocus()
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_FocusSwitch_Full(t *testing.T) {
	cfg := setupChatTestConfig(t)
	cfg.Options.TUI.CompactMode = false
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(130, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Switch focus to chat panel
	p.changeFocus()
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_ToggleCompact_FromFull(t *testing.T) {
	cfg := setupChatTestConfig(t)
	cfg.Options.TUI.CompactMode = false
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(130, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Toggle to compact
	_, _ = p.Update(commands.ToggleCompactModeMsg{})
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_ToggleCompact_FromCompact(t *testing.T) {
	cfg := setupChatTestConfig(t)
	cfg.Options.TUI.CompactMode = true
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(130, 40)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	// Toggle to full
	_, _ = p.Update(commands.ToggleCompactModeMsg{})
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_Session_Compact_Details_Focus(t *testing.T) {
	cfg := setupChatTestConfig(t)
	cfg.Options.TUI.CompactMode = true
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(80, 20)
	sess := session.Session{ID: "s1", Title: "Session One"}
	_, _ = p.Update(chatcmp.SessionSelectedMsg(sess))
	p.showingDetails = true
	// Switch focus while details are open
	p.changeFocus()
	golden.RequireEqual(t, []byte(p.View()))
}

func TestChatPage_NoSession_Compact_Initialized(t *testing.T) {
	cfg := setupChatTestConfig(t)
	// Simulate initialized project so onboarding and init prompts are skipped
	_ = os.MkdirAll(cfg.Options.DataDirectory, 0o755)
	if err := config.MarkProjectInitialized(); err != nil {
		t.Fatalf("mark initialized: %v", err)
	}
	p := New(minimalApp(cfg)).(*chatPage)
	_ = p.Init()
	p.SetSize(80, 20)
	golden.RequireEqual(t, []byte(p.View()))
}
