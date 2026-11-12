package gui

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"sourcesight/pkg/applog"
	"sourcesight/pkg/core"
	"sourcesight/pkg/launcher"
)

// Playlist for the previous / next incident hotkeys. Pressing "Ver" sets it: the
// incidents of that player when the moment belongs to one, otherwise every incident
// of the demo, in tick order. The hotkeys walk it, each step seeking the open game
// and putting the camera on the player in first person.

type playItem struct {
	tick    int
	steamID uint64
	label   string
}

var (
	plMu    sync.Mutex
	plItems []playItem
	plIndex int
	plDemo  string
	plRate  int
)

// watch shows a moment in the game and makes it the current spot of the playlist.
func (a *app) watch(rep *core.DemoReport, demoPath string, tick int, steamIDText string) error {
	steamID, _ := strconv.ParseUint(steamIDText, 10, 64)
	items := playlistFor(rep, steamID)

	plMu.Lock()
	plItems, plDemo, plRate = items, demoPath, rep.Header.Tickrate
	// The current spot is the last incident at or before the moment shown, so "next"
	// goes to the one after it
	plIndex = -1
	for i, it := range items {
		if it.tick <= tick {
			plIndex = i
		}
	}
	note := ""
	if plIndex >= 0 && items[plIndex].tick == tick {
		note = fmt.Sprintf("incidente %d de %d: %s", plIndex+1, len(items), items[plIndex].label)
	} else if name := playerName(rep, steamID); name != "" {
		note = name
	}
	plMu.Unlock()

	return launcher.LaunchInGame(demoPath, launcher.Target{Tick: tick, SteamID: steamID, Note: note}, rep.Header.Tickrate)
}

// step moves through the playlist (hotkeys): -1 previous, +1 next.
func (a *app) step(d int) {
	plMu.Lock()
	if len(plItems) == 0 {
		plMu.Unlock()
		a.toast("Tocá Ver en un incidente primero: las teclas recorren los incidentes desde ahí")
		return
	}
	i := plIndex + d
	if i < 0 || i >= len(plItems) {
		plMu.Unlock()
		a.toast("No hay más incidentes en esa dirección")
		return
	}
	plIndex = i
	it := plItems[i]
	demo, rate, n := plDemo, plRate, len(plItems)
	plMu.Unlock()

	note := fmt.Sprintf("incidente %d de %d: %s", i+1, n, it.label)
	applog.Printf("teclas: %s (tick %d)", note, it.tick)
	if err := launcher.LaunchInGame(demo, launcher.Target{Tick: it.tick, SteamID: it.steamID, Note: note}, rate); err != nil {
		a.toast("Error al abrir CS:GO: " + err.Error())
	}
}

func playlistFor(rep *core.DemoReport, steamID uint64) []playItem {
	var items []playItem
	if steamID != 0 {
		for _, p := range rep.Players {
			if p.SteamID != steamID {
				continue
			}
			for _, inc := range p.Incidents {
				items = append(items, playItem{inc.Tick, steamID, inc.Type + " de " + p.Name})
			}
		}
	}
	if len(items) == 0 {
		for _, inc := range rep.Timeline {
			items = append(items, playItem{inc.Tick, inc.SteamID, inc.Type + " de " + inc.Player})
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].tick < items[j].tick })
	return items
}

func playerName(rep *core.DemoReport, steamID uint64) string {
	for _, p := range rep.Players {
		if p.SteamID == steamID {
			return p.Name
		}
	}
	return ""
}
