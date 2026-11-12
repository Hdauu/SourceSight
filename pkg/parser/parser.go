package parser

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	demoinfocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"

	"sourcesight/pkg/core"
	"sourcesight/pkg/db"
	"sourcesight/pkg/modules"
)

// ErrCanceled is returned when ParserOptions.Cancel was closed mid-analysis.
var ErrCanceled = errors.New("análisis cancelado")

type ParserOptions struct {
	Verbose  bool
	DemoPath string
	// OnProgress, if set, receives the parsed fraction (0..1) a few times per second.
	OnProgress func(fraction float64)
	// Cancel, if set, aborts the analysis as soon as it is closed.
	Cancel <-chan struct{}
}

func ParseDemo(demoPath string, opts ParserOptions) (*core.DemoReport, error) {
	opts.DemoPath = demoPath
	f, err := os.Open(demoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open demo: %w", err)
	}
	defer f.Close()

	return ParseDemoReader(f, opts)
}

func ParseDemoReader(r io.Reader, opts ParserOptions) (*core.DemoReport, error) {
	cfg := demoinfocs.DefaultParserConfig
	cfg.IgnoreErrBombsiteIndexNotFound = true

	p := demoinfocs.NewParserWithConfig(r, cfg)
	defer p.Close()

	header, err := p.ParseHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to parse demo header: %w", err)
	}

	tickrate := 64
	if header.PlaybackTime > 0 && header.PlaybackTicks > 0 {
		inferred := int(math.Round(float64(header.PlaybackTicks) / header.PlaybackTime.Seconds()))
		if inferred >= 60 && inferred <= 140 {
			tickrate = inferred
		}
	}

	if opts.Verbose {
		fmt.Printf("[+] Map: %s | Tickrate: %d | PlaybackTicks: %d\n", header.MapName, tickrate, header.PlaybackTicks)
		fmt.Printf("[+] Server: %s\n", header.ServerName)
	}

	states := make(map[string]*core.PlayerState)
	getState := func(pl *common.Player) *core.PlayerState {
		if pl == nil || pl.Name == "" || pl.IsBot {
			return nil
		}
		st, ok := states[pl.Name]
		if !ok {
			st = core.NewPlayerState(pl.Name, pl.SteamID64)
			states[pl.Name] = st
		}
		return st
	}

	chat := modules.NewChatInspector()

	var chatLog []core.ChatMessageEntry
	var adminLog []core.AdminEvent
	var roundStarts []int
	var recentDeaths []modules.DeathRecord
	var activeSmokes []modules.SmokeVolume
	roundBullets := make(map[string]int)
	currentRound := 0

	// Event Handlers
	p.RegisterEventHandler(func(e events.RoundStart) {
		currentRound++
		// Keep deaths inside the revenge window: admins often slay right after the round flips
		curTick := p.GameState().IngameTick()
		roundStarts = append(roundStarts, curTick)
		var kept []modules.DeathRecord
		for _, d := range recentDeaths {
			if curTick-d.Tick <= core.RevengeSlayWindowTicks {
				kept = append(kept, d)
			}
		}
		recentDeaths = kept
		for _, st := range states {
			modules.ResetPeekRound(st)
			st.Aim.PrevPitch = nil
			st.Aim.PrevYaw = nil
			st.Aim.AliveTicks = 0
			st.Bhop.CurrentStreak = 0
		}
	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		gs := p.GameState()
		for _, pl := range gs.Participants().Playing() {
			st := getState(pl)
			if st != nil {
				pos := pl.Position()
				modules.MeasureRoundDisplacement(st, core.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z})
			}
		}
	})

	p.RegisterEventHandler(func(e events.SmokeStart) {
		activeSmokes = append(activeSmokes, modules.SmokeVolume{
			Pos: core.Vec3{
				X: e.Position.X,
				Y: e.Position.Y,
				Z: e.Position.Z,
			},
			StartTick:  p.GameState().IngameTick(),
			ExpiryTick: p.GameState().IngameTick() + (18 * tickrate),
		})
	})

	p.RegisterEventHandler(func(e events.SmokeExpired) {
		curTick := p.GameState().IngameTick()
		var kept []modules.SmokeVolume
		for _, s := range activeSmokes {
			if s.ExpiryTick > curTick {
				kept = append(kept, s)
			}
		}
		activeSmokes = kept
	})

	p.RegisterEventHandler(func(e events.WeaponFire) {
		if e.Shooter != nil && e.Shooter.Name != "" {
			roundBullets[e.Shooter.Name]++
			if ast := getState(e.Shooter); ast != nil {
				pos := e.Shooter.Position()
				eyeOffset := 64.0
				if e.Shooter.IsDucking() {
					eyeOffset = 46.0
				}
				ast.LastFirePitch = ast.Pitch
				ast.LastFireYaw = ast.Yaw
				ast.LastFireEye = core.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z + eyeOffset}
				ast.LastFireTick = p.GameState().IngameTick()
			}
		}
	})

	p.RegisterEventHandler(func(e events.PlayerHurt) {
		if e.Attacker == nil || e.Player == nil || e.Attacker == e.Player {
			return
		}
		ast := getState(e.Attacker)
		if ast == nil {
			return
		}

		// Only inspect bullet weapons for Silent Aim and bullet-wall penetration anomalies
		// Exclude knives, HE grenades, molotovs, and fall damage
		if e.Weapon != nil {
			class := e.Weapon.Class()
			isBulletWeapon := class == common.EqClassRifle ||
				class == common.EqClassPistols ||
				class == common.EqClassSMG ||
				class == common.EqClassHeavy

			if isBulletWeapon && !strings.Contains(strings.ToLower(e.Weapon.String()), "taser") {
				vicPos := e.Player.Position()
				victimVec := core.Vec3{X: vicPos.X, Y: vicPos.Y, Z: vicPos.Z + 36.0}
				tick := p.GameState().IngameTick()

				modules.CheckSilentAim(
					ast, tick,
					victimVec,
					e.Player.Name, e.HealthDamage,
				)
				if e.Attacker.Team != e.Player.Team {
					modules.RegisterPeekDamage(ast, e.Player.SteamID64, e.Player.Name, tick, tickrate)
				}
			}
		}
	})

	p.RegisterEventHandler(func(e events.Kill) {
		if e.Victim == nil {
			return
		}
		vst := getState(e.Victim)
		tick := p.GameState().IngameTick()

		if vst != nil && e.Killer != nil {
			kpos := e.Killer.Position()
			vpos := e.Victim.Position()
			modules.RegisterDeath(vst, tick, e.Killer.Name,
				core.Vec3{X: kpos.X, Y: kpos.Y, Z: kpos.Z},
				core.Vec3{X: vpos.X, Y: vpos.Y, Z: vpos.Z})
		}

		if e.Killer != nil {
			recentDeaths = append(recentDeaths, modules.DeathRecord{
				KillerName: e.Killer.Name,
				VictimName: e.Victim.Name,
				Tick:       tick,
			})

			// Check kill anomalies (flash/smoke) strictly on actual KILL events
			kast := getState(e.Killer)
			if kast != nil && e.Weapon != nil {
				class := e.Weapon.Class()
				if class == common.EqClassRifle || class == common.EqClassPistols || class == common.EqClassSMG || class == common.EqClassHeavy {
					kpos := e.Killer.Position()
					eyeOffset := 64.0
					if e.Killer.IsDucking() {
						eyeOffset = 46.0
					}
					killerEye := core.Vec3{X: kpos.X, Y: kpos.Y, Z: kpos.Z + eyeOffset}
					vpos := e.Victim.Position()
					victimVec := core.Vec3{X: vpos.X, Y: vpos.Y, Z: vpos.Z + 36.0}

					flashRem := float32(e.Killer.FlashDurationTimeRemaining().Seconds())
					flashAlpha := float32(0)
					if flashRem > 2.5 { // Only consider fully white blind kills
						flashAlpha = 255.0
					}

					modules.CheckKillAnomalies(
						kast, tick,
						flashAlpha, flashRem,
						killerEye, victimVec,
						e.Victim.Name, activeSmokes,
						roundBullets[e.Killer.Name],
					)
				}
			}
		}

		delete(roundBullets, e.Victim.Name)
	})

	p.RegisterEventHandler(func(e events.PlayerFlashed) {
		if e.Attacker == nil || e.Player == nil || e.Attacker == e.Player {
			return
		}
		if e.Attacker.Team != e.Player.Team {
			return
		}
		// Attacker threw the flashbang that blinded a teammate
		ast := getState(e.Attacker)
		if ast != nil {
			modules.RegisterTeamflash(ast, e.FlashDuration().Seconds(), false)
		}
	})

	p.RegisterEventHandler(func(e events.ChatMessage) {
		if e.Sender == nil {
			return
		}
		chatMsg := chat.Inspect(
			e.Sender.Name, e.Sender.SteamID64,
			p.GameState().IngameTick(), currentRound,
			e.Text, !e.IsChatAll, e.Sender.IsAlive(),
		)
		if chatMsg != nil {
			chatLog = append(chatLog, *chatMsg)
		}

		if cmd, target, ok := modules.ParseAdminChatCommand(e.Text); ok {
			tick := p.GameState().IngameTick()
			if !isDuplicateAdminEvent(adminLog, e.Sender.Name, cmd, tick) {
				adminLog = append(adminLog, modules.BuildAdminEvent(e.Sender.Name, cmd, target, tick, currentRound, recentDeaths))
			}
		}
	})

	// SourceMod activity broadcasts ("[SM] Admin: Slayed X.") arrive as server SayText
	p.RegisterEventHandler(func(e events.SayText) {
		admin, cmd, target, ok := modules.ParseAdminActivity(e.Text)
		if !ok {
			return
		}
		tick := p.GameState().IngameTick()
		if !isDuplicateAdminEvent(adminLog, admin, cmd, tick) {
			adminLog = append(adminLog, modules.BuildAdminEvent(admin, cmd, target, tick, currentRound, recentDeaths))
		}
	})

	lastProgress := time.Now()

	// Frame Hook: optimized per-tick engine processing
	p.RegisterEventHandler(func(e events.FrameDone) {
		if time.Since(lastProgress) > 150*time.Millisecond {
			lastProgress = time.Now()
			if opts.OnProgress != nil {
				opts.OnProgress(float64(p.Progress()))
			}
			select {
			case <-opts.Cancel:
				p.Cancel()
			default:
			}
		}

		gs := p.GameState()
		tick := gs.IngameTick()

		type cachedPlayer struct {
			pl          *common.Player
			st          *core.PlayerState
			buttons     int
			isJumping   bool
			isAttacking bool
			onGround    bool
			inWater     bool
			onLadder    bool
			speed       float64
			origin      core.Vec3
		}

		var alivePlayers []cachedPlayer

		for _, pl := range gs.Participants().Playing() {
			if pl == nil || pl.Name == "" || pl.IsBot {
				continue
			}
			st := getState(pl)
			if st == nil {
				continue
			}

			pos := pl.Position()
			vel := pl.Velocity()
			flags := int(pl.Flags())
			yaw := pl.ViewDirectionX()
			pitch := pl.ViewDirectionY()
			if pitch > 180 {
				pitch -= 360
			}

			st.Pitch = pitch
			st.Yaw = yaw
			st.Origin = core.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z}
			st.Team = int(pl.Team)
			st.IsAlive = pl.IsAlive()

			var buttons int
			if pl.Entity != nil {
				if prop := pl.Entity.Property("localdata.m_Local.m_nOldButtons"); prop != nil {
					buttons = prop.Value().Int()
				}
			}

			if !pl.IsAlive() {
				modules.ProcessAimTick(st, tick, pitch, yaw, tickrate, false)
				continue
			}

			flashRem := float32(pl.FlashDurationTimeRemaining().Seconds())
			flashAlpha := float32(0)
			if flashRem > 2.5 {
				flashAlpha = 255.0
			}
			st.FlashAlpha = flashAlpha
			st.FlashDuration = flashRem

			moveType := 2
			if pl.Entity != nil {
				if prop := pl.Entity.Property("movetype"); prop != nil {
					moveType = prop.Value().Int()
				}
			}

			speed := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)

			cp := cachedPlayer{
				pl:          pl,
				st:          st,
				buttons:     buttons,
				isJumping:   (buttons & core.IN_JUMP) != 0,
				isAttacking: (buttons & core.IN_ATTACK) != 0,
				onGround:    (flags & core.FL_ONGROUND) != 0,
				inWater:     (flags & core.FL_INWATER) != 0,
				onLadder:    moveType == core.MOVETYPE_LADDER,
				speed:       speed,
				origin:      st.Origin,
			}
			alivePlayers = append(alivePlayers, cp)
		}

		// Pre-compute enemy hitboxes per team once
		enemyHitboxes := make(map[int][]modules.EnemyHitbox)
		for _, cp := range alivePlayers {
			team := int(cp.pl.Team)
			h := 72.0
			if cp.pl.IsDucking() {
				h = 54.0
			}
			box := modules.EnemyHitbox{
				Name:  cp.pl.Name,
				Team:  team,
				Mins:  core.Vec3{X: cp.origin.X - 16, Y: cp.origin.Y - 16, Z: cp.origin.Z},
				Maxs:  core.Vec3{X: cp.origin.X + 16, Y: cp.origin.Y + 16, Z: cp.origin.Z + h},
				Pos:   cp.origin,
				Speed: cp.speed,
			}
			for otherTeam := 2; otherTeam <= 3; otherTeam++ {
				if otherTeam != team {
					enemyHitboxes[otherTeam] = append(enemyHitboxes[otherTeam], box)
				}
			}
		}

		playersByName := make(map[string]*common.Player, len(alivePlayers))
		for _, cp := range alivePlayers {
			playersByName[cp.pl.Name] = cp.pl
		}

		// Prefire: who sees whom. One spotted-mask read per player per tick, then bit checks.
		masks := make([]uint64, len(alivePlayers))
		maskOK := make([]bool, len(alivePlayers))
		for i, cp := range alivePlayers {
			masks[i], maskOK[i] = spottedMask(cp.pl)
		}
		for _, a := range alivePlayers {
			if a.st.SteamID == 0 || a.pl.EntityID <= 0 {
				continue
			}
			bit := uint(a.pl.EntityID - 1)
			for j, v := range alivePlayers {
				if v.pl.Team == a.pl.Team || !maskOK[j] || v.pl.SteamID64 == 0 {
					continue
				}
				modules.UpdatePeekVisibility(a.st, v.pl.SteamID64, masks[j]&(1<<bit) != 0, tick)
			}
		}

		// Run detection modules
		for _, cp := range alivePlayers {
			st := cp.st

			// Module A: Bhop
			modules.ProcessBhopTick(st, tick, cp.onGround, cp.isJumping, cp.onLadder, cp.inWater, cp.speed)

			// Module B: Triggerbot
			attackRisingEdge := cp.isAttacking && !st.PrevAttack
			st.PrevAttack = cp.isAttacking

			eyeOffset := 64.0
			if cp.pl.IsDucking() {
				eyeOffset = 46.0
			}
			eyePos := core.Vec3{X: cp.origin.X, Y: cp.origin.Y, Z: cp.origin.Z + eyeOffset}
			enemies := enemyHitboxes[st.Team]
			modules.ProcessTriggerTick(st, tick, st.Pitch, st.Yaw, eyePos, enemies, attackRisingEdge, tickrate)

			// Module W: Wallhack — only enemies the tracker has NOT spotted (no line of sight).
			// Runs before Module C so PrevPitch/PrevYaw still hold last tick's angles.
			tracker := cp.pl
			modules.ProcessWallTrackTick(st, tick, eyePos, st.Pitch, st.Yaw, cp.speed, enemies, func(name string) bool {
				enPl := playersByName[name]
				return enPl == nil || isSpottedBy(enPl, tracker)
			})

			// Module G (ghosting) stays unwired: dead teammates calling out over voice is legal,
			// so a prefire on the killer's spot is not proof — see modules.CheckGhosting.

			// Module C: Aimlock
			modules.ProcessAimTick(st, tick, st.Pitch, st.Yaw, tickrate, true)

			// Module E: Macro
			if attackRisingEdge && cp.pl.ActiveWeapon() != nil {
				modules.ProcessMacroTick(st, tick, true, cp.pl.ActiveWeapon().String(), tickrate)
			}
		}
	})

	// Parse stream
	parseErr := p.ParseToEnd()
	if errors.Is(parseErr, demoinfocs.ErrCancelled) {
		return nil, ErrCanceled
	}
	if parseErr != nil && parseErr != demoinfocs.ErrUnexpectedEndOfDemo {
		// Log but continue to finalize collected data
		if opts.Verbose {
			fmt.Printf("[!] Parsing warning: %v\n", parseErr)
		}
	}

	// Consolidate final player stats
	var playerStats []core.PlayerStats
	for _, st := range states {
		// Skip spectators / non-active players with 0 jumps and 0 engagements
		if st.Bhop.TotalJumps < 2 && len(st.Trigger.ReactionDeltas) == 0 &&
			len(st.Aim.SnapEvents) == 0 && len(st.Aim.SilentEvents) == 0 &&
			st.Grief.TeamflashSecs < 5.0 && len(st.Ghost.Incidents) == 0 {
			continue
		}

		s := st.ComputeStats(tickrate)
		playerStats = append(playerStats, s)
	}

	// Sort: cheaters / suspicious first, then by perf ratio descending. Name and
	// SteamID break ties so the same demo always yields the same order (players come
	// from a map, whose iteration order is random).
	sort.Slice(playerStats, func(i, j int) bool {
		a, b := playerStats[i], playerStats[j]
		if len(a.Flags) != len(b.Flags) {
			return len(a.Flags) > len(b.Flags)
		}
		if a.PerfRatio != b.PerfRatio {
			return a.PerfRatio > b.PerfRatio
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.SteamID < b.SteamID
	})

	// Consolidate and sort full Overwatch timeline
	var timeline []core.TimelineIncident
	for _, ps := range playerStats {
		timeline = append(timeline, ps.Incidents...)
	}
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Tick < timeline[j].Tick
	})

	// Fingerprint the match so duplicates are recognized, then look up reoffenders.
	// Recording the match in the database is up to the caller (db.RecordReport), once
	// it has decided the demo is not a duplicate.
	var steamIDs []uint64
	for _, ps := range playerStats {
		if ps.SteamID != 0 {
			steamIDs = append(steamIDs, ps.SteamID)
		}
	}
	matchID := core.MatchFingerprint(header.MapName, header.PlaybackTicks, len(roundStarts), steamIDs)

	pdb := db.GetDatabase()
	for i := range playerStats {
		aliases, infractions := pdb.LookupPlayer(playerStats[i].SteamID, matchID)
		// Only cheat or suspect detections make a reoffender; past conduct issues don't
		for _, inf := range infractions {
			if l := core.LevelOf(inf.Flags); l == core.LevelCheat || l == core.LevelSuspect {
				playerStats[i].PreviousInfractions++
			}
		}
		playerStats[i].PreviousNames = aliases

		// Attach player chat messages
		for _, msg := range chatLog {
			if msg.SenderSteamID == playerStats[i].SteamID {
				playerStats[i].ChatMessages = append(playerStats[i].ChatMessages, msg)
			}
		}
	}

	report := &core.DemoReport{
		Header: core.DemoHeader{
			MapName:     header.MapName,
			ServerName:  header.ServerName,
			TotalTicks:  header.PlaybackTicks,
			Tickrate:    tickrate,
			Date:        time.Now().Format("2006-01-02"),
			DemoFile:    filepath.Base(opts.DemoPath),
			RoundStarts: roundStarts,
			MatchID:     matchID,
		},
		Players:  playerStats,
		AdminLog: adminLog,
		ChatLog:  chatLog,
		Timeline: timeline,
	}

	return report, nil
}

