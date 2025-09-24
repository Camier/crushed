package providerstatus

import (
	"context"
	"testing"
)

func TestBuildShellCommand_Posix(t *testing.T) {
	cmd := buildShellCommand(context.Background(), "echo hi")
	if cmd.Path != "bash" && cmd.Args[0] != "bash" { // Path may include full path in some envs
		t.Fatalf("expected bash, got path=%q args=%v", cmd.Path, cmd.Args)
	}
	// Args should include -lc and the command string
	foundLC := false
	foundCmd := false
	for _, a := range cmd.Args {
		if a == "-lc" {
			foundLC = true
		}
		if a == "echo hi" {
			foundCmd = true
		}
	}
	if !foundLC || !foundCmd {
		t.Fatalf("expected args to contain -lc and command, got %v", cmd.Args)
	}
}
