package tui

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentleman-programming/gentle-ai/v4/internal/cli"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/screens"
)

// The ordinary Welcome starts with the agreed product preset. Incoming requests
// retain their explicit OFF/inherited choices; the transport defaults stay inert.
func shellWelcomeExperience() shellinstaller.UserExperience {
	profile := shellinstaller.DefaultUserExperience()
	profile.StrictTDD = "on"
	profile.GlobalRDD = "on"
	profile.CodeGraph = "on"
	profile.Context7 = "on"
	profile.GGA = "off"
	return profile
}

// shellDetectPi reports the single local Pi install Shared may reuse. Detection
// lives in the installer package; the TUI never invents paths.
func shellDetectPi() (string, string, bool, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", false, "missing"
	}
	detection := shellinstaller.DetectUserPi(home)
	note := detection.Problem
	if detection.Detail != "" {
		note = detection.Detail
	}
	return detection.Prefix, detection.Agent, detection.Found, note
}

// shellDefaultDestination keeps the operator out of path editing. The parent is
// created private 0700; an existing install refuses a blind overwrite.
func shellDefaultDestination() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(home, ".local", "share", "gentle-shell")
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", err
	}
	info, err := os.Lstat(base)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
		return "", errors.New("default install parent is not a private 0700 directory: " + base)
	}
	destination := filepath.Join(base, "install")
	if _, err := os.Lstat(destination); err == nil {
		return "", errors.New("Gentle-Shell is already installed at " + destination + "; remove it before reinstalling")
	}
	return destination, nil
}

func shellAgentUsable(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir()
}

func shellDiscoverInstalls() []shellinstaller.UserShellInstall {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	installs := shellinstaller.DiscoverUserShellInstalls(home)
	if len(installs) > 1 {
		return nil // Ambiguous discovery never offers a choice.
	}
	return installs
}

func (m *Model) beginShellExperience() {
	m.shellReadyCursor, m.shellReadyOffset = 0, 0
	profile := shellWelcomeExperience()
	prefix, agent, found, note := shellDetectPi()
	m.shellExperience = &shellExperienceConfig{profile: profile, mode: "separate", choice: "separate", prefix: prefix, agent: agent, piFound: found, piNote: note, existing: shellDiscoverInstalls()}
	m.shellInstallModel = cli.NewShellInstallModelWithExperience(nil, profile)
}

type shellExperienceConfig struct {
	profile     shellinstaller.UserExperience
	cursor      int
	err         error
	mode        string
	choice      string
	prefix      string
	agent       string
	piFound     bool
	piNote      string
	existing    []shellinstaller.UserShellInstall
	reconfigure bool
	target      string
}

