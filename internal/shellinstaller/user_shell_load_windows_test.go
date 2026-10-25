package shellinstaller

import (
	"os"
	"path/filepath"
	"testing"
)

// Every preference file is optional: a Custom persona, a partial profile or a
// CLI install without --experience leaves some absent, and each absent file
// loads as its default value instead of failing Manage. A present but
// non-regular entry stays a refusal, never a silent default.
func TestWindowsLoadUserShellExperienceMissingOptionalFiles(t *testing.T) {
	_, target, install := writeWindowsChannelFixture(t)
	want := UserExperience{Persona: "unmanaged", BackgroundSubagents: "on", GGA: "off"}
	loaded, err := LoadUserShellExperience(install)
	if err != nil || loaded != want {
		t.Fatalf("absent preference files did not load as defaults: %+v %v", loaded, err)
	}
	// Only some files present: the rest keep their defaults.
	if err := userWindowsWrite(filepath.Join(target, "config", "background-subagents.json"), []byte(`{"schema":"gentle-pi.background-subagents/v1","policy":"off"}`)); err != nil {
		t.Fatal(err)
	}
	want.BackgroundSubagents = "off"
	if loaded, err = LoadUserShellExperience(install); err != nil || loaded != want {
		t.Fatalf("partial profile differs: %+v %v", loaded, err)
	}
	if err := os.Mkdir(filepath.Join(target, "config", "persona.json"), 0700); err != nil {
		t.Fatal(err)
	}
	if loaded, err = LoadUserShellExperience(install); err == nil {
		t.Fatalf("a non-regular persona entry loaded as a default: %+v", loaded)
	}
}
