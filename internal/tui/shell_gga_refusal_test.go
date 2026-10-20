package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// GGA ON has no qualified consumer in this installer route. Both channels must
// say so while continuing instead of advancing to Destination and failing there.
func TestShellGGAONRefusedAtContinue(t *testing.T) {
	for _, sdk := range []string{"", "main-preview"} {
		t.Run("sdk-"+sdk, func(t *testing.T) {
			m := openAgreedConfiguration(t)
			profile := m.shellExperience.profile
			profile.SDKSource, profile.GGA, profile.GGAProvider, profile.CodeGraph = sdk, "on", "codex", "off"
			m.shellExperience.profile = profile
			next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
			m = next.(Model)
			next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			m = next.(Model)
			if cmd != nil {
				t.Fatal("GGA refusal started a worker")
			}
			if m.shellExperience == nil {
				t.Fatal("GGA ON advanced to Destination without a qualified consumer")
			}
			if err := m.shellExperience.err; err == nil || !strings.Contains(err.Error(), "GGA ON") {
				t.Fatalf("GGA refusal was not named precisely: %v", err)
			}
			if m.shellExperience.profile.GGA != "on" || m.shellExperience.profile.SDKSource != sdk {
				t.Fatal("GGA refusal silently changed the explicit choice")
			}
		})
	}
}
