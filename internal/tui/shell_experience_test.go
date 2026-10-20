package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/system"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/screens"
)

func experienceKey(m Model, key tea.KeyMsg) Model {
	model, _ := m.Update(key)
	return model.(Model)
}

func experienceFocus(t *testing.T, m Model, control screens.ShellExperienceControl, choice string) Model {
	t.Helper()
	if m.shellExperience == nil {
		t.Fatal("configuration is not active")
	}
	for _, entry := range screens.ShellExperienceOptions(m.shellExperience.profile) {
		if entry.Control == control && (choice == "" || entry.Choice == choice) {
			m = experienceKey(m, tea.KeyMsg{Type: tea.KeyHome})
			for i := 0; i < entry.Cursor; i++ {
				m = experienceKey(m, tea.KeyMsg{Type: tea.KeyDown})
			}
			return m
		}
	}
	t.Fatal("supported configuration action is missing", control, choice)
	return m
}

func experienceStart(t *testing.T) Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	m := NewModel(system.DetectionResult{}, "dev").WithShellInstaller()
	m.Width, m.Height = 120, 30
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if !strings.Contains(m.View(), "Configure Gentle-Shell experience") {
		t.Fatal("ordinary installer skips Configure")
	}
	return m
}

// Legacy requests remain independently testable without rewriting fresh defaults.
func experienceInherited(t *testing.T) Model {
	t.Helper()
	m := experienceStart(t)
	m.shellExperience.profile = shellinstaller.DefaultUserExperience()
	return m
}

func TestShellExperienceConfigureDestinationReady(t *testing.T) {
	m := experienceFocus(t, experienceInherited(t), screens.ShellExperiencePersona, "neutral")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = experienceFocus(t, m, screens.ShellExperiencePersona, "unmanaged")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = experienceFocus(t, m, screens.ShellExperienceBackground, "")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter}) // background off
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || state.Request.Experience == nil || state.Request.Experience.Persona != "unmanaged" || state.Request.Experience.BackgroundSubagents != "off" {
		t.Fatal("Configure lost profile at Destination")
	}
	target := shellExperienceTarget(t)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(target)})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	state, _ = cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !state.Review || state.Request.Confirmation == "" {
		t.Fatalf("working profile cannot reach Ready: %+v", state)
	}
	for _, text := range []string{"Ready", "Selected experience", "Changes + scopes", "unmanaged", "gentle-shell"} {
		if !strings.Contains(m.View(), text) {
			t.Fatalf("Ready hides %q", text)
		}
	}
	for _, size := range [][2]int{{120, 30}, {80, 14}, {30, 7}, {5, 1}} {
		m.Width, m.Height = size[0], size[1]
		for _, key := range []tea.KeyType{tea.KeyHome, tea.KeyEnd} {
			m = experienceKey(m, tea.KeyMsg{Type: key})
			lines := strings.Split(m.View(), "\n")
			if len(lines) > m.Height {
				t.Fatal("Ready exceeds terminal height")
			}
			for _, line := range lines {
				if ansi.StringWidth(line) > m.Width {
					t.Fatal("Ready exceeds terminal width")
				}
			}
		}
	}
	m.Width, m.Height = 120, 30
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter}) // gentleman
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	state, _ = cli.ShellInstallPresentationOf(m.shellInstallModel)
	if state.Review || state.Request.Confirmation != "" || state.Request.Destination != target || state.Request.Experience.Persona != "gentleman" {
		t.Fatal("editing configuration retains old consent or loses destination")
	}
}

func TestShellExperiencePreservesUnsupportedChoicesAndCancel(t *testing.T) {
	m := experienceFocus(t, experienceStart(t), screens.ShellExperienceGGA, "")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter}) // OFF → inherit.
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter}) // Inherit → ON is still unqualified.
	m = experienceFocus(t, m, screens.ShellExperienceProvider, "")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("codex")})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.shellExperience == nil || m.shellExperience.err == nil || !strings.Contains(m.shellExperience.err.Error(), "GGA ON") {
		t.Fatal("unsupported choice was dropped or admitted")
	}
	if m.shellExperience.profile.GGA != "on" || m.shellExperience.profile.GGAProvider != "codex" {
		t.Fatal("unsupported choice was cleared")
	}
	state, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if state.Review || state.Request.Confirmation != "" {
		t.Fatal("GGA refusal forged consent or advanced")
	}
	for _, key := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		cancel := experienceStart(t)
		cancel = experienceKey(cancel, tea.KeyMsg{Type: key})
		if cancel.Screen != ScreenWelcome || cancel.ShellInstallSelection().Confirmed {
			t.Fatal("Configure cancel installs or leaves Welcome")
		}
	}
}
