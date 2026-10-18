//go:build linux && amd64

package shellinstaller

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// PublishNativePNPMInstallation stages the original producer's retained DATA and
// publishes it without replacing an existing destination. The caller must own
// the qualified private execution boundary and the source's custody. This does
// not execute suppliers or establish launcher/installed-product acceptance.
func PublishNativePNPMInstallation(ctx context.Context, source, stage, destination string, location NativeInstallationLocation) (UserInstallResult, error) {
	return PublishNativePNPMInstallationWithExperience(ctx, source, stage, destination, location, nil)
}

// PublishNativePNPMInstallationWithExperience applies canonical base preferences
// only to the copied, unpublished stage. It never changes the retained producer
// or overwrites existing preferences. Unsupported optional effects are refused
// before any staging; nil preserves the unprofiled publication contract.
func PublishNativePNPMInstallationWithExperience(ctx context.Context, source, stage, destination string, location NativeInstallationLocation, profile *UserExperience) (result UserInstallResult, err error) {
	if err = userValidateStageExperience(profile); err != nil {
		return result, err
	}
	self, err := userExecutable()
	if err != nil {
		return result, err
	}
	staged, err := StageNativePNPMInstallation(ctx, source, stage, destination, location)
	if err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			// Rename consumes stage. Never remove a published destination, even
			// when its final readback or parent-directory synchronization fails.
			err = errors.Join(err, os.RemoveAll(stage))
		}
	}()
	if err = userStageExperience(ctx, stage, profile); err != nil {
		return result, err
	}
	supervisorSHA, err := userCopySupervisor(ctx, self, filepath.Join(stage, "supervisor"))
	if err != nil {
		return result, err
	}
	if err := os.Mkdir(filepath.Join(stage, "bin"), 0700); err != nil {
		return result, err
	}
	prefix, agent := filepath.Join(stage, "pnpm"), filepath.Join(stage, "agent")
	manifest := userManifest{
		Schema: userSchema, Destination: destination, Mode: "separate",
		Prefix: filepath.Join(destination, "pnpm"), Agent: filepath.Join(destination, "agent"), SupervisorSHA: supervisorSHA,
		PNPM: &userPNPMManifest{PackageRoot: strings.TrimSuffix(staged.Installation.ShellCLI, "/bin/gentle-shell.mjs"), ShellVersion: location.ShellVersion, PeerVersion: location.PiVersion, CodeSeal: staged.CodeSeal},
	}
	manifest.PrefixIdentity, err = userIdentity(prefix)
	if err == nil {
		manifest.AgentIdentity, err = userIdentity(agent)
	}
	if err != nil {
		return result, err
	}
	return userPublishManifest(ctx, stage, prefix, agent, manifest)
}
