package core

import "sort"

// Ping as the scoreboard shows it, which the server updates about once a second.
//
// Reaction times (prefire, triggerbot) are measured between two things that both
// reach the server through the player's own connection: what they did and when they
// did it. A steady ping delays both equally, so it cancels out. A lag spike doesn't:
// commands held back during the spike arrive bunched, and a normal reaction can look
// impossibly fast. Reactions measured during a spike are therefore discarded.

// PingSpikeMs: a sample this far above the player's usual ping counts as a spike.
const PingSpikeMs = 50

// pingBaselineWeight is how fast the usual ping follows changes (moving average).
const pingBaselineWeight = 0.1

type PingData struct {
	Current  int
	Sum      int
	Count    int
	Max      int
	baseline float64
	samples  []pingSample
}

// pingSample is one scoreboard reading kept so incidents can show the ping they
// happened with (about one per second: a few thousand per player at most).
type pingSample struct {
	tick  int
	ms    int
	spike bool
}

// SampleAt records the scoreboard ping read on a tick.
func (p *PingData) SampleAt(tick, ms int) {
	if ms <= 0 {
		return
	}
	p.Sample(ms)
	p.samples = append(p.samples, pingSample{tick: tick, ms: ms, spike: p.Spiking()})
}

// At returns the ping read last before tick (or the first one, if tick comes before
// any reading) and whether it was a spike. 0 if there are no readings.
func (p *PingData) At(tick int) (ms int, spike bool) {
	if len(p.samples) == 0 {
		return 0, false
	}
	i := sort.Search(len(p.samples), func(i int) bool { return p.samples[i].tick > tick })
	if i > 0 {
		i--
	}
	return p.samples[i].ms, p.samples[i].spike
}

// Sample records the scoreboard ping. Zero means unknown (bots, not connected).
func (p *PingData) Sample(ms int) {
	if ms <= 0 {
		return
	}
	// A spike must not drag the baseline up with it, or the next spike sample would
	// look normal: only samples near the baseline update it.
	if p.Count == 0 {
		p.baseline = float64(ms)
	} else if float64(ms) <= p.baseline+PingSpikeMs {
		p.baseline += pingBaselineWeight * (float64(ms) - p.baseline)
	}
	p.Current = ms
	p.Sum += ms
	p.Count++
	p.Max = max(p.Max, ms)
}

// Spiking reports whether the current ping is well above the player's usual one.
// It needs a few samples first to know what usual is.
func (p *PingData) Spiking() bool {
	return p.Count >= 3 && float64(p.Current) > p.baseline+PingSpikeMs
}

// Average is the mean ping over the demo, 0 if unknown.
func (p *PingData) Average() int {
	if p.Count == 0 {
		return 0
	}
	return (p.Sum + p.Count/2) / p.Count
}
