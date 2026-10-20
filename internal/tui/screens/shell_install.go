package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/styles"
)

// ShellBloomFrames is the number of rose-opening frames on success.
const ShellBloomFrames = 4

var shellRoseFrames = [][]string{
	{
		"      ╭─────╮",
		"      │  ●  │",
		"      ╰──┬──╯",
		"         │",
		"        ╱ ╲",
	},
	{
		"     ╭───────╮",
		"     │   ●   │",
		"    ╰───┬───╯",
		"         │",
		"        ╱ ╲",
	},
	{
		"    ╭─────────╮",
		"    │    ●    │",
		"   ╰────┬────╯",
		"         │",
		"        ╱ ╲",
	},
	{
		"      ╭─────╮",
		"    ╭─┤  ●  ├─╮",
		"    │ ╰──┬──╯ │",
		"    ╰────┼────╯",
		"         │",
		"        ╱ ╲",
	},
}

// ShellInstallStep is one ordered installation step. State is pending,
// running, done or failed. It carries no installation authority.
type ShellInstallStep struct {
	Label   string
	State   string
	Details []string
}

var shellInstallSpinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// RenderShellInstall keeps the operator inside the TUI while the owned worker
// runs. It never claims success without a completed worker result.
func RenderShellInstall(steps []ShellInstallStep, failure string, done bool, frame, bloom int, destinationPublished bool, mode string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	title := "Installing Gentle-Shell"
	if done && failure == "" {
		title = "Installation complete"
	}
	if done && failure != "" {
		title = "Installation failed"
	}
	completed := 0
	for _, step := range steps {
		if step.State == "done" {
			completed++
		}
	}
	barWidth := 12
	filled := 0
	if len(steps) > 0 {
		filled = barWidth * completed / len(steps)
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	lines := []string{styles.TitleStyle.Render(fmt.Sprintf("%s  [%s] %d/%d", title, bar, completed, len(steps))), ""}
	current := ""
	for _, step := range steps {
		icon, suffix := "[ ]", ""
		switch step.State {
		case "running":
			icon = "[" + shellInstallSpinner[frame%len(shellInstallSpinner)] + "]"
			suffix = "  running"
		case "done":
			icon, suffix = "[x]", "  done"
		case "failed":
			icon, suffix = "[!]", "  failed"
		}
		lines = append(lines, fmt.Sprintf("  %s %s%s", icon, step.Label, suffix))
		if step.State == "running" {
			for _, detail := range step.Details {
				lines = append(lines, "        › "+detail)
				current = detail
			}
		}
	}
	if failure != "" {
		lines = append(lines, "")
		switch {
		case !destinationPublished && mode == "separate":
			lines = append(lines, styles.WarningStyle.Render("No installation was applied: the process did not complete and no target was published."))
		case !destinationPublished:
			lines = append(lines, styles.WarningStyle.Render("No target was published. Agent/Shared side effects may need inspection before retrying."))
		default:
			lines = append(lines, styles.WarningStyle.Render("A destination exists; the transaction may be partial or uncertain. Inspect the preserved evidence."))
		}
		lines = append(lines, styles.ErrorStyle.Render("The installation stopped before publishing a verified result:"))
		for _, line := range strings.Split(failure, "\n") {
			for len(line) > width-2 && width > 2 {
				lines = append(lines, "  "+line[:width-2])
				line = line[width-2:]
			}
			lines = append(lines, "  "+line)
		}
	}
	if done && failure == "" {
		bloom = min(max(0, bloom), len(shellRoseFrames)-1)
		rose := lipgloss.NewStyle().Foreground(styles.ColorMauve)
		lines = append(lines, "")
		for _, line := range shellRoseFrames[bloom] {
			lines = append(lines, rose.Render("  "+line))
		}
		lines = append(lines, "", styles.TitleStyle.Render("Welcome to the gentle-tailored experience"), styles.SubtextStyle.Render("Your Gentle Shell is ready."))
	}
	lines = append(lines, "")
	if !done {
		if current != "" {
			lines = append(lines, styles.TitleStyle.Render("CURRENT TASK"), "  "+current, "")
		}
		lines = append(lines, styles.HelpStyle.Render("The worker is running; do not close the terminal"))
	} else if failure != "" {
		lines = append(lines, styles.HelpStyle.Render("Enter: back to menu"))
	} else {
		lines = append(lines, styles.HelpStyle.Render("Enter: finish"))
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for index, line := range lines {
		if len(line) > width {
			lines[index] = line[:width]
		}
	}
	return strings.Join(lines, "\n")
}
