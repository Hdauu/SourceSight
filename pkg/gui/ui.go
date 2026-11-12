package gui

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

// The interface lives in ui/ as plain files: index.html holds the markup, css/ and
// js/ hold the styles and scripts split by area. Files load in name order (hence the
// numeric prefixes) and are inlined into one page at startup, so the webview gets a
// single document exactly as before and the Go bindings work unchanged.
//
//go:embed ui
var uiFiles embed.FS

// AppHTML is the complete page served to the webview.
var AppHTML = buildPage()

func buildPage() string {
	index, err := uiFiles.ReadFile("ui/index.html")
	if err != nil {
		panic("ui/index.html missing from the embedded files: " + err.Error())
	}
	page := string(index)
	page = strings.Replace(page, "/* @styles */", concatDir("ui/css"), 1)
	page = strings.Replace(page, "/* @scripts */", concatDir("ui/js"), 1)
	return page
}

// concatDir joins every file of an embedded directory in name order.
func concatDir(dir string) string {
	entries, err := fs.ReadDir(uiFiles, dir)
	if err != nil {
		panic("embedded ui directory missing: " + dir)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		data, _ := uiFiles.ReadFile(dir + "/" + name)
		b.WriteString("/* ── ")
		b.WriteString(name)
		b.WriteString(" ── */\n")
		b.Write(data)
		b.WriteString("\n")
	}
	return b.String()
}
