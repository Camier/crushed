package editor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/tui/util"
)

type fakeExec struct{ called bool }

func (f *fakeExec) Exec(_ *exec.Cmd, done func(error) tea.Msg) tea.Cmd {
	f.called = true
	return func() tea.Msg { return done(nil) }
}

type fakeFactory struct {
	name string
	args []string
}

func (f *fakeFactory) New(_ context.Context, name string, args ...string) *exec.Cmd {
	f.name = name
	f.args = append([]string{}, args...)
	// Return a benign command; it won't actually run under fakeExec
	return exec.CommandContext(context.Background(), "true")
}

func newTestEditor() *editorCmp {
	e := New(Dependencies{
		Exec:    &fakeExec{},
		Command: &fakeFactory{},
	}).(*editorCmp)
	return e
}

func TestOpenEditor_UsesVISUALAndAppendsFile(t *testing.T) {
	t.Setenv("VISUAL", "code --wait")
	t.Setenv("EDITOR", "")

	e := newTestEditor()
	// Intercept injected fakes
	fe := e.exec.(*fakeExec)
	ff := e.cmdFactory.(*fakeFactory)

	// Seed some initial content
	e.textarea.SetValue("hello world")

	// Trigger open editor via command message
	cmd := e.openEditor(e.textarea.Value())
	if cmd == nil {
		t.Fatalf("expected tea.Cmd, got nil")
	}
	msg := cmd()
	if !fe.called {
		t.Fatalf("expected executor to be called")
	}
	// Assert command name and that a tmp path was appended
	if ff.name != "code" {
		t.Fatalf("expected command 'code', got %q", ff.name)
	}
	if len(ff.args) < 1 {
		t.Fatalf("expected at least 1 arg (tmp file), got %d", len(ff.args))
	}
	tmp := ff.args[len(ff.args)-1]
	if !strings.Contains(filepath.Base(tmp), "msg_") {
		t.Fatalf("expected tmp file to be appended, got %q", tmp)
	}

	// Message should be OpenEditorMsg with textarea content
	oem, ok := msg.(OpenEditorMsg)
	if !ok {
		t.Fatalf("expected OpenEditorMsg, got %T", msg)
	}
	if strings.TrimSpace(oem.Text) != "hello world" {
		t.Fatalf("unexpected text: %q", oem.Text)
	}
}

func TestOpenEditor_DefaultsToNvim(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	e := newTestEditor()
	fe := e.exec.(*fakeExec)
	ff := e.cmdFactory.(*fakeFactory)

	e.textarea.SetValue("foo")
	msg := e.openEditor(e.textarea.Value())()
	if !fe.called {
		t.Fatalf("expected executor to be called")
	}
	if ff.name != "nvim" {
		t.Fatalf("expected default editor 'nvim', got %q", ff.name)
	}
	if _, ok := msg.(OpenEditorMsg); !ok {
		t.Fatalf("expected OpenEditorMsg, got %T", msg)
	}
}

func TestOpenEditor_CommandWithArgsFromEDITOR(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "emacs -nw")

	e := newTestEditor()
	fe := e.exec.(*fakeExec)
	ff := e.cmdFactory.(*fakeFactory)

	e.textarea.SetValue("bar")
	_ = e.openEditor(e.textarea.Value())()
	if !fe.called {
		t.Fatalf("expected executor to be called")
	}
	if ff.name != "emacs" {
		t.Fatalf("expected 'emacs', got %q", ff.name)
	}
	if len(ff.args) < 2 || ff.args[0] != "-nw" {
		t.Fatalf("expected first arg '-nw', got %v", ff.args)
	}
}

func TestOpenEditor_ExecErrorReported(t *testing.T) {
	t.Setenv("VISUAL", "nvim")

	badExec := func(_ *exec.Cmd, done func(error) tea.Msg) tea.Cmd {
		return func() tea.Msg { return done(os.ErrPermission) }
	}
	// build editor with bad executor
	e := New(Dependencies{Exec: ProcessExecutorFunc(badExec), Command: &fakeFactory{}}).(*editorCmp)
	e.textarea.SetValue("x")
	msg := e.openEditor(e.textarea.Value())()
	// util.ReportError wraps error in tea.Msg; we just ensure not OpenEditorMsg
	if _, ok := msg.(OpenEditorMsg); ok {
		t.Fatalf("expected an error message, got OpenEditorMsg")
	}
}

