package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
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
		}
	}
	m.shellInstallModel, _ = m.shellInstallModel.Update(msg)
	after, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if before.Review != after.Review {
		m.shellReadyCursor, m.shellReadyOffset = 0, 0
	}
	if key, ok := msg.(tea.KeyMsg); ok && (key.Type == tea.KeyEsc || key.Type == tea.KeyCtrlC) {
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
		return m, tea.Quit
	}
	// This child performs synchronous selection only. Its sole command is
	// Quit, handled above; never run an installer from inside the parent TUI.
	return m, nil
}
