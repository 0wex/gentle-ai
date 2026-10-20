//go:build linux || darwin

package shellinstaller

import (
	"os"
	"path/filepath"
	"strings"
)

// UserPiDetection reports the single local Pi installation the installer may
// reuse. Detection never guesses: ambiguous or pnpm-style layouts stay absent.
type UserPiDetection struct {
	Prefix  string
	Agent   string
	Found   bool
	Problem string // "missing" or "ambiguous"; empty when found
}

// DetectUserPi looks only at bounded local candidates: NPM_CONFIG_PREFIX and
// the conventional per-user prefixes, paired with ~/.pi/agent. Exactly one
// npm-style Pi bundle must match; otherwise the caller must not select Shared.
func DetectUserPi(home string) UserPiDetection {
	if home == "" || !filepath.IsAbs(home) {
		return UserPiDetection{Problem: "missing"}
	}
	agent := filepath.Join(home, ".pi", "agent")
	if !userDetectDirectory(agent) {
		return UserPiDetection{Problem: "missing"}
	}
	var prefixes []string
	seen := map[string]bool{}
	for _, candidate := range []string{os.Getenv("NPM_CONFIG_PREFIX"), filepath.Join(home, ".npm-global"), filepath.Join(home, ".local")} {
		if candidate == "" || !filepath.IsAbs(candidate) || filepath.Clean(candidate) != candidate || seen[candidate] {
			continue
		}
		seen[candidate] = true
		bundle := filepath.Join(candidate, "lib/node_modules/@earendil-works/pi-coding-agent/dist/bundle/cli.js")
		if userDetectRegular(bundle) {
			prefixes = append(prefixes, candidate)
		}
	}
	switch len(prefixes) {
	case 0:
		return UserPiDetection{Agent: agent, Problem: "missing"}
	case 1:
		return UserPiDetection{Prefix: prefixes[0], Agent: agent, Found: true}
	default:
		return UserPiDetection{Agent: agent, Problem: "ambiguous"}
	}
}

func userDetectDirectory(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir() && !strings.HasSuffix(path, "/")
}

func userDetectRegular(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() > 0
}
