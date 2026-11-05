//go:build !windows

package main

import (
	"os/exec"
	"runtime"
	"syscall"
)

func configureDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func shellOpen(target string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", target).Start()
	}
	return exec.Command("xdg-open", target).Start()
}
func notifyLaunchError(err error) {}

func initializeWindowsInstall(root string) error { return nil }
