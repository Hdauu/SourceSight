package report

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"unicode"

	"github.com/fogleman/gg"
	"sourcesight/pkg/core"
	"sourcesight/pkg/version"
)

var (
	colorBg      = color.RGBA{14, 14, 14, 255}    // #0e0e0e
	colorHeader  = color.RGBA{10, 10, 10, 255}    // #0a0a0a
	colorRow1    = color.RGBA{26, 26, 26, 255}    // #1a1a1a
	colorRow2    = color.RGBA{17, 17, 17, 255}    // #111111
	colorSecBar  = color.RGBA{22, 22, 22, 255}    // #161616
	colorBorder  = color.RGBA{38, 38, 38, 255}    // #262626
	colorText    = color.RGBA{243, 244, 246, 255} // #f3f4f6
	colorDim     = color.RGBA{156, 163, 175, 255} // #9ca3af
	colorMid     = color.RGBA{209, 213, 219, 255} // #d1d5db
	colorColHead = color.RGBA{203, 213, 225, 255} // #cbd5e1
	colorRed     = color.RGBA{255, 107, 107, 255} // #ff6b6b
	colorAmber   = color.RGBA{255, 193, 69, 255}  // #ffc145
	colorGreen   = color.RGBA{74, 222, 128, 255}  // #4ade80
	colorBlue    = color.RGBA{56, 189, 248, 255}  // #38bdf8
	colorPurple  = color.RGBA{99, 102, 241, 255}  // #6366f1
)

