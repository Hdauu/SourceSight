package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"sourcesight/pkg/parser"
	"sourcesight/pkg/report"
	"sourcesight/pkg/version"
)

type BotConfig struct {
	Token      string   `json:"token"`
	Prefix     string   `json:"prefix"`
	ChannelIDs []string `json:"channelIds"`
	MaxDemoMB  int64    `json:"maxDemoMB"`
}

func loadConfig(path string) (*BotConfig, error) {
	cfg := &BotConfig{
		Token:      "PEGA_AQUI_TU_TOKEN_DE_DISCORD",
		Prefix:     "!sourcesight",
		ChannelIDs: []string{},
		MaxDemoMB:  150,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		out, _ := json.MarshalIndent(cfg, "", "    ")
		_ = os.WriteFile(path, out, 0644)
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	cfgPath := "bot_config.json"
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		fmt.Printf("[-] Error cargando configuración: %v\n", err)
		return
	}

	if cfg.Token == "" || cfg.Token == "PEGA_AQUI_TU_TOKEN_DE_DISCORD" {
		fmt.Println("═══════════════════════════════════════════════════")
		fmt.Println(" [!] DISCORD BOT: Token no configurado aún.")
		fmt.Println("     Abre 'bot_config.json' y coloca tu Token.")
		fmt.Println("═══════════════════════════════════════════════════")
		return
	}

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		fmt.Printf("[-] Error creando sesión de Discord: %v\n", err)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("SourceSight %s conectado como %s. Esperando demos.\n", version.Version, s.State.User.String())
		_ = s.UpdateGameStatus(0, "demos de CS:GO | "+cfg.Prefix)
	})

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot {
			return
		}

		if len(cfg.ChannelIDs) > 0 {
			allowed := false
			for _, id := range cfg.ChannelIDs {
				if id == m.ChannelID {
					allowed = true
					break
				}
			}
			if !allowed {
				return
			}
		}

		content := strings.TrimSpace(m.Content)
		if content == cfg.Prefix || content == cfg.Prefix+" help" {
			embed := &discordgo.MessageEmbed{
				Title:       "SourceSight",
				Description: "Analiza demos de CS:GO y marca jugadores sospechosos: autobunny, triggerbot, aimlock y spinbot, silent aim, wallhack, macros, kills cegado o a través de humo, abuso de admins y AFK.",
				Color:       0xe07b2a,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Cómo se usa",
						Value: "Subí un archivo `.dem` a este canal. En unos segundos responde con el reporte y un botón por cada jugador marcado.",
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "SourceSight " + version.Version + " · " + version.RepoURL,
				},
			}
			_, _ = s.ChannelMessageSendEmbed(m.ChannelID, embed)
			return
		}

		var demAtt *discordgo.MessageAttachment
		for _, att := range m.Attachments {
			if strings.HasSuffix(strings.ToLower(att.Filename), ".dem") {
				demAtt = att
				break
			}
		}
		if demAtt == nil {
			return
		}

		sizeMB := float64(demAtt.Size) / (1024.0 * 1024.0)
		if cfg.MaxDemoMB > 0 && int64(sizeMB) > cfg.MaxDemoMB {
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("La demo pesa %.1f MB y el límite es %d MB.", sizeMB, cfg.MaxDemoMB))
			return
		}

		statusMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Analizando `%s` (%.1f MB)...", demAtt.Filename, sizeMB))

		// Download demo
		resp, err := http.Get(demAtt.URL)
		if err != nil {
			_, _ = s.ChannelMessageEdit(m.ChannelID, statusMsg.ID, fmt.Sprintf("No se pudo descargar la demo: %v", err))
			return
		}
		defer resp.Body.Close()

		tempDir := "temp"
		_ = os.MkdirAll(tempDir, 0755)
		tempDem := filepath.Join(tempDir, fmt.Sprintf("demo_%d.dem", time.Now().UnixNano()))
		outDem, err := os.Create(tempDem)
		if err != nil {
			_, _ = s.ChannelMessageEdit(m.ChannelID, statusMsg.ID, fmt.Sprintf("No se pudo guardar la demo: %v", err))
			return
		}
		_, _ = io.Copy(outDem, resp.Body)
		outDem.Close()
		defer os.Remove(tempDem)

		// Parse
		start := time.Now()
		rep, err := parser.ParseDemo(tempDem, parser.ParserOptions{Verbose: false})
		if err != nil {
			_, _ = s.ChannelMessageEdit(m.ChannelID, statusMsg.ID, fmt.Sprintf("No se pudo analizar la demo: %v", err))
			return
		}
		elapsed := time.Since(start)

		// Render PNG
		tempPng := filepath.Join(tempDir, fmt.Sprintf("report_%d.png", time.Now().UnixNano()))
		if err := report.RenderReportPNG(rep, tempPng); err != nil {
			_, _ = s.ChannelMessageEdit(m.ChannelID, statusMsg.ID, fmt.Sprintf("No se pudo generar la imagen: %v", err))
			return
		}
		defer os.Remove(tempPng)

		// Build embed
		var flaggedNames []string
		for _, p := range rep.Players {
			if len(p.Flags) > 0 {
				flaggedNames = append(flaggedNames, fmt.Sprintf("• **%s** — `%s`", p.Name, strings.Join(p.Flags, " + ")))
			}
		}

		color := 0x57ab5a // green
		title := "Sin detecciones"
		if len(flaggedNames) > 0 {
			color = 0xe5534b // red
			title = fmt.Sprintf("%d jugador(es) marcado(s)", len(flaggedNames))
		}

		topInfo := "Sin jugadores suficientes."
		if len(rep.Players) > 0 {
			tp := rep.Players[0]
			topInfo = fmt.Sprintf("**%s**\n• Saltos: `%d` | Perfs: `%.1f%%`\n• Veredicto: `%s`",
				tp.Name, tp.TotalJumps, tp.PerfRatio, tp.Verdict)
		}

		flagsStr := "Ninguna detección."
		if len(flaggedNames) > 0 {
			flagsStr = strings.Join(flaggedNames, "\n")
		}

		var tlLines []string
		maxTl := len(rep.Timeline)
		if maxTl > 5 {
			maxTl = 5
		}
		for _, inc := range rep.Timeline[:maxTl] {
			tlLines = append(tlLines, fmt.Sprintf("`%s` — %s [%s]", inc.GotoCmd, inc.Player, inc.Type))
		}
		tlStr := "—"
		if len(tlLines) > 0 {
			tlStr = strings.Join(tlLines, "\n")
		}

		embed := &discordgo.MessageEmbed{
			Title: title,
			Color: color,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Mapa", Value: fmt.Sprintf("`%s`", rep.Header.MapName), Inline: true},
				{Name: "Tickrate", Value: fmt.Sprintf("`%d`", rep.Header.Tickrate), Inline: true},
				{Name: "Ticks", Value: fmt.Sprintf("`%d`", rep.Header.TotalTicks), Inline: true},
				{Name: "Top Jugador", Value: topInfo, Inline: false},
				{Name: "Flags Detectadas", Value: flagsStr, Inline: false},
				{Name: "Timeline (pegar en CS:GO)", Value: tlStr, Inline: false},
			},
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://reporte.png",
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("SourceSight %s · %.1f s", version.Version, elapsed.Seconds()),
			},
		}

		pngFile, err := os.Open(tempPng)
		if err != nil {
			_, _ = s.ChannelMessageEdit(m.ChannelID, statusMsg.ID, fmt.Sprintf("No se pudo adjuntar el reporte: %v", err))
			return
		}
		defer pngFile.Close()

		msgSend := &discordgo.MessageSend{
			Content:    fmt.Sprintf("`%s`", demAtt.Filename),
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: playerButtons(storeReport(rep), rep),
			Files: []*discordgo.File{
				{
					Name:   "reporte.png",
					Reader: pngFile,
				},
			},
		}

		_, _ = s.ChannelMessageSendComplex(m.ChannelID, msgSend)
		_ = s.ChannelMessageDelete(m.ChannelID, statusMsg.ID)
	})

	dg.AddHandler(handlePlayerButton)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	if err := dg.Open(); err != nil {
		fmt.Printf("[-] Error conectando a Discord: %v\n", err)
		return
	}
	defer dg.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	fmt.Println("\n[+] Cerrando bot...")
}
