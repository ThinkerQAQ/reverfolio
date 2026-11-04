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

	"github.com/ThinkerQAQ/reverfolio/internal/echo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func configuredAPIToken() string {
	if token := os.Getenv("REVERFOLIO_API_TOKEN"); token != "" {
		return token
	}
	return os.Getenv("OPENECHO_API_TOKEN")
}

func apiToken(path string) (string, error) {
	if token := configuredAPIToken(); token != "" {
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
	if len(os.Args) > 1 && (os.Args[1] == "mcp" || os.Args[1] == "tunnel-run") {
		if err := runMCPCommand(os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	addr := flag.String("listen", echo.DefaultConfig().Server.Listen, "loopback listen address")
	dbPath := flag.String("db", "data/reverfolio.sqlite", "private SQLite database path")
	restore := flag.String("restore", "", "restore a backup to a NEW -db path and exit")
	stdio := flag.Bool("mcp-stdio", false, "serve MCP over local stdio instead of HTTP")
	githubAuth := flag.String("github-auth", "env", "GitHub credentials: env or gh (explicit CLI fallback)")
	interval := flag.Duration("collect-interval", 24*time.Hour, "scheduled GitHub interval; 0 disables scheduling; minimum 1 minute")
	configPath := flag.String("config", "", "path to private TOML configuration (auto-detect next to bin directory)")
	syncExtension := flag.Bool("sync-extension-config", false, "synchronize nonsecret extension endpoint from config.toml and exit")
	flag.Parse()
	followCollectionConfig := true
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "collect-interval" {
			followCollectionConfig = false
		}
	})
	visitedDB := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "db" {
			visitedDB = true
		}
	})
	if !visitedDB {
		if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
			if _, legacyErr := os.Stat("data/openecho.sqlite"); legacyErr == nil {
				*dbPath = "data/openecho.sqlite"
			}
		}
	}
	if *configPath == "" {
		if exe, exeErr := os.Executable(); exeErr == nil {
			candidate := filepath.Join(filepath.Dir(exe), "..", "config.toml")
			if _, statErr := os.Stat(candidate); statErr == nil {
				*configPath = candidate
			}
		}
	}
	if *configPath != "" {
		if err := echo.SetMCPConfigPath(*configPath); err != nil {
			log.Fatal(err)
		}
	}
	settings := echo.NewConfigManager(*configPath)
	config, err := settings.Read()
	if err != nil {
		log.Fatal("invalid Reverfolio config.toml: ", err)
	}
	if *configPath != "" {
		visited := map[string]bool{}
		flag.Visit(func(f *flag.Flag) { visited[f.Name] = true })
		if !visited["listen"] {
			*addr = config.Server.Listen
		}
		if !visited["github-auth"] {
			*githubAuth = config.GitHub.Auth
		}
		if !visited["collect-interval"] {
			parsed, parseErr := time.ParseDuration(config.Collection.Interval)
			if parseErr != nil {
				log.Fatal("invalid TOML collection.interval")
			}
			*interval = parsed
		}
		if !visited["db"] {
			*dbPath = config.Storage.Database
			if !filepath.IsAbs(*dbPath) {
				*dbPath = filepath.Join(filepath.Dir(*configPath), *dbPath)
			}
		}
	}
	if *syncExtension {
		if *configPath == "" {
			log.Fatal("config path required")
		}
		if err := echo.SyncExtensionEndpoint(*configPath, config); err != nil {
			log.Fatal("extension endpoint synchronization failed: ", err)
		}
		return
	}
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
			// GitHub authentication is optional for the local workspace.
			// Report missing CLI credentials through collection results rather
			// than preventing the UI and MCP service from starting.
			log.Print("GitHub CLI credentials unavailable; authorized GitHub metrics will require authentication")
		} else {
			ghToken = strings.TrimSpace(string(out))
		}
	}
	service := echo.Service{Store: store, GitHub: echo.GitHub{Token: ghToken}, Config: settings}
	if !*stdio {
		go service.RunScheduler(ctx, *interval, followCollectionConfig, func(err error) { log.Print(err) })
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
	// Local extension capability lives beside SQLite, independent of MCP credentials.
	extPath := filepath.Join(filepath.Dir(*dbPath), "extension-access.key")
	if data, readErr := os.ReadFile(extPath); readErr == nil {
		extToken := strings.TrimSpace(string(data))
		if len(extToken) < 48 || len(extToken) > 256 {
			log.Fatal("invalid private extension access key")
		}
		service.ExtensionToken = extToken
	} else if !os.IsNotExist(readErr) {
		log.Fatal("could not read private extension access key")
	}
	if configuredAPIToken() == "" {
		log.Print("Local API token file: ", tokenPath)
	}
	server := &http.Server{Addr: *addr, Handler: echo.Handler(service, token), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, WriteTimeout: 100 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	log.Printf("Reverfolio listening at http://%s (MCP: /mcp)", *addr)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
