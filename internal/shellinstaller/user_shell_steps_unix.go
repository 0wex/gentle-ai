//go:build linux || darwin

package shellinstaller

import (
	"context"
	"fmt"
	"os"
)

// UserInstallProgress reports one ordered installation step to an in-process
// owner. It carries no installation authority and never substitutes for the
// inspected confirmation.
type UserInstallProgress struct {
	Index, Total int
	Label, State string
}

type userInstallProgressKey struct{}

// WithUserInstallProgress attaches an ordered progress observer. The callback
// runs synchronously on the installer goroutine and must not block.
func WithUserInstallProgress(ctx context.Context, report func(UserInstallProgress)) context.Context {
	if report == nil {
		return ctx
	}
	return context.WithValue(ctx, userInstallProgressKey{}, report)
}

func userProgressEnvironment(ctx context.Context) []string {
	if os.Getenv("GENTLE_AI_SHELL_STEPS") == "1" {
		return []string{"GENTLE_AI_SHELL_STEPS=1"}
	}
	if ctx == nil {
		return nil
	}
	if _, ok := ctx.Value(userInstallProgressKey{}).(func(UserInstallProgress)); ok {
		// The owned service is a separate process: its stdout is piped back and
		// parsed by the in-process owner.
		return []string{"GENTLE_AI_SHELL_STEPS=1"}
	}
	return nil
}

func userInstallReport(ctx context.Context, progress UserInstallProgress) {
	if ctx == nil {
		return
	}
	if report, ok := ctx.Value(userInstallProgressKey{}).(func(UserInstallProgress)); ok {
		report(progress)
	}
}

// userShellStep prints one completed plan step only when the TUI requested the
// plan output. It never changes installation authority or ordering.
func userShellStepBegin(ctx context.Context, index, total int, label string) {
	userInstallReport(ctx, UserInstallProgress{Index: index, Total: total, Label: label, State: "running"})
}

func userShellStep(ctx context.Context, index, total int, label string) {
	userInstallReport(ctx, UserInstallProgress{Index: index, Total: total, Label: label, State: "done"})
	if os.Getenv("GENTLE_AI_SHELL_STEPS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "✓ %d/%d %s\n", index, total, label)
}
