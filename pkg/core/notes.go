package core

import (
	"fmt"
	"math"
)

// Notes are plain-language observations about numbers that stand out compared to
// other players: where the bullets land, accuracy, reaction, lag. They are only
// informative: they never change a player's level or flags (see level.go), they
// point the staff at something worth a look in the demo.
//
// Each player is compared with a baseline that excludes them: the rest of the match
// for one demo, everyone else in the history for the player page. Every rule needs
// enough data on both sides, so a lucky round or a short demo says nothing.

// NoteCounts are the raw totals the rules work with. Counts, not percentages, so they
// can be added up across players and demos.
type NoteCounts struct {
	Hits, HeadHits int // bullets that hit an enemy, of them on the head or neck
	Shots          int
	Kills, HSKills int
	ReactSumMs     float64 // triggerbot reaction samples, summed
	ReactN         int
	Engagements    int // peeks measured (prefire module)
	Prefires       int
	PingSum, PingN int // demo ping averages, summed
	PingMax        int
	LagDiscarded   int
}

// CountsFromStats extracts the counts of one player in one demo.
func CountsFromStats(p *PlayerStats) NoteCounts {
	c := NoteCounts{
		Engagements:  p.PeekEngagements,
		Prefires:     p.Prefires,
		PingMax:      p.PingMax,
		LagDiscarded: p.LagDiscarded,
	}
	if h := p.Hits; h != nil {
		c.Hits = h.Zones.Total().Hits
		c.HeadHits = h.Zones.Head.Hits + h.Zones.Neck.Hits
		c.Shots, c.Kills, c.HSKills = h.Shots, h.Kills, h.HeadshotKills
	}
	if p.TriggerAvgMs != nil && p.TriggerSamples > 0 {
		c.ReactSumMs = float64(*p.TriggerAvgMs) * float64(p.TriggerSamples)
		c.ReactN = p.TriggerSamples
	}
	if p.PingAvg > 0 {
		c.PingSum, c.PingN = p.PingAvg, 1
	}
	return c
}

// Add sums another player's or demo's counts into these.
func (c *NoteCounts) Add(o NoteCounts) {
	c.Hits += o.Hits
	c.HeadHits += o.HeadHits
	c.Shots += o.Shots
	c.Kills += o.Kills
	c.HSKills += o.HSKills
	c.ReactSumMs += o.ReactSumMs
	c.ReactN += o.ReactN
	c.Engagements += o.Engagements
	c.Prefires += o.Prefires
	c.PingSum += o.PingSum
	c.PingN += o.PingN
	c.PingMax = max(c.PingMax, o.PingMax)
	c.LagDiscarded += o.LagDiscarded
}

// Minimum data for each comparison, for the player and for the baseline.
const (
	noteMinHits        = 20
	noteMinShots       = 80
	noteMinKills       = 8
	noteMinReactions   = 8
	noteMinEngagements = 10
	noteBaseFactor     = 3 // the baseline needs this many times the player's minimum
)

func pct(part, total int) float64 { return 100 * float64(part) / float64(total) }

func fmtP(v float64) string { return fmt.Sprintf("%.0f%%", math.Round(v)) }

