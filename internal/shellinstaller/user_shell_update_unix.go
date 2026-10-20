//go:build linux || darwin

package shellinstaller

import (
	"context"
	"errors"
	"os"
)

// UserShellUpdate is channel-scoped update availability. It never crosses the
// installed Stable/Main channel and never changes preferences.
type UserShellUpdate struct {
	Channel    string
	Available  bool
	Components []string
	Detail     string
}

func CheckUserShellUpdate(ctx context.Context, install UserShellInstall) (UserShellUpdate, error) {
	manifest, err := userReadInstallationManifest(install.Target)
	if err != nil {
		return UserShellUpdate{}, err
	}
	update := UserShellUpdate{Channel: "stable"}
	if manifest.MainSDK != nil {
		update.Channel = "main-preview"
		latest, _, err := userMainSnapshot(ctx)
		if err != nil {
			return update, err
		}
		if latest.SHA != manifest.MainSDK.SHA {
			update.Available = true
			update.Components = []string{"SDK Main snapshot"}
			update.Detail = "new Main snapshot available"
		}
		return update, nil
	}
	self, err := os.Executable()
	if err != nil {
		return update, err
	}
	supervisor, err := userSupervisorSHA(ctx, self)
	if err != nil {
		return update, err
	}
	if supervisor != manifest.SupervisorSHA {
		update.Available = true
		update.Components = append(update.Components, "Gentle Shell supervisor")
	}
	if manifest.NodeSHA != privateNativeNodeSHA {
		update.Available = true
		update.Components = append(update.Components, "Node runtime")
	}
	if update.Available {
		update.Detail = "stable channel components differ from this build"
	}
	return update, nil
}

var ErrUserShellUpdateUnavailable = errors.New("update check unavailable")
