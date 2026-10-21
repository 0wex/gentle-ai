//go:build windows

package shellinstaller

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUninstallWindowsLeavesMinimalBackup(t *testing.T) {
	root := t.TempDir()
	if err := userWindowsPrivate(root); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "install")
	for _, relative := range []string{
		`config`, `agent`, `home`, `home/go`, `runtime/node`, `runtime/archives`, `runtime/cache`,
		`prefix`, `state`, `tmp`, `bin`, `source`, `project`,
	} {
		if err := os.MkdirAll(filepath.Join(target, relative), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, relative := range []string{
		`config/persona.json`, `agent/mcp.json`, `home/.gentle-ai/state.json`,
		`selection.json`, `manifest.json`, `runtime/archives/node.zip`,
		`prefix/lib/x.js`, `state/old`, `tmp/scratch`, `bin/gentle-shell.cmd`,
		`source/package-lock.json`, `provision.mjs`, `supervisor.exe`,
		`complete-generated-lock-sri.mjs`, `main.json`, `home/go/pkg/mod/m`,
	} {
		if err := userWindowsWrite(filepath.Join(target, relative), []byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	home := t.TempDir()
	backup, err := UninstallUserShellInstall(home, UserShellInstall{Target: target, Mode: "separate"})
	if err != nil || backup == "" {
		t.Fatalf("Windows uninstall failed: %v %q", err, backup)
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatal("target still exists after Windows uninstall")
	}
	if filepath.Dir(backup) != root || filepath.Base(backup) == "install" {
		t.Fatalf("backup path %q is not the renamed target", backup)
	}
	entries, err := os.ReadDir(backup)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		if !userWindowsBackupKeep[entry.Name()] {
			t.Fatalf("backup retains unpruned entry %q", entry.Name())
		}
		seen[entry.Name()] = true
	}
	for _, name := range []string{"config", "agent", "home", "selection.json", "manifest.json"} {
		if !seen[name] {
			t.Fatalf("minimal backup lost %q", name)
		}
	}
	if _, err := os.Lstat(filepath.Join(backup, `home`, `go`)); !os.IsNotExist(err) {
		t.Fatalf("backup retained the Go module cache inside home: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(backup, `home`, `.gentle-ai`, `state.json`)); err != nil {
		t.Fatalf("uninstall lost the retained home state: %v", err)
	}
}

func TestUninstallWindowsRefusesUncertainTargets(t *testing.T) {
	home := t.TempDir()
	if _, err := UninstallUserShellInstall(home, UserShellInstall{Target: filepath.Join(home, "absent"), Mode: "separate"}); err == nil {
		t.Fatal("uninstall accepted an absent target")
	}
	real := t.TempDir()
	if err := userWindowsPrivate(real); err != nil {
		t.Fatal(err)
	}
	link := real + "-link"
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink creation requires privileges: %v", err)
	}
	if _, err := UninstallUserShellInstall(home, UserShellInstall{Target: link, Mode: "separate"}); err == nil {
		t.Fatal("uninstall accepted a reparse-point target")
	}
	if _, err := os.Lstat(real); err != nil {
		t.Fatalf("refused target was modified: %v", err)
	}
	if _, err := UninstallUserShellInstall(home, UserShellInstall{Target: real, Mode: "agent"}); err == nil {
		t.Fatal("uninstall accepted a non-Separate mode")
	}
}
