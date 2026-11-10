**English** · [Español](como-funciona.md)

# How it works

SourceSight reads the demo with [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang) and walks through the match tick by tick. For every living player it computes position, view angles, pressed buttons and speed, and each module accumulates evidence. At the end the thresholds are applied and the report is built. All thresholds live in `pkg/core/constants.go`.

A 35-minute demo at 128 tick (about 270,000 ticks) takes around 4 seconds.

## Autobunny (`pkg/modules/bhop.go`)

A jump is *perfect* if it leaves the ground within the first tick after landing: `Jump()` then runs before friction and no speed is lost. A person scrolling the mouse wheel hits 20–50%. It flags at 75% or more (minimum 10 jumps) or 5 perfect hops in a row. 55% or more is marked suspicious.

## Triggerbot (`pkg/modules/trigger.go`)

A ray is cast from the player's eyes along their view and intersected with each enemy's collision box. When the ray enters an enemy a timer starts and, on the shot, the reaction time is recorded. People average 150–250 ms with a lot of spread. It flags with at least 4 samples, an average under 80 ms and a deviation under 15 ms.

## Aimlock and spinbot (`pkg/modules/aim.go`)

It measures how much the view angles turn between one tick and the next, correcting the 359° to 0° wrap. A turn of more than 45° in one tick is a *snap*. A human flick can produce an isolated one, so it only flags when there are 5 or more within the same second (and 10 in total). That is what a spinbot or anti-aim looks like: yaw jumps to a different value every tick, usually with pitch pinned at ±89°.

The first 1.5 s after spawning are ignored, as are changes to a pitch of exactly 0, which are angles the server forces on spawn or teleport.

## Silent aim (`pkg/modules/aim.go`)

When a shot deals damage, the shooter's view direction is compared with the actual direction to the victim. Shooting on the move, spread reaches about 35°; beyond that, 3 times, it flags.

## Wallhack (`pkg/modules/wallhack.go`)

The demo records which enemies each player has "spotted" (the same data that feeds the radar). It looks for the crosshair following an enemy that is **not** spotted, within 4° of their chest, for at least one second, travelling 20° or more. The travel is what matters: holding the crosshair still on an angle where someone is about to peek is good crosshair placement, not a wallhack. It flags after 3 such tracks.

## Prefire (`pkg/modules/peek.go`)

For every fight it measures how long the player takes to deal damage after the enemy becomes visible to them, using the same "spotted" data that feeds the radar. A person needs about 200–400 ms to react; damaging in under 100 ms again and again suggests they already knew where the enemy would appear (wallhack) or a triggerbot. It is flagged as **suspicious** with 35% or more of fights under 100 ms, a median under 150 ms and at least 15 fights.

It was calibrated on 13 real demos: it only flagged a player reported as a cheater. The player page compares their average prefire across all demos with everyone else's, because the pattern shows best over several matches.

## Macro (`pkg/modules/macro.go`)

For semi-automatic pistols it measures the interval between shots. A finger has natural variation; a macro fires at an almost constant rhythm. It flags when the deviation between shots is under 1.5 ms, measured over at least 10.

## Blind and smoke kills (`pkg/modules/flash_smoke.go`)

- **Blind:** kills while the shooter is fully flashed. Flags at 4.
- **Smoke:** kills whose path crosses an active smoke grenade (a 144-unit sphere) with few bullets fired. Flags at 5.

## Admins and chat (`pkg/modules/staff.go`, `pkg/modules/chat.go`)

It recognizes punishment commands typed in chat (`!slay`, `/kick`, `sm_ban`, …) and the server's `[SM] Admin: Slayed X.` notices. If the admin punishes whoever killed them within the previous 5 seconds, it's flagged as revenge.

Chat is scanned for severe insults and repeated binds.

## What it doesn't detect (yet)

- **Discord ghosting:** a dead teammate calling positions over voice is legal, so shooting where the killer was proves nothing. The code is in `CheckGhosting` but isn't wired in.
- Real visibility against map geometry: the wallhack check uses the game's "spotted" data, which is an approximation.

## Going to an incident

Every incident comes with `demo_gototick <tick>`, with a few seconds of margin before the play. Open the demo with `playdemo` and paste the command into the console, or use **Watch** in the app to open the game right at that tick.
