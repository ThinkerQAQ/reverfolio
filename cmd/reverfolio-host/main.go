package main

// One-shot Chromium native messaging host. This executable deliberately does not
// read the user's SQLite database or API token; it only checks and starts the
// separate loopback Reverfolio service on demand.
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ThinkerQAQ/reverfolio/internal/echo"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	nativeHostName   = "com.reverfolio.bridge"
	maxNativeMessage = 8 * 1024
	extensionOrigin  = "chrome-extension://khdciakflcmilakmaiejamcgmggnkego/"
)

type nativeRequest struct {
	Command string `json:"command"`
}
type nativeReply struct {
	OK    bool   `json:"ok"`
	State string `json:"state,omitempty"`
	Error string `json:"error,omitempty"`
}

func readFrame(r io.Reader) ([]byte, error) {
	var n uint32
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return nil, err
	}
	if n == 0 || n > maxNativeMessage {
		return nil, errors.New("invalid message length")
	}
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}
func writeFrame(w io.Writer, v nativeReply) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err = binary.Write(w, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
func serve(r io.Reader, w io.Writer, handle func(string) nativeReply) error {
	buf, err := readFrame(r)
	if err != nil {
		return err
	}
	var request nativeRequest
	if err = json.Unmarshal(buf, &request); err != nil {
		return writeFrame(w, nativeReply{Error: "invalid request"})
	}
	return writeFrame(w, handle(request.Command))
}
func healthURL(configPath string) (string, error) {
	c, err := echo.LoadConfig(configPath)
	if err != nil {
		return "", err
	}
	origin, err := echo.LocalAPIOrigin(c.Server.Listen)
	return origin + "/healthz", err
}
func healthy() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	endpoint, err := healthURL(filepath.Join(filepath.Dir(filepath.Dir(exe)), "config.toml"))
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 750 * time.Millisecond}
	resp, err := client.Get(endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var d map[string]string
	if err = json.NewDecoder(io.LimitReader(resp.Body, 256)).Decode(&d); err != nil {
		return false
	}
	return d["status"] == "ok"
}
func ensure() nativeReply {
	if healthy() {
		return nativeReply{OK: true, State: "running"}
	}
	host, err := os.Executable()
	if err != nil {
		return nativeReply{Error: "host executable unavailable"}
	}
	root := filepath.Dir(filepath.Dir(host))
	exe := filepath.Join(root, "bin", "reverfolio.exe")
	cfg := filepath.Join(root, "config.toml")
	if _, err = os.Stat(exe); err != nil {
		return nativeReply{Error: "Reverfolio executable not installed"}
	}
	if _, err = os.Stat(cfg); err != nil {
		return nativeReply{Error: "Reverfolio config.toml not found"}
	}
	cmd := exec.Command(exe, "-config", cfg)
	cmd.Dir = root
	configureDetached(cmd)
	// Never inherit the native-messaging stdout. A server writing to it would
	// corrupt the framed protocol response.
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		return nativeReply{Error: "unable to start local service"}
	}
	_ = cmd.Process.Release()
	deadline := time.Now().Add(9 * time.Second)
	for time.Now().Before(deadline) {
		if healthy() {
			return nativeReply{OK: true, State: "started"}
		}
		time.Sleep(250 * time.Millisecond)
	}
	return nativeReply{Error: "service startup timed out; check the local config and port"}
}
func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--open-ui" || os.Args[1] == "--guide-extension") {
		if err := openHostGUI(os.Args[1]); err != nil {
			notifyLaunchError(err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(echo.CurrentVersion)
		return
	}
	if len(os.Args) == 2 && os.Args[1] == "--ensure-service" {
		if !ensure().OK {
			os.Exit(1)
		}
		return
	}
	// The OS/browser manifest pins allowed_origins, and this extra check rejects
	// an unexpected caller in Chromium's documented argv[1] form.
	if len(os.Args) > 1 && os.Args[1] != "--native-host" && os.Args[1] != extensionOrigin {
		_ = writeFrame(os.Stdout, nativeReply{Error: "extension origin rejected"})
		return
	}
	if err := serve(os.Stdin, os.Stdout, func(command string) nativeReply {
		switch command {
		case "status":
			if healthy() {
				return nativeReply{OK: true, State: "running"}
			}
			return nativeReply{OK: true, State: "stopped"}
		case "ensure_bridge":
			return ensure()
		default:
			return nativeReply{Error: "unsupported command"}
		}
	}); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "native host request failed")
	}
}

// openHostGUI uses a native GUI executable to avoid powershell.exe, cmd.exe,
// or an intermediate console when launched from Qt IFW or the Start Menu.
func openHostGUI(action string) error {
	host, err := os.Executable()
	if err != nil {
		return err
	}
	root := filepath.Dir(filepath.Dir(host))
	if action == "--guide-extension" {
		folder := filepath.Join(root, "extension")
		if _, err := os.Stat(folder); err != nil {
			return err
		}
		if err := shellOpen(folder); err != nil {
			return err
		}
		guide := filepath.Join(root, "README_ZH.md")
		if _, err := os.Stat(guide); err == nil {
			return shellOpen(guide)
		}
		return nil
	}
	if action != "--open-ui" {
		return errors.New("unsupported GUI action")
	}
	cfg := filepath.Join(root, "config.toml")
	if _, err := os.Stat(cfg); errors.Is(err, os.ErrNotExist) {
		sample, readErr := os.ReadFile(filepath.Join(root, "config.example.toml"))
		if readErr != nil {
			return readErr
		}
		output, createErr := os.OpenFile(cfg, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if errors.Is(createErr, os.ErrExist) {
		} else if createErr != nil {
			return createErr
		} else {
			if _, writeErr := output.Write(sample); writeErr != nil {
				output.Close()
				return writeErr
			}
			if closeErr := output.Close(); closeErr != nil {
				return closeErr
			}
		}
	} else if err != nil {
		return err
	}
	status := ensure()
	if !status.OK {
		return fmt.Errorf("Reverfolio local service: %s", status.Error)
	}
	endpoint, err := healthURL(cfg)
	if err != nil {
		return err
	}
	return shellOpen(strings.TrimSuffix(endpoint, "/healthz"))
}
