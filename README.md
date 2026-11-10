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

A tool for CS:GO server staff. Open a demo (`.dem`) and it flags players whose behavior doesn't look human: autobunny, triggerbot, spinbot, wallhack, macros. Every incident comes with its exact tick so you can go and watch it in-game.

Runs on Windows, fully local. Nothing is uploaded anywhere.

## What it detects

| Detection | What it looks at | When it flags |
|---|---|---|
| Autobunny | Jumps on the first tick of touching the ground | 75% perfect hops or more over 10+ jumps. With 30+ jumps, 55% or 5 in a row is suspicious |
| Triggerbot | Time between an enemy entering the crosshair and the shot | Average under 80 ms with a deviation under 15 ms |
| Aimlock / spinbot | Turns of more than 45° in a single tick | Several within one second: 50+ per match is a cheat, 10–49 is suspicious |
| Silent aim | Gap between where the player looks and where the hit lands | 3 hits more than 35° off |
| Wallhack | Tracking an enemy the player cannot see | 1 s of tracking with 20° of travel, 3 times |
| Prefire | Damage dealt right after an enemy becomes visible to the player | 35% or more of their fights under 100 ms, median under 150 ms, at least 15 fights (suspicious) |
| Macro | Pistol fire rhythm that is too regular | Deviation under 1.5 ms over 10 shots |
| Blind / smoke kills | Kills while fully flashed or through smoke without prior sight | 4 blind or 5 through smoke |
| Admins | `!slay`, `sm_kick`, etc., and `[SM]` server notices | Punishing whoever killed the admin less than 5 s earlier is flagged as revenge |
| Chat | Severe insults, spam binds | — |

Results are **leads, not proof**. Each incident includes its `demo_gototick` command so you can confirm it by watching the demo before banning anyone.

Every player gets a level:

| Level | Meaning |
|---|---|
| **Cheat detected** | Crossed a strong threshold: autobunny, triggerbot, aimlock/spinbot, silent aim, wallhack or macro |
| **Suspicious** | Weaker signs worth a look: high bhop or perfect-hop streaks, a few aim snaps, prefire, blind or through-smoke kills |
| **Conduct** | Not cheating, but relevant to staff: AFK, team-flashing |
| **Clean** | Nothing found |
| **Confirmed** | Set by a staff member after watching the demo, never by the app. Saved per SteamID |

## Install

1. Download **`SourceSight-Setup-vX.Y.Z.exe`** from the [latest release](https://github.com/Hdauu/SourceSight/releases/latest).
2. Run it. It installs for your user only, no administrator rights needed.
3. Windows may show *"Windows protected your PC"* because the installer isn't code-signed. Click **More info → Run anyway**.

Only need the Discord bot? The release also has `SourceSightBot.exe` on its own, no install required.

Requires the WebView2 runtime, which ships with Windows 10 and 11.

**Updates:** on launch the app checks GitHub for a new version and offers to install it with one click. It downloads the installer, checks it against the SHA-256 published by GitHub, updates the app, CLI and bot, and reopens SourceSight.

## Usage

- Drop one or more demos (or a whole folder) onto the window, or use **Open demos**. The installer also adds *Open with → SourceSight* for `.dem` files.
- Every analyzed demo stays in the **history**; reopening it is instant. Filter it by level (cheat, suspicious, clean) or by detection.
- If a demo is the same file or the same match as one already in the history (renamed, copied or recorded twice), the app asks whether to replace it, keep both or discard it, so no match is counted twice.
- Click a player to see their metrics and incidents per round, and mark them as a confirmed cheater after reviewing the demo.
- The **Players** tab averages every player across all your demos and compares them with everyone else. Search by name, SteamID64, profile link, `STEAM_0:X:Y` or `[U:1:N]`.
- **Export evidence** creates a printable HTML file with the incidents and the demo's SHA-256, to hand to another admin.
- **Watch** opens CS:GO at the incident's tick (set the game folder in Settings).

Data lives in `%APPDATA%\SourceSight`, exports in `Documents\SourceSight`, and a diagnostics log in `%APPDATA%\SourceSight\sourcesight.log`.

## Discord bot

1. Create an application and a bot in the [Discord developer portal](https://discord.com/developers/applications) and enable *Message Content Intent*.
2. Run `SourceSightBot.exe` once: it creates `bot_config.json` next to the executable. Paste the token there. **Never share or upload that file.**
3. Run it again. Upload a demo to the channel and it replies with the report and one button per flagged player.

## Building

Requires Go 1.27 or newer.

```
go build -ldflags "-H windowsgui" -o SourceSight.exe ./cmd/sourcesight
go build -o SourceSightCLI.exe ./cmd/sourcesight
go build -o SourceSightBot.exe ./cmd/bot
go test ./...
```

Without `-X sourcesight/pkg/version.Version=vX.Y.Z` in `-ldflags` the version is `dev` and update checks are off.

`third_party/go-webview2` is a copy of [jchv/go-webview2](https://github.com/jchv/go-webview2) (MIT) with a small patch to receive the real path of files dropped onto the window. See `third_party/go-webview2/SOURCESIGHT_PATCH.md`.

The icon is generated with `go run ./tools/icongen`. Release builds embed it with [go-winres](https://github.com/tc-hib/go-winres); the installer is built with [Inno Setup](https://jrsoftware.org/isinfo.php) from `installer/SourceSight.iss`.

## Releasing

```
git tag v2.1.0
git push origin v2.1.0
```

GitHub Actions builds the installer and the standalone bot and publishes the release. Installed copies see it next time they open.

## How it works

See [docs/how-it-works.md](docs/how-it-works.md).

## License

[MIT with the Commons Clause](LICENSE): you can use, modify and share SourceSight freely, including on servers that charge for VIP or skins. You **can't sell it** or charge for a product or service based on it.
