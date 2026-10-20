package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/styles"
)

// ShellInstallStep is one ordered installation step. State is pending,
// running, done or failed. It carries no installation authority.
type ShellInstallStep struct {
	Label string
	State string
}

// RenderShellInstall keeps the operator inside the TUI while the owned worker
// runs. It never claims success without a completed worker result.
func RenderShellInstall(steps []ShellInstallStep, failure string, done bool, width, height int) string {
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
	lines := []string{styles.TitleStyle.Render(title), ""}
	for _, step := range steps {
		icon, suffix := "[ ]", ""
		switch step.State {
		case "running":
			icon, suffix = "[>]", "  running"
		case "done":
			icon, suffix = "[x]", "  done"
		case "failed":
			icon, suffix = "[!]", "  failed"
		}
		lines = append(lines, fmt.Sprintf("  %s %s%s", icon, step.Label, suffix))
	}
	if failure != "" {
		lines = append(lines, "", styles.ErrorStyle.Render("The installation stopped before publishing a verified result:"))
		for _, line := range strings.Split(failure, "\n") {
			lines = append(lines, "  "+line)
		}
	}
	lines = append(lines, "")
	if !done {
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
