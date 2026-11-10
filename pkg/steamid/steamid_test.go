package steamid

import "testing"

func TestParse(t *testing.T) {
	const want = 76561198000000123
	cases := []string{
		"76561198000000123",
		"https://steamcommunity.com/profiles/76561198000000123/",
		"STEAM_0:1:19867197",
		"STEAM_1:1:19867197",
		"[U:1:39734395]",
		"U:1:39734395",
	}
	for _, c := range cases {
		if got, ok := Parse(c); !ok || got != want {
			t.Errorf("Parse(%q) = %d, %v; want %d", c, got, ok, want)
		}
	}
	for _, bad := range []string{"", "pepito", "https://steamcommunity.com/id/pepito"} {
		if _, ok := Parse(bad); ok {
			t.Errorf("Parse(%q) should fail", bad)
		}
	}
}
