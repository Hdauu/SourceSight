package core

import "testing"

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
