package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// GGA ON is installed by the existing native installer path, which owns its
// acquisition and prints its own setup instructions; the TUI only carries the
// explicit selection and provider and must let Continue reach Destination.
func TestShellGGAONReachesDestination(t *testing.T) {
	for _, sdk := range []string{"", "main-preview"} {
		t.Run("sdk-"+sdk, func(t *testing.T) {
			m := openAgreedConfiguration(t)
			profile := m.shellExperience.profile
			profile.SDKSource, profile.GGA = sdk, "on"
			m.shellExperience.profile = profile
			next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
			m = next.(Model)
			next, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			m = next.(Model)
			if cmd != nil || m.shellPlan == nil {
				t.Fatalf("GGA ON did not reach the installation plan: %+v", m.shellExperience)
			}
			request := m.shellPlanRequest
			if request.Experience == nil {
				t.Fatal("GGA ON dropped the request")
			}
			got := *request.Experience
			if got.GGA != "on" || got.GGAProvider != "" || got.SDKSource != sdk || got.GlobalRDD != "on" || got.CodeGraph != "on" {
				t.Fatalf("GGA selection or stock parity was lost: %+v", got)
			}
			if request.Confirmation == "" || m.ShellInstallSelection().Confirmed {
				t.Fatal("GGA plan lacks inspection or forged consent")
			}
		})
	}
}
