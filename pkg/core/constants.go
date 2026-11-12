package core

// Source Engine bitmasks, physics constants and detection thresholds.
const (
	// Netprops & input masks
	FL_ONGROUND = 1  // bit 0 (0x1)
	FL_DUCKING  = 2  // bit 1 (0x2)
	FL_INWATER  = 16 // bit 4 (0x10)

	IN_ATTACK = 1
	IN_JUMP   = 2
	IN_DUCK   = 4

	MOVETYPE_LADDER = 9

	// Module A: Bhop
	BhopPerfWindow        = 1  // 0 or 1 ticks on ground
	BhopMaxGroundWindow   = 15 // hops with >15 ground ticks are walking/stopping
	BhopMinJumps          = 5  // minimum jumps required for evaluation
	BhopCheatMinJumps     = 10 // AUTOBUNNY needs the ratio over at least this many jumps
	BhopSuspMinJumps      = 30 // fewer jumps are too small a sample to call suspicious
	BhopCheatPerfRatio    = 75.0
	BhopSuspPerfRatio     = 55.0
	BhopCheatStreak       = 5 // consecutive perfect hops (P < 0.2%)
	BhopOneTickCheatRatio = 80.0

	// Module B: Triggerbot
	TriggerHumanMinMs = 140.0
	TriggerBotMaxMs   = 80.0
	TriggerBotSigma   = 15.0
	TriggerMinEvents  = 4

	// Module C: Aimlock & Silent Aim
	AimSnapThresholdDeg = 45.0
	AimWrapThreshold    = 180.0
	AimAliveWarmup      = 192 // ticks after spawn to ignore warmup flick
	AimMinSnapsToFlag   = 10
	AimCheatMinSnaps    = 50   // spinbots: hundreds of snaps; fewer stays suspicious
	AimBurstMinSnaps    = 5    // snaps inside a 1s window: isolated human flicks never cluster like this
	SilentAimDivDeg     = 25.0 // exceeds maximum running spray inaccuracy cone

	// Module D: Anti-Flash & Smoke
	FlashBlindAlpha  = 255.0
	FlashDurationMin = 1.5
	SmokeRadiusUnits = 144.0

	// Module E: Macro
	MacroSigmaThreshold = 1.5 // ms, std-dev of the interval between shots
	MacroMinSamples     = 10

	// Module W: Wallhack / Occluded ESP tracking
	WallTrackMinDist     = 150.0  // closer engagements are trivially trackable
	WallTrackMaxDist     = 3000.0 // beyond this, crosshair overlap is noise
	WallTrackMaxAngleDeg = 4.0    // crosshair cone around the hidden enemy's chest
	WallTrackMinTicks    = 64     // 1.0 s of sustained lock, in 64-tick ticks (scaled to the demo)
	WallTrackMinSweepDeg = 20.0   // crosshair must actually follow the target, not hold an angle

	// Staff Modules (G, H, I)
	GhostWindowTicks       = 128   // 2.0s @ 64-tick
	GhostMinDelayMs        = 500.0 // faster shots are kill-feed trades
	GhostMinDistUnits      = 800.0 // nearer than this the shooter could hear the fight
	RevengeSlayWindowTicks = 320   // 5.0 s, in 64-tick ticks (scaled to the demo)
	AfkDisplacementUnits   = 50.0
	AfkRoundsThreshold     = 3
	TeamflashThresholdSecs = 10.0
)

// Prefire: damage right after an enemy becomes visible to the shooter.
// Calibrated on 13 demos: only a reported cheater crossed all three limits.
const (
	PrefireWindowMs       = 100.0  // faster than human visual reaction
	PrefireMaxSightMs     = 2000.0 // damage later than this isn't a reaction to the peek
	PrefireMinEngagements = 15
	PrefireMinRatio       = 35.0 // % of engagements that are prefires
	PrefireMaxMedianMs    = 150.0
)
