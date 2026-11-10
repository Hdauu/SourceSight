package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"sourcesight/pkg/core"
	"sourcesight/pkg/evidence"
	"sourcesight/pkg/version"
)

// Reports stay in memory so the per-player buttons keep working for a while after
// the analysis; after reportTTL a click answers that the report expired.
const reportTTL = 6 * time.Hour

const playerButtonPrefix = "ss:p:"

type storedReport struct {
	rep     *core.DemoReport
	created time.Time
}

var (
	reportsMu sync.Mutex
	reports   = map[string]storedReport{}
)

func storeReport(rep *core.DemoReport) string {
	reportsMu.Lock()
	defer reportsMu.Unlock()
	for k, v := range reports {
		if time.Since(v.created) > reportTTL {
			delete(reports, k)
		}
	}
	key := strconv.FormatInt(time.Now().UnixNano(), 36)
	reports[key] = storedReport{rep: rep, created: time.Now()}
	return key
}

func loadReport(key string) *core.DemoReport {
	reportsMu.Lock()
	defer reportsMu.Unlock()
	if v, ok := reports[key]; ok && time.Since(v.created) <= reportTTL {
		return v.rep
	}
	return nil
}

// playerButtons builds one button per flagged player (Discord allows 5 rows of 5).
func playerButtons(key string, rep *core.DemoReport) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent
	var row []discordgo.MessageComponent
	count := 0
	for _, p := range rep.Players {
		if len(p.Flags) == 0 || p.SteamID == 0 {
			continue
		}
		if count == 25 {
			break
		}
		label := p.Name
		if len([]rune(label)) > 70 {
			label = string([]rune(label)[:70])
		}
		row = append(row, discordgo.Button{
			Label:    label,
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("%s%s:%d", playerButtonPrefix, key, p.SteamID),
		})
		count++
		if len(row) == 5 {
			rows = append(rows, discordgo.ActionsRow{Components: row})
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, discordgo.ActionsRow{Components: row})
	}
	return rows
}

// handlePlayerButton answers a player button with a detail embed only the clicker sees.
func handlePlayerButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	id := i.MessageComponentData().CustomID
	if !strings.HasPrefix(id, playerButtonPrefix) {
		return
	}
	parts := strings.SplitN(strings.TrimPrefix(id, playerButtonPrefix), ":", 2)

	reply := func(data *discordgo.InteractionResponseData) {
		data.Flags = discordgo.MessageFlagsEphemeral
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: data,
		})
	}

	if len(parts) != 2 {
		return
	}
	rep := loadReport(parts[0])
	if rep == nil {
		reply(&discordgo.InteractionResponseData{Content: "Este reporte ya expiró. Volvé a subir la demo para verlo."})
		return
	}
	steamID, _ := strconv.ParseUint(parts[1], 10, 64)
	for _, p := range rep.Players {
		if p.SteamID == steamID {
			reply(&discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{playerEmbed(rep, p)}})
			return
		}
	}
	reply(&discordgo.InteractionResponseData{Content: "No encontré a ese jugador en el reporte."})
}

func playerEmbed(rep *core.DemoReport, p core.PlayerStats) *discordgo.MessageEmbed {
	reaction := "–"
	if p.TriggerAvgMs != nil {
		reaction = fmt.Sprintf("%d ms", *p.TriggerAvgMs)
	}
	metrics := fmt.Sprintf(
		"Perfs `%.1f%%` (%d/%d) · racha `%d`\nSnaps `%d` · silent `%d` · wallhack `%d`\nReacción `%s` · macro `%d` · cegado/humo `%d/%d`",
		p.PerfRatio, p.PerfectHops, p.TotalJumps, p.MaxStreak,
		p.SnapCount, p.SilentCount, p.WallTrackCount,
		reaction, p.MacroHits, p.FlashKills, p.SmokeKills,
	)

	var lines []string
	for idx, inc := range p.Incidents {
		if idx == 10 {
			lines = append(lines, fmt.Sprintf("y %d más", len(p.Incidents)-10))
			break
		}
		lines = append(lines, fmt.Sprintf("R%d %s · `%s`", evidence.RoundAt(rep.Header.RoundStarts, inc.Tick), inc.Type, inc.GotoCmd))
	}
	incStr := "Sin incidentes."
	if len(lines) > 0 {
		incStr = strings.Join(lines, "\n")
	}

	fields := []*discordgo.MessageEmbedField{
		{Name: "Detecciones", Value: "`" + strings.Join(p.Flags, " + ") + "`"},
		{Name: "Métricas", Value: metrics},
		{Name: fmt.Sprintf("Incidentes (%d)", len(p.Incidents)), Value: incStr},
	}
	if p.PreviousInfractions > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "Antecedentes",
			Value: fmt.Sprintf("%d detección(es) en demos anteriores", p.PreviousInfractions),
		})
	}

	return &discordgo.MessageEmbed{
		Title:  p.Name,
		URL:    fmt.Sprintf("https://steamcommunity.com/profiles/%d", p.SteamID),
		Color:  0xe5534b,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("SteamID64 %d · SourceSight %s", p.SteamID, version.Version)},
	}
}
