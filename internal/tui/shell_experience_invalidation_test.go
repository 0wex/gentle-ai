package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/screens"
)

func shellExperienceTarget(t *testing.T) string {
	t.Helper()
	parent, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(parent, "target")
}

func TestShellExperienceEditingImmediatelyInvalidatesInspection(t *testing.T) {
	m := experienceInherited(t)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	target := shellExperienceTarget(t)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(target)})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	before, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || !before.Review || before.Request.Confirmation == "" || before.Request.Experience == nil {
		t.Fatalf("fixture did not reach physical review: %+v", before)
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	after, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || m.shellExperience == nil || after.Review || after.Request.Confirmation != "" || m.ShellInstallSelection().Confirmed {
		t.Fatalf("entering Configure retained an inspected selection: %+v", after)
	}
	if after.Request.Destination != target || after.Request.Mode != before.Request.Mode || after.Request.Experience == nil || *after.Request.Experience != *before.Request.Experience {
		t.Fatal("editing discarded the selected destination, mode or complete profile")
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatal("entering Configure created a destination", err)
	}
}

func TestShellExperienceEditingDoesNotEraseExplicitProvider(t *testing.T) {
	m := experienceStart(t)
	// Exercise inherited-toggle editing independently of the fresh ON/OFF preset.
	m.shellExperience.profile.GGA = ""
	m.shellExperience.profile.StrictTDD = ""
	focus := func(control screens.ShellExperienceControl) {
		m = experienceFocus(t, m, control, "")
	}
	focus(screens.ShellExperienceGGA)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter}) // Explicit ON reveals provider input.
	focus(screens.ShellExperienceProvider)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("codex")})
	focus(screens.ShellExperienceTDD)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.shellExperience == nil || m.shellExperience.profile.GGAProvider != "codex" || m.shellExperience.profile.StrictTDD != "on" {
		t.Fatal("editing another preference silently cleared the explicit provider")
	}
	focus(screens.ShellExperienceGGA)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter}) // ON → OFF retains the explicit value.
	if m.shellExperience.profile.GGA != "off" || m.shellExperience.profile.GGAProvider != "codex" {
		t.Fatal("editing GGA silently discarded the explicitly typed provider")
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.shellExperience == nil || m.shellExperience.err == nil || m.shellExperience.profile.GGAProvider != "codex" {
		t.Fatal("incompatible explicit choices were advanced or silently replaced")
	}
	focus(screens.ShellExperienceProvider) // Retained invalid data must remain correctable.
	for range "codex" {
		m = experienceKey(m, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	if m.shellExperience.profile.GGAProvider != "" || m.shellExperience.profile.GGA != "off" || m.shellExperience.profile.Validate() != nil || m.ShellInstallSelection().Confirmed {
		t.Fatal("clearing the retained provider lost GGA choice or acquired consent")
	}
	for _, entry := range screens.ShellExperienceOptions(m.shellExperience.profile) {
		if entry.Control == screens.ShellExperienceProvider {
			t.Fatal("cleared inactive provider row did not disappear")
		}
	}
}

func TestShellExperienceNonPrivateParentRefusesBeforeConsent(t *testing.T) {
	m := experienceStart(t)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	target := shellExperienceTarget(t)
	parent := filepath.Dir(target)
	if err := os.Chmod(parent, 0755); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(parent)
	if err != nil {
		t.Fatal(err)
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(target)})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || state.Error == nil || state.Review || state.Request.Confirmation != "" || m.ShellInstallSelection().Confirmed {
		t.Fatalf("nonprivate destination parent acquired consent: %+v", state)
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatal("refusal created a destination", err)
	}
	after, err := os.Stat(parent)
	if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("refusal changed the destination parent", err)
	}
}
