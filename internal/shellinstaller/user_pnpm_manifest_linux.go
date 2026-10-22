//go:build linux && amd64

package shellinstaller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

func userReadPNPMManifest(ctx context.Context, root string, manifest userManifest) (userManifest, error) {
	fail := func(err error) (userManifest, error) { return manifest, privateError("readback", err) }
	p := manifest.PNPM
	if p == nil || manifest.Mode != "separate" || manifest.Prefix != filepath.Join(root, "pnpm") || manifest.Agent != filepath.Join(root, "agent") || manifest.NodeSHA != "" || manifest.MainSDK != nil || manifest.GGA || manifest.CodeGraph {
		return fail(nil)
	}
	if !filepath.IsLocal(p.PackageRoot) || filepath.Clean(p.PackageRoot) != p.PackageRoot || !strings.HasPrefix(p.PackageRoot, "pnpm/") || p.ShellVersion != "4.0.0" || p.PeerVersion != "1.2.0" {
		return fail(nil)
	}
	for _, pin := range []string{p.CodeSeal, manifest.SupervisorSHA} {
		decoded, err := hex.DecodeString(pin)
		if err != nil || len(decoded) != sha256.Size {
			return fail(err)
		}
	}
	for path, expected := range map[string]string{manifest.Prefix: manifest.PrefixIdentity, manifest.Agent: manifest.AgentIdentity} {
		actual, err := userIdentity(path)
		if err != nil || actual != expected {
			return fail(err)
		}
	}
	if _, err := privateNativeFile(ctx, filepath.Join(root, "supervisor"), 0700, -1, manifest.SupervisorSHA); err != nil {
		return fail(err)
	}
	for _, name := range []string{"gentle-shell", "pi"} {
		binding := userManifestBinding(root, name, manifest)
		if _, err := privateNativeFile(ctx, filepath.Join(root, "bin", name), 0700, int64(len(binding)), fmt.Sprintf("%x", sha256.Sum256([]byte(binding)))); err != nil {
			return fail(err)
		}
	}
	location := NativeInstallationLocation{PackageRoot: filepath.Join(root, p.PackageRoot), PNPMHome: manifest.Prefix, GlobalBin: filepath.Join(manifest.Prefix, "bin"), ShellVersion: p.ShellVersion, PiVersion: p.PeerVersion}
	if _, err := ReadNativePNPMInstallation(ctx, root, location); err != nil {
		return fail(err)
	}
	seal, err := PNPMCodeSeal(ctx, root)
	if err != nil || seal != p.CodeSeal {
		return fail(err)
	}
	return manifest, nil
}
