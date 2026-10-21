package screens

import (
	"strings"
	"testing"
)

func TestShellRoseFramesDecode(t *testing.T) {
	if ShellBloomFrames != len(shellRoseFrames) || ShellBloomFrames == 0 {
		t.Fatalf("bloom frames = %d, data = %d", ShellBloomFrames, len(shellRoseFrames))
	}
	for frame := 0; frame < ShellBloomFrames; frame++ {
		lines := shellRoseFrameLines(frame)
		if len(lines) != len(shellRoseFrames[frame]) || !strings.Contains(lines[0], "\x1b[") {
			t.Fatalf("frame %d did not decode: %d lines", frame, len(lines))
		}
	}
	steps := []ShellInstallStep{{Label: "Validate", State: "done"}, {Label: "Prepare", State: "done"}, {Label: "Runtime", State: "done"}, {Label: "Optional", State: "done"}, {Label: "Publish", State: "done"}}
	view := RenderShellInstall(steps, "", true, 0, ShellBloomFrames-1, true, false, "separate", "gentle-shell", 120, 35)
	if !strings.Contains(view, "Welcome to the gentle-tailored experience") || !strings.Contains(view, "gentle-shell") || !strings.Contains(view, "▀") || !strings.Contains(view, "Enter: finish") {
		t.Fatalf("truecolor rose or completion help is missing:\n%s", view)
	}
}
