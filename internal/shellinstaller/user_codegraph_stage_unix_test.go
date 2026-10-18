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
					// Qualified Separate now has paired positive coverage; the generic
					// writer must still reject the same selection without effects.
					files, err := userStageExperienceFiles(&profile)
					if files != nil {
						t.Fatal("generic CodeGraph refusal returned staged files", files)
					}
					refusal = err
				} else {
					refusal = userPrepareInstallStage(context.Background(), root, mode, &profile)
				}
				if refusal == nil || !strings.Contains(refusal.Error(), "CodeGraph") {
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
