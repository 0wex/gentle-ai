//go:build windows

package shellinstaller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

func windowsLocalAppData(home string) string {
	if value := os.Getenv("LOCALAPPDATA"); value != "" {
		return value
	}
	return filepath.Join(home, "AppData", "Local")
}

func DiscoverUserShellInstalls(home string) []UserShellInstall {
	target := filepath.Join(windowsLocalAppData(home), "gentle-shell", "install")
	if _, err := userWindowsManifestRead(target); err != nil {
		return nil
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil
	}
	channel := "stable"
	if data, readErr := userWindowsRead(filepath.Join(target, "selection.json"), 4096); readErr == nil {
		var selection struct{ Channel string }
		if json.Unmarshal(data, &selection) == nil && selection.Channel == "main" {
			channel = "main"
		}
	}
	return []UserShellInstall{{
		Target: target, Mode: "separate", Channel: channel, UpdatedAt: info.ModTime(),
		Prefix: filepath.Join(target, "prefix"), Agent: filepath.Join(target, "agent"),
	}}
}

func ReadUserShellInstalls(home string) []UserShellInstall { return DiscoverUserShellInstalls(home) }

func loadWindowsJSON(path string, value any) (bool, error) {
	data, err := userWindowsRead(path, 1<<20)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, value); err != nil {
		return false, err
	}
	return true, nil
}

func LoadUserShellExperience(install UserShellInstall) (UserExperience, error) {
	profile := DefaultUserExperience()
	profile.Persona = "unmanaged"
	profile.GGA = "off"
	var persona struct{ Mode string }
	if found, err := loadWindowsJSON(filepath.Join(install.Target, "config", "persona.json"), &persona); err != nil {
		return UserExperience{}, err
	} else if found {
		profile.Persona = persona.Mode
	}
	var background struct{ Policy string }
	if found, err := loadWindowsJSON(filepath.Join(install.Target, "config", "background-subagents.json"), &background); err != nil {
		return UserExperience{}, err
	} else if found {
		profile.BackgroundSubagents = background.Policy
	}
	var strict struct{ Policy string }
	if found, err := loadWindowsJSON(filepath.Join(install.Target, "config", "strict-tdd.json"), &strict); err != nil {
		return UserExperience{}, err
	} else if found {
		profile.StrictTDD = strict.Policy
	}
	var state struct {
		RDD string `json:"rdd_mode"`
	}
	if found, err := loadWindowsJSON(filepath.Join(install.Target, "home", ".gentle-ai", "state.json"), &state); err != nil {
		return UserExperience{}, err
	} else if found && (state.RDD == "on" || state.RDD == "off") {
		profile.GlobalRDD = state.RDD
	}
	var mcp struct {
		Servers map[string]struct{ Enabled bool } `json:"mcpServers"`
	}
	if found, err := loadWindowsJSON(filepath.Join(install.Target, "agent", "mcp.json"), &mcp); err != nil {
		return UserExperience{}, err
	} else if found {
		if server, ok := mcp.Servers["context7"]; ok {
			profile.Context7 = "off"
			if server.Enabled {
				profile.Context7 = "on"
			}
		}
	}
	if info, err := os.Stat(filepath.Join(install.Target, "runtime", "codegraph")); err == nil && info.IsDir() {
		profile.CodeGraph = "on"
	}
	if install.Channel == "main" || install.Channel == "main-preview" {
		profile.SDKSource = "main-preview"
	}
	return profile, nil
}

func RecordUserShellInstall(string, UserShellInstall) error { return nil }

func RemoveUserShellInstall(string, string) error { return nil }

func ApplyUserExperiencePreferences(UserShellInstall, UserExperience) error {
	return errors.New("managed Windows preference updates are not available in this build")
}

func UninstallUserShellInstall(_ string, install UserShellInstall) (string, error) {
	if install.Mode != "separate" || !filepath.IsAbs(install.Target) {
		return "", errors.New("Windows uninstall requires an absolute Separate installation")
	}
	if _, err := userWindowsIdentity(install.Target, true); err != nil {
		return "", err
	}
	backup := install.Target + ".removed-" + time.Now().UTC().Format("20060102T150405Z")
	if _, err := os.Lstat(backup); !os.IsNotExist(err) {
		return "", errors.Join(err, errors.New("Windows uninstall backup already exists"))
	}
	if err := os.Rename(install.Target, backup); err != nil {
		return "", err
	}
	// The active install is already fully moved aside; prune the backup to
	// the minimal retained set. A pruning failure never deletes the backup.
	pruneErr := userWindowsPruneRemovedBackup(backup)
	launcherErr := userWindowsRemoveLauncher(install.Target)
	if pruneErr != nil || launcherErr != nil {
		return backup, fmt.Errorf("installation moved to %s but minimal-backup cleanup failed; the backup keeps more than config, agent, home, selection.json and manifest.json: %w", backup, errors.Join(pruneErr, launcherErr))
	}
	return backup, nil
}

func userWindowsRemoveLauncher(target string) error {
	local, err := userWindowsLocalAppData()
	if err != nil {
		return err
	}
	link := filepath.Join(local, "Microsoft", "WindowsApps", "gentle-shell.cmd")
	data, readErr := os.ReadFile(link)
	if os.IsNotExist(readErr) {
		return nil
	}
	if readErr != nil {
		return readErr
	}
	if !bytes.Contains(data, []byte(target)) {
		return nil
	}
	return os.Remove(link)
}

func CheckUserShellUpdate(context.Context, UserShellInstall) (UserShellUpdate, error) {
	return UserShellUpdate{}, errors.New("managed Windows updates are not available in this build")
}
