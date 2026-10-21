//go:build windows

package shellinstaller

import (
	"context"
	"errors"
	"time"
)

// UserShellInstall is the shared desktop projection. The Windows native
// installer owns its own manifest; registry-backed Manage actions are not part
// of the first Windows slice.
type UserShellInstall struct {
	Target    string    `json:"target"`
	Mode      string    `json:"mode"`
	Channel   string    `json:"channel"`
	UpdatedAt time.Time `json:"updated_at"`
	Prefix    string    `json:"-"`
	Agent     string    `json:"-"`
}

type UserShellUpdate struct {
	Channel    string
	Available  bool
	Components []string
	Detail     string
}

type UserPiDetection struct {
	Prefix  string
	Agent   string
	Found   bool
	Problem string
	Detail  string
}

func DetectUserPi(string) UserPiDetection { return UserPiDetection{Problem: "missing"} }

func DiscoverUserShellInstalls(string) []UserShellInstall { return nil }

func ReadUserShellInstalls(string) []UserShellInstall { return nil }

func LoadUserShellExperience(UserShellInstall) (UserExperience, error) {
	return UserExperience{}, errors.New("managed Windows installations are not available in this build")
}

func RecordUserShellInstall(string, UserShellInstall) error { return nil }

func RemoveUserShellInstall(string, string) error { return nil }

func ApplyUserExperiencePreferences(UserShellInstall, UserExperience) error {
	return errors.New("managed Windows preference updates are not available in this build")
}

func UninstallUserShellInstall(string, UserShellInstall) (string, error) {
	return "", errors.New("managed Windows uninstall is not available in this build")
}

func CheckUserShellUpdate(context.Context, UserShellInstall) (UserShellUpdate, error) {
	return UserShellUpdate{}, errors.New("managed Windows updates are not available in this build")
}
