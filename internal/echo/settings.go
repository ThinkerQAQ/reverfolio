package echo

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Setting struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int64  `json:"version"`
}
type UpdateSettingInput struct {
	RequestID       string `json:"request_id"`
	Key             string `json:"key"`
	Value           string `json:"value"`
	ExpectedVersion int64  `json:"expected_version"`
}
type SettingChange struct {
	Key             string `json:"key"`
	Value           string `json:"value"`
	ExpectedVersion int64  `json:"expected_version"`
}
type UpdateSettingsInput struct {
	RequestID string          `json:"request_id"`
	Updates   []SettingChange `json:"updates"`
}

type ConfigFile struct {
	Sources []SiteSource `toml:"sources"`
	Version int64        `toml:"version"`
	Server  struct {
		Listen string `toml:"listen"`
	} `toml:"server"`
	Storage struct {
		Database string `toml:"database"`
	} `toml:"storage"`
	Collection struct {
		Interval string `toml:"interval"`
	} `toml:"collection"`
	GitHub struct {
		Auth          string `toml:"auth"`
		DownloadScope string `toml:"download_scope"`
	} `toml:"github"`
	Interface struct {
		Compact  bool   `toml:"compact"`
		Language string `toml:"language"`
	} `toml:"interface"`
	Overview struct {
		ShowGitHub    bool   `toml:"show_github"`
		ShowEvidence  bool   `toml:"show_evidence"`
		ShowHistory   bool   `toml:"show_history"`
		EvidenceLimit string `toml:"evidence_limit"`
	} `toml:"overview"`
	Evidence struct {
		DefaultFilter string `toml:"default_filter"`
		Reviewer      string `toml:"reviewer"`
	} `toml:"evidence"`
	Discovery struct {
		Worker string `toml:"worker"`
	} `toml:"discovery"`
	Display struct {
		Timezone string `toml:"timezone"`
	} `toml:"display"`
}

func DefaultConfig() ConfigFile {
	var c ConfigFile
	c.Version = 1
	c.Server.Listen = "127.0.0.1:19452"
	c.Storage.Database = "Data/state.sqlite"
	c.Collection.Interval = "24h"
	c.GitHub.Auth = "env"
	c.GitHub.DownloadScope = "all"
	c.Interface.Language = "zh-CN"
	c.Overview.ShowGitHub = true
	c.Overview.ShowEvidence = true
	c.Overview.ShowHistory = true
	c.Overview.EvidenceLimit = "4"
	c.Evidence.DefaultFilter = "all"
	c.Evidence.Reviewer = "local-ui"
	c.Discovery.Worker = "Reverfolio web UI"
	c.Display.Timezone = "local"
	return c
}

var preferenceKeys = []string{
	"server.listen", "github.auth", "storage.database",
	"overview.show_github", "overview.show_evidence", "overview.show_history", "overview.evidence_limit",
	"interface.language", "appearance.compact", "github.download_scope", "evidence.default_filter", "evidence.reviewer",
	"discovery.worker", "display.timezone", "collection.interval",
}

func validateSetting(key, value string) error {
	switch key {
	case "server.listen":
		host, port, err := net.SplitHostPort(value)
		ip := net.ParseIP(host)
		n, numberErr := strconv.Atoi(port)
		if err == nil && numberErr == nil && ip != nil && ip.IsLoopback() && n > 0 && n <= 65535 {
			return nil
		}
	case "github.auth":
		if value == "env" || value == "gh" {
			return nil
		}
	case "storage.database":
		if strings.TrimSpace(value) == value && value != "" && len(value) <= 1000 && !strings.ContainsAny(value, "\x00\r\n") {
			return nil
		}
	case "collection.interval":
		duration, err := time.ParseDuration(value)
		if err == nil && (duration == 0 || (duration >= time.Minute && duration <= 30*24*time.Hour)) {
			return nil
		}
	case "overview.show_github", "overview.show_evidence", "overview.show_history", "appearance.compact":
		if value == "true" || value == "false" {
			return nil
		}
	case "overview.evidence_limit":
		if value == "2" || value == "4" || value == "8" || value == "12" {
			return nil
		}
	case "github.download_scope":
		if value == "all" || value == "installers" {
			return nil
		}
	case "evidence.default_filter":
		if value == "all" || value == "pending" || value == "verified" {
			return nil
		}
	case "evidence.reviewer", "discovery.worker":
		if len(value) > 0 && len(value) <= 80 && strings.TrimSpace(value) == value &&
			!strings.ContainsAny(value, "\x00\r\n") {
			return nil
		}
	case "interface.language":
		if value == "zh-CN" || value == "en" {
			return nil
		}
	case "display.timezone":
		if value == "local" || value == "UTC" {
			return nil
		}
	}
	return ErrInvalid
}

