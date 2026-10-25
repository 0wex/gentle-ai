//go:build linux || darwin

package shellinstaller

import (
	"os"
	"path/filepath"
	"testing"
)

// A HOME spelled through a symlink (the macOS /var -> /private/var temp root,
// or a linked home directory) still finds the install whose canonical
// manifest records the resolved destination, reported in canonical form.
func TestDiscoverUserShellInstallsThroughSymlinkedHome(t *testing.T) {
	parent := userUnixTempRoot(t)
	home := filepath.Join(parent, "home")
	if err := os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, ".local", "share", "gentle-shell", "install")
	writeInstallFixture(t, target, userManifest{Schema: userSchema, Destination: target, Prefix: filepath.Join(target, "prefix"), Agent: filepath.Join(target, "agent"), Mode: "separate"})
	linked := filepath.Join(parent, "linked-home")
	if err := os.Symlink(home, linked); err != nil {
		t.Fatal(err)
	}
	for _, spelled := range []string{home, linked} {
		installs := DiscoverUserShellInstalls(spelled)
		if len(installs) != 1 || installs[0].Target != target || installs[0].Mode != "separate" || installs[0].Channel != "stable" || installs[0].Agent != filepath.Join(target, "agent") {
			t.Fatalf("HOME %q did not discover the canonical install: %+v", spelled, installs)
		}
	}
}
