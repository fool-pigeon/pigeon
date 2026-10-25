// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

//go:build darwin

package service

import "os/exec"

func (r revealer) Open(path string) error {
	return exec.Command("open", "-R", path).Start()
}
