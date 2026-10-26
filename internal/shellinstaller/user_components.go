package shellinstaller

import (
	"sort"
	"strings"
)

// Component scope names carried by UserInstallRequest.Scope and
// UserShellUpdate.Scope. An empty scope is a full install.
const (
	UserComponentSupervisor = "supervisor"
	UserComponentNode       = "node"
	UserComponentSDK        = "sdk"
	UserComponentCodeGraph  = "codegraph"
	UserComponentContext7   = "context7"
)

// CanonicalUpdateScope renders report component names as the canonical
// request scope string for UserInstallRequest.Scope.
func CanonicalUpdateScope(components []string) string {
	set := map[string]bool{}
	for _, name := range components {
		set[name] = true
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}
