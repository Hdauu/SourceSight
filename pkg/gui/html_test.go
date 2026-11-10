package gui

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

// TestPageScriptSyntax catches JavaScript syntax errors in the embedded page, which
// otherwise only show up as a blank window. Needs Node.js; skipped without it.
func TestPageScriptSyntax(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	m := regexp.MustCompile(`(?s)<script>(.*)</script>`).FindStringSubmatch(AppHTML)
	if m == nil {
		t.Fatal("page has no <script> block")
	}
	path := filepath.Join(t.TempDir(), "page.js")
	if err := os.WriteFile(path, []byte(m[1]), 0644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "--check", path).CombinedOutput(); err != nil {
		t.Fatalf("page script has a syntax error:\n%s", out)
	}
}
