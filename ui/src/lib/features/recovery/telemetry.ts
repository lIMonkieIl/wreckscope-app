export type SpeedUnit = 'km/h' | 'mph';
export type DisplayMode = 'system' | 'light' | 'dark';
export type DashboardView = 'race' | 'recovery';
export type CarState = 'moving' | 'slow' | 'caution' | 'alert' | 'out';

export interface DetectionSettings {
	cautionSeconds: number;
	alertSeconds: number;
	speedThreshold: number;
	hideBots: boolean;
	speedUnit: SpeedUnit;
	muteAlertSound: boolean;
	drawTrackOutline: boolean;
	theme: string;
	mode: DisplayMode;
	/** race | recovery (PC + LAN). */
	view: DashboardView;
	setupCompleted?: boolean;
	telemetryConfigPath?: string;
}

export interface AppearanceSettings {
	theme: string;
	mode: DisplayMode;
}

/** Packet nests (layout.json / Wreckfest2Structs); flat fields kept for UI. */
export interface ParticipantLeaderboard {
	status: number;
	trackStatus: number;
	lapCurrent: number;
	position: number;
	health: number;
	wrecks: number;
	frags: number;
	assists: number;
	score: number;
	points: number;
	deltaLeader: number;
	lapTiming: number;
}

export interface ParticipantTiming {
	lapTimeCurrent: number;
	lapTimePenaltyCurrent: number;
	lapTimeLast: number;
	lapTimeBest: number;
	lapBest: number;
	deltaAhead: number;
	deltaBehind: number;
	lapProgress: number;
}

export interface ParticipantTimingSectors {
	sectorTimeCurrentLap1: number;
	sectorTimeCurrentLap2: number;
	sectorTimeLastLap1: number;
	sectorTimeLastLap2: number;
	sectorTimeBestLap1: number;
	sectorTimeBestLap2: number;
	sectorTimeBest1: number;
	sectorTimeBest2: number;
	sectorTimeBest3: number;
}

export interface ParticipantInfo {
	carId: string;
	carName: string;
	playerName: string;
	participantIndex: number;
	lastNormalTrackStatusTime: number;
	lastCollisionTime: number;
	lastResetTime: number;
}

export interface ParticipantDamage {
	damageStates: number[];
	parts?: number[];
}

export interface Standing {
	slot: number;
	name: string;
	carId?: string;
	carName?: string;
	status: number;
	place: number;
	lap: number;
	health: number;
	trackStatus?: number;
	frags?: number;
	wrecks?: number;
	assists?: number;
	score?: number;
	points?: number;
	progress?: number;
	bestSectors?: [number | null, number | null, number | null];
	gap?: number;
	current?: number;
	last?: number;
	best?: number;
	leaderboard?: ParticipantLeaderboard;
	timing?: ParticipantTiming;
	sectors?: ParticipantTimingSectors;
	info?: ParticipantInfo;
	damage?: ParticipantDamage;
}

export interface RecoveryCar extends Standing {
	x: number;
	y?: number;
	extents?: [number, number, number];
	orientation?: [number, number, number, number];
	z: number;
	speed: number;
	tilt: number;
	state: CarState;
	reason: string;
	stopped: number;
	updated: string;
	score?: number;
	scoreFactors?: { label: string; points: number }[];
	incidentId?: string;
	incidentSince?: string;
}

export interface RaceDetails {
	trackId: string;
	track: string;
	trackName?: string;
	distance?: number;
	trackLength?: number;
	laps: number;
	eventLength?: number;
	mode: number;
	gameMode?: number;
	damage: number;
	damageMode?: number;
	grid: number;
	gridSize?: number;
	remaining: number;
	gridSizeRemaining?: number;
	sectorCount?: number;
	sectorFract1?: number;
	sectorFract2?: number;
	sessionStatus?: number;
	status?: number;
	flags: number;
	gameStatus?: number;
	sessionMs?: number;
	clock: number;
}

/** PacketMain player-side block. */
export interface PlayerTelemetry {
	marshalFlagsPlayer: number;
	participantPlayerLeaderboard: ParticipantLeaderboard;
	participantPlayerTiming: ParticipantTiming;
	participantPlayerTimingSectors: ParticipantTimingSectors;
	participantPlayerInfo: ParticipantInfo;
	participantPlayerDamage: ParticipantDamage;
	carPlayer: Record<string, unknown>;
	playerStatusFlags: number;
	inputExtended: { flags: number; ffbForce: number };
}

export interface TelemetrySnapshot {
	mode: 'live' | 'demo';
	connection: 'waiting' | 'live' | 'stale' | 'demo' | 'offline';
	cars: RecoveryCar[];
	standings?: Standing[];
	settings: DetectionSettings;
	sessionRacing?: boolean;
	raceEpoch?: number;
	localSlot?: number;
	/** Medic slots (PC + LAN; no cap). */
	medics?: number[];
	race?: RaceDetails;
	player?: PlayerTelemetry;
	participantVisibility?: number;
	/** Bind / status line from the host (UDP/HTTP failures). */
	message?: string;
}

export interface PhoneAccess {
	lan: boolean;
	urls: string[];
	readOnly?: boolean;
	/** True only when the LAN HTTP server bound successfully. */
	listening?: boolean;
	error?: string;
}
