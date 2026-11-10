package core

import "testing"

func TestLevelOf(t *testing.T) {
	cases := map[string][]string{
		LevelClean:   nil,
		LevelConduct: {"TEAMFLASH"},
		LevelSuspect: {"TEAMFLASH", "SOSPECHOSO-BHOP"},
		LevelCheat:   {"ANTI-FLASH", "AIMLOCK"},
	}
	for want, flags := range cases {
		if got := LevelOf(flags); got != want {
			t.Errorf("LevelOf(%v) = %s, want %s", flags, got, want)
		}
	}
}
