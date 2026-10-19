package screens

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/styles"
)

func RenderShellExperience(p shellinstaller.UserExperience, cursor int, err error, width, height int) string {
	width, height = shellViewport(width, height)
	view := renderShellExperienceConfig(p, cursor, width, height)
	if err != nil {
		warning := wrapShellExperienceConfigText(err.Error(), width)
		warning = warning[:min(len(warning), max(1, min(3, height-1)))]
		for i := range warning {
			warning[i] = styles.WarningStyle.Render(warning[i])
		}
		return joinShellExperienceConfigLines(append(warning, strings.Split(view, "\n")...), height)
	}
	return view
}

func shellReadyLines(req shellinstaller.UserInstallRequest, preview string, width int) []string {
	data, err := json.MarshalIndent(req, "", "  ")
	left := "Selected experience\n" + shellSDKSourceLabel(req.Experience) + "\n" + string(data)
	if err != nil {
		left = "Selected experience unavailable\n" + err.Error()
	}
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".cmd"
	}
	right := strings.Join([]string{
		"Changes + scopes",
		"New selected runtime; personal Pi/projects unchanged.",
		"Config: " + filepath.Join(req.Destination, "config"),
		"Target HOME/state: " + filepath.Join(req.Destination, "home"),
		"Commands: " + filepath.Join(req.Destination, "bin", "gentle-shell"+suffix) + " and " + filepath.Join(req.Destination, "bin", "pi"+suffix),
		"Project overrides remain; no automatic indexing/GGA hooks.",
		preview,
	}, "\n")
	if req.Experience != nil && req.Experience.SDKSource == "main-preview" {
		right = "SDK Main + stock provider: Linux amd64 fresh Separate only.\nReady reviews choices, not installed health.\nMain installer is inert DATA; stock native/pins remain.\nSource + selected SDK: " + filepath.Join(req.Destination, "runtime") + "\n" + right
	}
	content := left + "\n\n" + right
	if width >= 100 {
		panel := (width-2)/2 - 4
		box := styles.PanelStyle.Width(panel)
		content = lipgloss.JoinHorizontal(lipgloss.Top, box.Render(ansi.Hardwrap(left, panel, true)), box.Render(ansi.Hardwrap(right, panel, true)))
	}
	return strings.Split(ansi.Hardwrap(content, width, true), "\n")
}

func RenderShellReady(req shellinstaller.UserInstallRequest, preview string, width, height, offset int) string {
	width, height = shellViewport(width, height)
	return shellViewportRender("Ready — review before installation", shellReadyLines(req, preview, width), offset, width, height, "y confirms; n edits target; Ctrl-E configures; Esc cancels; PgUp/PgDn/Home/End scroll")
}

func ShellReadyLastOffset(req shellinstaller.UserInstallRequest, preview string, width, height int) int {
	width, height = shellViewport(width, height)
	return max(0, len(shellReadyLines(req, preview, width))-max(1, height-2))
}

func shellViewport(width, height int) (int, int) {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 30
	}
	return width, height
}

func shellViewportRender(title string, lines []string, offset, width, height int, footer string) string {
	if height == 1 {
		return ansi.Truncate(lines[min(max(0, offset), len(lines)-1)], width, "")
	}
	start := min(max(0, offset), max(0, len(lines)-max(1, height-2)))
	end := min(len(lines), start+max(1, height-2))
	view := []string{ansi.Truncate(styles.TitleStyle.Render(title), width, "")}
	view = append(view, lines[start:end]...)
	if height > 2 {
		view = append(view, ansi.Truncate(styles.HelpStyle.Render(footer), width, ""))
	}
	return strings.Join(view, "\n")
}

func shellSDKSourceLabel(p *shellinstaller.UserExperience) string {
	if p == nil || p.SDKSource == "" {
		return "SDK source: stock (default)"
	}
	if p.SDKSource == "main-preview" {
		return "SDK Main + stock provider (Linux amd64 fresh Separate only)"
	}
	return "SDK source: stock (" + p.SDKSource + ")"
}
