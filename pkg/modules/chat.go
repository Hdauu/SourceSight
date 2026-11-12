package modules

import (
	"fmt"
	"strings"

	"sourcesight/pkg/core"
)

// Severe toxicity and slur patterns (multilingual: ES / EN / PT)
var toxicPatterns = []string{
	// Slurs & severe insults
	"negro de m", "negro sucio", "mono de m", "bolita de m", "paragua de m",
	"nigger", "nigga", "faggot", "fag", "maricon", "maricon de m",
	"mogolico", "autista de m", "retrasado",
	// Severe threats / harassment
	"matate", "suicidate", "kys", "kill yourself", "cancer de m",
	"hijo de mil puta", "hdp",
}

// All-chat ghosting callout patterns (dead player revealing info to enemy or giving away team)
var ghostingPatterns = []string{
	"esta en b", "estan en b", "van a b", "es b", "son b",
	"esta en a", "estan en a", "van a a", "es a", "son a",
	"site b", "site a", "bomb b", "bomba b", "bomb a", "bomba a",
	"defuseando", "defuse", "planted", "fake", "last a", "last b",
}

type ChatTracker struct {
	LastMsgText string
	LastMsgTick int
	RepeatCount int
}

// ChatInspector keeps the per-player spam state of one demo. Each analysis creates
// its own, so demos analyzed in parallel never share it.
type ChatInspector struct {
	trackers map[uint64]*ChatTracker
}

func NewChatInspector() *ChatInspector {
	return &ChatInspector{trackers: map[uint64]*ChatTracker{}}
}

// Inspect analyzes a chat message for toxicity, spam binds, and all-chat ghosting.
func (ci *ChatInspector) Inspect(
	senderName string,
	senderSteamID uint64,
	tick, round int,
	text string,
	isTeamOnly bool,
	senderIsAlive bool,
) *core.ChatMessageEntry {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	lower := strings.ToLower(trimmed)

	isToxic := false
	isGhost := false
	category := ""

	// 1. Check severe slurs & toxicity
	for _, pat := range toxicPatterns {
		if strings.Contains(lower, pat) {
			isToxic = true
			category = "TOXICITY"
			break
		}
	}

	// 2. Check all-chat ghosting (dead player leaking positions in global chat)
	if !isTeamOnly && !senderIsAlive {
		for _, pat := range ghostingPatterns {
			if strings.Contains(lower, pat) {
				isGhost = true
				category = "ALL-CHAT GHOSTING"
				break
			}
		}
	}

	// 3. Check spam bind (repeated text within 64 ticks)
	tr, ok := ci.trackers[senderSteamID]
	if !ok {
		tr = &ChatTracker{}
		ci.trackers[senderSteamID] = tr
	}

	if tr.LastMsgText == trimmed && (tick-tr.LastMsgTick) <= 64 {
		tr.RepeatCount++
		if tr.RepeatCount >= 3 && category == "" {
			category = "SPAM BIND"
		}
	} else {
		tr.RepeatCount = 1
	}
	tr.LastMsgText = trimmed
	tr.LastMsgTick = tick

	preroll := tick - 128
	if preroll < 128 {
		preroll = 128
	}

	return &core.ChatMessageEntry{
		Tick:          tick,
		Round:         round,
		SenderName:    senderName,
		SenderSteamID: senderSteamID,
		Text:          trimmed,
		IsTeamOnly:    isTeamOnly,
		IsToxicity:    isToxic,
		IsGhosting:    isGhost,
		Category:      category,
		GotoCmd:       fmt.Sprintf("demo_gototick %d", preroll),
	}
}
