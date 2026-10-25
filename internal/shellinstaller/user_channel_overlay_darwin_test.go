package shellinstaller

import (
	"context"
	"path/filepath"
	"testing"
)

// macOS has no Main composition runtime yet, so a channel-only reconfigure
// never enters the in-place overlay: it reports applied=false with a nil
// error and the caller keeps the full retire-and-reinstall reconfigure.
func TestUserDarwinChannelOverlayFallsBackToFullReconfigure(t *testing.T) {
	root := filepath.Join(userUnixTempRoot(t), "install")
	writeInstallFixture(t, root, userManifest{Schema: userSchema, Destination: root, Prefix: filepath.Join(root, "prefix"), Agent: filepath.Join(root, "agent"), Mode: "separate"})
	t.Setenv("GENTLE_AI_RECONFIGURE_TARGET", root)
	previousVerify, previousAcquire, previousCompose := userUnixOverlayVerify, userUnixOverlayAcquire, userUnixOverlayCompose
	t.Cleanup(func() {
		userUnixOverlayVerify, userUnixOverlayAcquire, userUnixOverlayCompose = previousVerify, previousAcquire, previousCompose
	})
	userUnixOverlayVerify = func(context.Context, string, userManifest) error {
		t.Fatal("darwin overlay verified the installation")
		return nil
	}
	userUnixOverlayAcquire = func(context.Context, string) (userSnapshotArtifact, error) {
		t.Fatal("darwin overlay acquired a Main snapshot")
		return userSnapshotArtifact{}, nil
	}
	userUnixOverlayCompose = func(context.Context, string, userManifest, userSnapshotArtifact) error {
		t.Fatal("darwin overlay composed the Main SDK")
		return nil
	}
	for _, source := range []string{"main-preview", "stable"} {
		req := UserInstallRequest{Destination: root, Mode: "separate", Experience: &UserExperience{Persona: "gentleman", BackgroundSubagents: "on", SDKSource: source}}
		result, applied, err := userUnixChannelOverlay(context.Background(), req)
		if err != nil || applied || result != (UserInstallResult{}) {
			t.Fatalf("%s reconfigure did not fall back to the full path: %+v applied=%v err=%v", source, result, applied, err)
		}
	}
}