func TestOpenEditor_QuotedPathWithSpace(t *testing.T) {
	t.Setenv("VISUAL", "\"/opt/My Editor/editor\" -w")
	t.Setenv("EDITOR", "")

	e := newTestEditor()
	fe := e.exec.(*fakeExec)
	ff := e.cmdFactory.(*fakeFactory)

	e.textarea.SetValue("abc")
	_ = e.openEditor(e.textarea.Value())()

	if !fe.called {
		t.Fatalf("expected executor to be called")
	}
	if ff.name != "/opt/My Editor/editor" {
		t.Fatalf("expected quoted path preserved, got %q", ff.name)
	}
	if len(ff.args) < 1 || ff.args[0] != "-w" {
		t.Fatalf("expected first arg '-w', got %v", ff.args)
	}
	if len(ff.args) < 2 {
		t.Fatalf("expected tmp file appended, got args=%v", ff.args)
	}
}

func TestOpenEditor_EmptyMessageWarn(t *testing.T) {
	t.Setenv("VISUAL", "nvim")
	e := newTestEditor()

	// whitespace only becomes empty after TrimSpace
	e.textarea.SetValue("   \n  ")
	m := e.openEditor(e.textarea.Value())()

	// Should not return OpenEditorMsg; may return a tea.Cmd that yields InfoMsg
	if _, ok := m.(OpenEditorMsg); ok {
		t.Fatalf("expected warn message, got OpenEditorMsg")
	}
	var im util.InfoMsg
	switch v := m.(type) {
	case util.InfoMsg:
		im = v
	case tea.Cmd:
		mm := v()
		var ok bool
		im, ok = mm.(util.InfoMsg)
		if !ok {
			t.Fatalf("expected util.InfoMsg from tea.Cmd, got %#v", mm)
		}
	default:
		t.Fatalf("unexpected message type %T", m)
	}
	if im.Type != util.InfoTypeWarn {
		t.Fatalf("expected warn type, got %#v", im)
	}
	if strings.ToLower(im.Msg) != "message is empty" {
		t.Fatalf("expected 'Message is empty' warn, got %q", im.Msg)
	}
}

func TestOpenEditor_SingleQuotedPath(t *testing.T) {
	t.Setenv("VISUAL", "'/opt/My Editor/editor' --wait")
	t.Setenv("EDITOR", "")

	e := newTestEditor()
	fe := e.exec.(*fakeExec)
	ff := e.cmdFactory.(*fakeFactory)

	e.textarea.SetValue("abc")
	_ = e.openEditor(e.textarea.Value())()

	if !fe.called {
		t.Fatalf("expected executor to be called")
	}
	if ff.name != "/opt/My Editor/editor" {
		t.Fatalf("expected single-quoted path preserved, got %q", ff.name)
	}
	if len(ff.args) < 1 || ff.args[0] != "--wait" {
		t.Fatalf("expected first arg '--wait', got %v", ff.args)
	}
}

func TestOpenEditor_EscapedSpacePath(t *testing.T) {
	t.Setenv("VISUAL", "/opt/My\\ Editor/editor --wait")
	t.Setenv("EDITOR", "")

	e := newTestEditor()
	fe := e.exec.(*fakeExec)
	ff := e.cmdFactory.(*fakeFactory)

	e.textarea.SetValue("abc")
	_ = e.openEditor(e.textarea.Value())()

	if !fe.called {
		t.Fatalf("expected executor to be called")
	}
	if ff.name != "/opt/My Editor/editor" {
		t.Fatalf("expected escaped space path preserved, got %q", ff.name)
	}
	if len(ff.args) < 1 || ff.args[0] != "--wait" {
		t.Fatalf("expected first arg '--wait', got %v", ff.args)
	}
}

func TestOpenEditor_NoEditorConfigured(t *testing.T) {
	// Whitespace-only should result in no fields
	t.Setenv("VISUAL", "   ")
	t.Setenv("EDITOR", "")

	e := newTestEditor()
	e.textarea.SetValue("abc")
	m := e.openEditor(e.textarea.Value())()
	// Expect an error message (util.InfoMsg error) from util.ReportError
	if _, ok := m.(OpenEditorMsg); ok {
		t.Fatalf("expected error message, got OpenEditorMsg")
	}
}