func loadFont(dc *gg.Context, points float64) {
	candidates := []string{
		`C:\Windows\Fonts\segoeui.ttf`,
		`C:\Windows\Fonts\arial.ttf`,
		`/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf`,
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := dc.LoadFontFace(p, points); err == nil {
				return
			}
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + "…"
	}
	return s
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 128 {
			if r >= 32 {
				b.WriteRune(r)
			}
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsPunct(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func RenderReportPNG(rep *core.DemoReport, outputPath string) error {
	const (
		W    = 960.0
		M    = 32.0
		IW   = W - M*2.0
		rowH = 32.0
	)

	// Filter sections
	var bhopPlayers []core.PlayerStats
	for _, p := range rep.Players {
		if p.TotalJumps >= 3 {
			bhopPlayers = append(bhopPlayers, p)
		}
	}

	var flagPlayers []core.PlayerStats
	for _, p := range rep.Players {
		hasMultiModule := false
		for _, f := range p.Flags {
			if f != "AUTOBUNNY" && f != "SOSPECHOSO-BHOP" {
				hasMultiModule = true
				break
			}
		}
		if hasMultiModule {
			flagPlayers = append(flagPlayers, p)
		}
	}

	topTimeline := rep.Timeline
	if len(topTimeline) > 10 {
		topTimeline = topTimeline[:10]
	}

	topAdmin := rep.AdminLog
	if len(topAdmin) > 5 {
		topAdmin = topAdmin[:5]
	}

	// Calculate layout height
	cursor := M
	headerH := 84.0
	cursor += headerH + 16.0

	// Bhop section
	bhopSecH := 26.0 + 26.0 + float64(len(bhopPlayers))*rowH + 16.0
	bhopSecY := cursor
	cursor += bhopSecH

	// Multi-module flags section
	var flagSecY, flagSecH float64
	if len(flagPlayers) > 0 {
		flagSecH = 26.0 + 26.0 + float64(len(flagPlayers))*rowH + 16.0
		flagSecY = cursor
		cursor += flagSecH
	}

	// Timeline section
	var tlSecY, tlSecH float64
	if len(topTimeline) > 0 {
		tlSecH = 26.0 + 26.0 + float64(len(topTimeline))*rowH + 16.0
		tlSecY = cursor
		cursor += tlSecH
	}

	// Admin log section
	var admSecY, admSecH float64
	if len(topAdmin) > 0 {
		admSecH = 26.0 + float64(len(topAdmin))*rowH + 16.0
		admSecY = cursor
		cursor += admSecH
	}

	footerY := cursor
	totalH := footerY + 36.0

	// Start drawing
	dc := gg.NewContext(int(W), int(totalH))
	dc.SetColor(colorBg)
	dc.Clear()

	// ── HEADER CARD ──
	dc.SetColor(colorHeader)
	dc.DrawRoundedRectangle(M, M, IW, headerH-4.0, 6.0)
	dc.Fill()

	// Accent left bar
	dc.SetColor(colorPurple)
	dc.DrawRoundedRectangle(M, M, 4.0, headerH-4.0, 2.0)
	dc.Fill()

	// Brand title
	loadFont(dc, 20.0)
	dc.SetColor(color.RGBA{255, 255, 255, 255})
	dc.DrawString("SourceSight", M+20.0, M+28.0)

	loadFont(dc, 14.0)
	dc.SetColor(colorPurple)
	dc.DrawString("— CS:GO / CS2 Demo Auditor (Go Engine)", M+144.0, M+28.0)

	// Subtitles
	loadFont(dc, 12.0)
	dc.SetColor(colorDim)
	metaStr := fmt.Sprintf("%s  ·  %d-tick  ·  %s ticks  ·  %s",
		rep.Header.MapName, rep.Header.Tickrate,
		formatNumber(rep.Header.TotalTicks), rep.Header.Date)
	dc.DrawString(metaStr, M+20.0, M+50.0)

	loadFont(dc, 11.0)
	dc.SetColor(colorDim)
	dc.DrawString(sanitize(truncate(rep.Header.ServerName, 75)), M+20.0, M+68.0)

	// Verdict Pill
	totalInfractions := 0
	for _, p := range rep.Players {
		if len(p.Flags) > 0 {
			totalInfractions++
		}
	}
	hasInfractions := totalInfractions > 0
	pillColor := colorGreen
	pillText := "PARTIDA LIMPIA"
	if hasInfractions {
		pillColor = colorRed
		pillText = fmt.Sprintf("%d INFRACCIÓN(ES)", totalInfractions)
	}
	pillW := 150.0
	pillH := 24.0
	pillX := M + IW - pillW - 10.0
	pillY := M + 14.0

	dc.SetColor(color.RGBA{pillColor.R, pillColor.G, pillColor.B, 30})
	dc.DrawRoundedRectangle(pillX, pillY, pillW, pillH, 4.0)
	dc.Fill()
	dc.SetColor(pillColor)
	dc.SetLineWidth(1.0)
	dc.DrawRoundedRectangle(pillX, pillY, pillW, pillH, 4.0)
	dc.Stroke()

	loadFont(dc, 11.0)
	dc.SetColor(pillColor)
	dc.DrawStringAnchored(pillText, pillX+pillW/2.0, pillY+pillH/2.0, 0.5, 0.5)

	// Helper for section title bar
	drawSectionBar := func(y float64, title string) {
		dc.SetColor(colorSecBar)
		dc.DrawRoundedRectangle(M, y, IW, 26.0, 3.0)
		dc.Fill()

		loadFont(dc, 11.0)
		dc.SetColor(colorColHead)
		dc.DrawString(title, M+10.0, y+18.0)

		dc.SetColor(colorBorder)
		dc.SetLineWidth(1.0)
		dc.DrawLine(M, y+26.0, M+IW, y+26.0)
		dc.Stroke()
	}

	// ── SECTION 1: BHOP ──
	drawSectionBar(bhopSecY, "BHOP PHYSICS VALIDATOR — Módulo A")
	bColY := bhopSecY + 26.0
	dc.SetColor(color.RGBA{20, 20, 20, 255})
	dc.DrawRectangle(M, bColY, IW, 26.0)
	dc.Fill()

	loadFont(dc, 11.0)
	dc.SetColor(colorColHead)
	dc.DrawString("Jugador", M+10.0, bColY+18.0)
	dc.DrawStringAnchored("Saltos", M+240.0, bColY+18.0, 1.0, 0.0)
	dc.DrawStringAnchored("Perfs %", M+340.0, bColY+18.0, 1.0, 0.0)
	dc.DrawStringAnchored("1-Tick %", M+450.0, bColY+18.0, 1.0, 0.0)
	dc.DrawStringAnchored("Racha", M+540.0, bColY+18.0, 1.0, 0.0)
	dc.DrawStringAnchored("Vel.Máx", M+630.0, bColY+18.0, 1.0, 0.0)
	dc.DrawStringAnchored("Estado", M+IW-10.0, bColY+18.0, 1.0, 0.0)

	for i, p := range bhopPlayers {
		y := bColY + 26.0 + float64(i)*rowH
		bg := colorRow1
		if i%2 != 0 {
			bg = colorRow2
		}
		dc.SetColor(bg)
		dc.DrawRectangle(M, y, IW, rowH)
		dc.Fill()

		// Row border
		dc.SetColor(colorBorder)
		dc.SetLineWidth(1.0)
		dc.DrawLine(M, y+rowH, M+IW, y+rowH)
		dc.Stroke()

		// Text
		loadFont(dc, 12.0)
		nameCol := colorText
		if strings.Contains(p.BhopVerdict, "AUTOBUNNY") {
			nameCol = colorRed
		} else if strings.Contains(p.BhopVerdict, "Sospechoso") {
			nameCol = colorAmber
		}
		dc.SetColor(nameCol)
		dc.DrawString(sanitize(truncate(p.Name, 24)), M+10.0, y+21.0)

		dc.SetColor(colorMid)
		dc.DrawStringAnchored(fmt.Sprintf("%d", p.TotalJumps), M+240.0, y+21.0, 1.0, 0.0)

		perfCol := colorMid
		if p.PerfRatio >= 75.0 {
			perfCol = colorRed
		} else if p.PerfRatio >= 55.0 {
			perfCol = colorAmber
		}
		dc.SetColor(perfCol)
		dc.DrawStringAnchored(fmt.Sprintf("%.1f%% (%d)", p.PerfRatio, p.PerfectHops), M+340.0, y+21.0, 1.0, 0.0)

		dc.SetColor(colorMid)
		dc.DrawStringAnchored(fmt.Sprintf("%.1f%%", p.OneTickRatio), M+450.0, y+21.0, 1.0, 0.0)

		streakCol := colorMid
		if p.MaxStreak >= 5 {
			streakCol = colorRed
		}
		dc.SetColor(streakCol)
		dc.DrawStringAnchored(fmt.Sprintf("%d", p.MaxStreak), M+540.0, y+21.0, 1.0, 0.0)

		dc.SetColor(colorMid)
		dc.DrawStringAnchored(fmt.Sprintf("%du/s", p.MaxSpeed), M+630.0, y+21.0, 1.0, 0.0)

		statusCol := colorGreen
		if strings.Contains(p.BhopVerdict, "AUTOBUNNY") {
			statusCol = colorRed
		} else if strings.Contains(p.BhopVerdict, "Sospechoso") {
			statusCol = colorAmber
		}
		dc.SetColor(statusCol)
		dc.DrawStringAnchored(p.BhopVerdict, M+IW-10.0, y+21.0, 1.0, 0.0)
	}

	// ── SECTION 2: MULTI-MODULE FLAGS ──
	if len(flagPlayers) > 0 {
		drawSectionBar(flagSecY, "DETECCIONES MULTI-MÓDULO — Triggerbot / Aimlock / Macro / Flash")
		fColY := flagSecY + 26.0
		dc.SetColor(color.RGBA{20, 20, 20, 255})
		dc.DrawRectangle(M, fColY, IW, 26.0)
		dc.Fill()

		loadFont(dc, 11.0)
		dc.SetColor(colorColHead)
		dc.DrawString("Jugador", M+10.0, fColY+18.0)
		dc.DrawStringAnchored("Triggerbot", M+250.0, fColY+18.0, 1.0, 0.0)
		dc.DrawStringAnchored("Aim / Silent", M+400.0, fColY+18.0, 1.0, 0.0)
		dc.DrawStringAnchored("Macro", M+540.0, fColY+18.0, 1.0, 0.0)
		dc.DrawStringAnchored("Blind / Smoke", M+660.0, fColY+18.0, 1.0, 0.0)
		dc.DrawStringAnchored("Veredicto", M+IW-10.0, fColY+18.0, 1.0, 0.0)

		for i, p := range flagPlayers {
			y := fColY + 26.0 + float64(i)*rowH
			bg := colorRow1
			if i%2 != 0 {
				bg = colorRow2
			}
			dc.SetColor(bg)
			dc.DrawRectangle(M, y, IW, rowH)
			dc.Fill()

			dc.SetColor(colorBorder)
			dc.SetLineWidth(1.0)
			dc.DrawLine(M, y+rowH, M+IW, y+rowH)
			dc.Stroke()

			loadFont(dc, 12.0)
			dc.SetColor(colorRed)
			dc.DrawString(sanitize(truncate(p.Name, 24)), M+10.0, y+21.0)

			// Trigger
			trigStr := "—"
			trigCol := colorMid
			if p.TriggerAvgMs != nil {
				trigStr = fmt.Sprintf("%dms (σ%.1f)", *p.TriggerAvgMs, *p.TriggerSigma)
				if *p.TriggerAvgMs < 80 {
					trigCol = colorRed
				}
			}
			dc.SetColor(trigCol)
			dc.DrawStringAnchored(trigStr, M+250.0, y+21.0, 1.0, 0.0)

			// Aim
			aimStr := fmt.Sprintf("%d snaps / %d silent", p.SnapCount, p.SilentCount)
			aimCol := colorMid
			if p.SnapCount >= 10 || p.SilentCount >= 1 {
				aimCol = colorRed
			}
			dc.SetColor(aimCol)
			dc.DrawStringAnchored(aimStr, M+400.0, y+21.0, 1.0, 0.0)

			// Macro
			macStr := "—"
			macCol := colorMid
			if p.MacroHits > 0 {
				macStr = fmt.Sprintf("%d series", p.MacroHits)
				macCol = colorAmber
			}
			dc.SetColor(macCol)
			dc.DrawStringAnchored(macStr, M+540.0, y+21.0, 1.0, 0.0)

			// Flash / Smoke
			fsStr := fmt.Sprintf("%dB / %dS", p.FlashKills, p.SmokeKills)
			fsCol := colorMid
			if p.FlashKills > 0 || p.SmokeKills > 0 {
				fsCol = colorAmber
			}
			dc.SetColor(fsCol)
			dc.DrawStringAnchored(fsStr, M+660.0, y+21.0, 1.0, 0.0)

			// Verdict
			dc.SetColor(colorRed)
			dc.DrawStringAnchored(sanitize(truncate(p.Verdict, 26)), M+IW-10.0, y+21.0, 1.0, 0.0)
		}
	}

	// ── SECTION 3: OVERWATCH TIMELINE ──
	if len(topTimeline) > 0 {
		drawSectionBar(tlSecY, "OVERWATCH TIMELINE — Incidentes ordenados por tick (pre-rollback 3s)")
		tColY := tlSecY + 26.0
		dc.SetColor(color.RGBA{20, 20, 20, 255})
		dc.DrawRectangle(M, tColY, IW, 26.0)
		dc.Fill()

		loadFont(dc, 11.0)
		dc.SetColor(colorColHead)
		dc.DrawString("Tipo", M+10.0, tColY+18.0)
		dc.DrawString("Jugador", M+150.0, tColY+18.0)
		dc.DrawString("Comando Consola", M+320.0, tColY+18.0)
		dc.DrawStringAnchored("Evidencia / Nota", M+IW-10.0, tColY+18.0, 1.0, 0.0)

		for i, inc := range topTimeline {
			y := tColY + 26.0 + float64(i)*rowH
			bg := colorRow1
			if i%2 != 0 {
				bg = colorRow2
			}
			dc.SetColor(bg)
			dc.DrawRectangle(M, y, IW, rowH)
			dc.Fill()

			dc.SetColor(colorBorder)
			dc.SetLineWidth(1.0)
			dc.DrawLine(M, y+rowH, M+IW, y+rowH)
			dc.Stroke()

			typeCol := colorBlue
			if strings.Contains(inc.Type, "TRIGGER") || strings.Contains(inc.Type, "SILENT") {
				typeCol = colorRed
			} else if strings.Contains(inc.Type, "AIM") || strings.Contains(inc.Type, "MACRO") {
				typeCol = colorAmber
			}

			loadFont(dc, 11.0)
			dc.SetColor(typeCol)
			dc.DrawString(inc.Type, M+10.0, y+21.0)

			dc.SetColor(colorMid)
			dc.DrawString(sanitize(truncate(inc.Player, 20)), M+150.0, y+21.0)

			dc.SetColor(colorGreen)
			dc.DrawString(inc.GotoCmd, M+320.0, y+21.0)

			dc.SetColor(colorDim)
			dc.DrawStringAnchored(sanitize(truncate(inc.Note, 52)), M+IW-10.0, y+21.0, 1.0, 0.0)
		}
	}

	// ── SECTION 4: ADMIN LOG ──
	if len(topAdmin) > 0 {
		drawSectionBar(admSecY, "AUDITORÍA DE STAFF — Admin Abuse Log")
		for i, ev := range topAdmin {
			y := admSecY + 26.0 + float64(i)*rowH
			bg := colorRow1
			if i%2 != 0 {
				bg = colorRow2
			}
			dc.SetColor(bg)
			dc.DrawRectangle(M, y, IW, rowH)
			dc.Fill()

			dc.SetColor(colorBorder)
			dc.SetLineWidth(1.0)
			dc.DrawLine(M, y+rowH, M+IW, y+rowH)
			dc.Stroke()

			adminCol := colorAmber
			if ev.IsRevenge {
				adminCol = colorRed
			}

			loadFont(dc, 12.0)
			dc.SetColor(adminCol)
			dc.DrawString(ev.AdminName, M+10.0, y+21.0)

			dc.SetColor(colorMid)
			dc.DrawString(fmt.Sprintf("%s %s (Ronda %d)", ev.Command, ev.TargetName, ev.Round), M+220.0, y+21.0)

			if ev.IsRevenge {
				dc.SetColor(colorRed)
				dc.DrawStringAnchored("VENGANZA", M+IW-10.0, y+21.0, 1.0, 0.0)
			}
		}
	}

	// ── FOOTER ──
	dc.SetColor(colorBorder)
	dc.SetLineWidth(1.0)
	dc.DrawLine(M, footerY, M+IW, footerY)
	dc.Stroke()

	loadFont(dc, 10.0)
	dc.SetColor(colorDim)
	footerNotes := "Humano bhop: 20–50% perfs · Autobunny: ≥75% perfs + 1-tick · Triggerbot: avg<80ms σ<15ms · Silent Aim: divergencia >35°"
	dc.DrawString(footerNotes, M, footerY+20.0)

	dc.SetColor(color.RGBA{100, 100, 100, 255})
	dc.DrawStringAnchored("SourceSight "+version.Version, M+IW, footerY+20.0, 1.0, 0.0)

	return dc.SavePNG(outputPath)
}

func formatNumber(n int) string {
	in := fmt.Sprintf("%d", n)
	out := ""
	for i, c := range in {
		if i > 0 && (len(in)-i)%3 == 0 {
			out += "."
		}
		out += string(c)
	}
	return out
}