func b(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
func (c ConfigFile) get(key string) string {
	switch key {
	case "server.listen":
		return c.Server.Listen
	case "github.auth":
		return c.GitHub.Auth
	case "storage.database":
		return c.Storage.Database
	case "collection.interval":
		return c.Collection.Interval
	case "overview.show_github":
		return b(c.Overview.ShowGitHub)
	case "overview.show_evidence":
		return b(c.Overview.ShowEvidence)
	case "overview.show_history":
		return b(c.Overview.ShowHistory)
	case "overview.evidence_limit":
		return c.Overview.EvidenceLimit
	case "interface.language":
		return c.Interface.Language
	case "appearance.compact":
		return b(c.Interface.Compact)
	case "github.download_scope":
		return c.GitHub.DownloadScope
	case "evidence.default_filter":
		return c.Evidence.DefaultFilter
	case "evidence.reviewer":
		return c.Evidence.Reviewer
	case "discovery.worker":
		return c.Discovery.Worker
	case "display.timezone":
		return c.Display.Timezone
	}
	return ""
}
func (c *ConfigFile) set(key, v string) {
	switch key {
	case "server.listen":
		c.Server.Listen = v
	case "github.auth":
		c.GitHub.Auth = v
	case "storage.database":
		c.Storage.Database = v
	case "collection.interval":
		c.Collection.Interval = v
	case "overview.show_github":
		c.Overview.ShowGitHub = v == "true"
	case "overview.show_evidence":
		c.Overview.ShowEvidence = v == "true"
	case "overview.show_history":
		c.Overview.ShowHistory = v == "true"
	case "overview.evidence_limit":
		c.Overview.EvidenceLimit = v
	case "interface.language":
		c.Interface.Language = v
	case "appearance.compact":
		c.Interface.Compact = v == "true"
	case "github.download_scope":
		c.GitHub.DownloadScope = v
	case "evidence.default_filter":
		c.Evidence.DefaultFilter = v
	case "evidence.reviewer":
		c.Evidence.Reviewer = v
	case "discovery.worker":
		c.Discovery.Worker = v
	case "display.timezone":
		c.Display.Timezone = v
	}
}

func LoadConfig(path string) (ConfigFile, error) {
	c := DefaultConfig()
	if path == "" {
		return c, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	meta, err := toml.Decode(string(data), &c)
	if err != nil {
		return c, fmt.Errorf("invalid Reverfolio TOML: %w", err)
	}
	if len(meta.Undecoded()) > 0 {
		return c, fmt.Errorf("%w: unknown TOML keys", ErrInvalid)
	}
	if c.Version < 1 {
		return c, ErrInvalid
	}
	for _, key := range preferenceKeys {
		if validateSetting(key, c.get(key)) != nil {
			return c, fmt.Errorf("%w: invalid setting %s", ErrInvalid, key)
		}
	}
	seenSources := map[string]bool{}
	for _, source := range c.Sources {
		if source.validate() != nil || seenSources[source.ID] {
			return c, ErrInvalid
		}
		seenSources[source.ID] = true
	}
	return c, nil
}

type ConfigManager struct {
	Path string
	mu   sync.Mutex
}

func NewConfigManager(path string) *ConfigManager { return &ConfigManager{Path: path} }
func (m *ConfigManager) Read() (ConfigFile, error) {
	if m == nil {
		return DefaultConfig(), nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return LoadConfig(m.Path)
}
func (m *ConfigManager) Settings(_ context.Context) ([]Setting, error) {
	c, err := m.Read()
	if err != nil {
		return nil, err
	}
	out := make([]Setting, 0, len(preferenceKeys))
	for _, key := range preferenceKeys {
		out = append(out, Setting{Key: key, Value: c.get(key), Version: c.Version})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

func saveConfig(path string, c ConfigFile) error {
	if path == "" {
		return fmt.Errorf("%w: config path required", ErrInvalid)
	}
	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(c); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".reverfolio-config-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(buffer.Bytes()); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
func (m *ConfigManager) UpdateSettings(_ context.Context, in UpdateSettingsInput) ([]Setting, error) {
	if m == nil || m.Path == "" || len(in.RequestID) < 1 || len(in.RequestID) > 128 {
		return nil, ErrInvalid
	}
	if len(in.Updates) == 0 || len(in.Updates) > 16 {
		return nil, ErrInvalid
	}
	seen := map[string]bool{}
	for _, u := range in.Updates {
		if u.ExpectedVersion < 1 || seen[u.Key] {
			return nil, ErrInvalid
		}
		if err := validateSetting(u.Key, u.Value); err != nil {
			return nil, err
		}
		seen[u.Key] = true
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	c, err := LoadConfig(m.Path)
	if err != nil {
		return nil, err
	}
	for _, u := range in.Updates {
		if u.ExpectedVersion != c.Version {
			return nil, ErrConflict
		}
	}
	previous := c
	updated := false
	for _, u := range in.Updates {
		if c.get(u.Key) != u.Value {
			c.set(u.Key, u.Value)
			updated = true
		}
	}
	if updated {
		if c.Storage.Database != previous.Storage.Database {
			path := c.Storage.Database
			if !filepath.IsAbs(path) {
				path = filepath.Join(filepath.Dir(m.Path), path)
			}
			f, openErr := os.Open(path)
			if openErr != nil {
				return nil, fmt.Errorf("%w: choose an existing SQLite backup; database is not moved automatically", ErrInvalid)
			}
			header := make([]byte, 16)
			_, readErr := f.Read(header)
			f.Close()
			if readErr != nil || string(header) != "SQLite format 3\x00" {
				return nil, fmt.Errorf("%w: existing SQLite database required", ErrInvalid)
			}
			absolute, absErr := filepath.Abs(path)
			if absErr != nil {
				return nil, ErrInvalid
			}
			uri := url.URL{Scheme: "file", Path: filepath.ToSlash(absolute), RawQuery: "mode=ro"}
			if !strings.HasPrefix(uri.Path, "/") {
				uri.Path = "/" + uri.Path
			}
			db, dbErr := sql.Open("sqlite", uri.String())
			if dbErr != nil {
				return nil, ErrInvalid
			}
			var tables int
			dbErr = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('works','mentions','commands')").Scan(&tables)
			db.Close()
			if dbErr != nil || tables != 3 {
				return nil, fmt.Errorf("%w: select a Reverfolio database backup", ErrInvalid)
			}
		}
		if c.Server.Listen != previous.Server.Listen {
			if err = SyncExtensionEndpoint(m.Path, c); err != nil {
				return nil, err
			}
		}
		c.Version++
		if err = saveConfig(m.Path, c); err != nil {
			if c.Server.Listen != previous.Server.Listen {
				_ = SyncExtensionEndpoint(m.Path, previous)
			}
			return nil, err
		}
	}
	out := []Setting{}
	for _, u := range in.Updates {
		out = append(out, Setting{Key: u.Key, Value: c.get(u.Key), Version: c.Version})
	}
	return out, nil
}

// RuntimeConfigStatus exposes safe startup configuration, never credentials.
type RuntimeConfigStatus struct {
	Version                 string `json:"version"`
	GitHubCredentialLoaded  bool   `json:"github_credential_loaded"`
	Listen                  string `json:"listen"`
	Database                string `json:"database"`
	CollectionInterval      string `json:"collection_interval"`
	GitHubAuthMode          string `json:"github_auth_mode"`
	GitHubCollectionEnabled bool   `json:"github_collection_enabled"`
	ConfigFileAvailable     bool   `json:"config_file_available"`
}

func (m *ConfigManager) RuntimeStatus(_ context.Context) (RuntimeConfigStatus, error) {
	c, err := m.Read()
	if err != nil {
		return RuntimeConfigStatus{}, err
	}
	hasConfig := m != nil && m.Path != ""
	return RuntimeConfigStatus{
		Version: "0.9.2", Listen: c.Server.Listen, Database: c.Storage.Database,
		CollectionInterval: c.Collection.Interval, GitHubAuthMode: c.GitHub.Auth,
		GitHubCollectionEnabled: c.Collection.Interval != "0" && c.Collection.Interval != "0s",
		ConfigFileAvailable:     hasConfig,
	}, nil
}

// RuntimeStatus reports credential presence only; snapshot errors determine API permissions.
func (s Service) RuntimeStatus(ctx context.Context) (RuntimeConfigStatus, error) {
	status, err := s.Config.RuntimeStatus(ctx)
	status.GitHubCredentialLoaded = s.GitHub.Token != ""
	return status, err
}
