// Package steamid normalizes the ways people paste a Steam account into a SteamID64.
package steamid

import (
	"regexp"
	"strconv"
	"strings"
)

// base is the SteamID64 of account number 0 in the public universe.
const base uint64 = 76561197960265728

var (
	reID64   = regexp.MustCompile(`7656119\d{10}`)
	reSteam2 = regexp.MustCompile(`(?i)STEAM_[0-5]:([01]):(\d+)`)
	reSteam3 = regexp.MustCompile(`(?i)\[?U:1:(\d+)\]?`)
	reVanity = regexp.MustCompile(`(?i)steamcommunity\.com/id/`)
)

// Parse accepts a SteamID64, a profile URL (/profiles/...), STEAM_0:Y:Z or [U:1:N].
// Vanity URLs (/id/name) cannot be resolved offline and return ok=false.
func Parse(input string) (id uint64, ok bool) {
	s := strings.TrimSpace(input)
	if s == "" || reVanity.MatchString(s) {
		return 0, false
	}
	if m := reID64.FindString(s); m != "" {
		v, err := strconv.ParseUint(m, 10, 64)
		return v, err == nil
	}
	if m := reSteam2.FindStringSubmatch(s); m != nil {
		y, _ := strconv.ParseUint(m[1], 10, 64)
		z, err := strconv.ParseUint(m[2], 10, 64)
		return base + z*2 + y, err == nil
	}
	if m := reSteam3.FindStringSubmatch(s); m != nil {
		n, err := strconv.ParseUint(m[1], 10, 64)
		return base + n, err == nil
	}
	return 0, false
}

// IsVanityURL reports whether the input is a custom profile URL that needs the Steam
// API to resolve.
func IsVanityURL(input string) bool {
	return reVanity.MatchString(input)
}
