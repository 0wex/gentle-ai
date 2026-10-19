package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
	"github.com/gentleman-programming/gentle-ai/v4/internal/tui/styles"
	"github.com/rivo/uniseg"
)

const shellExperienceConfigHelp = "j/k: navigate • enter: select • End: continue • esc: back"

// These layouts restore e1f6bbe's configuration presentation. Their entries are
// projections of UserExperience, not the retired profile or review controller.
func renderShellExperienceConfig(p shellinstaller.UserExperience, cursor, width, height int) string {
	entries := shellExperienceEntries(p)
	cursor = min(max(0, cursor), len(ShellExperienceOptions(p))-1)
	if width <= 5 || height <= 4 {
		return renderShellExperienceConfigMinimum(entries, cursor, width, height)
	}
	rows := len(entries) + 3
	for i := 1; i < len(entries); i++ {
		if entries[i].Category != entries[i-1].Category {
			rows++
		}
	}
	if width >= 100 && height >= rows {
		return renderShellExperienceConfigWide(entries, cursor, width, height)
	}
	return renderShellExperienceConfigCompact(entries, cursor, width, height)
}

func renderShellExperienceConfigWide(entries []ShellExperienceEntry, cursor, width, height int) string {
	leftWidth := (width - 1) / 2
	rightWidth := width - leftWidth - 1
	left := make([]string, 0, len(entries)+len(shellExperienceCategories())-1)
	for index, entry := range entries {
		if index > 0 && entry.Category != entries[index-1].Category {
			left = append(left, "")
		}
		left = append(left, renderShellExperienceConfigEntry(entry, entry.Selectable && entry.Cursor == cursor, leftWidth))
	}
	focused, _ := shellExperienceConfigEntryAtCursor(entries, cursor)
	right := renderShellExperienceConfigPanel(entries, focused, rightWidth)
	rows := max(len(left), len(right))
	lines := []string{
		renderShellExperienceConfigLine(styles.TitleStyle, "Configure Gentle-Shell experience", width),
		renderShellExperienceConfigLine(styles.SubtextStyle, strings.Join(shellExperienceCategories(), " • "), width),
	}
	for i := 0; i < rows; i++ {
		var l, r string
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		lines = append(lines, padShellExperienceConfigStyled(l, leftWidth)+styles.SubtextStyle.Render("│")+padShellExperienceConfigStyled(r, rightWidth))
	}
	lines = append(lines, renderShellExperienceConfigLine(styles.HelpStyle, shellExperienceConfigHelp, width))
	return joinShellExperienceConfigLines(lines, height)
}

func renderShellExperienceConfigCompact(entries []ShellExperienceEntry, cursor, width, height int) string {
	focused, _ := shellExperienceConfigEntryAtCursor(entries, cursor)
	lines := []string{renderShellExperienceConfigLine(styles.TitleStyle, "Configure Gentle-Shell experience", width)}
	headerLines := 1
	if height >= 6 {
		lines = append(lines, renderShellExperienceConfigLine(styles.HeadingStyle, focused.Category+": "+focused.Label, width))
		headerLines++
	}
	available := height - headerLines - 1
	if available < 1 {
		return joinShellExperienceConfigLines(lines, height)
	}
	descriptionSlots := 0
	if available >= 4 {
		descriptionSlots = 2
	} else if available >= 3 {
		descriptionSlots = 1
	}
	listRows := available - descriptionSlots
	showHint := listRows >= 2 && listRows < len(entries)
	if showHint {
		listRows--
	}
	if listRows < 1 {
		listRows, descriptionSlots, showHint = 1, 0, false
	}
	listRows = min(listRows, len(entries))
	start := min(max(0, shellExperienceConfigEntryIndex(entries, cursor)-listRows/2), len(entries)-listRows)
	for _, entry := range entries[start : start+listRows] {
		lines = append(lines, renderShellExperienceConfigEntry(entry, entry.Selectable && entry.Cursor == cursor, width))
	}
	if showHint {
		lines = append(lines, renderShellExperienceConfigLine(styles.HelpStyle, "↑/↓: more", width))
	}
	if descriptionSlots > 0 {
		description := wrapShellExperienceConfigText("About: "+focused.Description, width)
		for _, line := range description[:min(descriptionSlots, len(description))] {
			lines = append(lines, renderShellExperienceConfigLine(styles.SubtextStyle, line, width))
		}
	}
	lines = append(lines, renderShellExperienceConfigLine(styles.HelpStyle, shellExperienceConfigHelp, width))
	return joinShellExperienceConfigLines(lines, height)
}

