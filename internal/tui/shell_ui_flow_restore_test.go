package tui

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/screens"
)

func shellFlowDestination(t *testing.T) Model {
	t.Helper()
	m := experienceInherited(t)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	return experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
}

func shellFlowReady(t *testing.T) Model {
	t.Helper()
	m := shellFlowDestination(t)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(shellExperienceTarget(t))})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok || !state.Review || state.Request.Confirmation == "" {
		t.Fatalf("private fixture did not reach inspected review: %+v", state)
	}
	return m
}

func shellFlowContains(t *testing.T, m Model, parts ...string) {
	t.Helper()
	view := ansi.Strip(m.View())
	for _, part := range parts {
		if !strings.Contains(view, part) {
			t.Fatalf("historical flow omits %q:\n%s", part, view)
		}
	}
}

func TestShellFlowDestinationContext(t *testing.T) {
	m := shellFlowDestination(t)
	shellFlowContains(t, m, "Choose installation destination", "About: Target", "│")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyTab})
	shellFlowContains(t, m, "About: Mode")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRight})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyTab})
	shellFlowContains(t, m, "About: Shared prefix")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyTab})
	shellFlowContains(t, m, "About: Shared agent")
	if m.ShellInstallSelection().Confirmed {
		t.Fatal("field navigation acquired installation consent")
	}
}

func TestShellFlowReviewActionsAreInert(t *testing.T) {
	m := shellFlowReady(t)
	before, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	for _, action := range []string{"Install Gentle-Shell", "Back", "Cancel"} {
		shellFlowContains(t, m, "Plan", ">"+action, "About: "+action)
		after, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
		if !after.Review || !reflect.DeepEqual(after.Request, before.Request) || m.ShellInstallSelection().Confirmed {
			t.Fatal("action focus changed the inspected request or granted consent")
		}
		m = experienceKey(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyUp})
	shellFlowContains(t, m, ">Back", "About: Back")
}

func TestShellFlowReviewPosition(t *testing.T) {
	m := shellFlowReady(t)
	indicator := regexp.MustCompile(`\[(\d+)/(\d+)\]`)
	position := indicator.FindStringSubmatch(ansi.Strip(m.View()))
	if len(position) != 3 || position[1] != "0" {
		t.Fatal("historical review line-position indicator missing", m.View())
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	position = indicator.FindStringSubmatch(ansi.Strip(m.View()))
	if len(position) != 3 || position[1] != position[2] {
		t.Fatal("End did not reach the last review line offset", m.View())
	}
}

func TestShellFlowInstallEnterOnlySelectsHandoff(t *testing.T) {
	m := shellFlowReady(t)
	before, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(Model)
	if cmd == nil || !got.ShellInstallSelection().Confirmed || len(cli.ShellInstallArguments(got.ShellInstallSelection())) == 0 {
		t.Fatal("Install action did not use the existing physical selection path")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("Install action must only quit for the host's restored-terminal handoff")
	}
	if _, err := os.Lstat(before.Request.Destination); !os.IsNotExist(err) {
		t.Fatal("selection ran an installer or created a destination", err)
	}
}

func TestShellFlowBackPreservesRequestAndClearsConsent(t *testing.T) {
	m := shellFlowReady(t)
	before, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	after, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	want := before.Request
	want.Confirmation = ""
	if cmd != nil || !ok || after.Review || !reflect.DeepEqual(after.Request, want) || m.ShellInstallSelection().Confirmed {
		t.Fatalf("Back lost the complete request or retained inspection consent: %+v", after)
	}
	shellFlowContains(t, m, "About: Target")
	if _, err := os.Lstat(want.Destination); !os.IsNotExist(err) {
		t.Fatal("Back created a destination", err)
	}
}

func TestShellFlowCancelReturnsWelcome(t *testing.T) {
	m := shellFlowReady(t)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if cmd != nil || m.Screen != ScreenWelcome || m.ShellInstallSelection().Confirmed || len(cli.ShellInstallArguments(m.ShellInstallSelection())) != 0 {
		t.Fatal("Cancel retained a selection or failed to return to Welcome")
	}
}

func TestShellFlowConfigureReentryResetsReview(t *testing.T) {
	m := shellFlowReady(t)
	before, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyCtrlE})
	editing, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if editing.Review || editing.Request.Confirmation != "" || m.ShellInstallSelection().Confirmed {
		t.Fatal("Configure retained physical consent")
	}
	m = experienceFocus(t, m, screens.ShellExperiencePersona, "neutral")
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnd})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	after, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !after.Review || after.Request.Destination != before.Request.Destination || after.Request.Confirmation == "" || after.Request.Confirmation == before.Request.Confirmation {
		t.Fatal("reentry did not freshly inspect the preserved destination and edited profile")
	}
	shellFlowContains(t, m, ">Install Gentle-Shell", "About: Install Gentle-Shell", "[0/")
}

func TestShellFlowLegacyConsentKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyRunes, Runes: []rune{'y'}}, {Type: tea.KeyRunes, Runes: []rune{'n'}}, {Type: tea.KeyEsc}, {Type: tea.KeyCtrlC}} {
		t.Run(strings.ReplaceAll(key.String(), "+", "-"), func(t *testing.T) {
			m := shellFlowReady(t)
			next, cmd := m.Update(key)
			got := next.(Model)
			if key.String() == "y" {
				if cmd == nil || !got.ShellInstallSelection().Confirmed {
					t.Fatal("legacy physical y confirmation changed")
				}
				return
			}
			if cmd != nil || got.ShellInstallSelection().Confirmed {
				t.Fatal("legacy edit/cancel key acquired consent")
			}
			if key.String() == "n" {
				state, _ := cli.ShellInstallPresentationOf(got.shellInstallModel)
				if state.Review {
					t.Fatal("legacy n did not return to editing")
				}
			} else if got.Screen != ScreenWelcome {
				t.Fatal("legacy cancellation did not return to Welcome")
			}
		})
	}
}
