//go:build windows

package reveal

import "os/exec"

func Open(path string) error {
	return exec.Command("explorer", "/select,", path).Start()
}
