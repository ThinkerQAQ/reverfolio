package echo

import (
	"encoding/json"
	"regexp"
	"strings"
)

const AssetClassificationVersion = "filename-v1"

var sourceName = regexp.MustCompile(`(^|[._ -])(source|sources|src)([._ -]|$)`)
var platformName = regexp.MustCompile(`(^|[._ -])(windows|win32|win64|linux|darwin|macos|osx|amd64|arm64|x64|x86_64|aarch64|extension)([._ -]|$)`)

// ClassifyAsset is a conservative filename heuristic, not proof of installation.
// Ambiguous archives remain other instead of silently inflating package counts.
func ClassifyAsset(name string) (string, string) {
	n := strings.ToLower(name)
	for _, suffix := range []string{".sha256", ".sha512", ".sha1", ".md5", ".asc", ".sig", ".sigstore", ".spdx.json", ".cdx.json"} {
		if strings.HasSuffix(n, suffix) {
			return "auxiliary", "checksum_signature_or_sbom"
		}
	}
	if strings.Contains(n, "sha256sum") || strings.Contains(n, "sha512sum") || strings.Contains(n, "checksums") || strings.Contains(n, "sbom") || strings.HasPrefix(n, "license") || strings.HasPrefix(n, "notice") {
		return "auxiliary", "checksum_license_or_sbom"
	}
	if sourceName.MatchString(n) {
		return "source", "source_filename_marker"
	}
	for _, suffix := range []string{".exe", ".msi", ".msix", ".dmg", ".pkg", ".deb", ".rpm", ".appimage", ".vsix", ".crx"} {
		if strings.HasSuffix(n, suffix) {
			return "package", "package_extension"
		}
	}
	for _, suffix := range []string{".zip", ".tar.gz", ".tar.xz", ".tgz", ".7z", ".tar.bz2", ".gz", ".xz"} {
		if strings.HasSuffix(n, suffix) {
			if platformName.MatchString(n) {
				return "package", "platform_archive_filename"
			}
			return "other", "ambiguous_archive"
		}
	}
	if platformName.MatchString(n) && !strings.Contains(n, ".") {
		return "package", "platform_binary_filename"
	}
	return "other", "unrecognized_filename"
}

type DownloadBreakdown struct {
	Total                 int64  `json:"total"`
	Packages              int64  `json:"packages"`
	Source                int64  `json:"source"`
	Auxiliary             int64  `json:"auxiliary"`
	Other                 int64  `json:"other"`
	ClassificationVersion string `json:"classification_version"`
}

func (d *DownloadBreakdown) add(category string, downloads int64) {
	d.Total += downloads
	switch category {
	case "package":
		d.Packages += downloads
	case "source":
		d.Source += downloads
	case "auxiliary":
		d.Auxiliary += downloads
	default:
		d.Other += downloads
	}
}
func annotateSnapshot(s *Snapshot) {
	for i := range s.Assets {
		s.Assets[i].Category, s.Assets[i].ClassificationReason = ClassifyAsset(s.Assets[i].Name)
	}
	s.Downloads = nil
	if !s.ReleasesOK {
		return
	}
	s.Downloads = &DownloadBreakdown{ClassificationVersion: AssetClassificationVersion}
	for _, a := range s.Assets {
		s.Downloads.add(a.Category, a.Downloads)
	}
}

// Recompute derived fields for historical snapshots; never rewrite raw observations.
func (s *Snapshot) UnmarshalJSON(data []byte) error {
	type plain Snapshot
	if err := json.Unmarshal(data, (*plain)(s)); err != nil {
		return err
	}
	annotateSnapshot(s)
	return nil
}
