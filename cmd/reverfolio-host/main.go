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
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	nativeHostName   = "com.reverfolio.bridge"
	localHealth      = "http://127.0.0.1:19452/healthz"
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
func healthy() bool {
	client := &http.Client{Timeout: 750 * time.Millisecond}
	resp, err := client.Get(localHealth)
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