func (m Model) configureShellExperience(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if key.Type == tea.KeyEsc || key.Type == tea.KeyCtrlC {
		m.shellExperience, m.shellInstallModel = nil, nil
		m.shellInstallSelection = cli.ShellInstallSelection{}
		m.setScreen(ScreenWelcome)
		return m, nil
	}
	config := *m.shellExperience
	config.err = nil
	reuseAgent := ""
	if shellAgentUsable(config.agent) {
		reuseAgent = config.agent
	}
	targets := screens.ShellInstallTargetEntries(config.choice, config.piFound, config.piNote, reuseAgent, config.existing)
	options := screens.ShellExperienceOptions(config.profile, targets...)
	config.cursor = min(max(0, config.cursor), len(options)-1)
	entry := options[config.cursor]
	if entry.Control == screens.ShellExperienceProvider && key.Type == tea.KeyRunes {
		text := string(key.Runes)
		if len(config.profile.GGAProvider)+len(text) <= 64 && strings.IndexFunc(text, func(r rune) bool { return r < 32 || r > 126 }) < 0 {
			config.profile.GGAProvider += text
		}
	} else {
		pressed := key.String()
		if strings.EqualFold(pressed, "x") {
			return m.configureShellExperience(tea.KeyMsg{Type: tea.KeyEsc})
		}
		if strings.EqualFold(pressed, "c") {
			for i, option := range options {
				if option.Control == screens.ShellExperienceContinue {
					config.cursor, entry = i, option
				}
			}
			pressed = "enter"
		}
		switch pressed {
		case "up", "k":
			config.cursor = max(0, config.cursor-1)
		case "down", "j", "tab":
			config.cursor = min(len(options)-1, config.cursor+1)
		case "home":
			config.cursor = 0
		case "end":
			for i, option := range options {
				if option.Control == screens.ShellExperienceContinue {
					config.cursor = i
				}
			}
		case "backspace":
			if entry.Control == screens.ShellExperienceProvider && len(config.profile.GGAProvider) > 0 {
				config.profile.GGAProvider = config.profile.GGAProvider[:len(config.profile.GGAProvider)-1]
			}
		case "enter":
			switch entry.Control {
			case screens.ShellExperiencePersona:
				config.profile.Persona = entry.Choice
			case screens.ShellExperienceBackground:
				if config.profile.BackgroundSubagents == "on" {
					config.profile.BackgroundSubagents = "off"
				} else {
					config.profile.BackgroundSubagents = "on"
				}
			case screens.ShellExperienceInstallTarget:
				config.reconfigure, config.target = false, ""
				if entry.Choice == "shared" {
					if !config.piFound {
						detail := "No single npm-style Pi install was detected; choose Separate install"
						if config.piNote != "missing" && config.piNote != "ambiguous" && config.piNote != "" {
							detail = config.piNote
						}
						config.err = errors.New(detail)
					} else {
						config.mode, config.choice = "shared", "shared"
					}
				} else if entry.Choice == "agent" {
					if !shellAgentUsable(config.agent) {
						config.err = errors.New("No Pi agent was detected; choose Separate install")
					} else {
						config.mode, config.choice, config.prefix = "agent", "agent", ""
					}
				} else if entry.Choice == "existing" && len(config.existing) == 1 {
					install := config.existing[0]
					profile, loadErr := shellinstaller.LoadUserShellExperience(install)
					if loadErr != nil {
						config.err = loadErr
					} else {
						config.profile, config.mode, config.choice = profile, install.Mode, "existing"
						config.prefix, config.agent = install.Prefix, install.Agent
						config.reconfigure, config.target = true, install.Target
					}
				} else {
					config.mode, config.choice = "separate", "separate"
				}
			case screens.ShellExperienceProvider: // Typed explicitly, never inferred from Pi.
			case screens.ShellExperienceSDK:
				// The channel never consumes an optional preference. Only the
				// platform qualification locks it; unsupported preference
				// pairings are named when continuing.
				if entry.Choice == "main-preview" {
					config.err = shellMainPlatformError()
				}
				if config.err == nil {
					config.profile.SDKSource = entry.Choice
				}
			case screens.ShellExperienceContinue:
				config.err = config.profile.Validate()
				if config.err == nil && config.profile.SDKSource == "main-preview" {
					config.err = shellMainSDKChoiceError(config.profile)
				}
				if config.err == nil {
					destination, destErr := shellDefaultDestination()
					if destErr != nil {
						config.err = destErr
					} else {
						state, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
						request := state.Request
						request.Destination = destination
						request.Mode = config.mode
						if config.mode == "shared" {
							request.SharedPrefix, request.SharedAgent = config.prefix, config.agent
						} else if config.mode == "agent" {
							request.SharedPrefix, request.SharedAgent = "", config.agent
						} else {
							request.SharedPrefix, request.SharedAgent = "", ""
						}
						request.Experience = &config.profile
						token, inspectErr := shellinstaller.InspectUserInstall(request)
						if inspectErr != nil {
							config.err = inspectErr
						} else {
							request.Confirmation = token
							m.shellPlanRequest = request
							m.shellPlanConfig = &config
							m.shellPlan = append([]string(nil), shellInstallStepLabels...)
							m.shellExperience = nil
							m.shellInstallModel, _ = m.shellInstallModel.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height})
							return m, nil
						}
					}
				}
			case screens.ShellExperienceCancel:
				return m.configureShellExperience(tea.KeyMsg{Type: tea.KeyEsc})
			default:
				fields := []*string{&config.profile.StrictTDD, &config.profile.GlobalRDD, &config.profile.CodeGraph, &config.profile.Context7, &config.profile.GGA}
				field := fields[entry.Control-screens.ShellExperienceTDD]
				if *field == "on" {
					*field = "off"
				} else {
					*field = "on"
				}
			}
		}
	}
	m.shellExperience = &config
	return m, nil
}

