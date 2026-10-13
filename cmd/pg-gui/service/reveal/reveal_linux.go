//go:build !windows && !darwin

package reveal

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func Open(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}

	out, err := exec.Command("xdg-mime", "query", "default", "inode/directory").Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return exec.Command("xdg-open", filepath.Dir(abs)).Start()
	}

	name := strings.TrimSuffix(strings.TrimSpace(string(out)), ".desktop")
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}

	switch name {
	case "nautilus", "dolphin", "nemo", "thunar", "pcmanfm", "caja":
		return exec.Command(name, "--select", abs).Start()
	}

	if err := exec.Command(name, "--select", abs).Start(); err != nil {
		return exec.Command("xdg-open", filepath.Dir(abs)).Start()
	}
	return nil
}
