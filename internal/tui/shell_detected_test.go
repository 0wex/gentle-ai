package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/system"
)

func managedInstallFixture(t *testing.T) (string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	target := filepath.Join(home, ".local", "share", "gentle-shell", "install")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"Schema": "gentle-shell-user-install/v1", "Destination": target,
		"Prefix": filepath.Join(target, "prefix"), "Agent": filepath.Join(target, "agent"),
		"Mode": "separate", "SupervisorSHA": "sha", "NodeSHA": "sha",
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "installation.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	return home, target
}

func TestShellDetectsManagedInstallAndManage(t *testing.T) {
	_, target := managedInstallFixture(t)
	got := shellinstaller.DiscoverUserShellInstalls(os.Getenv("HOME"))
	if len(got) != 1 || got[0].Target != target {
		t.Fatalf("default managed install not discovered: %+v", got)
	}
	if profile, err := shellinstaller.LoadUserShellExperience(got[0]); err != nil {
		t.Fatalf("installed profile did not load: %v", err)
	} else {
		t.Logf("loaded profile: %+v", profile)
	}
	m := NewModel(system.DetectionResult{}, "detected").WithShellInstaller()
	m.Width, m.Height = 120, 35
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if !strings.Contains(m.View(), "Found a managed Gentle-Shell installation") || !strings.Contains(m.View(), "Manage this installation") || !strings.Contains(m.View(), target) {
		t.Fatalf("detection screen is incomplete:\n%s", m.View())
	}
	if m.ShellInstallSelection().Confirmed {
		t.Fatal("detection forged installation consent")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Manage.
	m = next.(Model)
	if m.shellExperience == nil || !m.shellExperience.reconfigure || m.shellExperience.target != target || m.shellExperience.choice != "existing" {
		t.Fatalf("manage did not prefill the discovered install: %+v", m.shellExperience)
	}
	if !strings.Contains(m.View(), "Configure Gentle-Shell experience") {
		t.Fatalf("manage did not open the customization screen:\n%s", m.View())
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.shellPlan == nil || m.shellPlanRequest.Destination != target {
		t.Fatalf("manage did not plan against the discovered target: %+v err=%v", m.shellPlanRequest, m.shellExperience.err)
	}
}

func TestShellDetectionBackReturnsToMenu(t *testing.T) {
	managedInstallFixture(t)
	m := NewModel(system.DetectionResult{}, "detected").WithShellInstaller()
	m.Width, m.Height = 120, 35
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.Screen != ScreenWelcome || len(m.View()) == 0 {
		t.Fatalf("back did not return to Welcome: screen=%v", m.Screen)
	}
}
