package cli

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v4/internal/shellinstaller"
)

func TestShellStepLineParsesOnlyOrderedMarkers(t *testing.T) {
	for _, test := range []struct {
		line  string
		index int
		total int
		label string
		ok    bool
	}{
		{"✓ 1/5 Validate the selection and target", 1, 5, "Validate the selection and target", true},
		{"✓ 5/5 Publish, verify and register the installation", 5, 5, "Publish, verify and register the installation", true},
		{"Installed /x; commands /x/bin/gentle-shell", 0, 0, "", false},
		{"✓ 0/5 refused", 0, 0, "", false},
		{"✓ 6/5 refused", 0, 0, "", false},
		{"noise", 0, 0, "", false},
	} {
		index, total, label, ok := shellStepLine(test.line)
		if index != test.index || total != test.total || label != test.label || ok != test.ok {
			t.Fatalf("shellStepLine(%q) = %d,%d,%q,%v", test.line, index, total, label, ok)
		}
	}
}

func TestRunShellInstallInlineRefusesUninspectedSelection(t *testing.T) {
	request := shellinstaller.UserInstallRequest{
		Destination: filepath.Join(t.TempDir(), "install"),
		Mode:        "separate",
	}
	err := RunShellInstallInline(context.Background(), ShellInstallSelection{Request: request, Confirmed: true, Executed: true}, nil, io.Discard)
	if err == nil {
		t.Fatal("inline install accepted a request without the inspected confirmation")
	}
}
