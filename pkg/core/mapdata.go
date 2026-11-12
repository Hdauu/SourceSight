package core

// Positions for the minimap: where each kill happened and where each player moved.
// Kept compact (integers, one sample per second) because it is stored with every
// report: about 300–400 KB of JSON for a full match, a third of that compressed.

// MapKill is one kill with both players' positions, in world units.
type MapKill struct {
	Tick       int    `json:"tick"`
	Killer     string `json:"killer"`
	KillerID   uint64 `json:"killerId,string"`
	KillerTeam int    `json:"killerTeam"` // 2 = T, 3 = CT
	Victim     string `json:"victim"`
	VictimID   uint64 `json:"victimId,string"`
	VictimTeam int    `json:"victimTeam"`
	Weapon     string `json:"weapon"`
	Headshot   bool   `json:"headshot,omitempty"`
	KX         int    `json:"kx"`
	KY         int    `json:"ky"`
	KZ         int    `json:"kz"`
	VX         int    `json:"vx"`
	VY         int    `json:"vy"`
	VZ         int    `json:"vz"`
}

// PlayerPath is where a player was, sampled once per second while alive: a flat list
// of tick, x, y, z, tick, x, y, z…
type PlayerPath struct {
	SteamID uint64 `json:"steamId,string"`
	Name    string `json:"name"`
	Points  []int  `json:"points"`
}

// MapData is everything the minimap draws.
type MapData struct {
	Kills []MapKill    `json:"kills"`
	Paths []PlayerPath `json:"paths"`
	Every int          `json:"every"` // ticks between path samples
}

// AddSample appends one position to a player's path.
func (p *PlayerPath) AddSample(tick int, pos Vec3) {
	p.Points = append(p.Points, tick, int(pos.X), int(pos.Y), int(pos.Z))
}
