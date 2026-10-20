package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v4/internal/components/gga"
	"github.com/gentleman-programming/gentle-ai/v4/internal/model"
)

func TestSyncGGAProviderNotices(t *testing.T) {
	for _, original := range []string{"", "PROVIDER=$(touch sentinel)\n", "PROVIDER=claude\nPROVIDER=codex\n"} {
		t.Run(original, func(t *testing.T) {
			home := t.TempDir()
			setSyncTestHome(t, home)
			for _, key := range []string{"HOME", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME", "XDG_RUNTIME_DIR", "PI_CODING_AGENT_DIR", "TEMP", "TMP"} {
				t.Setenv(key, home)
			}
			path := gga.ConfigPath(home)
			original = strings.ReplaceAll(original, "sentinel", "'"+filepath.ToSlash(filepath.Join(home, "sentinel"))+"'")
			if original != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(original), 0600); err != nil {
					t.Fatal(err)
				}
			}
			selection := model.Selection{Agents: []model.AgentID{model.AgentCursor}, Components: []model.ComponentID{model.ComponentGGA}}
			for i := 0; i < 2; i++ {
				result, err := RunSyncWithSelection(home, selection)
				if err != nil {
					t.Fatal(err)
				}
				if i == 1 && result.FilesChanged != 0 {
					t.Fatalf("repeat changed %d files: %v", result.FilesChanged, result.ChangedFiles)
				}
				rendered := RenderSyncReport(result)
				if !strings.Contains(rendered, "GGA config unchanged:") || !strings.Contains(rendered, "rerun gentle-ai sync") || !strings.Contains(rendered, path) {
					t.Fatalf("notice lost on run %d: %s", i, rendered)
				}
				for _, changed := range result.ChangedFiles {
					if changed == path {
						t.Fatal("skipped config counted as changed")
					}
				}
				got, err := os.ReadFile(path)
				if original == "" {
					if !os.IsNotExist(err) {
						t.Fatalf("config created: %s (%v)", got, err)
					}
				} else if err != nil || string(got) != original {
					t.Fatalf("unsafe config modified: %q (%v)", got, err)
				}
			}
			if _, err := os.Stat(filepath.Join(home, "sentinel")); !os.IsNotExist(err) {
				t.Fatal("dynamic shell executed")
			}
		})
	}
}

func TestSyncGGAProviderPreserved(t *testing.T) {
	home := t.TempDir()
	setSyncTestHome(t, home)
	t.Setenv("APPDATA", filepath.Join(home, "roaming"))
	// Seed using the very same Inject implementation called by installation.
	if _, err := gga.Inject(home, []model.AgentID{model.AgentCodex}); err != nil {
		t.Fatal(err)
	}
	selection := model.Selection{Agents: []model.AgentID{model.AgentCursor, model.AgentClaudeCode}, Components: []model.ComponentID{model.ComponentGGA}}
	result, err := RunSyncWithSelection(home, selection)
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(gga.ConfigPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `PROVIDER="codex"`) {
		t.Fatalf("shared install/sync provider overwritten: %s", content)
	}
	for _, changed := range result.ChangedFiles {
		if changed == gga.ConfigPath(home) {
			t.Fatal("unchanged provider config counted as changed")
		}
	}
}
