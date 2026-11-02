//go:build !windows

package echo

import "os/exec"

func configureMCPProcess(cmd *exec.Cmd, setup bool) {}
