package executil

import "os/exec"

type OSRunner struct{}

func (OSRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}
