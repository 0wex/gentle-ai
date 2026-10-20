package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"os"
)

// WithShellInstaller opts into selection within this existing TUI program.
func (m Model) WithShellInstaller() Model {
	m.shellInstallEnabled = true
	return m
}

// ShellInstallSelection is consumed by the host after its program exits.
func (m Model) ShellInstallSelection() cli.ShellInstallSelection {
	return m.shellInstallSelection
}

func (m Model) updateShellInstall(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.shellPlan != nil {
		key, ok := msg.(tea.KeyMsg)
		if !ok {
			return m, nil
		}
		switch key.Type {
		case tea.KeyEnter:
			_ = os.Setenv("GENTLE_AI_SHELL_STEPS", "1")
			return m.finishShellInstallSelection(cli.ShellInstallSelection{Request: m.shellPlanRequest, Confirmed: true}, nil)
		case tea.KeyEsc:
			m.shellExperience, m.shellPlanConfig = m.shellPlanConfig, nil
			m.shellPlan, m.shellPlanRequest = nil, shellinstaller.UserInstallRequest{}
			return m, nil
		case tea.KeyCtrlC:
			m.shellPlan, m.shellPlanConfig, m.shellExperience = nil, nil, nil
			m.shellPlanRequest, m.shellInstallModel = shellinstaller.UserInstallRequest{}, nil
			_ = os.Unsetenv("GENTLE_AI_RECONFIGURE_TARGET")
			m.setScreen(ScreenWelcome)
			return m, nil
		}
		return m, nil
	}
	if m.shellExperience != nil {
		return m.configureShellExperience(msg)
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		if next, handled := m.shellExperienceNavigation(key); handled {
			return next, nil
		}
	}
	before, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if key, ok := msg.(tea.KeyMsg); ok && before.Review && key.Type == tea.KeyEnter {
		switch m.shellReadyCursor {
		case 0:
			// Reuse the existing physical-confirmation path, never a worker.
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
		case 1:
			m.shellInstallModel = cli.NewShellInstallModelWithRequest(nil, before.Request)
			m.shellInstallModel, _ = m.shellInstallModel.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height})
			m.shellInstallSelection = cli.ShellInstallSelection{}
			m.shellReadyCursor, m.shellReadyOffset = 0, 0
			return m, nil
		case 2:
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		}
	}
	if key, ok := msg.(tea.KeyMsg); ok {
		if state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel); ok && !state.Review {
			if state.Field == 1 && (key.Type == tea.KeyLeft || key.Type == tea.KeyRight) {
				return m, nil // Install target is chosen in configuration.
			}
			if state.Field >= 2 && (key.Type == tea.KeyRunes || key.Type == tea.KeyBackspace) {
				return m, nil // Detected Shared paths are read-only.
			}
			if m.shellReconfigure != "" && state.Field == 0 && (key.Type == tea.KeyRunes || key.Type == tea.KeyBackspace) {
				return m, nil // The existing install target is fixed.
			}
		}
	}
	m.shellInstallModel, _ = m.shellInstallModel.Update(msg)
	after, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if before.Review != after.Review {
		m.shellReadyCursor, m.shellReadyOffset = 0, 0
	}
	if key, ok := msg.(tea.KeyMsg); ok && (key.Type == tea.KeyEsc || key.Type == tea.KeyCtrlC) {
		m.shellReconfigure = ""
		_ = os.Unsetenv("GENTLE_AI_RECONFIGURE_TARGET")
		m.shellInstallSelection = cli.ShellInstallSelection{}
		m.shellInstallModel = nil
		m.shellReadyCursor, m.shellReadyOffset = 0, 0
		m.setScreen(ScreenWelcome)
		return m, nil
	}
	selection, err := cli.ShellInstallOutcome(m.shellInstallModel)
	return m.finishShellInstallSelection(selection, err)
}

func (m Model) finishShellInstallSelection(selection cli.ShellInstallSelection, err error) (tea.Model, tea.Cmd) {
	if err == nil && selection.Confirmed {
		m.shellInstallSelection = selection
		if home, homeErr := os.UserHomeDir(); homeErr == nil {
			_ = os.Setenv("GENTLE_AI_SHELL_REGISTRY_HOME", home)
		}
		if m.shellReconfigure != "" {
			_ = os.Setenv("GENTLE_AI_RECONFIGURE_TARGET", m.shellReconfigure)
		} else {
			_ = os.Unsetenv("GENTLE_AI_RECONFIGURE_TARGET")
		}
		return m, tea.Quit
	}
	// This child performs synchronous selection only. Its sole command is
	// Quit, handled above; never run an installer from inside the parent TUI.
	return m, nil
}
