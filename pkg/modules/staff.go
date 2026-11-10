package modules

import (
	"fmt"
	"regexp"
	"strings"

	"sourcesight/pkg/core"
)

type DeathRecord struct {
	KillerName string
	VictimName string
	Tick       int
}

func RegisterDeath(st *core.PlayerState, tick int, enemyName string, enemyPos, deathPos core.Vec3) {
	t := tick
	st.Ghost.DeathTick = &t
	st.Ghost.EnemyName = enemyName
	pos := enemyPos
	st.Ghost.EnemyPos = &pos
	dpos := deathPos
	st.Ghost.DeathPos = &dpos
	st.IsAlive = false
}

func CheckGhosting(deadState, livingState *core.PlayerState, tick int, livingAimPos core.Vec3, tickrate int) {
	if deadState.IsAlive || deadState.Ghost.DeathTick == nil || deadState.Ghost.EnemyPos == nil {
		return
	}

	// Window constants are expressed in 64-tick ticks; scale to the demo's tickrate
	elapsed := tick - *deadState.Ghost.DeathTick
	elapsedMs := float64(elapsed) * (1000.0 / float64(tickrate))
	windowMs := float64(core.GhostWindowTicks) * (1000.0 / 64.0)
	if elapsedMs < core.GhostMinDelayMs || elapsedMs > windowMs {
		return // immediate shots are trades off the kill feed / sound, not leaked info
	}
	// Close enough to have heard the fight
	if deadState.Ghost.DeathPos != nil && core.Dist3D(livingState.Origin, *deadState.Ghost.DeathPos) < core.GhostMinDistUnits {
		return
	}

	dist := core.Dist3D(livingAimPos, *deadState.Ghost.EnemyPos)
	if dist < 150.0 {
		livingState.Ghost.Incidents = append(livingState.Ghost.Incidents, core.TimelineIncident{
			Tick: tick,
			Type: "GHOSTING",
			Note: fmt.Sprintf("Prefired hidden enemy (%s) %.0fms after %s died — DISCORD GHOSTING", deadState.Ghost.EnemyName, elapsedMs, deadState.Name),
		})
		// One incident per death: consume the leaked position
		deadState.Ghost.EnemyPos = nil
	}
}

func RegisterTeamflash(teammate *core.PlayerState, durationSecs float64, isEarlyRound bool) {
	weight := 1.0
	if isEarlyRound {
		weight = 2.0
	}
	teammate.Grief.TeamflashSecs += durationSecs * weight
}

func MeasureRoundDisplacement(st *core.PlayerState, curPos core.Vec3) {
	if st.Grief.StartPos == nil {
		p := curPos
		st.Grief.StartPos = &p
		return
	}
	d := core.Dist2D(*st.Grief.StartPos, curPos)
	st.Grief.RoundDisplace = append(st.Grief.RoundDisplace, d)
	p := curPos
	st.Grief.StartPos = &p
}

// Punitive commands audited for abuse; the rest of SourceMod's admin menu is ignored.
var adminCommands = map[string]bool{
	"slay": true, "slap": true, "kick": true, "ban": true, "burn": true,
	"freeze": true, "beacon": true, "mute": true, "gag": true, "silence": true,
	"timebomb": true, "firebomb": true, "freezebomb": true, "drug": true,
}

// Commands that punish the target in-round — the only ones that can be a "revenge".
var revengeCommands = map[string]bool{
	"slay": true, "slap": true, "kick": true, "ban": true, "burn": true,
	"timebomb": true, "firebomb": true,
}

// SourceMod ShowActivity2 broadcast, e.g. "[SM] Juan: Slayed Pedro."
var smActivityRe = regexp.MustCompile(`(?i)^\[SM\]\s*(.+?):\s*([a-z]+)\s+(.+?)\.?$`)

// ParseAdminChatCommand recognizes "!slay foo", "/slay foo" and "sm_slay foo" typed in chat.
func ParseAdminChatCommand(text string) (cmd, target string, ok bool) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "", "", false
	}
	head := strings.ToLower(fields[0])
	switch {
	case strings.HasPrefix(head, "sm_"):
		head = head[3:]
	case strings.HasPrefix(head, "!"), strings.HasPrefix(head, "/"):
		head = head[1:]
	default:
		return "", "", false
	}
	if !adminCommands[head] {
		return "", "", false
	}
	return head, strings.Join(fields[1:], " "), true
}

// ParseAdminActivity recognizes SourceMod's server-side activity broadcast.
func ParseAdminActivity(text string) (admin, cmd, target string, ok bool) {
	m := smActivityRe.FindStringSubmatch(strings.TrimSpace(text))
	if m == nil {
		return "", "", "", false
	}
	verb := strings.ToLower(m[2])
	// "Slayed" -> "slay", "Kicked" -> "kick", "Banned" -> "ban"
	for c := range adminCommands {
		if verb == c || verb == c+"ed" || verb == c+"d" || verb == c+c[len(c)-1:]+"ed" {
			return m[1], c, m[3], true
		}
	}
	return "", "", "", false
}

// BuildAdminEvent correlates an admin command with the admin's recent deaths.
// A punitive command against the player who killed the admin within
// RevengeSlayWindowTicks is flagged as a revenge action.
func BuildAdminEvent(adminName, cmd, target string, tick, round int, deaths []DeathRecord) core.AdminEvent {
	ev := core.AdminEvent{
		Tick:       tick,
		Round:      round,
		AdminName:  adminName,
		Command:    "sm_" + cmd,
		TargetName: target,
	}

	needle := strings.ToLower(strings.Trim(target, "\"' "))
	if needle != "" && revengeCommands[cmd] {
		for i := len(deaths) - 1; i >= 0; i-- {
			d := deaths[i]
			elapsed := tick - d.Tick
			if elapsed < 0 || elapsed > core.RevengeSlayWindowTicks {
				continue
			}
			// SourceMod targets by case-insensitive partial name match
			if d.VictimName == adminName && strings.Contains(strings.ToLower(d.KillerName), needle) {
				ev.IsRevenge = true
				ev.TargetName = d.KillerName
				ev.Note = fmt.Sprintf("%s mató al admin %d ticks antes — posible abuso por venganza", d.KillerName, elapsed)
				break
			}
		}
	}

	preroll := tick - 192
	if preroll < 128 {
		preroll = 128
	}
	ev.GotoCmd = fmt.Sprintf("demo_gototick %d", preroll)
	return ev
}
