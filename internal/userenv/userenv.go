package userenv

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// SetUserXGOROOT sets the Windows USER environment variable XGOROOT (visible in new sessions).
// It does not modify the current process environment.
func SetUserXGOROOT(root string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("persisting XGOROOT to user env is only implemented on Windows")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return fmt.Errorf("empty XGOROOT path")
	}
	esc := strings.ReplaceAll(root, `'`, `''`)
	ps := fmt.Sprintf(`[Environment]::SetEnvironmentVariable('XGOROOT','%s','User')`, esc)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
