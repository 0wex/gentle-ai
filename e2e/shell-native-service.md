# Linux native worker custody builder

This internal helper prepares a service argument vector; it does not launch a worker or protect the ordinary installation route by itself.

## Ordering

1. Finish private source/archive and temporary Node preparation before capturing custody.
2. Obtain the existing manager runtime through `userBusEnvironment`; create a distinct owned0700 empty mask directory under trusted parent custody.
3. Call `prepareUserNativeService`, then its `args` method immediately before the supervised launch. Both repeat the actual kernel/resource checks and authenticate the complete source against the caller-bound descriptor.
4. The later ordinary bridge must authenticate re-entry, verify active source/runtime mounts, and retain custody until all children settle. Internal selectors and environment tokens are not authority.

The builder preserves `userServiceArgs` unchanged, including terminal handling, capability drop, resource limits and stop timeout. Linux-native additions are PrivateUsers, an empty capability bounding set, RuntimeMaxSec300s, read-only retained/extracted source and cgroup control view, and an empty read-only overmount of the selected manager runtime. The bus environment must agree with that runtime.

Source/archive and mask identities are captured and rechecked; identical-content inode replacements, symlinks, unsafe ownership/modes, nonempty masks, overlapping paths and cancellation refuse without modifying data. Runtime contents may change while its physical directory identity remains fixed.

The existing462626 qualification demonstrated the mount mechanism and denied the tested parent-root/FD/namespace aliases. These unit tests validate argument construction and physical custody, not new mount execution, installer execution, publication, TUI wiring, or installed acceptance. Generic/Darwin supervision remains unchanged.