// isSpottedBy wraps the m_bSpottedByMask lookup, which panics on demos missing the netprop.
func isSpottedBy(target, spotter *common.Player) (spotted bool) {
	defer func() {
		if recover() != nil {
			spotted = true // unknown visibility: assume visible to avoid false positives
		}
	}()
	return target.IsSpottedBy(spotter)
}

// isDuplicateAdminEvent drops the SourceMod broadcast that echoes a chat command already logged.
func isDuplicateAdminEvent(log []core.AdminEvent, admin, cmd string, tick int) bool {
	for i := len(log) - 1; i >= 0; i-- {
		ev := log[i]
		if tick-ev.Tick > 64 {
			break
		}
		if ev.AdminName == admin && ev.Command == "sm_"+cmd {
			return true
		}
	}
	return false
}

// spottedMask returns the bitmask of players (by EntityID-1) that currently see pl.
// CS:GO splits it in two 32-bit netprops; ok is false when the demo lacks them.
func spottedMask(pl *common.Player) (mask uint64, ok bool) {
	defer func() {
		if recover() != nil {
			mask, ok = 0, false
		}
	}()
	if pl.Entity == nil {
		return 0, false
	}
	lo := pl.Entity.Property("m_bSpottedByMask.000")
	if lo == nil {
		return 0, false
	}
	mask = uint64(uint32(lo.Value().IntVal))
	if hi := pl.Entity.Property("m_bSpottedByMask.001"); hi != nil {
		mask |= uint64(uint32(hi.Value().IntVal)) << 32
	}
	return mask, true
}
