package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/system"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/screens"
)

func focusMainPreview(t *testing.T, m Model) (Model, int) {
	t.Helper()
	if m.shellExperience == nil {
		t.Fatal("configuration is not open")
	}
	options := screens.ShellExperienceOptions(m.shellExperience.profile)
	for index, option := range options {
		if option.Control == screens.ShellExperienceSDK && option.Choice == "main-preview" {
			m.shellExperience.cursor = index
			return m, index
		}
	}
	t.Fatal("Main preview choice is missing")
	return m, -1
}

func openAgreedConfiguration(t *testing.T) Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	m := NewModel(system.DetectionResult{}, "main-preview-test").WithShellInstaller()
	m.Width, m.Height = 209, 61
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.shellExperience == nil {
		t.Fatal("Welcome did not open the configuration")
	}
	return m
}

// The channel radio is a channel, not a preference consumer: explicit ON/OFF
// values must not lock it. Unsupported pairings are named when continuing.
func TestShellMainChannelIsSelectable(t *testing.T) {
	m := openAgreedConfiguration(t)
	m, _ = focusMainPreview(t, m)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil {
		t.Fatal("selecting Main preview started a worker")
	}
	if m.shellExperience.profile.SDKSource != "main-preview" {
		t.Fatalf("explicit OFF/ON preferences blocked the channel: %+v", m.shellExperience.profile)
	}

	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = next.(Model)
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil || m.shellExperience != nil {
		t.Fatalf("Main lost the stock preference parity: %+v", m.shellExperience)
	}
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || state.Request.Experience == nil || state.Request.Experience.SDKSource != "main-preview" || state.Request.Experience.CodeGraph != "on" {
		t.Fatalf("Main request dropped a qualified preference: %+v", state.Request.Experience)
	}
	if state.Request.Confirmation != "" || m.ShellInstallSelection().Confirmed {
		t.Fatal("Main channel reused or forged consent")
	}
}

func TestShellMainAcceptsQualifiedSelections(t *testing.T) {
	m := openAgreedConfiguration(t)
	m, _ = focusMainPreview(t, m)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil || m.shellExperience.profile.SDKSource != "main-preview" {
		t.Fatalf("qualified Main channel was refused: %+v", m.shellExperience.profile)
	}
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = next.(Model)
	next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil {
		t.Fatal("Continue started a worker")
	}
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || m.shellExperience != nil {
		t.Fatalf("qualified Main configuration did not continue: %v", state)
	}
	want := shellinstaller.UserExperience{
		Persona: "gentleman", BackgroundSubagents: "on", StrictTDD: "on",
		GlobalRDD: "on", Context7: "on", CodeGraph: "on", GGA: "off",
		SDKSource: "main-preview",
	}
	if state.Request.Experience == nil || *state.Request.Experience != want {
		t.Fatalf("Main request dropped selections: %+v", state.Request.Experience)
	}
	if state.Request.Confirmation != "" || m.ShellInstallSelection().Confirmed {
		t.Fatal("Main channel reused or forged consent")
	}
}
