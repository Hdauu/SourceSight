package core

// Hits and damage per body zone, from the hit group the game reports with each
// bullet that damages an enemy. Only bullet weapons count (no knife, grenades, fire
// or zeus) and only damage to enemies.

// Hit groups as CS:GO numbers them in player_hurt.
const (
	HitGroupGeneric  = 0
	HitGroupHead     = 1
	HitGroupChest    = 2
	HitGroupStomach  = 3
	HitGroupLeftArm  = 4
	HitGroupRightArm = 5
	HitGroupLeftLeg  = 6
	HitGroupRightLeg = 7
	HitGroupNeck     = 8
	HitGroupGear     = 10
)

// Weapon classes used to split hits.
const (
	WeaponRifle  = "rifle"
	WeaponSniper = "sniper"
	WeaponPistol = "pistol"
	WeaponSMG    = "smg"
	WeaponHeavy  = "heavy" // shotguns and machine guns
)

// ZoneCount is how many bullets hit a zone and the health they took.
type ZoneCount struct {
	Hits   int `json:"hits"`
	Damage int `json:"damage"`
}

func (z *ZoneCount) add(o ZoneCount) {
	z.Hits += o.Hits
	z.Damage += o.Damage
}

// BodyZones holds one count per zone. Other covers generic and gear hits, which the
// game reports for some wallbangs and edge cases.
type BodyZones struct {
	Head     ZoneCount `json:"head"`
	Neck     ZoneCount `json:"neck"`
	Chest    ZoneCount `json:"chest"`
	Stomach  ZoneCount `json:"stomach"`
	LeftArm  ZoneCount `json:"leftArm"`
	RightArm ZoneCount `json:"rightArm"`
	LeftLeg  ZoneCount `json:"leftLeg"`
	RightLeg ZoneCount `json:"rightLeg"`
	Other    ZoneCount `json:"other"`
}

func (b *BodyZones) zone(hitGroup int) *ZoneCount {
	switch hitGroup {
	case HitGroupHead:
		return &b.Head
	case HitGroupNeck:
		return &b.Neck
	case HitGroupChest:
		return &b.Chest
	case HitGroupStomach:
		return &b.Stomach
	case HitGroupLeftArm:
		return &b.LeftArm
	case HitGroupRightArm:
		return &b.RightArm
	case HitGroupLeftLeg:
		return &b.LeftLeg
	case HitGroupRightLeg:
		return &b.RightLeg
	default:
		return &b.Other
	}
}

func (b *BodyZones) all() []*ZoneCount {
	return []*ZoneCount{&b.Head, &b.Neck, &b.Chest, &b.Stomach, &b.LeftArm, &b.RightArm, &b.LeftLeg, &b.RightLeg, &b.Other}
}

// Add sums another set of zones into this one.
func (b *BodyZones) Add(o BodyZones) {
	dst, src := b.all(), o.all()
	for i := range dst {
		dst[i].add(*src[i])
	}
}

// Total sums every zone.
func (b BodyZones) Total() ZoneCount {
	var t ZoneCount
	for _, z := range b.all() {
		t.add(*z)
	}
	return t
}

// WeaponHits is the hit profile of one weapon class.
type WeaponHits struct {
	Shots int       `json:"shots"`
	Zones BodyZones `json:"zones"`
}

// HitProfile is everything a player hit in a demo (or across demos, when merged).
type HitProfile struct {
	Shots         int                    `json:"shots"` // bullets fired with bullet weapons
	Kills         int                    `json:"kills"` // enemies killed with bullet weapons
	HeadshotKills int                    `json:"headshotKills"`
	Zones         BodyZones              `json:"zones"`
	ByClass       map[string]*WeaponHits `json:"byClass,omitempty"`
}

func (h *HitProfile) class(c string) *WeaponHits {
	if h.ByClass == nil {
		h.ByClass = map[string]*WeaponHits{}
	}
	w := h.ByClass[c]
	if w == nil {
		w = &WeaponHits{}
		h.ByClass[c] = w
	}
	return w
}

// AddShot records one bullet fired.
func (h *HitProfile) AddShot(class string) {
	h.Shots++
	h.class(class).Shots++
}

// AddHit records one bullet that damaged an enemy.
func (h *HitProfile) AddHit(class string, hitGroup, damage int) {
	z := ZoneCount{Hits: 1, Damage: damage}
	h.Zones.zone(hitGroup).add(z)
	h.class(class).Zones.zone(hitGroup).add(z)
}

// AddKill records an enemy killed with a bullet weapon.
func (h *HitProfile) AddKill(headshot bool) {
	h.Kills++
	if headshot {
		h.HeadshotKills++
	}
}

// Merge adds another profile (another demo) into this one.
func (h *HitProfile) Merge(o *HitProfile) {
	if o == nil {
		return
	}
	h.Shots += o.Shots
	h.Kills += o.Kills
	h.HeadshotKills += o.HeadshotKills
	h.Zones.Add(o.Zones)
	for c, w := range o.ByClass {
		dst := h.class(c)
		dst.Shots += w.Shots
		dst.Zones.Add(w.Zones)
	}
}

// Empty reports whether nothing was recorded.
func (h *HitProfile) Empty() bool {
	return h == nil || (h.Shots == 0 && h.Zones.Total().Hits == 0 && h.Kills == 0)
}
