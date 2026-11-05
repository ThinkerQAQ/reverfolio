//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	createNewProcessGroup = 0x00000200
	createNoWindow        = 0x08000000
)

func configureDetached(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNoWindow | createNewProcessGroup,
		HideWindow:    true,
	}
}

var shell32 = syscall.NewLazyDLL("shell32.dll")
var shellExecute = shell32.NewProc("ShellExecuteW")
var user32 = syscall.NewLazyDLL("user32.dll")
var messageBox = user32.NewProc("MessageBoxW")

func shellOpen(target string) error {
	verb, _ := syscall.UTF16PtrFromString("open")
	item, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	r, _, callErr := shellExecute.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(item)), 0, 0, 1)
	if r <= 32 {
		return fmt.Errorf("ShellExecuteW error %d: %v", r, callErr)
	}
	return nil
}
func notifyLaunchError(err error) {
	if err == nil {
		return
	}
	title, _ := syscall.UTF16PtrFromString("Reverfolio")
	msg, _ := syscall.UTF16PtrFromString(err.Error())
	messageBox.Call(0, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0x10)
}
