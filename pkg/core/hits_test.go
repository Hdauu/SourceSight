package core

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHitProfile(t *testing.T) {
	var h HitProfile
	if !h.Empty() {
		t.Fatal("new profile should be empty")
	}
	h.AddShot(WeaponRifle)
	h.AddShot(WeaponRifle)
	h.AddShot(WeaponPistol)
	h.AddHit(WeaponRifle, HitGroupHead, 100)
	h.AddHit(WeaponRifle, HitGroupChest, 27)
	h.AddHit(WeaponPistol, HitGroupLeftLeg, 15)
	h.AddHit(WeaponPistol, HitGroupGear, 5) // unknown groups go to Other
	h.AddKill(true)

	if h.Shots != 3 || h.Kills != 1 || h.HeadshotKills != 1 {
		t.Fatalf("shots=%d kills=%d hs=%d", h.Shots, h.Kills, h.HeadshotKills)
	}
	if h.Zones.Head != (ZoneCount{1, 100}) || h.Zones.LeftLeg != (ZoneCount{1, 15}) || h.Zones.Other != (ZoneCount{1, 5}) {
		t.Fatalf("zones = %+v", h.Zones)
	}
	if tot := h.Zones.Total(); tot != (ZoneCount{4, 147}) {
		t.Fatalf("total = %+v", tot)
	}
	if r := h.ByClass[WeaponRifle]; r.Shots != 2 || r.Zones.Total().Hits != 2 {
		t.Fatalf("rifle = %+v", r)
	}

	// Merging another demo adds everything, per class too, without aliasing it
	var sum HitProfile
	sum.Merge(&h)
	sum.Merge(&h)
	sum.Merge(nil)
	if sum.Shots != 6 || sum.Zones.Head.Hits != 2 || sum.ByClass[WeaponPistol].Shots != 2 {
		t.Fatalf("merged = %+v", sum)
	}
	if h.ByClass[WeaponPistol].Shots != 1 {
		t.Fatal("merge modified the source profile")
	}
}

func TestPingSpikes(t *testing.T) {
	var p PingData
	p.Sample(0) // unknown: ignored
	if p.Count != 0 || p.Spiking() {
		t.Fatal("zero sample must be ignored")
	}
	for range 10 {
		p.Sample(30)
	}
	if p.Spiking() {
		t.Fatal("steady ping is not a spike")
	}
	p.Sample(140)
	if !p.Spiking() {
		t.Fatal("140 ms over a 30 ms baseline should be a spike")
	}
	p.Sample(150) // a long spike must not become the new normal
	if !p.Spiking() {
		t.Fatal("baseline followed the spike")
	}
	p.Sample(35)
	if p.Spiking() {
		t.Fatal("back to normal")
	}
	if p.Max != 150 || p.Average() != (300+140+150+35+6)/13 {
		t.Fatalf("max=%d avg=%d", p.Max, p.Average())
	}

	// High but steady ping is not a spike: it delays everything equally
	var high PingData
	for range 10 {
		high.Sample(180)
	}
	if high.Spiking() {
		t.Fatal("steady high ping is not a spike")
	}
}

// Shotgun pellets would push accuracy over 100%: they are left out of it.
func TestAccuracyWithoutShotguns(t *testing.T) {
	var h HitProfile
	for range 4 {
		h.AddShot(WeaponRifle)
	}
	h.AddHit(WeaponRifle, HitGroupChest, 27)
	h.AddShot(WeaponShotgun)
	for range 6 { // six pellets of one shell
		h.AddHit(WeaponShotgun, HitGroupChest, 20)
	}
	if hits, shots := h.Accuracy(); hits != 1 || shots != 4 {
		t.Fatalf("accuracy = %d/%d, want 1/4", hits, shots)
	}
	if h.Zones.Chest.Hits != 7 {
		t.Fatalf("pellets must still count for the body zones, chest = %d", h.Zones.Chest.Hits)
	}
}

// Zones that were never hit are left out of the JSON and come back as zero.
func TestBodyZonesJSON(t *testing.T) {
	var h HitProfile
	h.AddShot(WeaponRifle)
	h.AddHit(WeaponRifle, HitGroupHead, 100)
	b, err := json.Marshal(&h)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "chest") || !strings.Contains(string(b), `"head":{"hits":1,"damage":100}`) {
		t.Fatalf("json = %s", b)
	}
	var back HitProfile
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Zones.Head != (ZoneCount{1, 100}) || back.Zones.Chest != (ZoneCount{}) || back.ByClass[WeaponRifle].Zones.Head.Hits != 1 {
		t.Fatalf("round trip = %+v", back)
	}
}

func TestPingAt(t *testing.T) {
	var p PingData
	if ms, _ := p.At(100); ms != 0 {
		t.Fatalf("no readings: %d", ms)
	}
	for i, ms := range []int{30, 30, 30, 30, 140, 30} {
		p.SampleAt(1000+i*128, ms)
	}
	cases := []struct {
		tick  int
		ms    int
		spike bool
	}{
		{500, 30, false},               // before the first reading: the first one
		{1000 + 4*128 + 50, 140, true}, // during the spike
		{1000 + 5*128, 30, false},      // exactly on a reading
		{99999, 30, false},             // after the last
	}
	for _, c := range cases {
		ms, spike := p.At(c.tick)
		if ms != c.ms || spike != c.spike {
			t.Errorf("At(%d) = %d,%v want %d,%v", c.tick, ms, spike, c.ms, c.spike)
		}
	}
}