func renderShellExperienceConfigPanel(entries []ShellExperienceEntry, focused ShellExperienceEntry, width int) []string {
	contentWidth := width - styles.PanelStyle.GetHorizontalFrameSize()
	if contentWidth < 1 {
		return []string{renderShellExperienceConfigLine(styles.SubtextStyle, focused.Label, width)}
	}
	content := []string{"Focused: " + focused.Label}
	if focused.Value != "" && focused.Kind != "foundation" {
		content = append(content, "Entry value: "+focused.Value)
	}
	persona, sdk, gga := "", "", ""
	for _, entry := range entries {
		switch entry.Control {
		case ShellExperiencePersona:
			persona = entry.Value
		case ShellExperienceSDK:
			sdk = entry.Value
		case ShellExperienceGGA:
			gga = entry.Value
		}
	}
	if persona == "" {
		persona = "unset"
	}
	content = append(content, "Current: "+persona, sdk, "GGA: "+gga, "")
	content = append(content, wrapShellExperienceConfigText(focused.Description, contentWidth)...)
	return strings.Split(styles.PanelStyle.Width(contentWidth).Render(strings.Join(content, "\n")), "\n")
}

func renderShellExperienceConfigEntry(entry ShellExperienceEntry, focused bool, width int) string {
	prefix, style := "  ", styles.UnselectedStyle
	if focused {
		prefix, style = styles.Cursor, styles.SelectedStyle
	} else if entry.Kind == "foundation" {
		style = styles.SubtextStyle
	}
	label := entry.Label
	switch entry.Kind {
	case "foundation":
		label = "• " + label
	case "persona", "radio":
		marker := "( )"
		if entry.Current {
			marker = "(*)"
		}
		label = marker + " " + label
	case "toggle":
		marker := "[ ]"
		if entry.Checked {
			marker = "[x]"
		}
		label = marker + " " + label + ": " + entry.Value
	case "text":
		label += ": " + entry.Value
	}
	return renderShellExperienceConfigLine(style, prefix+label, width)
}

func shellExperienceConfigEntryAtCursor(entries []ShellExperienceEntry, cursor int) (ShellExperienceEntry, bool) {
	for _, entry := range entries {
		if entry.Selectable && entry.Cursor == cursor {
			return entry, true
		}
	}
	return ShellExperienceEntry{}, false
}

func shellExperienceConfigEntryIndex(entries []ShellExperienceEntry, cursor int) int {
	for i, entry := range entries {
		if entry.Selectable && entry.Cursor == cursor {
			return i
		}
	}
	return 0
}

func renderShellExperienceConfigMinimum(entries []ShellExperienceEntry, cursor, width, height int) string {
	if height <= 2 {
		if width == 1 {
			return joinShellExperienceConfigLines([]string{"C", "X"}, height)
		}
		return joinShellExperienceConfigLines([]string{"CX"}, height)
	}
	focused, _ := shellExperienceConfigEntryAtCursor(entries, cursor)
	label := focused.Label
	if focused.Control == ShellExperienceContinue {
		label = "C"
	} else if focused.Control == ShellExperienceCancel {
		label = "X"
	} else if g := uniseg.NewGraphemes(label); g.Next() {
		label = g.Str()
	}
	lines := []string{
		renderShellExperienceConfigLine(styles.TitleStyle, "Config", width),
		renderShellExperienceConfigLine(styles.SelectedStyle, "▸ "+label, width),
		renderShellExperienceConfigLine(styles.HelpStyle, "CX", width),
	}
	if height >= 4 {
		lines = append(lines, renderShellExperienceConfigLine(styles.HelpStyle, "j/k", width))
	}
	return joinShellExperienceConfigLines(lines, height)
}

func renderShellExperienceConfigLine(style lipgloss.Style, text string, width int) string {
	return style.Render(clipShellExperienceConfigDisplay(text, width))
}

func padShellExperienceConfigStyled(text string, width int) string {
	return text + strings.Repeat(" ", max(0, width-lipgloss.Width(text)))
}

func joinShellExperienceConfigLines(lines []string, height int) string {
	return strings.Join(lines[:min(len(lines), height)], "\n")
}

func wrapShellExperienceConfigText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	lines := wrapPlainLine(text, width)
	for i := range lines {
		lines[i] = clipShellExperienceConfigDisplay(lines[i], width)
	}
	return lines
}

func clipShellExperienceConfigDisplay(text string, width int) string {
	if width <= 0 || lipgloss.Width(text) <= width {
		return text
	}
	var b strings.Builder
	used := 0
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		cluster := g.Str()
		n := lipgloss.Width(cluster)
		if n == 0 || used+n <= width {
			b.WriteString(cluster)
			used += n
		} else {
			break
		}
	}
	return b.String()
}
