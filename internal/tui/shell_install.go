package tui

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/screens"
)

var shellInstallStepLabels = []string{
	"Validate the selection and target",
	"Prepare the private stage and preferences",
	"Install the runtime and pinned suppliers",
	"Apply optional runtimes and the SDK channel",
	"Publish, verify and register the installation",
}

type shellInstallStep struct {
	Label   string
	State   string
	Details []string
}

type shellInstallRun struct {
	steps   []shellInstallStep
	done    bool
	err     error
	spinner int
}

type shellInstallTick struct{}

type shellInstallEvent struct {
	progress shellinstaller.UserInstallProgress
	done     bool
	err      error
}

// WithShellInstallRunner lets the owning application inject its validated
// in-process worker. The runner receives the inspected request, never a new
// confirmation.
func (m Model) WithShellInstallRunner(runner func(context.Context, shellinstaller.UserInstallRequest, func(shellinstaller.UserInstallProgress)) error) Model {
	m.shellInstallRunner = runner
	return m
}

func waitShellInstallEvent(events <-chan shellInstallEvent) tea.Cmd {
	return func() tea.Msg {
		select {
		case event, ok := <-events:
			if !ok {
				return shellInstallEvent{done: true, err: errors.New("installation ended without a result")}
			}
			return event
		case <-time.After(120 * time.Millisecond):
			return shellInstallTick{}
		}
	}
}

func (m Model) beginShellInstall() (tea.Model, tea.Cmd) {
	request := m.shellPlanRequest
	m.shellInstallSelection = cli.ShellInstallSelection{Request: request, Confirmed: true, Executed: true}
	steps := make([]shellInstallStep, len(shellInstallStepLabels))
	for index, label := range shellInstallStepLabels {
		steps[index] = shellInstallStep{Label: label, State: "pending"}
	}
	steps[0].State = "running"
	events := make(chan shellInstallEvent, 32)
	m.shellInstallRun = &shellInstallRun{steps: steps}
	m.shellInstallEvents = events
	m.shellPlan, m.shellPlanRequest = nil, shellinstaller.UserInstallRequest{}
	runner := m.shellInstallRunner
	if runner == nil {
		runner = func(ctx context.Context, req shellinstaller.UserInstallRequest, report func(shellinstaller.UserInstallProgress)) error {
			return cli.RunShellInstallInline(ctx, cli.ShellInstallSelection{Request: req, Confirmed: true, Executed: true}, report, io.Discard)
		}
	}
	go func() {
		// The inspected request is immutable here; the runner revalidates its
		// token and physical selection before any effect.
		err := runner(context.Background(), request, func(progress shellinstaller.UserInstallProgress) {
			events <- shellInstallEvent{progress: progress}
		})
		events <- shellInstallEvent{done: true, err: err}
		close(events)
	}()
	return m, waitShellInstallEvent(events)
}

type shellManageState struct {
	install     shellinstaller.UserShellInstall
	config      shellExperienceConfig
	preferences []screens.ShellManageChange
	components  []screens.ShellManageChange
	cursor      int
	applied     bool
	err         error
}

func manageDisplay(value, fallback string) string {
	if value == "" {
		return fallback
	}
	switch value {
	case "on":
		return "On"
	case "off":
		return "Off"
	case "gentleman":
		return "Gentleman"
	case "neutral":
		return "Neutral"
	case "unmanaged":
		return "Custom"
	case "main-preview":
		return "Main preview"
	case "stable":
		return "Stable"
	default:
		return value
	}
}

func (m Model) beginManageChanges(config shellExperienceConfig) (tea.Model, tea.Cmd) {
	if len(config.existing) != 1 {
		config.err = errors.New("the managed installation is no longer available")
		m.shellExperience = &config
		return m, nil
	}
	install := config.existing[0]
	current, err := shellinstaller.LoadUserShellExperience(install)
	if err != nil {
		config.err = err
		m.shellExperience = &config
		return m, nil
	}
	selected := config.profile
	inPlace := install.Mode == "separate"
	preferences, components := manageChanges(current, selected, inPlace)
	m.shellManage = &shellManageState{install: install, config: config, preferences: preferences, components: components}
	m.shellExperience = nil
	m.shellPlan = nil
	return m, nil
}

func manageChanges(current, selected shellinstaller.UserExperience, inPlace bool) ([]screens.ShellManageChange, []screens.ShellManageChange) {
	var preferences, components []screens.ShellManageChange
	add := func(label, from, to string, component bool) {
		if from == to {
			return
		}
		change := screens.ShellManageChange{Label: label, From: manageDisplay(from, "Inherit"), To: manageDisplay(to, "Inherit"), Component: component}
		if component || !inPlace {
			components = append(components, change)
		} else {
			preferences = append(preferences, change)
		}
	}
	add("Persona", current.Persona, selected.Persona, false)
	add("Background", current.BackgroundSubagents, selected.BackgroundSubagents, false)
	add("Strict TDD", current.StrictTDD, selected.StrictTDD, false)
	add("Global RDD", current.GlobalRDD, selected.GlobalRDD, false)
	add("Context7", current.Context7, selected.Context7, false)
	add("GGA", current.GGA, selected.GGA, true)
	add("GGA provider", current.GGAProvider, selected.GGAProvider, true)
	add("CodeGraph", current.CodeGraph, selected.CodeGraph, true)
	add("SDK source", current.SDKSource, selected.SDKSource, true)
	return preferences, components
}

