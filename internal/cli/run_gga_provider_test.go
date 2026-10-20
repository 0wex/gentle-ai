package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v4/internal/components/gga"
	"github.com/gentleman-programming/gentle-ai/v4/internal/model"
	"github.com/gentleman-programming/gentle-ai/v4/internal/planner"
)

func TestGGAProviderVerificationUsesActualSkip(t *testing.T) {
	for _, agent := range []model.AgentID{model.AgentCursor, model.AgentClaudeCode} {
		t.Run(string(agent), func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("APPDATA", filepath.Join(home, "roaming"))
			selection := model.Selection{Agents: []model.AgentID{agent}, Components: []model.ComponentID{model.ComponentGGA}}
			res, err := gga.Inject(home, selection.Agents)
			if err != nil {
				t.Fatal(err)
			}
			wantSkip := agent == model.AgentCursor
			if res.ConfigCreationSkipped != wantSkip {
				t.Fatalf("creation skip = %v, want %v", res.ConfigCreationSkipped, wantSkip)
			}
			state := &runtimeState{}
			state.recordGGAInjection(res)
			input := postApplyVerificationInput{HomeDir: home, Scope: ScopeGlobal, Selection: selection, Resolved: planner.ResolvedPlan{Agents: selection.Agents, OrderedComponents: selection.Components}, State: state}
			report := runPostApplyVerification(input)
			if !report.Ready {
				t.Fatalf("install verification rejected actual injection: %+v", report)
			}
			configChecked := false
			for _, check := range report.Checks {
				if check.ID == "verify:file:"+gga.ConfigPath(home) {
					configChecked = true
				}
			}
			if configChecked == wantSkip {
				t.Fatalf("config verification requirement lost: skip=%v checks=%+v", wantSkip, report.Checks)
			}
			if report := runPostSyncVerificationWithState(home, "", ScopeGlobal, selection, state); !report.Ready {
				t.Fatalf("sync verification rejected actual injection: %+v", report)
			}
			// Snapshot/rollback inventory must retain even intentionally absent config.
			paths := componentPathsWithWorkspaceScoped(home, "", ScopeGlobal, selection, resolveAdapters(selection.Agents), model.ComponentGGA)
			if !containsPath(paths, gga.ConfigPath(home)) || !containsPath(paths, gga.AgentsTemplatePath(home)) {
				t.Fatalf("managed inventory weakened: %v", paths)
			}
			// Missing files without an observed skip must still fail, including
			// an unmapped selection that never actually ran injection.
			if wantSkip {
				input.State = nil
				if report := runPostApplyVerification(input); report.Ready {
					t.Fatal("absence allowed without actual skip")
				}
				if report := runPostSyncVerificationScoped(home, "", ScopeGlobal, selection); report.Ready {
					t.Fatal("sync absence allowed without actual skip")
				}
			}
			// Even an actual skip cannot exempt the required starter template.
			emptyHome := t.TempDir()
			t.Setenv("APPDATA", filepath.Join(emptyHome, "roaming"))
			input.HomeDir, input.State = emptyHome, state
			if report := runPostApplyVerification(input); report.Ready {
				t.Fatal("missing required files accepted")
			}
			if report := runPostSyncVerificationWithState(emptyHome, "", ScopeGlobal, selection, state); report.Ready {
				t.Fatal("sync accepted missing required files")
			}
		})
	}
}

// No installer is invoked: exercise the shared injection result and the exact
// installation output channel without persistent Windows PATH side effects.
func TestGGAProviderInstallNotice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("APPDATA", filepath.Join(home, "roaming"))
	for i := 0; i < 2; i++ {
		res, err := gga.Inject(home, []model.AgentID{model.AgentCursor})
		if err != nil {
			t.Fatal(err)
		}
		var stderr bytes.Buffer
		writeGGAProviderNotices(&stderr, res)
		if !strings.Contains(stderr.String(), "WARNING: GGA config unchanged: no provider") || !strings.Contains(stderr.String(), gga.ConfigPath(home)) {
			t.Fatalf("installation discarded notice: %q", stderr.String())
		}
	}
}
