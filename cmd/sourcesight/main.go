package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sourcesight/pkg/core"
	"sourcesight/pkg/db"
	"sourcesight/pkg/gui"
	"sourcesight/pkg/parser"
	"sourcesight/pkg/report"
	"sourcesight/pkg/version"
)

func main() {
	execName := strings.ToLower(filepath.Base(os.Args[0]))
	defaultCLI := strings.Contains(execName, "cli")

	cliMode := flag.Bool("cli", defaultCLI, "Ejecutar en modo consola terminal")
	flag.Parse()

	args := flag.Args()
	var demoPath string
	if len(args) > 0 {
		demoPath = args[0]
	}

	if *cliMode {
		runCLI(demoPath)
		return
	}

	// Native Desktop GUI Mode
	if err := gui.LaunchApp(demoPath); err != nil {
		fmt.Printf("[-] Error iniciando interfaz de escritorio: %v\n", err)
		runCLI(demoPath)
	}
}

func runCLI(demoPath string) {
	if demoPath == "" {
		// Scan current working directory
		files, _ := filepath.Glob("*.dem")
		if len(files) > 0 {
			demoPath = files[0]
			fmt.Printf("[+] No se especificó demo. Usando: %s\n", demoPath)
		} else {
			// Scan parent folders (e.g. AnalizarBHOPs)
			parentFiles, _ := filepath.Glob("..\\AnalizarBHOPs\\*.dem")
			if len(parentFiles) > 0 {
				demoPath = parentFiles[0]
				fmt.Printf("[+] No se especificó demo. Usando: %s\n", demoPath)
			}
		}
	}

	if demoPath == "" {
		fmt.Println("[-] Error: No se encontró ningún archivo .dem.")
		fmt.Println("    Uso: Arrastra un archivo .dem arriba de SourceSight.exe -o- ejecuta:")
		fmt.Println("         SourceSight.exe \"ruta\\a\\tu\\demo.dem\"")
		fmt.Println("\nPresiona ENTER para salir...")
		fmt.Scanln()
		os.Exit(1)
	}

	if _, err := os.Stat(demoPath); os.IsNotExist(err) {
		fmt.Printf("[-] Error: El archivo \"%s\" no existe.\n", demoPath)
		fmt.Println("\nPresiona ENTER para salir...")
		fmt.Scanln()
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Println("  SourceSight " + version.Version)
	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Printf("  Demo  : %s\n", filepath.Base(demoPath))
	fmt.Printf("  Ruta  : %s\n", demoPath)
	fmt.Println("──────────────────────────────────────────────────────────")

	start := time.Now()

	rep, err := parser.ParseDemo(demoPath, parser.ParserOptions{Verbose: true})
	if err != nil {
		fmt.Printf("[-] Error analizando la demo: %v\n", err)
		fmt.Println("\nPresiona ENTER para salir...")
		fmt.Scanln()
		os.Exit(1)
	}
	// Count the match for reoffender tracking; the same match is never counted twice
	db.GetDatabase().RecordReport(rep)

	elapsed := time.Since(start)

	// Print Console Table
	fmt.Println()
	fmt.Println("  RESULTADOS BHOP & INPUT ANALYSIS")
	fmt.Println("──────────────────────────────────────────────────────────")
	fmt.Printf(" #  %-22s %7s %9s %8s %7s %7s  Estado\n",
		"Jugador", "Saltos", "Perfs%", "1-Tick", "Racha", "Vel")
	fmt.Println(strings.Repeat("─", 76))

	for idx, p := range rep.Players {
		if p.TotalJumps < 2 && len(p.Flags) == 0 {
			continue
		}

		fmt.Printf("%2d  %-22s %7d %8.1f%% %7.1f%% %7d %6du/s  %s\n",
			idx+1,
			truncate(p.Name, 22),
			p.TotalJumps,
			p.PerfRatio,
			p.OneTickRatio,
			p.MaxStreak,
			p.MaxSpeed,
			p.BhopVerdict,
		)
	}
	fmt.Println(strings.Repeat("─", 76))

	// Multi-module flags summary (excludes pure Bhop players already featured in the primary table)
	var flagged []core.PlayerStats
	for _, p := range rep.Players {
		hasMulti := false
		for _, f := range p.Flags {
			if f != "AUTOBUNNY" && f != "SOSPECHOSO-BHOP" {
				hasMulti = true
				break
			}
		}
		if hasMulti {
			flagged = append(flagged, p)
		}
	}

	if len(flagged) > 0 {
		fmt.Println()
		fmt.Println("  DETECCIONES ADICIONALES")
		fmt.Println("──────────────────────────────────────────────────────────")
		for _, p := range flagged {
			fmt.Printf("  %s:  %s\n", p.Name, strings.Join(p.Flags, " | "))
			if p.TriggerAvgMs != nil {
				fmt.Printf("    → Triggerbot: avg %dms  σ=%.1fms  muestras=%d\n", *p.TriggerAvgMs, *p.TriggerSigma, p.TriggerSamples)
			}
			if p.SnapCount >= 10 {
				fmt.Printf("    → Aimlock: %d angular snaps detectados\n", p.SnapCount)
			}
			if p.SilentCount >= 1 {
				fmt.Printf("    → Silent Aim: %d disparo(s) con divergencia >5°\n", p.SilentCount)
			}
			if p.MacroHits > 0 {
				fmt.Printf("    → Macro: σ plana detectada en %d serie(s) de disparos\n", p.MacroHits)
			}
			if p.FlashKills > 0 {
				fmt.Printf("    → Anti-Flash: %d kill(s) con visión 100%% bloqueada\n", p.FlashKills)
			}
			if p.SmokeKills > 0 {
				fmt.Printf("    → Smoke: %d kill(s) a través de humo activo\n", p.SmokeKills)
			}
		}
	}

	// Overwatch Timeline
	if len(rep.Timeline) > 0 {
		fmt.Println()
		fmt.Println("  OVERWATCH TIMELINE — Pegar en CS:GO para ir al incidente")
		fmt.Println("──────────────────────────────────────────────────────────")
		maxTl := len(rep.Timeline)
		if maxTl > 15 {
			maxTl = 15
		}
		for _, inc := range rep.Timeline[:maxTl] {
			fmt.Printf("  [%-14s] %-20s %s\n", inc.Type, truncate(inc.Player, 20), inc.GotoCmd)
			if inc.Note != "" {
				fmt.Printf("                                 %s\n", truncate(inc.Note, 65))
			}
		}
	}

	// Admin Log
	if len(rep.AdminLog) > 0 {
		fmt.Println()
		fmt.Println("  ADMIN AUDIT LOG")
		fmt.Println("──────────────────────────────────────────────────────────")
		for _, ev := range rep.AdminLog {
			tag := ""
			if ev.IsRevenge {
				tag = " [VENGANZA]"
			}
			fmt.Printf("  Ronda %d  %s → %s %s%s\n", ev.Round, ev.AdminName, ev.Command, ev.TargetName, tag)
		}
	}

	// Render PNG report
	outDir := filepath.Dir(demoPath)
	baseName := strings.TrimSuffix(filepath.Base(demoPath), filepath.Ext(demoPath))
	ts := time.Now().Format("2006-01-02T15-04-05")
	pngPath := filepath.Join(outDir, fmt.Sprintf("reporte_%s_%s.png", baseName, ts))

	fmt.Println()
	fmt.Println("[+] Generando reporte visual PNG...")
	if err := report.RenderReportPNG(rep, pngPath); err != nil {
		fmt.Printf("[-] Error generando PNG: %v\n", err)
	} else {
		fmt.Printf("[OK] Reporte guardado: %s\n", filepath.Base(pngPath))
		_ = exec.Command("cmd", "/c", "start", "", pngPath).Start()
	}

	fmt.Printf("\nTiempo total de análisis: %.2fs (Go Engine)\n", elapsed.Seconds())
	time.Sleep(300 * time.Millisecond)
}

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
}
