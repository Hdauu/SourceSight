package gui

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// nodeOrSkip returns the Node.js binary used to syntax-check the page scripts.
func nodeOrSkip(t *testing.T) string {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	return node
}

func checkScript(t *testing.T, node, name, code string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "page.js")
	if err := os.WriteFile(path, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(node, "--check", path).CombinedOutput(); err != nil {
		t.Errorf("%s has a syntax error:\n%s", name, out)
	}
}

// TestPageScriptSyntax catches JavaScript syntax errors, which otherwise only show
// up as a blank window. Each file is checked alone (to name the culprit) and then
// the assembled page script (to catch clashes between files, such as a duplicate
// top-level declaration).
func TestPageScriptSyntax(t *testing.T) {
	node := nodeOrSkip(t)

	files, _ := fs.Glob(uiFiles, "ui/js/*.js")
	if len(files) == 0 {
		t.Fatal("no embedded scripts")
	}
	for _, f := range files {
		data, _ := uiFiles.ReadFile(f)
		checkScript(t, node, f, string(data))
	}

	m := regexp.MustCompile(`(?s)<script>(.*)</script>`).FindStringSubmatch(AppHTML)
	if m == nil {
		t.Fatal("assembled page has no <script> block")
	}
	checkScript(t, node, "assembled page script", m[1])
}

// TestPageAssembled makes sure every placeholder was filled.
func TestPageAssembled(t *testing.T) {
	for _, marker := range []string{"/* @styles */", "/* @scripts */"} {
		if strings.Contains(AppHTML, marker) {
			t.Errorf("placeholder %s was not replaced", marker)
		}
	}
	if !strings.Contains(AppHTML, "--accent:") || !strings.Contains(AppHTML, "function loadDemoReport") {
		t.Error("styles or scripts missing from the assembled page")
	}
}
