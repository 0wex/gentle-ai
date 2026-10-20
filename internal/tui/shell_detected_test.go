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

func addAgentInstall(t *testing.T, home string) string {
	t.Helper()
	target := filepath.Join(home, ".gentle-shell")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"Schema": "gentle-shell-user-install/v1", "Destination": target,
		"Prefix": filepath.Join(target, "prefix"), "Agent": filepath.Join(home, ".pi", "agent"),
		"Mode": "agent", "SupervisorSHA": "sha", "NodeSHA": "sha",
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "installation.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	return target
}

func TestShellMultipleInstallsRequireChoice(t *testing.T) {
	home, _ := managedInstallFixture(t)
	agent := addAgentInstall(t, home)
	installs := shellinstaller.DiscoverUserShellInstalls(home)
	if len(installs) != 2 {
		t.Fatalf("both installations must be discovered: %+v", installs)
	}
	m := NewModel(system.DetectionResult{}, "detected").WithShellInstaller()
	m.Width, m.Height = 120, 35
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	view := m.View()
	for _, want := range []string{"installations detected", "Separate install", "On my existing Pi", "~/.local/share/gentle-shell/install", "~/.gentle-shell"} {
		if !strings.Contains(view, want) {
			t.Fatalf("multi-install selector hides %q:\n%s", want, view)
		}
	}
	if m.ShellInstallSelection().Confirmed {
		t.Fatal("multi-install detection forged consent")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.shellExperience == nil || m.shellExperience.target != agent || m.shellExperience.mode != "agent" {
		t.Fatalf("selected install did not open Manage: %+v", m.shellExperience)
	}
	if view := m.View(); !strings.Contains(view, "(*) On my existing Pi  · locked") || !strings.Contains(view, "( ) Separate install (Gentle-Shell)  · locked") {
		t.Fatalf("agent target is not the locked current target:\n%s", view)
	}
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
	if !strings.Contains(m.View(), "Found a managed Gentle-Shell installation") || !strings.Contains(m.View(), "Manage this installation") || !strings.Contains(m.View(), "~/.local/share/gentle-shell/install") {
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
	view := m.View()
	if !strings.Contains(view, "Configure Gentle-Shell experience") || !strings.Contains(view, "(*) Separate install (Gentle-Shell)  · locked") || !strings.Contains(view, "( ) On my existing Pi  · locked") {
		t.Fatalf("manage did not lock the managed target:\n%s", view)
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
