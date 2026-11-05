//go:build windows

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

// initializeWindowsInstall replaces the former PowerShell first-run scripts.
// Existing TOML, SQLite, and the extension access key are never replaced.
func initializeWindowsInstall(root string) error {
	if err := os.MkdirAll(filepath.Join(root, "Data"), 0700); err != nil {
		return err
	}
	cfg := filepath.Join(root, "config.toml")
	if _, err := os.Stat(cfg); errors.Is(err, os.ErrNotExist) {
		sample, readErr := os.ReadFile(filepath.Join(root, "config.example.toml"))
		if readErr != nil {
			return readErr
		}
		text := strings.ReplaceAll(string(sample), `interval = "24h"`, `interval = "0"`)
		text = strings.ReplaceAll(text, `auth = "gh"`, `auth = "env"`)
		file, createErr := os.OpenFile(cfg, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if errors.Is(createErr, os.ErrExist) {
		} else if createErr != nil {
			return createErr
		} else {
			if _, writeErr := file.WriteString(text); writeErr != nil {
				file.Close()
				return writeErr
			}
			if closeErr := file.Close(); closeErr != nil {
				return closeErr
			}
		}
	} else if err != nil {
		return err
	}
	extension := filepath.Join(root, "extension")
	manifest, err := os.ReadFile(filepath.Join(extension, "manifest.json"))
	if err != nil {
		return err
	}
	var info struct {
		ShortName   string   `json:"short_name"`
		Permissions []string `json:"permissions"`
	}
	if err := json.Unmarshal(manifest, &info); err != nil {
		return err
	}
	if info.ShortName != "Reverfolio" {
		return errors.New("unexpected extension identity")
	}
	permitted := false
	for _, v := range info.Permissions {
		if v == "nativeMessaging" {
			permitted = true
		}
	}
	if !permitted {
		return errors.New("nativeMessaging permission missing")
	}
	keyPath := filepath.Join(root, "Data", "extension-access.key")
	keyBytes, err := os.ReadFile(keyPath)
	if errors.Is(err, os.ErrNotExist) {
		generated := make([]byte, 32)
		if _, err = rand.Read(generated); err != nil {
			return err
		}
		value := hex.EncodeToString(generated) + "\n"
		file, createErr := os.OpenFile(keyPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if errors.Is(createErr, os.ErrExist) {
			keyBytes, err = os.ReadFile(keyPath)
			if err != nil {
				return err
			}
		} else if createErr != nil {
			return createErr
		} else {
			if _, err = file.WriteString(value); err != nil {
				file.Close()
				return err
			}
			if err = file.Close(); err != nil {
				return err
			}
			keyBytes = []byte(value)
		}
	} else if err != nil {
		return err
	}
	key := strings.TrimSpace(string(keyBytes))
	if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(key) {
		return errors.New("existing extension access key invalid; refusing to replace")
	}
	ui := filepath.Join(extension, "ui")
	if _, err := os.Stat(ui); err != nil {
		return err
	}
	localAuth := `"use strict";` + "\n" + `// Local-only extension UI capability. Generated from Data/extension-access.key.` + "\n" + `window.REVERFOLIO_EXTENSION_ACCESS = "` + key + `";` + "\n"
	localPath := filepath.Join(ui, "local-auth.js")
	// The extension needs to read this local capability, but data/ key is the
	// durable secret. Never print either value to stdout or error messages.
	if err := os.WriteFile(localPath, []byte(localAuth), 0600); err != nil {
		return err
	}
	host := filepath.Join(root, "bin", "reverfolio-host.exe")
	if _, err := os.Stat(host); err != nil {
		return err
	}
	nativeDir := filepath.Join(root, "Data", "native-host")
	if err := os.MkdirAll(nativeDir, 0700); err != nil {
		return err
	}
	nativePath := filepath.Join(nativeDir, nativeHostName+".json")
	data := map[string]any{
		"name": nativeHostName, "description": "Reverfolio local service launcher",
		"path": host, "type": "stdio",
		"allowed_origins": []string{extensionOrigin},
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(nativePath, append(encoded, '\n'), 0600); err != nil {
		return err
	}
	for _, name := range []string{
		"Software\\Microsoft\\Edge\\NativeMessagingHosts\\" + nativeHostName,
		"Software\\Google\\Chrome\\NativeMessagingHosts\\" + nativeHostName,
	} {
		key, _, keyErr := registry.CreateKey(registry.CURRENT_USER, name, registry.SET_VALUE|registry.QUERY_VALUE)
		if keyErr != nil {
			return keyErr
		}
		writeErr := key.SetStringValue("", nativePath)
		closeErr := key.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	// Run the existing built-in configuration sync, not the legacy .ps1,
	// and suppress the separate console process.
	mainExe := filepath.Join(root, "bin", "reverfolio.exe")
	cmd := exec.Command(mainExe, "--config", cfg, "--sync-extension-config")
	cmd.Dir = root
	configureDetached(cmd)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sync extension endpoint: %w", err)
	}
	return nil
}
