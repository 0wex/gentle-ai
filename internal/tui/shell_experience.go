package tui

import (
	"errors"
	"os"
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
	return detection.Prefix, detection.Agent, detection.Found, detection.Problem
}

func (m *Model) beginShellExperience() {
	m.shellReadyCursor, m.shellReadyOffset = 0, 0
	profile := shellWelcomeExperience()
	prefix, agent, found, note := shellDetectPi()
	m.shellExperience = &shellExperienceConfig{profile: profile, mode: "separate", prefix: prefix, agent: agent, piFound: found, piNote: note}
	m.shellInstallModel = cli.NewShellInstallModelWithExperience(nil, profile)
}

type shellExperienceConfig struct {
	profile shellinstaller.UserExperience
	cursor  int
	err     error
	mode    string
	prefix  string
	agent   string
	piFound bool
	piNote  string
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
	targets := screens.ShellInstallTargetEntries(config.mode, config.piFound)
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
				if entry.Choice == "shared" {
					if !config.piFound {
						config.err = errors.New("No single npm-style Pi install was detected; choose Separate install")
					} else {
						config.mode = "shared"
					}
				} else {
					config.mode = "separate"
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
				if config.err == nil && config.mode == "shared" && config.profile.SDKSource == "main-preview" {
					config.err = errors.New("Main preview requires a Separate install; choose Stable or Separate")
				}
				if config.err == nil {
					state, _ := cli.ShellInstallPresentationOf(m.shellInstallModel)
					state.Request.Mode = config.mode
					if config.mode == "shared" {
						state.Request.SharedPrefix, state.Request.SharedAgent = config.prefix, config.agent
					} else {
						state.Request.SharedPrefix, state.Request.SharedAgent = "", ""
					}
					state.Request.Experience = &config.profile
					m.shellInstallModel = cli.NewShellInstallModelWithRequest(nil, state.Request)
					m.shellInstallModel, _ = m.shellInstallModel.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height})
					m.shellExperience, m.shellReadyOffset, m.shellReadyCursor = nil, 0, 0
					return m, nil
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
	if m.shellExperience != nil {
		targets := screens.ShellInstallTargetEntries(m.shellExperience.mode, m.shellExperience.piFound)
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
		m.shellExperience = &shellExperienceConfig{profile: profile, mode: mode, prefix: prefix, agent: agent, piFound: found, piNote: note}
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
