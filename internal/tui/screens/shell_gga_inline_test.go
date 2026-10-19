package screens

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
)

func TestShellGGAInlineProvider(t *testing.T) {
	for _, tt := range []struct {
		name, choice, provider string
		visible                bool
	}{
		{"inherited", "", "", false},
		{"off", "off", "", false},
		{"on requires explicit input", "on", "", true},
		{"on with provider", "on", "codex", true},
		{"off keeps provider editable for recovery", "off", "codex", true},
		{"inherit keeps provider editable for recovery", "", "codex", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			profile := shellinstaller.DefaultUserExperience()
			profile.GGA, profile.GGAProvider = tt.choice, tt.provider
			before := profile
			options := ShellExperienceOptions(profile)
			provider, gga, next := -1, -1, -1
			for i, entry := range options {
				if entry.Cursor != i {
					t.Fatal("conditional entry broke continuous cursor positions")
				}
				switch entry.Control {
				case ShellExperienceGGA:
					gga = i
				case ShellExperienceProvider:
					provider = i
					if entry.Category != "Optional workflows" {
						t.Error("provider is not grouped with GGA")
					}
				case ShellExperienceContinue:
					next = i
				}
			}
			if (provider >= 0) != tt.visible {
				t.Errorf("unnecessary empty provider row remains or required explicit input disappeared: index=%d", provider)
			}
			wantNext := gga + 1
			if tt.visible {
				wantNext++
				if provider != gga+1 {
					t.Error("provider did not remain adjacent to GGA")
				}
			}
			if next != wantNext {
				t.Error("Continue is not reachable directly after the relevant GGA fields")
			}
			view := ansi.Strip(RenderShellExperience(profile, 0, nil, 209, 61))
			if strings.Contains(view, "ADVANCED") {
				t.Error("Advanced group remains in ordinary configuration")
			}
			if !tt.visible && strings.Contains(view, "GGA provider") {
				t.Error("inactive GGA shows an empty provider field")
			}
			for _, size := range [][2]int{{100, 31}, {100, 32}, {80, 14}, {30, 7}, {5, 1}} {
				for _, entry := range options {
					assertShellOriginalViewport(t, RenderShellExperience(profile, entry.Cursor, nil, size[0], size[1]), size[0], size[1])
				}
			}
			if profile != before {
				t.Fatal("rendering changed the selected GGA/provider values")
			}
		})
	}
}
