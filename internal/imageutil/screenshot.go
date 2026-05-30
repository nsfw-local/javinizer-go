package imageutil

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

var (
	dmmImageExtRegex = regexp.MustCompile(`(?i)\.jpe?g$`)
)

// IsDMMHost returns true if the hostname belongs to a DMM-owned domain
// (dmm.co.jp, dmm.com, and their subdomains).
func IsDMMHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "dmm.co.jp" || strings.HasSuffix(host, ".dmm.co.jp") ||
		host == "dmm.com" || strings.HasSuffix(host, ".dmm.com")
}

// NormalizeDMMScreenshotURL normalizes a DMM-hosted screenshot URL for
// consistent deduplication and higher-quality image retrieval.
//
// Applies the following transformations when the URL is on a DMM domain:
//   - Protocol-relative URLs (//...) are upgraded to https
//   - awsimgsrc.dmm.co.jp CDN paths are rewritten to pics.dmm.co.jp
//   - Query parameters and fragments are stripped
//   - Screenshot filenames missing the "jp" suffix get it inserted
//     (e.g., avsa00432-1.jpg -> avsa00432jp-1.jpg) for the larger
//     resolution version, while cover/poster URLs (pl.jpg, ps.jpg) are
//     left unchanged.
//
// Non-DMM URLs are returned unchanged.
func NormalizeDMMScreenshotURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}

	raw = strings.Replace(raw, "awsimgsrc.dmm.co.jp/pics_dig", "pics.dmm.co.jp", 1)

	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	if !IsDMMHost(u.Hostname()) {
		return u.String()
	}

	u.Host = strings.ToLower(u.Host)

	if strings.Contains(u.Path, "/digital/amateur/") {
		u.Path = strings.ToLower(u.Path)
	}

	u.RawQuery = ""
	u.Fragment = ""

	base := path.Base(u.Path)
	lowerBase := strings.ToLower(base)
	if dmmImageExtRegex.MatchString(lowerBase) &&
		strings.Contains(base, "-") &&
		!strings.Contains(lowerBase, "jp-") &&
		!strings.HasSuffix(lowerBase, "pl.jpg") &&
		!strings.HasSuffix(lowerBase, "ps.jpg") {
		base = strings.Replace(base, "-", "jp-", 1)
		u.Path = strings.TrimSuffix(u.Path, path.Base(u.Path)) + base
	}

	return u.String()
}

// UpgradeCoverResolution upgrades cover image URLs to their highest-resolution
// variant. It applies two transformations:
//   - ps.jpg → pl.jpg (for all URLs, including amateur)
//   - jp.jpg → pl.jpg (for non-amateur URLs only)
//
// Screenshot-style filenames (e.g., ipx00535jp-1.jpg) are left unchanged
// because the suffix check uses HasSuffix rather than Contains.
func UpgradeCoverResolution(rawURL string) string {
	if strings.HasSuffix(rawURL, "ps.jpg") {
		rawURL = rawURL[:len(rawURL)-len("ps.jpg")] + "pl.jpg"
	}
	if !strings.Contains(rawURL, "/amateur/") && strings.HasSuffix(rawURL, "jp.jpg") {
		rawURL = rawURL[:len(rawURL)-len("jp.jpg")] + "pl.jpg"
	}
	return rawURL
}
