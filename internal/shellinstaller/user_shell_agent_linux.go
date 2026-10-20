//go:build linux || darwin

package shellinstaller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// An operator-owned Pi agent may be group-readable but must never be writable
// by another user. The prefix is not reused in this mode, so Homebrew/system
// prefixes stay untouched.
func userOwnedIdentity(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0022 != 0 || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Getuid()) {
		return "", privateError("filesystem", err)
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil || canonical != path {
		return "", privateError("filesystem", err)
	}
	st := info.Sys().(*syscall.Stat_t)
	return fmt.Sprintf("%d:%d:%d:%v", st.Dev, st.Ino, st.Uid, info.Mode()), nil
}

// userStageOwnedAgentContext7 merges the selected Context7 entry into the
// existing Pi agent, after writing the untouched preimage into the published
// target's recovery evidence. Only the selected fields are replaced.
func userStageOwnedAgentContext7(ctx context.Context, agent, choice, stage string) error {
	if choice == "" {
		return nil
	}
	if ctx == nil || ctx.Err() != nil {
		return privateError("canceled", context.Canceled)
	}
	before, err := userOwnedIdentity(agent)
	if err != nil {
		return err
	}
	path := filepath.Join(agent, "mcp.json")
	original := []byte("{}\n")
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() || info.Sys().(*syscall.Stat_t).Nlink != 1 || info.Mode().Perm()&0022 != 0 || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Getuid()) || info.Size() > 1<<20 {
			return privateError("preimage", nil)
		}
		original, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(original, &fields); err != nil {
		return errors.New("agent mcp.json is not a JSON object")
	}
	if err := userContext7Update(choice)(fields); err != nil {
		return err
	}
	data, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}
	backup := filepath.Join(stage, "state/agent.preimage")
	if err := os.MkdirAll(backup, 0700); err != nil {
		return err
	}
	if err := userToolWrite(filepath.Join(backup, "mcp.json"), original, 0600); err != nil {
		return err
	}
	temporary := path + ".new"
	if _, err := os.Lstat(temporary); !os.IsNotExist(err) {
		return privateError("preimage", nil)
	}
	if err := os.WriteFile(temporary, append(data, '\n'), 0600); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		return err
	}
	directory, err := os.Open(agent)
	if err != nil {
		return err
	}
	if err := directory.Sync(); err != nil {
		directory.Close()
		return err
	}
	if err := directory.Close(); err != nil {
		return err
	}
	after, err := userOwnedIdentity(agent)
	if err != nil || after != before {
		return privateError("preimage", err)
	}
	return nil
}

// userSnapshotOwnedAgentSettings retains the operator agent's settings inside
// the target before the legacy provisioning writes its stock declaration.
func userSnapshotOwnedAgentSettings(ctx context.Context, agent, root string) error {
	if ctx == nil || ctx.Err() != nil {
		return privateError("canceled", context.Canceled)
	}
	if _, err := userOwnedIdentity(agent); err != nil {
		return err
	}
	path := filepath.Join(agent, "settings.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil || len(data) > 1<<20 {
		return privateError("preimage", err)
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0022 != 0 || info.Sys().(*syscall.Stat_t).Nlink != 1 || info.Sys().(*syscall.Stat_t).Uid != uint32(os.Getuid()) {
		return privateError("preimage", err)
	}
	backup := filepath.Join(root, "state/agent.preimage")
	if err := os.MkdirAll(backup, 0700); err != nil {
		return err
	}
	return userToolWrite(filepath.Join(backup, "settings.json"), data, 0600)
}
