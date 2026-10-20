package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/system"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/screens"
)

func detectedPiHome(t *testing.T) {
	t.Helper()
	t.Setenv("NPM_CONFIG_PREFIX", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".pi", "agent"), 0700); err != nil {
		t.Fatal(err)
	}
	bundle := filepath.Join(home, ".npm-global/lib/node_modules/@earendil-works/pi-coding-agent/dist/bundle")
	if err := os.MkdirAll(bundle, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "cli.js"), []byte("stock cli"), 0600); err != nil {
		t.Fatal(err)
	}
}

func openTargetConfiguration(t *testing.T) Model {
	t.Helper()
	m := NewModel(system.DetectionResult{}, "install-target-test").WithShellInstaller()
	m.Width, m.Height = 209, 61
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return next.(Model)
}

func TestShellInstallTargetSelectsDetectedPi(t *testing.T) {
	detectedPiHome(t)
	m := openTargetConfiguration(t)
	m = experienceFocus(t, m, screens.ShellExperienceInstallTarget, "shared")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil || m.shellExperience == nil || m.shellExperience.mode != "shared" {
		t.Fatalf("detected Pi was not selectable: %+v", m.shellExperience)
	}
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = next.(Model)
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil || m.shellExperience != nil {
		t.Fatalf("Shared defaults did not reach the destination: %v", m.shellExperience)
	}
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || state.Request.Mode != "shared" || state.Request.SharedPrefix == "" || state.Request.SharedAgent == "" {
		t.Fatalf("detected Shared paths were not carried: %+v", state.Request)
	}
}

func TestShellInstallTargetRefusesSharedWithoutPi(t *testing.T) {
	t.Setenv("NPM_CONFIG_PREFIX", "")
	t.Setenv("HOME", t.TempDir())
	m := openTargetConfiguration(t)
	m = experienceFocus(t, m, screens.ShellExperienceInstallTarget, "shared")
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.shellExperience.mode != "separate" || m.shellExperience.err == nil {
		t.Fatalf("Shared was accepted without a detected Pi: %+v", m.shellExperience)
	}
}

func TestShellInstallTargetRefusesSharedMain(t *testing.T) {
	detectedPiHome(t)
	m := openTargetConfiguration(t)
	m = experienceFocus(t, m, screens.ShellExperienceInstallTarget, "shared")
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	profile := m.shellExperience.profile
	profile.SDKSource = "main-preview"
	m.shellExperience.profile = profile
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.shellExperience == nil || m.shellExperience.err == nil || !strings.Contains(m.shellExperience.err.Error(), "Separate") {
		t.Fatalf("Shared Main preview was not refused precisely: %v", m.shellExperience.err)
	}
}
