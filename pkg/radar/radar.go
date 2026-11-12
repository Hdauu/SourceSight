// Package radar reads map radars from the CS:GO installation (resource/overviews):
// the radar image of each map and the numbers that place world coordinates on it.
// Nothing from Valve is shipped with SourceSight: the images come from the user's own
// game, are converted to PNG once and cached in the data folder.
package radar

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"sourcesight/pkg/appdata"
)

// Section is one floor of a map with several (de_nuke's lower level, de_vertigo):
// points with a height between AltMin and AltMax are drawn on its image.
type Section struct {
	Name   string  `json:"name"` // "" for the main image
	AltMin float64 `json:"altMin"`
	AltMax float64 `json:"altMax"`
	Image  string  `json:"image"` // PNG as a data: URL
}

// Radar places world coordinates on the radar image: a point (x, y) is at pixel
// ((x - PosX) / Scale, (PosY - y) / Scale) of a Size×Size image.
type Radar struct {
	Map      string    `json:"map"`
	PosX     float64   `json:"posX"`
	PosY     float64   `json:"posY"`
	Scale    float64   `json:"scale"`
	Size     int       `json:"size"`
	Sections []Section `json:"sections"`
}

var (
	cacheMu sync.Mutex
	kvRe    = regexp.MustCompile(`"([^"]+)"\s+"([^"]*)"`)
	safeMap = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
)

// Load returns the radar of a map from the CS:GO folder ("...\csgo").
func Load(csgoDir, mapName string) (*Radar, error) {
	if !safeMap.MatchString(mapName) {
		return nil, fmt.Errorf("nombre de mapa inválido: %q", mapName)
	}
	dir := filepath.Join(csgoDir, "resource", "overviews")
	txt, err := os.ReadFile(filepath.Join(dir, mapName+".txt"))
	if err != nil {
		return nil, fmt.Errorf("tu CS:GO no tiene el radar de %s", mapName)
	}
	r := &Radar{Map: mapName}
	if err := r.parse(string(txt)); err != nil {
		return nil, err
	}
	for i := range r.Sections {
		s := &r.Sections[i]
		file := mapName + "_radar.dds"
		if s.Name != "" {
			file = mapName + "_" + s.Name + "_radar.dds"
		}
		img, size, err := pngFor(filepath.Join(dir, file), mapName, s.Name)
		if err != nil {
			if s.Name == "" {
				return nil, err
			}
			continue // a missing extra floor: points stay on the main image
		}
		s.Image, r.Size = img, size
	}
	// Drop floors without an image
	kept := r.Sections[:0]
	for _, s := range r.Sections {
		if s.Image != "" {
			kept = append(kept, s)
		}
	}
	r.Sections = kept
	return r, nil
}

// parse reads pos_x, pos_y, scale and the vertical sections of an overview file.
func (r *Radar) parse(txt string) error {
	section := ""
	depth := 0
	alt := map[string]*Section{}
	var order []string
	for _, line := range strings.Split(txt, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		switch {
		case line == "{":
			depth++
			continue
		case line == "}":
			depth--
			if depth <= 2 {
				section = ""
			}
			continue
		}
		if m := kvRe.FindStringSubmatch(line); m != nil {
			k, v := strings.ToLower(m[1]), m[2]
			f, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
			switch {
			case section != "" && k == "altitudemin":
				alt[section].AltMin = f
			case section != "" && k == "altitudemax":
				alt[section].AltMax = f
			case depth == 1 && k == "pos_x":
				r.PosX = f
			case depth == 1 && k == "pos_y":
				r.PosY = f
			case depth == 1 && k == "scale":
				r.Scale = f
			}
			continue
		}
		// A bare quoted name opens a vertical section ("default", "lower"...)
		if depth == 2 && strings.HasPrefix(line, `"`) {
			name := strings.ToLower(strings.Trim(line, `" `))
			section = name
			alt[name] = &Section{AltMin: -1e9, AltMax: 1e9}
			order = append(order, name)
		}
	}
	if r.Scale <= 0 {
		return errors.New("archivo de radar sin escala")
	}
	if len(order) == 0 {
		r.Sections = []Section{{AltMin: -1e9, AltMax: 1e9}}
		return nil
	}
	for _, name := range order {
		s := *alt[name]
		if name != "default" {
			s.Name = name
		}
		r.Sections = append(r.Sections, s)
	}
	// Main image first, then the rest from top to bottom
	sort.SliceStable(r.Sections, func(i, j int) bool {
		if (r.Sections[i].Name == "") != (r.Sections[j].Name == "") {
			return r.Sections[i].Name == ""
		}
		return r.Sections[i].AltMax > r.Sections[j].AltMax
	})
	return nil
}

// pngFor converts a radar DDS to PNG once and keeps it in the data folder.
func pngFor(ddsPath, mapName, section string) (string, int, error) {
	name := mapName
	if section != "" {
		name += "_" + section
	}
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cached := filepath.Join(appdata.Sub("radar"), name+".png")
	src, err := os.Stat(ddsPath)
	if err != nil {
		return "", 0, fmt.Errorf("tu CS:GO no tiene la imagen del radar de %s", mapName)
	}
	if st, err := os.Stat(cached); err == nil && st.ModTime().After(src.ModTime()) {
		if b, err := os.ReadFile(cached); err == nil {
			if cfg, err := png.DecodeConfig(bytes.NewReader(b)); err == nil {
				return dataURL(b), cfg.Width, nil
			}
		}
	}
	raw, err := os.ReadFile(ddsPath)
	if err != nil {
		return "", 0, err
	}
	img, err := decodeDDS(raw)
	if err != nil {
		return "", 0, fmt.Errorf("radar de %s: %w", mapName, err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", 0, err
	}
	_ = os.WriteFile(cached, buf.Bytes(), 0644)
	return dataURL(buf.Bytes()), img.Rect.Dx(), nil
}

func dataURL(b []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
}
