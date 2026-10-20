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

func existingInstallHome(t *testing.T) (home, target string) {
	t.Helper()
	t.Setenv("NPM_CONFIG_PREFIX", "")
	home = t.TempDir()
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	target = filepath.Join(home, ".gentle-shell")
	manifest := `{"Schema":"gentle-shell-user-install/v1","Destination":"` + target + `","Prefix":"` + target + `/prefix","Agent":"` + target + `/agent","Mode":"separate","codegraph":true,"main_sdk":{"URL":"x"}}`
	for path, data := range map[string]string{
		"installation.json":          manifest,
		"config/persona.json":        `{"mode":"neutral"}`,
		"config/strict-tdd.json":     `{"policy":"on"}`,
		"home/.gentle-ai/state.json": `{"rdd_mode":"off"}`,
		"agent/mcp.json":             `{"mcpServers":{"context7":{"enabled":true}}}`,
	} {
		full := filepath.Join(target, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return home, target
}

func TestShellConfigureExistingPrefillsAndKeepsTarget(t *testing.T) {
	_, target := existingInstallHome(t)
	defer os.Unsetenv("GENTLE_AI_RECONFIGURE_TARGET")
	m := NewModel(system.DetectionResult{}, "configure-existing-test").WithShellInstaller()
	m.Width, m.Height = 209, 61
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	m = experienceFocus(t, m, screens.ShellExperienceInstallTarget, "existing")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.shellExperience.reconfigure != true || m.shellExperience.target != target || m.shellExperience.mode != "separate" {
		t.Fatalf("existing install was not selected: %+v", m.shellExperience)
	}
	profile := m.shellExperience.profile
	if profile.Persona != "neutral" || profile.StrictTDD != "on" || profile.GlobalRDD != "off" || profile.CodeGraph != "on" || profile.Context7 != "on" || profile.SDKSource != "main-preview" {
		t.Fatalf("existing selections were not prefilled: %+v", profile)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || state.Request.Destination != target || state.Request.Mode != "separate" {
		t.Fatalf("reconfigure target was not fixed: %+v", state.Request)
	}
	if os.Getenv("GENTLE_AI_RECONFIGURE_TARGET") != target {
		t.Fatal("reconfigure intent was not published for inspection")
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	state, _ = cli.ShellInstallPresentationOf(m.shellInstallModel)
	if state.Request.Destination != target {
		t.Fatalf("existing target became editable: %q", state.Request.Destination)
	}
}
