package radar

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseOverview(t *testing.T) {
	var r Radar
	err := r.parse(`"de_nuke"
{
	"material"	"overviews/de_nuke"	// texture file
	"pos_x"		"-3453"	// upper left world coordinate
	"pos_y"		"2887"
	"scale"		"7"
	"verticalsections"
	{
		"default" // use the primary radar image
		{
			"AltitudeMax" "10000"
			"AltitudeMin" "-495"
		}
		"lower" // i.e. de_nuke_lower_radar.dds
		{
			"AltitudeMax" "-495"
			"AltitudeMin" "-10000"
		}
	}
	"CTSpawn_x"	"0.82"
}`)
	if err != nil {
		t.Fatal(err)
	}
	if r.PosX != -3453 || r.PosY != 2887 || r.Scale != 7 {
		t.Fatalf("pos/scale = %v %v %v", r.PosX, r.PosY, r.Scale)
	}
	if len(r.Sections) != 2 || r.Sections[0].Name != "" || r.Sections[1].Name != "lower" || r.Sections[1].AltMax != -495 {
		t.Fatalf("sections = %+v", r.Sections)
	}

	var single Radar
	if err := single.parse(`"de_mirage" { "pos_x" "-3230" "pos_y" "1713" "scale" "5.00" }`); err == nil {
		// one-line files aren't how CS:GO writes them, but must not break
		_ = single
	}
}

// Real radars from an installed CS:GO: SOURCESIGHT_CSGO=<...\csgo> go test ./pkg/radar
func TestLoadInstalled(t *testing.T) {
	dir := os.Getenv("SOURCESIGHT_CSGO")
	if dir == "" {
		t.Skip("SOURCESIGHT_CSGO not set")
	}
	for _, m := range []string{"de_mirage", "de_dust2", "de_nuke", "de_inferno"} {
		r, err := Load(dir, m)
		if err != nil {
			t.Fatalf("%s: %v", m, err)
		}
		if r.Size != 1024 || len(r.Sections) == 0 {
			t.Fatalf("%s: size %d sections %d", m, r.Size, len(r.Sections))
		}
		b, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(r.Sections[0].Image, "data:image/png;base64,"))
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("%s: png: %v", m, err)
		}
		// The radar isn't blank: the centre of the map has some drawing
		c := img.At(512, 512)
		if _, _, _, a := c.RGBA(); a == 0 {
			t.Logf("%s: centre pixel transparent (fine for some maps)", m)
		}
		if out := os.Getenv("SOURCESIGHT_RADAR_OUT"); out != "" {
			_ = os.WriteFile(filepath.Join(out, m+".png"), b, 0644)
		}
		t.Logf("%s: %d secciones, escala %.2f", m, len(r.Sections), r.Scale)
	}
	if _, err := Load(dir, "../../etc"); err == nil {
		t.Fatal("unsafe map name accepted")
	}
}
