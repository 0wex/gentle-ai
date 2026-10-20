package screens

import (
	"fmt"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/styles"
)

// ShellManageChange is one detected preference or component difference.
type ShellManageChange struct {
	Label, From, To string
	Component       bool
}

// RenderShellManage summarizes the selected deltas before applying anything.
func RenderShellManage(target string, preferences, components []ShellManageChange, applied bool, applyErr string, cursor, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := []string{styles.TitleStyle.Render("Manage installation"), detectedSubtleStyle.Render(shellDetectedTarget(target)), ""}
	if len(preferences) == 0 && len(components) == 0 {
		lines = append(lines, "No changes selected.")
	}
	changeLines := func(changes []ShellManageChange) []string {
		var out []string
		for _, change := range changes {
			out = append(out, fmt.Sprintf("  %-14s %s → %s", change.Label, stringOr(change.From, "inherit"), stringOr(change.To, "inherit")))
		}
		return out
	}
	if len(preferences) > 0 {
		lines = append(lines, styles.HeadingStyle.Render("PREFERENCE CHANGES"))
		lines = append(lines, changeLines(preferences)...)
		lines = append(lines, "")
	}
	if len(components) > 0 {
		lines = append(lines, styles.WarningStyle.Render("COMPONENT CHANGES (not applied here)"))
		lines = append(lines, changeLines(components)...)
		lines = append(lines, detectedSubtleStyle.Render("These need a component update; changing preferences does not reinstall the stack."), "")
	}
	if applyErr != "" {
		lines = append(lines, styles.ErrorStyle.Render(applyErr), "")
	}
	if applied {
		lines = append(lines, styles.SuccessStyle.Render("Preferences updated."))
		if len(components) > 0 {
			lines = append(lines, styles.WarningStyle.Render("Component changes remain pending."))
		}
		lines = append(lines, "", detectedSubtleStyle.Render("Enter: back to menu"))
		return strings.Join(lines, "\n")
	}
	actions := []string{}
	if len(preferences) > 0 {
		actions = append(actions, "Apply preferences")
	}
	actions = append(actions, "Back")
	for index, action := range actions {
		marker, style := "   ", detectedValueStyle
		if index == cursor {
			marker, style = " ▸ ", detectedAccentStyle
		}
		lines = append(lines, marker+style.Render(action))
	}
	lines = append(lines, "", detectedSubtleStyle.Render("j/k: navigate  •  enter: select  •  esc: back"))
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

func stringOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
