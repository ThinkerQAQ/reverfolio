package echo

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LocalAPIOrigin derives the extension and native-launcher endpoint from the
// same loopback-only server setting. Never accepts a remote endpoint.
func LocalAPIOrigin(listen string) (string, error) {
	if err := validateSetting("server.listen", listen); err != nil {
		return "", err
	}
	host, port, _ := net.SplitHostPort(listen)
	return (&url.URL{Scheme: "http", Host: net.JoinHostPort(host, port)}).String(), nil
}

// SyncExtensionEndpoint edits only nonsecret local extension metadata. A browser
// extension reload and service restart are required after changing the listener.
func SyncExtensionEndpoint(configPath string, c ConfigFile) error {
	origin, err := LocalAPIOrigin(c.Server.Listen)
	if err != nil {
		return err
	}
	root := filepath.Dir(configPath)
	manifestPath := filepath.Join(root, "extension", "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var manifest map[string]any
	if err = json.Unmarshal(raw, &manifest); err != nil {
		return err
	}
	htmlPath := filepath.Join(root, "extension", "ui", "index.html")
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		return err
	}
	meta := regexp.MustCompile(`<meta name="reverfolio-api-origin" content="[^"]+">`)
	if !meta.Match(html) {
		return ErrInvalid
	}
	csp, ok := manifest["content_security_policy"].(map[string]any)
	if !ok {
		return ErrInvalid
	}
	policy, ok := csp["extension_pages"].(string)
	if !ok {
		return ErrInvalid
	}
	directives := strings.Split(policy, ";")
	found := false
	for i, d := range directives {
		if strings.HasPrefix(strings.TrimSpace(d), "connect-src ") {
			directives[i] = "connect-src 'self' " + origin
			found = true
		}
	}
	if !found {
		return ErrInvalid
	}
	csp["extension_pages"] = strings.Join(directives, ";")
	manifest["host_permissions"] = []string{origin + "/*"}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	originalHTML := append([]byte(nil), html...)
	html = meta.ReplaceAll(html, []byte(`<meta name="reverfolio-api-origin" content="`+origin+`">`))
	if err = os.WriteFile(htmlPath, html, 0600); err != nil {
		return err
	}
	if err = os.WriteFile(manifestPath, append(encoded, '\n'), 0600); err != nil {
		_ = os.WriteFile(htmlPath, originalHTML, 0600)
		return err
	}
	return nil
}
