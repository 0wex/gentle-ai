//go:build linux || darwin

package shellinstaller

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUserCodeGraphStageRefusesBeforeEffects(t *testing.T) {
	for _, mode := range []string{"separate", "shared"} {
		for _, choice := range []string{"on", "off"} {
			t.Run(mode+"/"+choice, func(t *testing.T) {
				profile, err := DecodeUserExperience(`{"persona":"unmanaged","background_subagents":"off","codegraph":"` + choice + `"}`)
				if err != nil {
					t.Fatal(err)
				}
				parent := t.TempDir()
				root := filepath.Join(parent, "stage")
				var refusal error
				if mode == "separate" && runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
					// Qualified Separate has paired positive coverage; an unsupported
					// Main+CodeGraph combination keeps the generic refusal.
					if files, err := userStageExperienceFiles(&profile); err != nil || len(files) == 0 {
						t.Fatal("qualified profile lost generic staging", files, err)
					}
					refused := profile
					refused.SDKSource, refused.CodeGraph = "main-preview", "on"
					files, err := userStageExperienceFiles(&refused)
					if files != nil {
						t.Fatal("generic CodeGraph refusal returned staged files", files)
					}
					refusal = err
				} else {
					refusal = userPrepareInstallStage(context.Background(), root, mode, &profile)
				}
				want := "CodeGraph"
				if mode == "separate" && runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
					// The stock PNPM stage has no Main producer, so it names Main
					// first on the unsupported combination.
					want = "Main SDK"
				}
				if refusal == nil || !strings.Contains(refusal.Error(), want) {
					t.Fatal("unqualified CodeGraph stage was admitted", refusal)
				}
				entries, err := os.ReadDir(parent)
				if err != nil || len(entries) != 0 {
					t.Fatal("CodeGraph refusal changed the destination parent", entries, err)
				}
			})
		}
	}
}
