package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ThinkerQAQ/openecho/internal/echo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func apiToken(path string) (string, error) {
	if token := os.Getenv("OPENECHO_API_TOKEN"); token != "" {
		if len(token) < 32 {
			return "", echo.ErrInvalid
		}
		return token, nil
	}
	saved, err := os.ReadFile(path)
	if err == nil {
		token := strings.TrimSpace(string(saved))
		if len(token) < 32 {
			return "", echo.ErrInvalid
		}
		return token, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	_, writeErr := f.WriteString(token + "\n")
	closeErr := f.Close()
	if writeErr != nil {
		return "", writeErr
	}
	return token, closeErr
}
func main() {
	addr := flag.String("listen", "127.0.0.1:18440", "loopback listen address")
	dbPath := flag.String("db", "data/openecho.sqlite", "private SQLite database path")
	restore := flag.String("restore", "", "restore a backup to a NEW -db path and exit")
	stdio := flag.Bool("mcp-stdio", false, "serve MCP over local stdio instead of HTTP")
	githubAuth := flag.String("github-auth", "env", "GitHub credentials: env or gh (explicit CLI fallback)")
	interval := flag.Duration("collect-interval", 24*time.Hour, "scheduled GitHub interval; 0 disables scheduling; minimum 1 minute")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *restore != "" {
		if err := echo.RestoreDatabase(ctx, *restore, *dbPath); err != nil {
			log.Fatal("restore failed: ", err)
		}
		log.Print("Backup restored to ", *dbPath)
		return
	}
	if *interval != 0 && *interval < time.Minute {
		log.Fatal("collection interval must be 0 or at least one minute")
	}
	if *githubAuth != "env" && *githubAuth != "gh" {
		log.Fatal("github-auth must be env or gh")
	}
	store, err := echo.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" && *githubAuth == "gh" {
		child, cancel := context.WithTimeout(ctx, 5*time.Second)
		out, err := exec.CommandContext(child, "gh", "auth", "token").Output()
		cancel()
		if err != nil {
			log.Fatal("could not read signed-in GitHub CLI credential")
		}
		ghToken = strings.TrimSpace(string(out))
	}
	service := echo.Service{Store: store, GitHub: echo.GitHub{Token: ghToken}}
	if *interval > 0 {
		go service.RunScheduler(ctx, *interval, func(err error) { log.Print(err) })
	}
	if *stdio {
		if err = service.MCPServer().Run(ctx, &mcp.StdioTransport{}); err != nil && ctx.Err() == nil {
			log.Fatal(err)
		}
		return
	}
	host, _, err := net.SplitHostPort(*addr)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		log.Fatal("listen must use a loopback IP address")
	}
	tokenPath := filepath.Join(filepath.Dir(*dbPath), "api-token")
	token, err := apiToken(tokenPath)
	if err != nil {
		log.Fatal("could not load private API token: ", err)
	}
	if os.Getenv("OPENECHO_API_TOKEN") == "" {
		log.Print("Local API token file: ", tokenPath)
	}
	server := &http.Server{Addr: *addr, Handler: echo.Handler(service, token), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, WriteTimeout: 100 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	log.Printf("OpenEcho listening at http://%s (MCP: /mcp)", *addr)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
