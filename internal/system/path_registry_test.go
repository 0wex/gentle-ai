package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Execute the public operations' actual PowerShell against an isolated registry
// key. Legacy Environment calls are redirected too, so RED never edits real PATH.
func TestUserPathPreservesRegistryValue(t *testing.T) {
	if runtime.GOOS != "windows" || testing.Short() {
		t.Skip("requires Windows registry and PowerShell")
	}
	host, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Fatal(err)
	}
	const target = `C:\O'Brien\bin`
	const raw = `%USERPROFILE%\bin;%GENTLE_AI_UNDEFINED_5581%\tools`
	for _, kind := range []string{"ExpandString", "String", "missing"} {
		for _, operation := range []string{"add", "add existing", "remove", "prioritize", "rollback"} {
			t.Run(kind+"/"+operation, func(t *testing.T) {
				t.Setenv("PATH", os.Getenv("PATH"))
				initial, want := raw, target+";"+raw
				if operation != "add" {
					initial = raw + ";" + target
				}
				if operation == "add existing" {
					want = initial
				}
				if operation == "remove" || operation == "rollback" {
					want = raw
				}
				wantKind := kind
				if kind == "missing" {
					initial = ""
					wantKind = "ExpandString"
					want = target
					if operation == "remove" || operation == "rollback" {
						want = ""
					}
				}
				key := fmt.Sprintf(`Software\GentleAI-PathTest-%d-%d`, os.Getpid(), time.Now().UnixNano())
				useWindowsUserPathSeam(t, userPathRunnerFunc(func(ctx context.Context, args ...string) ([]byte, error) {
					script := args[len(args)-1]
					// Redirect the old API for the regression's RED run.
					script = strings.ReplaceAll(script, `[Environment]::GetEnvironmentVariable('PATH', 'User')`, `$fixture.GetValue('Path')`)
					script = strings.ReplaceAll(script, `[Environment]::SetEnvironmentVariable('PATH', `, `$fixture.SetValue('Path', `)
					script = strings.ReplaceAll(script, `, 'User')`, `, [Microsoft.Win32.RegistryValueKind]::String)`)
					// Redirect the fixed registry implementation, not the real Environment key.
					script = strings.ReplaceAll(script, `CreateSubKey('Environment')`, `CreateSubKey('`+key+`')`)
					if strings.Contains(script, `GetEnvironmentVariable`) || strings.Contains(script, `SetEnvironmentVariable`) || strings.Contains(script, `SubKey('Environment'`) {
						t.Fatal("unsafe registry script: fixture redirection incomplete")
					}
					setup := ""
					if kind != "missing" {
						setup = fmt.Sprintf(`$fixture.SetValue('Path', '%s', [Microsoft.Win32.RegistryValueKind]::%s); `, escapePowerShellString(initial), kind)
					}
					wrapped := fmt.Sprintf(`$ErrorActionPreference = 'Stop'; $fixture = [Microsoft.Win32.Registry]::CurrentUser.CreateSubKey('%s'); try { %s & { %s }; $actual = $fixture.GetValue('Path', $null, [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames); if ($actual -cne '%s') { throw "raw PATH mismatch: $actual" }; if ($fixture.GetValueKind('Path').ToString() -ne '%s') { throw "kind mismatch: $($fixture.GetValueKind('Path'))" } } finally { $fixture.Close(); [Microsoft.Win32.Registry]::CurrentUser.DeleteSubKeyTree('%s') }`, key, setup, script, escapePowerShellString(want), wantKind, key)
					cmd := exec.CommandContext(ctx, host, "-NoProfile", "-NonInteractive", "-Command", wrapped)
					out, err := cmd.CombinedOutput()
					if err != nil {
						return out, fmt.Errorf("registry probe: %w: %s", err, out)
					}
					wantOutput := ""
					if operation == "add" || operation == "add existing" {
						wantOutput = "changed"
						if operation == "add existing" && kind != "missing" {
							wantOutput = "unchanged"
						}
					}
					if strings.TrimSpace(string(out)) != wantOutput {
						t.Fatalf("PowerShell output = %q, want %q", out, wantOutput)
					}
					return out, nil
				}))
				var err error
				switch operation {
				case "add", "add existing":
					var addition UserPathAddition
					addition, err = AddToUserPathWithResult(target)
					if err == nil && addition.PersistentAdded != (operation == "add" || kind == "missing") {
						t.Fatalf("unexpected ownership: %+v", addition)
					}
				case "remove":
					err = RemoveFromUserPath(target)
				case "prioritize":
					err = PrioritizeUserPath(target)
				case "rollback":
					err = RollbackUserPathAddition(target, UserPathAddition{PersistentAdded: true})
				}
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
