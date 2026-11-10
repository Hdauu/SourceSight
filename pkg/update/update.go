// Package update checks GitHub Releases for a newer build and swaps the running
// executable in place. Windows lets a running .exe be renamed but not overwritten,
// so the old binary is moved aside to "<exe>.old" and removed on the next launch.
package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sourcesight/pkg/version"
)

type Release struct {
	Version  string `json:"version"`
	Notes    string `json:"notes"`
	PageURL  string `json:"pageUrl"`
	assetURL string
	digest   string
}

type ghRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	Draft   bool   `json:"draft"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"` // "sha256:<hex>"
	} `json:"assets"`
}

var client = &http.Client{Timeout: 15 * time.Second}

// CleanupOld removes the binary left behind by a previous update.
func CleanupOld() {
	if exe, err := os.Executable(); err == nil {
		_ = os.Remove(exe + ".old")
	}
}

// Check returns the latest release when it is newer than this build, or nil.
func Check() (*Release, error) {
	if version.Version == "dev" {
		return nil, nil
	}
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/"+version.Repo+"/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "SourceSight/"+version.Version)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // no releases published yet
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub respondió %s", resp.Status)
	}

	var gh ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&gh); err != nil {
		return nil, err
	}
	if gh.Draft || !newer(gh.TagName, version.Version) {
		return nil, nil
	}

	rel := &Release{Version: gh.TagName, Notes: gh.Body, PageURL: gh.HTMLURL}
	if exe, err := os.Executable(); err == nil {
		want := strings.ToLower(filepath.Base(exe))
		for _, a := range gh.Assets {
			if strings.ToLower(a.Name) == want {
				rel.assetURL = a.BrowserDownloadURL
				rel.digest = strings.TrimPrefix(a.Digest, "sha256:")
			}
		}
	}
	return rel, nil
}

// CanApply reports whether the release ships a binary matching this executable.
func (r *Release) CanApply() bool {
	return r.assetURL != ""
}

// Apply downloads the new binary, verifies its SHA-256 when GitHub publishes one,
// swaps it in and relaunches. The caller should exit right after it returns nil.
func (r *Release) Apply() error {
	if r.assetURL == "" {
		return fmt.Errorf("la versión %s no incluye un ejecutable para este programa", r.Version)
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	tmp := exe + ".new"

	resp, err := (&http.Client{Timeout: 10 * time.Minute}).Get(r.assetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("descarga fallida: %s", resp.Status)
	}

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, h), resp.Body)
	f.Close()
	if err != nil {
		os.Remove(tmp)
		return err
	}
	if r.digest != "" && !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), r.digest) {
		os.Remove(tmp)
		return fmt.Errorf("el archivo descargado no coincide con el hash publicado; actualización cancelada")
	}

	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Rename(old, exe) // roll back
		return err
	}
	return exec.Command(exe).Start()
}

// newer compares "vMAJOR.MINOR.PATCH" tags numerically.
func newer(latest, current string) bool {
	l, c := parse(latest), parse(current)
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

func parse(v string) [3]int {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	for i, part := range strings.SplitN(v, ".", 3) {
		out[i], _ = strconv.Atoi(part)
	}
	return out
}
