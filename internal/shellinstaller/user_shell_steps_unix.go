//go:build linux || darwin

package shellinstaller

import (
	"fmt"
	"os"
)

// userShellStep prints one completed plan step only when the TUI requested the
// plan output. It never changes installation authority or ordering.
func userShellStep(index, total int, label string) {
	if os.Getenv("GENTLE_AI_SHELL_STEPS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "✓ %d/%d %s\n", index, total, label)
}
