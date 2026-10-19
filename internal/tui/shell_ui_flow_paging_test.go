package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestShellFlowReviewPageStride(t *testing.T) {
	m := shellFlowReady(t)
	m.Width, m.Height = 80, 6
	// Compact review has a title, focused action and footer. A page must
	// advance by the remaining visible data rows, without dropping a line.
	visible := len(strings.Split(ansi.Strip(m.View()), "\n")) - 3
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyPgDown})
	if visible <= 0 || m.shellReadyOffset != visible {
		t.Fatalf("paging skipped review lines: advanced %d, visible data rows %d", m.shellReadyOffset, visible)
	}
	m = experienceKey(m, tea.KeyMsg{Type: tea.KeyPgUp})
	if m.shellReadyOffset != 0 {
		t.Fatal("reverse review page did not return to the first data row")
	}
}