func (m Model) updateShellManage(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	manage := m.shellManage
	if manage == nil {
		return m, nil
	}
	if manage.applied {
		if key.String() == "enter" || key.String() == "esc" || key.String() == "ctrl+c" {
			m.shellManage = nil
			m.shellExperience, m.shellInstallModel = nil, nil
			m.shellInstallSelection = cli.ShellInstallSelection{}
			m.setScreen(ScreenWelcome)
		}
		return m, nil
	}
	last := 0
	if len(manage.preferences) > 0 {
		last = 1
	}
	switch key.String() {
	case "j", "down":
		manage.cursor = min(last, manage.cursor+1)
	case "k", "up":
		manage.cursor = max(0, manage.cursor-1)
	case "esc", "ctrl+c":
		m.shellExperience, m.shellManage = &manage.config, nil
		return m, nil
	case "enter":
		if manage.cursor == 1 || len(manage.preferences) == 0 {
			m.shellExperience, m.shellManage = &manage.config, nil
			return m, nil
		}
		if err := shellinstaller.ApplyUserExperiencePreferences(manage.install, manage.config.profile); err != nil {
			manage.err = err
			return m, nil
		}
		manage.applied, manage.err = true, nil
	}
	return m, nil
}

func (m Model) updateShellDetected(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	last := len(m.shellDetected) - 1
	if len(m.shellDetected) == 1 {
		// Single install: Manage is the focused action and Back is below it.
		last = 1
	}
	switch key.String() {
	case "j", "down":
		m.shellDetectedCursor = min(last, m.shellDetectedCursor+1)
	case "k", "up":
		m.shellDetectedCursor = max(0, m.shellDetectedCursor-1)
	case "enter":
		if len(m.shellDetected) == 1 && m.shellDetectedCursor == 1 {
			return m.leaveDetectedInstall(), nil
		}
		return m.manageDetectedInstall(), nil
	case "esc", "x", "ctrl+c":
		return m.leaveDetectedInstall(), nil
	}
	return m, nil
}

func (m Model) leaveDetectedInstall() Model {
	m.shellDetected, m.shellDetectedProfiles = nil, nil
	m.shellExperience, m.shellInstallModel = nil, nil
	m.shellDetectedCursor = 0
	m.setScreen(ScreenWelcome)
	return m
}

func (m Model) manageDetectedInstall() Model {
	if m.shellDetectedCursor < 0 || m.shellDetectedCursor >= len(m.shellDetected) {
		return m.leaveDetectedInstall()
	}
	install := m.shellDetected[m.shellDetectedCursor]
	profile := m.shellDetectedProfiles[m.shellDetectedCursor]
	if loaded, err := shellinstaller.LoadUserShellExperience(install); err == nil {
		profile = loaded
	}
	_, _, found, note := shellDetectPi()
	m.shellExperience = &shellExperienceConfig{
		profile: profile, mode: install.Mode, choice: "existing",
		prefix: install.Prefix, agent: install.Agent, piFound: found, piNote: note,
		existing: []shellinstaller.UserShellInstall{install}, reconfigure: true, target: install.Target,
	}
	m.shellInstallModel = cli.NewShellInstallModelWithExperience(nil, profile)
	m.shellDetected, m.shellDetectedProfiles = nil, nil
	m.shellDetectedCursor = 0
	return m
}

func (m Model) updateShellInstallRun(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(shellInstallTick); ok {
		if m.shellInstallRun == nil || m.shellInstallRun.done {
			return m, nil
		}
		m.shellInstallRun.spinner++
		return m, waitShellInstallEvent(m.shellInstallEvents)
	}
	event, ok := msg.(shellInstallEvent)
	if ok {
		run := m.shellInstallRun
		if run == nil {
			return m, nil
		}
		if event.done {
			run.done, run.err = true, event.err
			if event.err != nil {
				for index := range run.steps {
					if run.steps[index].State == "running" {
						run.steps[index].State = "failed"
					}
				}
			} else {
				for index := range run.steps {
					run.steps[index].State = "done"
				}
			}
			return m, nil
		}
		progress := event.progress
		if progress.Index >= 1 && progress.Index <= len(run.steps) {
			step := &run.steps[progress.Index-1]
			switch progress.State {
			case "running":
				step.State = "running"
			case "detail":
				if progress.Detail != "" {
					step.Details = append(step.Details, progress.Detail)
					if len(step.Details) > 3 {
						step.Details = step.Details[len(step.Details)-3:]
					}
				}
			case "done":
				step.State = "done"
				if progress.Index < len(run.steps) && run.steps[progress.Index].State == "pending" {
					run.steps[progress.Index].State = "running"
				}
			}
		}
		return m, waitShellInstallEvent(m.shellInstallEvents)
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	run := m.shellInstallRun
	if run == nil || !run.done {
		// The worker owns the operation; no key can claim completion or cancel
		// a possibly uncertain publication.
		return m, nil
	}
	if key.Type != tea.KeyEnter {
		return m, nil
	}
	if run.err != nil {
		m.shellInstallRun, m.shellInstallEvents = nil, nil
		m.shellInstallSelection = cli.ShellInstallSelection{}
		m.setScreen(ScreenWelcome)
		return m, nil
	}
	return m, tea.Quit
}

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
	if len(m.shellDetected) > 0 {
		return m.updateShellDetected(msg)
	}
	if m.shellManage != nil {
		return m.updateShellManage(msg)
	}
	if m.shellInstallRun != nil {
		return m.updateShellInstallRun(msg)
	}
	if m.shellPlan != nil {
		key, ok := msg.(tea.KeyMsg)
		if !ok {
			return m, nil
		}
		switch key.Type {
		case tea.KeyEnter:
			return m.beginShellInstall()
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
