//go:build linux && amd64 && shell_pnpm_readback

package shellinstaller

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Retained original installer DATA is mounted read-only at /out. No supplier is
// executed here; a successful readback is not publication or installed health.
func TestUserPNPMReadback(t *testing.T) {
	if testing.Short() {
		t.Skip("requires retained original installer DATA in the private sandbox")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	data, err := os.ReadFile("/out/verified-location.json")
	if err != nil {
		t.Fatal(err)
	}
	var location NativeInstallationLocation
	if err := json.Unmarshal(data, &location); err != nil {
		t.Fatal(err)
	}
	got, err := ReadNativePNPMInstallation(ctx, "/out", location)
	if err != nil {
		t.Fatal(err)
	}
	if got.PiCLI == got.PeerCLI || !strings.Contains(got.PiCLI, "/1.0.0/") || !strings.Contains(got.PeerCLI, "/1.2.0/") || !strings.Contains(got.Node, "/24.21.0/") {
		t.Fatal("persistent Node, primary Pi and adjacent peer were conflated", got)
	}
	for _, target := range []string{got.Node, got.PiCLI, got.PeerCLI, got.ShellCLI} {
		if filepath.IsAbs(target) || got.Files[target] == "" {
			t.Fatal("missing relative target identity", target)
		}
	}
	if info, err := os.Stat(filepath.Join("/out", got.Node)); err != nil || info.Size() <= 32<<20 {
		t.Fatal("genuine Node must exceed the consent-tree per-file bound", info, err)
	}
	selected := map[string]bool{}
	for name := range got.Files {
		selected[name] = true
	}
	for _, name := range []string{"gentle-ai", "integrity.json"} {
		selected[filepath.Join(got.SDKRoot, "v4.0.0", name)] = true
	}
	fixture := filepath.Join(t.TempDir(), "stage")
	if err := os.Mkdir(fixture, 0700); err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir("/out/pnpm", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel("/out", path)
		dest := filepath.Join(fixture, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Mkdir(dest, info.Mode().Perm())
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(strings.Replace(target, "/out/", fixture+"/", 1), dest)
		}
		if !selected[rel] {
			return nil
		}
		from, err := os.Open(path)
		if err != nil {
			return err
		}
		defer from.Close()
		to, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
		if err != nil {
			return err
		}
		_, err = io.Copy(to, from)
		closeErr := to.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"pi", "gentle-shell"} {
		path := filepath.Join(fixture, "pnpm/bin", name)
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(body), "/out/", fixture+"/")), 0755); err != nil {
			t.Fatal(err)
		}
	}
	location.PackageRoot = strings.Replace(location.PackageRoot, "/out/", fixture+"/", 1)
	location.PNPMHome, location.GlobalBin = fixture+"/pnpm", fixture+"/pnpm/bin"
	if _, err := ReadNativePNPMInstallation(ctx, fixture, location); err != nil {
		t.Fatal("relocated real target fixture", err)
	}
	for _, tt := range []struct{ name, path, old, new string }{
		{"shim-command", "pnpm/bin/pi", "  exec node ", "  exec false "},
		{"metadata-version", filepath.Dir(got.ShellCLI) + "/../package.json", `"4.0.0"`, `"4.0.1"`},
		{"metadata-bin", filepath.Dir(got.Node) + "/../package.json", `"bin/node"`, `"bin/other"`},
		{"sdk-integrity", got.SDKRoot + "/v4.0.0/integrity.json", `"4.0.0"`, `"4.0.1"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(fixture, tt.path)
			before, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(before), tt.old) {
				t.Fatal("fixture mismatch", err)
			}
			changed := strings.ReplaceAll(string(before), tt.old, tt.new)
			if err := os.WriteFile(path, []byte(changed), 0600); err != nil {
				t.Fatal(err)
			}
			defer os.WriteFile(path, before, 0600)
			if _, err := ReadNativePNPMInstallation(ctx, fixture, location); err == nil {
				t.Fatal("changed layout accepted")
			}
			if after, _ := os.ReadFile(path); string(after) != changed {
				t.Fatal("rejected readback changed DATA")
			}
		})
	}
	for _, name := range []string{"outside-node-link", "symlink-parent-node"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(fixture, "pnpm/bin/node")
			target := "/out/" + got.Node
			if name == "symlink-parent-node" {
				alias := filepath.Join(fixture, filepath.Dir(filepath.Dir(got.Node)), "redirect")
				if err := os.Symlink("/out/"+filepath.Dir(got.Node), alias); err != nil {
					t.Fatal(err)
				}
				target = alias + "/../bin/node"
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
			if physical, err := filepath.EvalSymlinks(path); err != nil || physical != "/out/"+got.Node {
				t.Fatal("fixture must resolve to the existing external Node", physical, err)
			}
			if _, err := ReadNativePNPMInstallation(ctx, fixture, location); err == nil {
				t.Fatal("external owner accepted")
			}
			if after, _ := os.Readlink(path); after != target {
				t.Fatal("rejected link changed")
			}
		})
	}
}