func (m Model) shellInstallView() string {
	if m.shellInstallRun != nil {
		steps := make([]screens.ShellInstallStep, len(m.shellInstallRun.steps))
		for index, step := range m.shellInstallRun.steps {
			steps[index] = screens.ShellInstallStep{Label: step.Label, State: step.State}
		}
		failure := ""
		if m.shellInstallRun.err != nil {
			failure = m.shellInstallRun.err.Error()
		}
		return screens.RenderShellInstall(steps, failure, m.shellInstallRun.done, m.Width, m.Height)
	}
	if m.shellPlan != nil {
		return screens.RenderShellPlan(m.shellPlan, m.Width, m.Height)
	}
	if m.shellExperience != nil {
		reuseAgent := ""
		if shellAgentUsable(m.shellExperience.agent) {
			reuseAgent = m.shellExperience.agent
		}
		targets := screens.ShellInstallTargetEntries(m.shellExperience.choice, m.shellExperience.piFound, m.shellExperience.piNote, reuseAgent, m.shellExperience.existing)
		return screens.RenderShellExperience(m.shellExperience.profile, m.shellExperience.cursor, m.shellExperience.err, m.Width, m.Height, targets...)
	}
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if ok && state.Review {
		return screens.RenderShellReady(state.Request, state.Preview, m.Width, m.Height, m.shellReadyOffset, m.shellReadyCursor)
	}
	if ok {
		return screens.RenderUserInstallDestination(state.Request, state.Field, state.Error, m.Width, m.Height)
	}
	return m.shellInstallModel.View()
}

func (m Model) shellExperienceNavigation(key tea.KeyMsg) (Model, bool) {
	state, ok := cli.ShellInstallPresentationOf(m.shellInstallModel)
	if !ok {
		return m, false
	}
	if key.Type == tea.KeyCtrlE {
		profile := shellWelcomeExperience()
		if state.Request.Experience != nil {
			profile = *state.Request.Experience
		}
		m.shellInstallModel = cli.NewShellInstallModelWithRequest(nil, state.Request)
		m.shellInstallModel, _ = m.shellInstallModel.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height})
		m.shellInstallSelection = cli.ShellInstallSelection{}
		m.shellReadyOffset, m.shellReadyCursor = 0, 0
		prefix, agent, found, note := shellDetectPi()
		mode := "separate"
		if state.Request.Mode == "shared" {
			mode, prefix, agent = "shared", state.Request.SharedPrefix, state.Request.SharedAgent
		}
		m.shellExperience = &shellExperienceConfig{profile: profile, mode: mode, choice: mode, prefix: prefix, agent: agent, piFound: found, piNote: note, existing: shellDiscoverInstalls()}
		return m, true
	}
	if !state.Review {
		return m, false
	}
	switch key.String() {
	case "up", "k":
		m.shellReadyCursor = max(0, m.shellReadyCursor-1)
	case "down", "j":
		m.shellReadyCursor = min(2, m.shellReadyCursor+1)
	case "pgup":
		m.shellReadyOffset -= screens.ShellReadyPageRows(m.Width, m.Height)
	case "pgdown":
		m.shellReadyOffset += screens.ShellReadyPageRows(m.Width, m.Height)
	case "home":
		m.shellReadyOffset = 0
	case "end":
		m.shellReadyOffset = screens.ShellReadyLastOffset(state.Request, state.Preview, m.Width, m.Height)
	default:
		return m, false
	}
	m.shellReadyOffset = min(max(0, m.shellReadyOffset), screens.ShellReadyLastOffset(state.Request, state.Preview, m.Width, m.Height))
	return m, true
}

// This visible choice gate never grants worker or physical-selection authority.
func shellMainPlatformError() error {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		return errors.New("SDK Main + stock provider requires Linux amd64 fresh Separate")
	}
	return nil
}

func shellMainSDKChoiceError(p shellinstaller.UserExperience) error {
	if err := shellMainPlatformError(); err != nil {
		return err
	}
	return nil
}
