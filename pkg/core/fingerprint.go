package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// MatchFingerprint identifies a match by its content: map, length, round count and
// the set of players. Two files of the same recording produce the same value no
// matter their name or location.
func MatchFingerprint(mapName string, totalTicks, rounds int, steamIDs []uint64) string {
	ids := append([]uint64(nil), steamIDs...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprint(id)
	}
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%d|%d|%s", strings.ToLower(mapName), totalTicks, rounds, strings.Join(parts, ","))))
	return hex.EncodeToString(sum[:8])
}

// Fingerprint computes the report's MatchFingerprint.
func (r *DemoReport) Fingerprint() string {
	ids := make([]uint64, 0, len(r.Players))
	for _, p := range r.Players {
		if p.SteamID != 0 {
			ids = append(ids, p.SteamID)
		}
	}
	return MatchFingerprint(r.Header.MapName, r.Header.TotalTicks, len(r.Header.RoundStarts), ids)
}
