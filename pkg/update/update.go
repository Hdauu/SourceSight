// Package update checks GitHub Releases for a newer build and installs it by running
// the release's installer silently. The installer closes the running app, replaces
// every executable (app, CLI and bot) and relaunches SourceSight.
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

// installerPrefix matches release assets like "SourceSight-Setup-v2.1.2.exe".
const installerPrefix = "sourcesight-setup-"

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

// CleanupOld removes leftovers from the pre-installer update mechanism (v2.1.0/v2.1.1).
func CleanupOld() {
	if exe, err := os.Executable(); err == nil {
		_ = os.Remove(exe + ".old")
	}
}

// Installed reports whether this copy was set up by the installer, which leaves its
// uninstaller next to the executable. Only installed copies can update themselves.
func Installed() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(filepath.Dir(exe), "unins000.exe"))
	return err == nil
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
	for _, a := range gh.Assets {
		name := strings.ToLower(a.Name)
		if strings.HasPrefix(name, installerPrefix) && strings.HasSuffix(name, ".exe") {
			rel.assetURL = a.BrowserDownloadURL
			rel.digest = strings.TrimPrefix(a.Digest, "sha256:")
		}
	}
	return rel, nil
}

// CanApply reports whether the update can be installed from inside the app.
func (r *Release) CanApply() bool {
	return r.assetURL != "" && Installed()
}

// Apply downloads the installer, verifies its SHA-256 when GitHub publishes one and
// starts it silently. The installer closes this process and relaunches the new version,
// so the caller should exit right after Apply returns nil.
func (r *Release) Apply() error {
	if r.assetURL == "" {
		return fmt.Errorf("la versión %s no incluye instalador", r.Version)
	}

	dir, err := os.MkdirTemp("", "sourcesight-update-")
	if err != nil {
		return err
	}
	setup := filepath.Join(dir, "SourceSight-Setup-"+r.Version+".exe")

	resp, err := (&http.Client{Timeout: 10 * time.Minute}).Get(r.assetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("descarga fallida: %s", resp.Status)
	}

	f, err := os.Create(setup)
	if err != nil {
		return err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(f, h), resp.Body)
	f.Close()
	if err != nil {
		os.RemoveAll(dir)
		return err
	}
	if r.digest != "" && !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), r.digest) {
		os.RemoveAll(dir)
		return fmt.Errorf("el instalador descargado no coincide con el hash publicado; actualización cancelada")
	}

	return exec.Command(setup, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/CLOSEAPPLICATIONS").Start()
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