// Notes compares a player with a baseline. baseLabel names the baseline in the
// sentence, e.g. "el resto de la partida".
func Notes(me, base NoteCounts, baseLabel string) []string {
	var out []string

	// Head share of hits: a legit player lands most bullets on the chest
	if me.Hits >= noteMinHits && base.Hits >= noteMinHits*noteBaseFactor {
		mine, avg := pct(me.HeadHits, me.Hits), pct(base.HeadHits, base.Hits)
		if mine >= 30 && mine >= avg+15 {
			out = append(out, fmt.Sprintf("El %s de sus impactos fueron a la cabeza (%d de %d); %s: %s.",
				fmtP(mine), me.HeadHits, me.Hits, baseLabel, fmtP(avg)))
		}
	}

	// Accuracy
	if me.Shots >= noteMinShots && base.Shots >= noteMinShots*noteBaseFactor {
		mine, avg := pct(me.Hits, me.Shots), pct(base.Hits, base.Shots)
		if mine >= 40 && mine >= avg*1.6 {
			out = append(out, fmt.Sprintf("Acertó el %s de sus tiros (%d de %d); %s: %s.",
				fmtP(mine), me.Hits, me.Shots, baseLabel, fmtP(avg)))
		}
	}

	// Kills with a headshot
	if me.Kills >= noteMinKills && base.Kills >= noteMinKills*noteBaseFactor {
		mine, avg := pct(me.HSKills, me.Kills), pct(base.HSKills, base.Kills)
		if mine >= 75 && mine >= avg+25 {
			out = append(out, fmt.Sprintf("%d de sus %d kills fueron con headshot (%s); %s: %s.",
				me.HSKills, me.Kills, fmtP(mine), baseLabel, fmtP(avg)))
		}
	}

	// Reaction from crosshair on enemy to shot
	if me.ReactN >= noteMinReactions && base.ReactN >= noteMinReactions*noteBaseFactor {
		mine, avg := me.ReactSumMs/float64(me.ReactN), base.ReactSumMs/float64(base.ReactN)
		if mine <= avg*0.6 {
			out = append(out, fmt.Sprintf("Dispara %.0f ms después de tener al rival en la mira, de promedio en %d disparos; %s: %.0f ms.",
				mine, me.ReactN, baseLabel, avg))
		}
	}

	// Damage right after the enemy appears (below the PREFIRE flag, which says it itself)
	if me.Engagements >= noteMinEngagements && base.Engagements >= noteMinEngagements*noteBaseFactor {
		mine, avg := pct(me.Prefires, me.Engagements), pct(base.Prefires, base.Engagements)
		if mine >= 20 && mine < PrefireMinRatio && mine >= avg*2.5 {
			out = append(out, fmt.Sprintf("En el %s de sus enfrentamientos dañó al rival a menos de %.0f ms de que apareciera; %s: %s.",
				fmtP(mine), PrefireWindowMs, baseLabel, fmtP(avg)))
		}
	}

	// Lag: context for everything above, not a sign of cheating. A single spike is
	// common and says nothing, so only a high average or reactions actually affected
	// by spikes are mentioned.
	if me.PingN > 0 {
		if avg := (me.PingSum + me.PingN/2) / me.PingN; avg >= 120 {
			out = append(out, fmt.Sprintf("Jugó con ping alto (%d ms de promedio).", avg))
		}
	}
	if me.LagDiscarded > 0 {
		out = append(out, fmt.Sprintf("Se descartaron %d reacciones medidas durante picos de lag.", me.LagDiscarded))
	}
	return out
}

// MatchNotes fills Notes for every player of a demo, each compared with the rest.
func MatchNotes(players []PlayerStats) {
	var total NoteCounts
	counts := make([]NoteCounts, len(players))
	for i := range players {
		counts[i] = CountsFromStats(&players[i])
		total.Add(counts[i])
	}
	for i := range players {
		players[i].Notes = Notes(counts[i], total.without(counts[i]), "el resto de la partida")
	}
}

// without subtracts one player's counts (PingMax can't be subtracted and isn't used
// for the baseline).
func (c NoteCounts) without(o NoteCounts) NoteCounts {
	c.Hits -= o.Hits
	c.HeadHits -= o.HeadHits
	c.Shots -= o.Shots
	c.Kills -= o.Kills
	c.HSKills -= o.HSKills
	c.ReactSumMs -= o.ReactSumMs
	c.ReactN -= o.ReactN
	c.Engagements -= o.Engagements
	c.Prefires -= o.Prefires
	c.PingSum -= o.PingSum
	c.PingN -= o.PingN
	c.LagDiscarded -= o.LagDiscarded
	return c
}
