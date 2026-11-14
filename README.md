<p align="center">
  <img src="cmd/sourcesight/winres/icon_128.png" width="88" alt="SourceSight icon">
</p>

<h1 align="center">SourceSight</h1>

<p align="center">
  <a href="README.md"><img src="https://img.shields.io/badge/lang-English-e07b2a?style=for-the-badge" alt="English"></a>
  <a href="README.es.md"><img src="https://img.shields.io/badge/idioma-Español-2a2c30?style=for-the-badge" alt="Español"></a>
</p>

<p align="center">
  <a href="https://github.com/Hdauu/SourceSight/releases/latest"><img src="https://img.shields.io/github/v/release/Hdauu/SourceSight?label=download&color=e07b2a" alt="Download"></a>
  <img src="https://img.shields.io/badge/platform-Windows-2a2c30" alt="Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT%20%2B%20Commons%20Clause-2a2c30" alt="License"></a>
</p>

A tool for CS:GO (Legacy) server staff. Open a demo (`.dem`) and it flags players whose behavior doesn't look human: autobunny, triggerbot, spinbot, wallhack, macros and more. Every incident comes with its exact tick so you can watch it in-game with one click.

Runs on Windows, fully local. Nothing is uploaded anywhere.

> **Results are leads, not proof.** Every detection explains what it measured and with which numbers, and comes with the command to go and watch it. Watch the demo before punishing anyone.

The app itself is in Spanish; this document describes it in English.

## Contents

