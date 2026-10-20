package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/styles"
)

// ShellDetectedInstall is the read-only discovery summary. It grants no
// installation or replacement authority.
type ShellDetectedInstall struct {
	Target, Mode, Channel              string
	Persona, Background                string
	TDD, RDD, CodeGraph, Context7, GGA string
}

var shellDetectedActions = []string{"Manage this installation", "Back"}

func shellDetectedValue(value string) string {
	if value == "" {
		return "inherit"
	}
	return value
}

// RenderShellDetected identifies an existing managed installation before the
// operator enters the same customization screen used for a fresh install.
func RenderShellDetected(install ShellDetectedInstall, cursor, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := []string{
		styles.TitleStyle.Render("Gentle-Shell detected"),
		"",
		"  Found a managed Gentle-Shell installation:",
		"  " + install.Target,
		"",
		fmt.Sprintf("  Mode: %s    Channel: %s", install.Mode, install.Channel),
		fmt.Sprintf("  Persona: %s    Background subagents: %s", shellDetectedValue(install.Persona), shellDetectedValue(install.Background)),
		fmt.Sprintf("  Strict TDD: %s    Global RDD: %s    CodeGraph: %s", shellDetectedValue(install.TDD), shellDetectedValue(install.RDD), shellDetectedValue(install.CodeGraph)),
		fmt.Sprintf("  Context7: %s    GGA: %s", shellDetectedValue(install.Context7), shellDetectedValue(install.GGA)),
		"",
		styles.TitleStyle.Render("MANAGE"),
	}
	for index, action := range shellDetectedActions {
		marker := "  "
		if index == cursor {
			marker = "▸ "
		}
		lines = append(lines, "  "+marker+action)
	}
	lines = append(lines, "", styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))
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
