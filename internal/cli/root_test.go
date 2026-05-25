package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpHidesInternalCommands(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}

	help := out.String()
	for _, hidden := range []string{"completion", "gen"} {
		if strings.Contains(help, hidden) {
			t.Fatalf("help should not include %q:\n%s", hidden, help)
		}
	}
	if !strings.Contains(help, "upgrade") {
		t.Fatalf("help should include upgrade:\n%s", help)
	}
}

func TestCompletionCommandIsHiddenButRegistered(t *testing.T) {
	cmd := NewRootCommand()
	completion, _, err := cmd.Find([]string{"completion"})
	if err != nil {
		t.Fatalf("find completion: %v", err)
	}
	if completion == nil || completion.Name() != "completion" {
		t.Fatalf("expected completion command, got %#v", completion)
	}
	if !completion.Hidden {
		t.Fatalf("completion command should be hidden")
	}
}
