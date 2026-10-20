package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/system"
)

func openInstallationPlan(t *testing.T) (Model, shellinstaller.UserInstallRequest) {
	t.Helper()
	m := NewModel(system.DetectionResult{}, "plan-test").WithShellInstaller()
	m.Width, m.Height = 120, 30
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	m, request := experiencePlan(t, m)
	return m, request
}

func TestShellPlanListsStepsAndDefaultTarget(t *testing.T) {
	m, request := openInstallationPlan(t)
	view := m.View()
	for _, part := range []string{"Installation plan", "1. Validate the selection and target", "2. Prepare the private stage and preferences", "3. Install the runtime and pinned suppliers", "4. Apply optional runtimes and the SDK channel", "5. Publish, verify and register the installation", "Enter: install"} {
		if !strings.Contains(view, part) {
			t.Fatalf("installation plan omits %q:\n%s", part, view)
		}
	}
	wantDestination := filepath.Join(os.Getenv("HOME"), ".local", "share", "gentle-shell", "install")
	if request.Destination != wantDestination || request.Mode != "separate" || request.Confirmation == "" {
		t.Fatalf("plan request differs: %+v", request)
	}
	base := filepath.Dir(request.Destination)
	if info, err := os.Lstat(base); err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("default parent is not private: %v %v", info, err)
	}
	if _, err := os.Lstat(request.Destination); !os.IsNotExist(err) {
		t.Fatal("plan created the destination before confirmation")
	}
}

func TestShellPlanEnterConfirmsHandoff(t *testing.T) {
	m, request := openInstallationPlan(t)
	if m.ShellInstallSelection().Confirmed {
		t.Fatal("plan inspection forged consent")
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	selection := m.ShellInstallSelection()
	if cmd == nil || !selection.Confirmed || selection.Request.Confirmation != request.Confirmation {
		t.Fatalf("plan enter did not confirm the inspected handoff: %+v", selection)
	}
	args := cli.ShellInstallArguments(selection)
	if len(args) != 8 || args[6] != "--experience" {
		t.Fatalf("plan handoff changed the worker ABI: %v", args)
	}
	if decoded, err := shellinstaller.UserInstallFromEntry(args[1:]); err != nil || decoded.Destination != request.Destination || decoded.Mode != request.Mode {
		t.Fatalf("plan handoff lost the request: %+v %v", decoded, err)
	}
}

func TestShellPlanEscapeAndCancel(t *testing.T) {
	m, _ := openInstallationPlan(t)
	back := experienceKey(m, tea.KeyMsg{Type: tea.KeyEsc})
	if back.shellExperience == nil || back.shellPlan != nil || back.shellPlanRequest.Confirmation != "" {
		t.Fatalf("plan escape did not return to Configure: %+v", back.shellExperience)
	}
	if back.ShellInstallSelection().Confirmed {
		t.Fatal("plan escape retained consent")
	}
	cancel := experienceKey(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cancel.Screen != ScreenWelcome || len(cli.ShellInstallArguments(cancel.ShellInstallSelection())) != 0 {
		t.Fatal("plan ctrl-c did not cancel to Welcome")
	}
}