- [Install](#install)
- [Quick start](#quick-start)
- [How a demo is analyzed](#how-a-demo-is-analyzed)
- [How the app is organized](#how-the-app-is-organized)
- [A demo report](#a-demo-report)
- [Levels](#levels)
- [Detections](#detections)
  - [Autobunny](#autobunny) · [Triggerbot](#triggerbot) · [Aimlock and spinbot](#aimlock-and-spinbot) · [Silent aim](#silent-aim) · [Wallhack](#wallhack) · [Prefire](#prefire) · [Macro](#macro) · [Blind kills](#blind-kills) · [Kills through smoke](#kills-through-smoke) · [AFK](#afk) · [Team flashes](#team-flashes)
- [Where they hit: the body figure](#where-they-hit-the-body-figure)
- [Ping and lag](#ping-and-lag)
- [Automatic observations](#automatic-observations)
- [SHIFT log](#shift-log)
- [Minimap](#minimap)
- [Admins and chat](#admins-and-chat)
- [History and duplicate demos](#history-and-duplicate-demos)
- [Players: averages across demos](#players-averages-across-demos)
- [Confirming a cheater](#confirming-a-cheater)
- [Watch in game](#watch-in-game)
- [Automatic demo archive](#automatic-demo-archive)
- [Exporting evidence](#exporting-evidence)
- [Discord bot](#discord-bot)
- [Updates](#updates)
- [Where data is stored](#where-data-is-stored)
- [What isn't detected](#what-isnt-detected)
- [Building and testing](#building-and-testing)
- [Releasing](#releasing)
- [License](#license)

## Install

1. Download **`SourceSight-Setup-vX.Y.Z.exe`** from the [latest release](https://github.com/Hdauu/SourceSight/releases/latest).
2. Run it. It installs for your user only, without asking for administrator rights.
3. *"Windows protected your PC"* may show up because the installer isn't signed. Click **More info → Run anyway**.

Only need the Discord bot? The same release has a standalone `SourceSightBot.exe`, nothing to install. See [Discord bot](#discord-bot).

Requires the WebView2 runtime, which ships with Windows 10 and 11.

## Quick start

1. Drop one or more demos (or a whole folder) on the window, or use **Abrir demos** (open demos). The installer also adds *Open with → SourceSight* for `.dem` files.
2. With one demo its report opens. With several, you see each one's progress and a summary at the end; click any to open it.
3. Click a player to see their metrics, where they hit, their incidents per round and the **Ver** (watch) button of each one.
4. The **Jugadores** (players) tab adds each player up across all your demos and compares them with everyone else.

To use **Ver**, set the CS:GO folder once in **Ajustes** (settings, bottom left of the menu).

## How a demo is analyzed

SourceSight reads the demo with [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) v4 (the last one that supports CS:GO; v5 is CS2 only) and walks through the match tick by tick. On every tick, for every living player, it takes their position, where they look, which buttons they press, their speed and which enemies they can see. Each detector gathers evidence as the match goes on. At the end the thresholds are applied and the report is built.

A 35-minute demo at 128 tick (about 270,000 ticks) takes 4 to 8 seconds.

**Several demos at once.** When you open several, they are analyzed in parallel:

- One analysis per CPU thread, minus one left free so the window stays responsive.
- Each analysis uses up to about 300 MB of memory (measured: a 207 MB demo took about 210 MB), so it never uses more than half the free RAM: with little memory it analyzes fewer at once.
- The heavy part (reading and analyzing) runs in parallel. Saving to the history, asking about duplicates and updating the reoffender database happen **one at a time**, in the order demos finish. So if you drop two copies of the same match together, the second is still recognized as a duplicate.
- The screen shows each demo with its state (queued, checking, analyzing, saving, done, already stored, duplicate, error, canceled), its progress bar, its result and how long it took. On top there is an overall bar with the time left. **Cancelar** (cancel) really stops the running analyses.
- At the end there is a summary ("2 con cheat · 1 sospechosa · 5 limpias…").

Measured with 12 real demos on a 12-thread CPU: 62 s one by one, 11 s in parallel. Reports come out identical either way, and a test checks it.

## How the app is organized

On the left, a menu with the sections:

- **Inicio** (home): what's worth reviewing first. The latest demos with cheats or suspects, the history totals, the automatic archive status and a notice if there are demos to re-analyze.
- **Demos:** the history on the left (search, filter by level or category) and the chosen demo's report on the right.
- **Jugadores** (players): the list of players across your demos and each one's page. With no player chosen it shows "Todas las demos", with totals and how everyone hits on average.
- **Archivo** (archive): the automatic archive of the server's demos: status, settings and the list of demos it kept.
- **Ajustes** (settings): game folder, seconds before the play, incident hotkeys, SHIFT log distance, thresholds and version.

On top, when a demo is open: its map, file and server, **Re-analizar** (re-analyze) and **Exportar** (export: PNG image or summary for Discord). **Abrir demos** (open demos) is always there. Demos can be dropped anywhere on the window.

## A demo report

On top, a summary (players, cheats, length, reoffenders). Below, tabs:

- **Jugadores** (players): one row per player with their result, perfs, reaction, prefire, head %, snaps, SHIFT cases and ping. Sortable by any column (worst result first by default) and filterable by level or name. Click a player for their detail. Names made of invisible characters show as "(nombre invisible)".
- **Incidentes** (incidents): each incident with its round, type, player, command and **Ver** button.
- **Mapa** (map): see [Minimap](#minimap).
- **SHIFT:** see [SHIFT log](#shift-log).
- **Chat y admins:** admin commands and chat.

The app remembers the last tab you used.

## Levels

Every player gets a level. A demo's level is the worst of its players.

| Level | Meaning |
|---|---|
| **Cheat detectado** (cheat detected) | Crossed a strong threshold: autobunny, triggerbot, aimlock/spinbot, silent aim, wallhack or macro |
| **Sospechoso** (suspicious) | Weaker signs worth a look: high bhop or perfect-hop streaks, some sharp turns, prefire, blind or through-smoke kills |
| **Conducta** (conduct) | Not cheating but relevant to staff: AFK, flashing their own team |
| **Limpio** (clean) | Nothing found |
| **Confirmado** (confirmed) | Set by an admin after watching the demo, never by the app. Stored per SteamID (see [Confirming a cheater](#confirming-a-cheater)) |

The history can be filtered by level and by detection category.

## Detections

All thresholds live in `pkg/core/constants.go`, and the rules that set the level in `pkg/core/state.go` and `pkg/core/level.go`. Times defined in ticks are meant for 64 tick and scaled to the demo's real tickrate, so the same rule holds for 64- and 128-tick demos.

Each detection explains: what it looks for, which data it uses, how it computes, when it flags and why that number, what can fool it and how to verify it.

### Autobunny

`pkg/modules/bhop.go` · level **cheat** (AUTOBUNNY) or **suspicious** (SOSPECHOSO-BHOP)

**What it looks for.** Scripts that jump on their own the exact moment the player touches the ground, to chain jumps without losing speed.

**Data.** Whether the player is on the ground (`FL_ONGROUND` flag), whether they press jump (`IN_JUMP`), ladders, water and speed, every tick.

**Computation.**
- When the player lands, the tick is noted. When they take off holding jump, the ticks spent on the ground are counted.
- Taking off after 0 or 1 tick on the ground is a **perfect** hop: the jump runs before friction and no speed is lost.
- Taking off after more than 15 ticks on the ground doesn't count: that's walking and jumping, not bunny hopping. Ladders and water don't count.
- Total jumps, perfect ones, the percentage (*perfs*) and the longest streak of perfect hops in a row are stored.

**When it flags.**
- **AUTOBUNNY (cheat):** 75% perfect or more, over at least 10 jumps.
- **SOSPECHOSO-BHOP:** with 30+ jumps, 55% perfect or a streak of 5 in a row.

**Why.** A person with the mouse wheel lands 20% to 50% perfect hops. A sustained 75% isn't doable by hand. A streak of 5 is very unlikely, but a good bhopper can pull it off once, so on its own it's only suspicious. Fewer than 30 jumps is too small a sample for "suspicious".

**Limits.** A server with its own autobunny (bhop plugin) makes everyone jump perfectly. On those servers this detection is useless.

**How to verify.** In the demo, with `host_timescale 0.25`, watch how they chain the jumps. With a script the player never loses or gains speed irregularly and never "misses" a hop. Every streak of 5+ is an incident with its tick.

### Triggerbot

`pkg/modules/trigger.go` · level **cheat** (TRIGGERBOT)

**What it looks for.** Programs that fire on their own as soon as the crosshair is over an enemy.

**Data.** Where the player looks, the position of their eyes and the collision box of every living enemy, every tick, and the moment they press fire.

**Computation.**
- A ray is cast from the player's eyes in the direction they look and intersected with each enemy's box.
- When the ray enters an enemy, a timer starts. When fire is pressed, the elapsed time (*reaction*) is recorded, between 0 and 1.5 s.
- If the crosshair leaves the enemy before the shot, the measurement is dropped.
- Reactions during a lag spike are discarded (see [Ping and lag](#ping-and-lag)).

**When it flags.** With at least 4 reactions measured, an average under **80 ms** and a deviation under **15 ms**.

**Why.** A person averages 150 to 250 ms, with a lot of variation from one time to the next. A triggerbot reacts fast and always the same. Both are required: speed and regularity.

**Limits.** It measures against the collision box, not the exact body shape. A player already spraying when the enemy crosses their crosshair doesn't produce a new reaction.

**How to verify.** Watch the incidents in slow motion: with a triggerbot the shot comes right as the crosshair touches the enemy, even when they barely peek.

### Aimlock and spinbot

`pkg/modules/aim.go` · level **cheat** (AIMLOCK) or **suspicious** (SOSPECHOSO-AIM, POSIBLE-LAG)

**What it looks for.** A view that jumps onto the target (aimlock) or spins nonstop (spinbot / anti-aim).

**Data.** The view angles (pitch and yaw) every tick.

**Computation.**
- How much the view turned between one tick and the next is measured, correcting the 359° → 0° wrap.
- A turn of more than **45° in a single tick** is a *snap* (at 128 tick, a tick is 7.8 ms).
- The first 3 seconds after spawning and changes to a pitch of exactly 0 are ignored: those are angles forced by the server on spawn or teleport.
- **Turns measured during a connection hiccup don't count.** With lag, the player's commands arrive bunched and several ticks of turning show up at once in one. A snap is dropped if on that tick: the server ran more than one of the player's commands (seen in their *tickbase*, which advances one per command), the player moved faster than 1200 u/s (a warp; nobody runs that fast, not even bunny hopping), or their ping spiked. Those turns are counted separately ("Giros con lag") and, if there are 5 or more, the observations say so. A high but steady ping causes no snaps and nothing is discounted.

**When it flags.** Only when snaps cluster: at least 5 within one second and 10 or more in the match.
- 50 snaps or more: **AIMLOCK (cheat)**.
- 10 to 49: **SOSPECHOSO-AIM**.
- **POSIBLE-LAG** instead of either if the player has a bad connection (average ping of 80 ms or more) and at least 20% of their sharp turns already landed on connection hiccups. With a connection like that, the lag filter can't see every hiccup and the remaining snaps are probably lag too. It stays suspicious so someone looks at the demo. It never applies with a low ping, however many lag turns.

**Why.** A fast human flick can produce a single snap, but never 5 in a second. A spinbot produces hundreds: the yaw jumps to a different value every tick, usually with the pitch stuck at ±89°. Between 10 and 49 it can be a cheat toggled briefly or odd flicks, worth a look.

**Limits.** A "smooth" aimbot that turns fast but spread over several ticks never exceeds 45° per tick and isn't caught this way.

**How to verify.** Go to the incident's tick and watch first-person: a spinbot shows as the view spinning out of control. From outside, the model spins or looks at the floor.

### Silent aim

`pkg/modules/aim.go` · level **cheat** (SILENT AIM)

**What it looks for.** Cheats that send the server a different firing angle from the one on screen: the bullet hits even though the crosshair isn't on the enemy.

**Data.** Where the shooter was looking when firing, the position of their eyes, and the victim's position when damaged.

**Computation.** When a shot deals damage (up to 16 ticks after the shot), the direction the shooter was looking is compared with the real direction from their eyes to the victim's chest.

**When it flags.** When the difference exceeds **35°**, 3 times in the match.

**Why.** Firing on the move or jumping with SMGs and shotguns, spread can reach about 35°. More than that isn't weapon inaccuracy: the bullet went somewhere other than the view.

**Limits.** Wallbangs and odd damage registration can occasionally give large differences, hence the 3 required.

**How to verify.** At the incident's tick, first-person: the enemy takes damage without being near the crosshair.

### Wallhack

`pkg/modules/wallhack.go` · level **cheat** (WALLHACK)

**What it looks for.** Players following with their crosshair an enemy who is behind a wall.

**Data.** The demo stores which enemies each player has "spotted" (`m_bSpottedByMask`, the same data that feeds the radar). Plus where each player looks, positions and speeds.

**Computation.**
- Every tick, it looks for a **non-spotted** enemy within 4° of the crosshair (aiming at the chest), 150 to 3000 units away.
- If neither of them moves it doesn't count: a still crosshair on a still enemy is crosshair placement.
- If the crosshair follows that same enemy for **1 second**, sweeping **20° or more**, it's a tracking.

**When it flags.** With 3 trackings in the match.

**Why.** The sweep is what matters: holding the crosshair on an angle someone is about to come out of is good placement, not a wallhack. Following someone you can't see across 20° for a second is something else. Under 150 units is dropped because up close you hear and follow people easily, and over 3000 because at that distance any crosshair passes near someone.

**Limits.** "Spotted" is the game's approximation, not the map's exact geometry. Following someone behind a thin wall by their footsteps is legitimate and can look similar. That's why 3 cases are needed.

**How to verify.** Watch each tracking first-person with `spec_show_xray 1`, or with the radar: the crosshair moves with the enemy through the wall.

### Prefire

`pkg/modules/peek.go` · level **suspicious** (PREFIRE)

**What it looks for.** Players who damage an enemy right as they appear, faster than a person can react, over and over. It suggests they already knew where the enemy would show up (wallhack) or a triggerbot.

**Data.** The same "spotted" data as the radar, and the damage of each bullet.

**Computation.**
- For each fight, it measures from the moment the enemy becomes visible to the player until the first bullet that damages them (2 s at most; later than that isn't a reaction to that appearance).
- A measurement under **100 ms** is a *prefire*.
- Measurements during a lag spike are discarded (see [Ping and lag](#ping-and-lag)).

**When it flags.** With at least 15 fights measured, **35% or more** prefires and a median under **150 ms**.

**Why.** A person needs about 200 to 400 ms to see and react. A single prefire happens (pre-firing a known angle is a legal technique), but not in a third of the fights. Calibrated on 13 real demos: it only flagged a player reported as a cheater.

**Limits.** Pre-firing common angles is legitimate and some maps have many. That's why it's only "suspicious" and needs many fights.

**How to verify.** The player page compares their prefire % across all demos with everyone else's: the pattern shows better over several matches. Watch the incidents: do they shoot before the enemy finishes peeking, at spots that aren't typical angles?

### Macro

`pkg/modules/macro.go` · level **cheat** (MACRO)

**What it looks for.** Software or mice with macros that click for the player with a semi-automatic pistol to fire as fast as the weapon allows.

**Data.** Every time the player presses fire (the rising edge of `IN_ATTACK`) and which weapon they hold.

**Computation.**
- Only pistols that fire once per click: Glock, USP-S, P2000, P250, Desert Eagle, Dual Berettas, Five-SeveN, Tec-9 and R8. The weapon is recognized by type. The CZ75 is left out because it's automatic.
- The time between consecutive clicks is measured. A pause of more than 500 ms or switching weapons starts a new burst, so the pause between fights doesn't mix in.
- The deviation of the last 10 intervals of the burst is checked.

**When it flags.** When those 10 intervals have a deviation under **1.5 ms**. One incident per burst, at most one every 5 s.

**Why.** A finger varies naturally: click to click it changes by several milliseconds. A macro clicks at an almost constant rhythm. At 128 tick a tick is 7.8 ms, so a deviation under 1.5 ms over 10 intervals means practically all of them were exactly the same.

**Limits.** Firing with the mouse wheel (`bind mwheeldown +attack`) on free-spinning wheels can be very regular. If a MACRO shows up, check whether the player uses the wheel.

**How to verify.** Watch the burst in slow motion and listen to the rhythm of the shots: with a macro it sounds like a metronome.

### Blind kills

`pkg/modules/flash_smoke.go` · level **suspicious** (ANTI-FLASH)

**What it looks for.** Cheats that remove the flashbang effect.

**Data.** At each firearm kill, how much blindness the killer has left.

**Computation.** A kill with more than 2.5 s of blindness left (fully white screen) counts as a blind kill.

**When it flags.** With 4 blind kills in the match.

**Why.** Fully blind you can kill by chance (spraying where the enemy was), but not 4 times.

**Limits.** Flashes from the side or behind blind less than the game reports in some cases.

**How to verify.** Watch the kill first-person: with anti-flash the player aims precisely with a white screen.

### Kills through smoke

`pkg/modules/flash_smoke.go` · level **suspicious** (SMOKE-VISION)

**What it looks for.** Cheats that let you see through smoke.

**Data.** Active smoke grenades (position, start and end), the killer's eye position and the victim's, and the bullets the killer fired in the round.

**Computation.**
- A segment is drawn between the killer's eyes and the victim. Only a smoke **between** them counts (a sphere of 144 units radius): one behind the victim blocks nothing.
- The smoke must exist at that moment: a smoke lasts 18 s from when it pops and disappears at round end. The game removes smokes on round change without recording it in the demo, so the duration is used rather than the smoke-end event.
- If the killer also fired 3 bullets or fewer in the round, it's a kill through smoke with few bullets: not a blind spray.

**When it flags.** With 5 such kills in the match.

**Why.** Firing through smoke at footsteps is legitimate, but usually takes many bullets. Killing 5 times with 3 bullets or fewer suggests they could see the enemy.

**Limits.** The smoke is modeled as a sphere, but the real one deforms with the map. An enemy visible at the edge of the smoke can count.

**How to verify.** At the kill, first-person: was the enemy visible? Did the player aim straight at them?

### AFK

`pkg/modules/staff.go` · level **conduct** (AFK/GRIEF)

**What it looks for.** Players who aren't playing.

**Computation.** At the end of each round, how far the player moved since the end of the previous round is measured. Under 50 units is a still round.

**When it flags.** With 3 still rounds in a row.

**Limits.** A player who always holds the same spot and dies quickly may barely move. That's why it's conduct, not cheating.

### Team flashes

`pkg/modules/staff.go` · level **conduct** (TEAMFLASH)

**What it looks for.** Players who blind their own team.

**Computation.** The seconds of blindness each of the player's flashes caused to teammates are added up.

**When it flags.** With 35 seconds or more in the match.

**Why.** Blinding a teammate now and then happens. 35 seconds is systematic.

## Where they hit: the body figure

`pkg/core/hits.go`, `pkg/gui/ui/js/16-body.js`

Every bullet that damages an enemy comes with the body zone it hit (*hit group*): head, neck, chest, stomach, each arm and each leg. SourceSight stores, per player:

- shots fired, hits and damage dealt per zone;
- the same split by weapon type: rifles, snipers (AWP, SSG 08, autos), pistols, SMGs, shotguns and machine guns;
- kills and headshot kills.

Only bullets to enemies count: knife, grenades, fire and zeus don't.

**The figure.** A flat, minimal silhouette with each zone as a separate piece. Each zone is colored by how much the player hits it, on a heat scale from grey (none) to orange and pale yellow (a lot), with its % next to it. The scale is fixed (45% of hits is already the top color) and has its reference bar underneath, so two figures side by side can be compared at a glance. Arms and legs are the soldier's own, as if facing them: their left arm is on your right.

**Modes.**
- **Player:** their hits in that demo (player detail) or across all their demos (player page).
- **What they are compared with, never counting the player:** in a demo's player detail, **the rest of that match** (same map, same opponents); on the player page, **the rest of the players** in the history; in "Todas las demos", everyone's average.
- **Comparison:** both figures side by side.

It can be filtered by weapon type. Below there's a table with hits, damage, % and the difference from the average in points. A head difference of 10 points or more is shown in yellow, purely as information. With fewer than 10 hits it warns that the percentages don't mean much yet.

**Accuracy.** Hits divided by shots, **without shotguns**: each pellet that lands counts as a hit, and with shotguns accuracy would go over 100% (in real demos the Nova gave 141%). Pellets do count for the figure, since each one lands on a zone.

**What's normal.** In CS:GO, 10 to 25% of an average player's hits land on the head, most on the chest and stomach. A player who hits the head much more than everyone else **proves nothing**: good players hit the head more. It's something to look at along with the rest.

## Ping and lag

`pkg/core/ping.go`

The demo includes every player's ping (the scoreboard one, which the server updates about once a second). SourceSight samples it once a second **while the player is alive**, to skip the spikes of connecting, loading the map or spectating.

**Does high ping change the detections? No.** Reactions (prefire, triggerbot) are measured between two things that reach the server through the same player's connection: what they saw and when they fired. A steady ping delays both equally and cancels out. A high-ping player doesn't "look faster"; if they react late, they reacted late.

**What does mislead is a lag spike.** During a spike the player's commands arrive bunched and a normal reaction can look impossibly fast. So when the ping is **50 ms or more above that player's usual**, reactions measured at that moment are discarded and counted separately ("Reacciones con lag: N descartadas"). "Usual" is a moving average that doesn't follow spikes, so a long spike doesn't become the new normal.

Across 25 real demos, spikes cover 0.09% of the time players are alive, and none of 17,538 reactions were discarded: the filter is there for the rare case and doesn't change normal results.

Each player's average and peak ping are shown in their detail and page. Ping alone never flags anyone.

**Ping of each incident.** Every incident shows the ping the player had at that moment (the latest scoreboard reading, up to a second earlier), in yellow if it was a spike. It appears in the incident list, the player detail, the Discord summary and the bot. It shows at a glance whether something happened on a broken connection. Being one reading per second, a very short hiccup may not show.

## Automatic observations

`pkg/core/notes.go`

When one of a player's numbers stands out, the app writes it as a sentence in a **"Para mirar"** (worth a look) box. For example:

> El 46% de sus impactos fueron a la cabeza (51 de 111); el resto de la partida: 22%.
> *(46% of their hits went to the head (51 of 111); rest of the match: 22%.)*

**They are informative only: they never change the level or the detections.** They tell you what to look at in the demo.

**What they compare against.** A player is never compared with themselves:
- in a demo, against **the rest of the match**;
- on the player page (all their demos), against **the rest of the players in the history**.

**Rules.** Each needs enough data from the player and from the comparison (the comparison needs 3 times the player's minimum), so one lucky round produces nothing.

| What | Minimum | When it shows |
|---|---|---|
| Head hits | 20 hits | 30% or more, and 15+ points above the comparison |
| Accuracy (no shotguns) | 80 shots | 40% or more, and 1.6 times the comparison |
| Headshot kills | 8 kills | 75% or more, and 25+ points above |
| Time from crosshair to shot | 8 reactions | 60% or less of the comparison |
| Damage right as the enemy appears | 10 fights | 20%+ prefires without reaching the detection's 35%, and 2.5 times the comparison |
| High ping | — | 120 ms average or more |
| Reactions discarded for lag | — | If any were discarded |

Across 25 real demos, observations show up for 24 of 339 players (7%).

They appear in the player detail, the player page and the Discord bot's message.

## SHIFT log

`pkg/core/shift.go`, `pkg/gui/ui/js/17-shift.js`

**What it is.** A list of every moment a player walked with shift (`+speed`, walking silently) with an enemy they couldn't see close by. Walking silently is how you sneak up without being heard, so doing it right next to an enemy you supposedly don't know is there is worth a look. It's also completely normal play, so the log **never changes the level or the detections**: it's a list to review, with no criteria other than distance and visibility.

**Data.** Every tick: whether the player holds shift (the `IN_SPEED` button), the distance to each living enemy and whether they can see them (the same "spotted" data the radar uses).

**Computation.**
- A case starts when the player walks with shift and there's at least one enemy **they can't see** within **1000 units** (roughly the distance at which running footsteps are heard).
- While that enemy (or any other unseen one) stays close, everything goes into the **same case**: every new shift press counts as one more press, not as a new case.
- The case closes after 1 second with no enemy close, when the player dies or when the round ends.
- Warmup and the freeze at the start of each round don't count.
- If the demo lacks visibility data for an enemy, they count as seen: a case is never added for lack of data.

**What each case shows.** Round, player, the close enemies they couldn't see, how many times they pressed shift, how long they walked, the closest an enemy got, the `demo_gototick` command and the **Ver** (watch) button.

**Where it is.**
- In the demo report, the **"Registro de SHIFT"** section, with a player filter (sorted by number of cases). It shows 25 and more can be loaded.
- In a player's detail, a box with how many cases they had and a **Ver en el registro** button that opens the list filtered to them.
- The Discord bot shows a summary: how many cases and the players with the most.

**Distance.** Changed in **Ajustes** (settings) (500, 750, 1000, 1500 or 2000 units) and in the bot with `shiftRadius` in `bot_config.json`. It applies to demos analyzed or re-analyzed after changing it.

**What to expect.** Normal matches have many cases: across 25 real demos there were from a few dozen to about 300 per match, and up to about 45 per player. What's worth watching is a player who walks with shift **straight toward** an enemy they can't see, again and again, or stops right behind a wall where someone is.

**Limits.** "Spotted" is the game's approximation, not exact geometry. And the log can't tell whether the player walks because they heard the enemy, because it's the team's play or because they know where the enemy is: only watching the demo shows that.

## Minimap

`pkg/radar`, `pkg/core/mapdata.go`, `pkg/gui/ui/js/18-map.js`

**What it is.** The match drawn on the map's radar, in each demo report (the **"Mapa"** section).

**Where the radar comes from.** From **your installed CS:GO**: the radar image and its position file are in `csgo\resource\overviews`. SourceSight ships no Valve images: it reads them from your game, converts them to PNG once and keeps them in `%APPDATA%\SourceSight\radar`. That's why the CS:GO folder must be set in **Ajustes** (settings) (otherwise it tries to find it on its own). Maps with several floors (like Nuke) have an **Arriba / Abajo** (upper / lower) selector, and each point goes on the floor that matches its height.

**What the analysis stores.**
- Every kill, with the exact position of killer and victim, the weapon and whether it was a headshot.
- Each player's path: one position sample per second while alive.
- Warmup doesn't count. A full match adds about 300 KB to the report (about 100 KB compressed).

**Layers** (toggled independently):
- **Kills and deaths:** a circle in the team color where the killer stood, a cross where the victim died and a line between them. With one player selected, their deaths are red.
- **Path:** where they moved. With one player selected, their path in white; with everyone, all paths fainter. The line breaks on death, a new round or a floor change.
- **Heatmap:** where they spent the most time, with the same color scale as the body figure.
- **Incidents:** a diamond where the player was at each incident (red for triggerbot, silent aim and wallhack; yellow for the rest).
- **SHIFT:** a triangle where each SHIFT log case started, with a dashed line to each enemy they couldn't see.

**Filters.** By player and by round. From a player's detail, **Ver en el mapa** opens the map filtered to them with their path.

**Click a point.** Shows what happened (who killed whom, with what weapon and from how far; or the incident; or the SHIFT case) and the **Ver** button to open it in the game.

**Limits.** Incident and SHIFT case positions come from the nearest path sample, so they can be off by up to half a second of movement. Kill positions are exact.

## Admins and chat

`pkg/modules/staff.go`, `pkg/modules/chat.go`

**Admin commands.** Punishment commands typed in chat (`!slay`, `/kick`, `sm_ban`…) and SourceMod's server notices (`[SM] Admin: Slayed X.`) are recognized. Logged: slay, slap, kick, ban, burn, freeze, beacon, mute, gag, silence, timebomb, firebomb, freezebomb and drug.

**Revenge.** If an admin punishes (slay, slap, kick, ban, burn, timebomb, firebomb) the player who killed them in the previous **5 seconds**, it's flagged as possible revenge abuse. SourceMod targets by partial name, and SourceSight does the same to find the target.

**Chat.** Every message goes into the report with its round and tick. Flagged:
- **Toxicity:** severe insults and threats (list in Spanish, English and Portuguese).
- **Chat ghosting:** a dead player writing positions in all-chat ("están en b", "planted"…).
- **Spam binds:** the same message 3 times in a row, each less than 1 second after the previous one.

Chat doesn't change a player's level: it's information for staff.

## History and duplicate demos

`pkg/history`

Every analyzed demo goes into the history, and reopening it is instant (the stored report opens). The last 300 are kept. You can search by demo, map or player, and filter by level and category.

**Duplicates.** So no match counts twice in the averages:
- **Same file:** the file's SHA-256 is compared. If it's an exact copy under another name or folder, it asks whether to open the stored one, store it anyway or skip it.
- **Same match:** a match fingerprint (map, length, rounds and SteamIDs) is compared. Also, if map and rounds match, the length differs by under 2% and 80% of the players are shared, it's considered the same match too (another recording or a trimmed copy). It asks whether to replace the previous one, keep both or discard.

**Re-analyzing.** When a new version changes the detectors, the app shows which demos were analyzed by an older version and offers to re-analyze them all.

## Players: averages across demos

`pkg/history/players.go`

The **Jugadores** tab adds each player (by SteamID) up across all your demos:

- how many demos they appear in, how many they were flagged in and why;
- every name they used;
- totals of incidents, blind, smoke and macro kills;
- **their average against the rest of the players** (everyone else, without them, so the comparison isn't pulled toward their own numbers): bhop perfs, reaction, prefire, head hits, ping, snaps, silent aim and wallhack per demo. Values well above are shown in red;
- the where-they-hit figure, compared with the average;
- the observations of their demos added up;
- the list of their demos, click to open each.

**Search.** By name, SteamID64, Steam profile link, `STEAM_0:X:Y` or `[U:1:N]`.

The **Todas las demos** (all demos) view shows the history totals and how all players hit on average.

## Confirming a cheater

In a player's detail or page, **Confirmar cheater** marks them as confirmed by staff, with an optional note. The app never confirms on its own. It's stored per SteamID, shows in all their demos, and can be removed.

## Watch in game

`pkg/launcher`

Each incident's **Ver** button opens CS:GO at that moment of the demo:

1. If the demo isn't inside the game's `csgo` folder, a *junction* to the demo's folder is created inside `csgo`. Nothing is copied, it works across drives and needs no administrator rights.
2. If CS:GO is already open and SourceSight started it, the demo is sent through its remote console (`-netconport 2121`). If it's the demo already playing, it just jumps to the tick.
3. If it isn't open, it's started through Steam (`steam.exe -applaunch <id>`, with the App ID read from Steam's files). It waits for the main menu, loads the demo with `playdemo` and, once the map has loaded, jumps with `demo_gototick`.

It jumps a few seconds before the play (configurable in **Ajustes** (settings), 2.5 s by default), converted to ticks with the demo's tickrate.

If CS:GO is open but you started it yourself from Steam, it has no remote console: close it once and press Ver again, or add `-netconport 2121` to its launch options.

Every incident also has its `demo_gototick` command to copy and paste by hand.


**First-person camera.** When the moment belongs to a player (an incident, a SHIFT case, a kill on the minimap, a chat message), after seeking the app sends the game `spec_player_by_accountid` with the player's account number (from their SteamID) and `spec_mode 4`: the demo watches from their eyes. The game console shows a `SourceSight: …` line with which incident it is.

**Previous / next incident.** With the game in front, **Ctrl + Alt + ←** and **Ctrl + Alt + →** jump to the previous or next incident, also in first person. They walk the list of the last **Ver** you pressed: that player's incidents, or every incident of the demo if the moment wasn't a player's. They are Windows global hotkeys received by SourceSight and forwarded to the game through the remote console: no bind or game file is touched. They can be changed in **Ajustes** (settings) (Ctrl + Shift + arrows, Ctrl + Alt + Page Up / Page Down, Alt + Page Up / Page Down) or turned off. Function keys aren't used because CS:GO already uses several (F10 opens the quit dialog).

**Safety.** Anyone picks their player name, so before a name is shown in the game console `;`, quotes and backslashes are removed: a name can't sneak commands in.

## Automatic demo archive

`pkg/archive`

**What for.** Servers delete their demos after a few days (3y33, after 3). The automatic archive checks the demo page, analyzes new demos and **keeps only those where some player came out as cheat or suspicious**. That way the demos that matter stay forever, even after the server deletes them.

**How it works.**
1. Every so often (30 minutes to 4 hours, 1 hour by default) it reads the demo page (`https://csgo.3y33.com.ar/demos` by default) and its next pages, until it finds a whole page it already checked.
2. It keeps the new demos that pass the filters: **server** (#1 to #8), **mode** (Competitivo, AutoMIX, Retake, Casual, DeathMatch), **map**, and skips those under 5 MB (warmups, empty servers).
3. It downloads them **one at a time**, oldest first (the closest to being deleted), straight into the chosen folder, and checks they arrived whole.
4. It analyzes them like any demo: they go into the history and count toward the players' averages. A match already in the history isn't counted twice.
5. If some player came out as cheat or suspicious, **the file stays** in the folder. Otherwise it's deleted right away: the disk doesn't fill up.
6. It remembers which demos it checked, so it doesn't download them again.

**Where it runs.** In the app (while open) and in the bot (24/7), wherever you turn it on. If both are on the same PC, a lock file keeps them from downloading the same demo at once. It's a per-PC option and comes off: it's best turned on **on one PC only**, the bot's, so the same demos aren't downloaded several times.

**Internet data.** Downloading doesn't take disk, but it does use internet: downloading everything would be about 15 GB a day on 3y33. That's what the filters are for; for example, Competitivo only downloads much less.

**Settings.**
- **App:** the **Archivo** (archive) section of the menu: folder, servers, modes, maps, how often to check, demo page and the **Revisar ahora** (check now) button, with the result of the last check.
- **Bot:** in `bot_config.json`, `archive` (`enabled`, `dir`, `servers`, `modes`, `maps`, `minMB`, `everyMin`, `url`) and `archiveChannelId`, the channel where each kept demo is announced (if missing, the first of `channelIds`).

**Limits.** It reads the page the way a browser sees it: if the site changes its layout, the reader (`ParsePage`) has to be adapted. It only works with pages that allow downloading demos without logging in.

## Exporting evidence

- **Exportar evidencia** (in the player detail) creates a printable HTML with the player's data, their incidents with tick and command, and the demo's SHA-256, so another admin can check it's the same file.
- **Exportar PNG** creates an image of the demo report, the same one the bot posts.

Exports are saved in `Documents\SourceSight`.

## Discord bot

`cmd/bot`

1. Create an application and a bot in the [Discord developer portal](https://discord.com/developers/applications) and enable *Message Content Intent*.
2. Run `SourceSightBot.exe` once: it creates `bot_config.json` next to the executable. Paste the token there. **Don't share or commit that file.**
3. Run it again. Upload a demo to the channel and it replies with the report.

**What it replies.** A message with the result (cheat, suspicious or clean), the flagged players and their level, the observations, the first incidents with their `demo_gototick` command, the report image and a button per flagged player to see their detail.

**Configuration (`bot_config.json`).**

| Field | What it does | Default |
|---|---|---|
| `token` | Bot token | — |
| `prefix` | Help command | `!sourcesight` |
| `channelIds` | Channels it listens to (empty = all) | all |
| `maxDemoMB` | Maximum demo size | 250 |
| `maxParallel` | Demos analyzed at once | 2 |
| `shiftRadius` | SHIFT log distance | 1000 |
| `archive` | Automatic archive (see [Automatic demo archive](#automatic-demo-archive)) | off |
| `archiveChannelId` | Channel where kept demos are announced | the first of `channelIds` |

If more demos arrive than `maxParallel`, the rest wait their turn and the bot says how many are ahead. The bot shares the history and the reoffender database with the app when both run on the same PC and user, and doesn't count a match twice.

## Updates

On startup, the app checks GitHub for a new version and offers to install it with one click. It downloads the installer, verifies it against the SHA-256 that GitHub publishes, updates the app, the command line and the bot, and reopens SourceSight. Only offered on copies installed with the installer.

## Where data is stored

Everything stays on your PC:

| What | Where |
|---|---|
| History, reports, reoffenders, confirmations and settings | `%APPDATA%\SourceSight` |
| Exports (evidence, PNG) | `Documents\SourceSight` |
| Diagnostic log | `%APPDATA%\SourceSight\sourcesight.log` |

The only internet connection is the update check against GitHub (and Discord, if you use the bot).

## What isn't detected

- **Smooth aimbots**, which turn fast but spread over several ticks: they don't cross the snap threshold.
- **Discord ghosting:** a dead teammate calling positions over voice is legal on many servers, so firing where the killer was proves nothing. The code exists (`CheckGhosting`) but isn't used.
- **Exact visibility:** wallhack and prefire use the game's "spotted" data, an approximation, not the map's real geometry.
- **Backtrack and no-recoil** aren't analyzed yet.

## Building and testing

Requires Go 1.27 or newer.

```
go build -ldflags "-H windowsgui" -o SourceSight.exe ./cmd/sourcesight
go build -o SourceSightCLI.exe ./cmd/sourcesight
go build -o SourceSightBot.exe ./cmd/bot
go test ./...
```

Without `-X sourcesight/pkg/version.Version=vX.Y.Z` in `-ldflags` the version is `dev` and no update check runs.

**Tests.**
- `go test ./...` runs everything without real demos. Tests use a temporary data folder (`SOURCESIGHT_DATA_DIR`) and refuse to run unless it took effect, so they never touch your history.
- With real demos: `SOURCESIGHT_TEST_DEMOS=<folder with demos> go test ./pkg/parser ./pkg/gui -run "Parallel|Batch"` checks that parallel analysis gives the same result as one by one, and runs the whole engine with a duplicate demo and a cancel halfway through.
- CI also runs `go test -race` on Linux to catch state shared between analyses.

**Layout.**

| Folder | Contents |
|---|---|
| `cmd/sourcesight` | The app (window) and the command line |
| `cmd/bot` | The Discord bot |
| `pkg/parser` | Reading the demo and walking it tick by tick |
| `pkg/modules` | One file per detector |
| `pkg/core` | Thresholds, levels, hits, ping, observations |
| `pkg/history`, `pkg/db` | History, per-player averages and reoffenders |
| `pkg/gui` | The window; the page is in `pkg/gui/ui` (separate HTML, CSS and JS) |
| `pkg/launcher` | Watch in game |
| `pkg/radar` | Radars from the installed game for the minimap |
| `pkg/archive` | Automatic archive of the server's demos |

`third_party/go-webview2` is a copy of [jchv/go-webview2](https://github.com/jchv/go-webview2) (MIT) with a small patch to receive the real path of files dropped on the window. Details in `third_party/go-webview2/SOURCESIGHT_PATCH.md`.

The icon is generated with `go run ./tools/icongen`. Releases embed it in the `.exe` files with [go-winres](https://github.com/tc-hib/go-winres), and the installer is built with [Inno Setup](https://jrsoftware.org/isinfo.php) from `installer/SourceSight.iss`.

## Releasing

```
git tag vX.Y.Z
git push origin vX.Y.Z
```

GitHub Actions builds the installer and the standalone bot and publishes the release. Installed copies see it the next time they open.

## License

[MIT with the Commons Clause](LICENSE): you can use, modify and share SourceSight freely, including on servers that charge for VIP or skins. What you **can't do is sell it** or charge for a product or service based on it.
