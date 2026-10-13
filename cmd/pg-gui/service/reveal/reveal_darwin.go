//go:build darwin

package reveal

import "os/exec"

func Open(path string) error {
	return exec.Command("open", "-R", path).Start()
}
