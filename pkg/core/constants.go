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
	WallTrackMinTicks    = 64     // ~1.0s sustained lock @ 64-tick
	WallTrackMinSweepDeg = 20.0   // crosshair must actually follow the target, not hold an angle

	// Staff Modules (G, H, I)
	GhostWindowTicks       = 128   // 2.0s @ 64-tick
	GhostMinDelayMs        = 500.0 // faster shots are kill-feed trades
	GhostMinDistUnits      = 800.0 // nearer than this the shooter could hear the fight
	RevengeSlayWindowTicks = 320   // 5.0s @ 64-tick
	AfkDisplacementUnits   = 50.0
	AfkRoundsThreshold     = 3
	TeamflashThresholdSecs = 10.0
)
