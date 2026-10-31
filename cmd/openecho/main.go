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
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/ThinkerQAQ/openecho/internal/echo"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:18440", "loopback listen address")
	dbPath := flag.String("db", "data/openecho.sqlite", "private SQLite database path")
	flag.Parse()
	host, _, err := net.SplitHostPort(*addr)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		log.Fatal("listen must use a loopback IP address")
	}
	store, err := echo.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	token := os.Getenv("OPENECHO_API_TOKEN")
	if token == "" {
		tokenPath := filepath.Join(filepath.Dir(*dbPath), "api-token")
		saved, readErr := os.ReadFile(tokenPath)
		if readErr == nil {
			token = strings.TrimSpace(string(saved))
		} else if os.IsNotExist(readErr) {
			buf := make([]byte, 32)
			if _, err = rand.Read(buf); err != nil {
				log.Fatal(err)
			}
			token = hex.EncodeToString(buf)
			f, createErr := os.OpenFile(tokenPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if createErr != nil {
				log.Fatal(createErr)
			}
			_, writeErr := f.WriteString(token + "\n")
			closeErr := f.Close()
			if writeErr != nil || closeErr != nil {
				log.Fatal("could not save local API token")
			}
		} else {
			log.Fatal("could not read local API token")
		}
		if len(token) < 32 {
			log.Fatal("local API token must contain at least 32 characters")
		}
		log.Printf("Local API token file: %s", tokenPath)
	}
	if len(token) < 32 {
		log.Fatal("OPENECHO_API_TOKEN must contain at least 32 characters")
	}
	service := echo.Service{Store: store, GitHub: echo.GitHub{Token: os.Getenv("GITHUB_TOKEN")}}
	server := &http.Server{Addr: *addr, Handler: echo.Handler(service, token), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, WriteTimeout: 100 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()
	log.Printf("OpenEcho listening at http://%s", *addr)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
