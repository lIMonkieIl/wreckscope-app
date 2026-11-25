package wf2

// Decoded Pino UDP structs. Field names track layout.json / Wreckfest2Structs.cs.
// Reserved blobs without documented meaning are omitted (skipped at decode).

type Header struct {
	Signature  string `json:"signature"`
	Type       byte   `json:"type"`
	StatusFlags byte  `json:"statusFlags"`
	SessionMS  int32  `json:"sessionMs"`
	RaceMS     int32  `json:"raceMs"`
}

type Session struct {
	TrackID       string   `json:"trackId"`
	TrackName     string   `json:"trackName"`
	TrackLength   float64  `json:"trackLength"`
	Laps          int16    `json:"laps"`
	EventLength   int16    `json:"eventLength"`
	GridSize      byte     `json:"gridSize"`
	GridRemaining byte     `json:"gridRemaining"`
	SectorCount   byte     `json:"sectorCount"`
	SectorFract1  float64  `json:"sectorFract1"`
	SectorFract2  float64  `json:"sectorFract2"`
	GameMode      byte     `json:"gameMode"`
	DamageMode    byte     `json:"damageMode"`
	Status        byte     `json:"status"` // SessionStatus from layout.json
}

type Leaderboard struct {
	Status      byte   `json:"status"`
	TrackStatus byte   `json:"trackStatus"`
	LapCurrent  uint16 `json:"lapCurrent"`
	Position    byte   `json:"position"`
	Health      byte   `json:"health"`
	Wrecks      uint16 `json:"wrecks"`
	Frags       uint16 `json:"frags"`
	Assists     uint16 `json:"assists"`
	Score       int32  `json:"score"`
	Points      int32  `json:"points"`
	DeltaLeader int32  `json:"deltaLeader"`
	LapTiming   uint16 `json:"lapTiming"` // layout.json; absent in older C# struct
}

type Timing struct {
	LapTimeCurrent        uint32  `json:"lapTimeCurrent"`
	LapTimePenaltyCurrent uint32  `json:"lapTimePenaltyCurrent"`
	LapTimeLast           uint32  `json:"lapTimeLast"`
	LapTimeBest           uint32  `json:"lapTimeBest"`
	LapBest               byte    `json:"lapBest"`
	DeltaAhead            int32   `json:"deltaAhead"`
	DeltaBehind           int32   `json:"deltaBehind"`
	LapProgress           float64 `json:"lapProgress"`
}

type TimingSectors struct {
	SectorTimeCurrentLap1 uint32 `json:"sectorTimeCurrentLap1"`
	SectorTimeCurrentLap2 uint32 `json:"sectorTimeCurrentLap2"`
	SectorTimeLastLap1    uint32 `json:"sectorTimeLastLap1"`
	SectorTimeLastLap2    uint32 `json:"sectorTimeLastLap2"`
	SectorTimeBestLap1    uint32 `json:"sectorTimeBestLap1"`
	SectorTimeBestLap2    uint32 `json:"sectorTimeBestLap2"`
	SectorTimeBest1       uint32 `json:"sectorTimeBest1"`
	SectorTimeBest2       uint32 `json:"sectorTimeBest2"`
	SectorTimeBest3       uint32 `json:"sectorTimeBest3"`
}

type Motion struct {
	X           float64    `json:"x"`
	Y           float64    `json:"y"`
	Z           float64    `json:"z"`
	QX          float64    `json:"-"`
	QY          float64    `json:"-"`
	QZ          float64    `json:"-"`
	QW          float64    `json:"-"`
	Orientation [4]float64 `json:"orientation"`
	Extents     [3]uint16  `json:"extents"`
	Speed       float64    `json:"speed"`
}

type Info struct {
	CarID                     string `json:"carId"`
	CarName                   string `json:"carName"`
	PlayerName                string `json:"playerName"`
	ParticipantIndex          byte   `json:"participantIndex"`
	LastNormalTrackStatusTime int32  `json:"lastNormalTrackStatusTime"`
	LastCollisionTime         int32  `json:"lastCollisionTime"`
	LastResetTime             int32  `json:"lastResetTime"`
	Bot                       bool   `json:"bot"`
}

// Damage holds unpacked 3-bit part states (0..4) for DamagePartCount entries.
type Damage struct {
	Parts [DamagePartCount]byte `json:"parts"`
	Raw   [DamageBytesPerDriver]byte `json:"raw"`
}

type Assists struct {
	Flags         byte `json:"flags"`
	AssistGearbox byte `json:"assistGearbox"`
	LevelABS      byte `json:"levelAbs"`
	LevelTCS      byte `json:"levelTcs"`
	LevelESC      byte `json:"levelEsc"`
}

type Chassis struct {
	TrackWidth             [2]float64 `json:"trackWidth"`
	WheelBase              float64    `json:"wheelBase"`
	SteeringWheelLockToLock int32     `json:"steeringWheelLockToLock"`
	SteeringLock           float64    `json:"steeringLock"`
	CornerWeights          [4]float64 `json:"cornerWeights"`
}

type Driveline struct {
	Type    byte    `json:"type"`
	Gear    byte    `json:"gear"`
	GearMax byte    `json:"gearMax"`
	Speed   float64 `json:"speed"`
}

type Engine struct {
	Flags            byte    `json:"flags"`
	RPM              int32   `json:"rpm"`
	RPMMax           int32   `json:"rpmMax"`
	RPMRedline       int32   `json:"rpmRedline"`
	RPMIdle          int32   `json:"rpmIdle"`
	Torque           float64 `json:"torque"`
	Power            float64 `json:"power"`
	TempBlock        float64 `json:"tempBlock"`
	TempWater        float64 `json:"tempWater"`
	PressureManifold float64 `json:"pressureManifold"`
	PressureOil      float64 `json:"pressureOil"`
}

type Input struct {
	Throttle  float64 `json:"throttle"`
	Brake     float64 `json:"brake"`
	Clutch    float64 `json:"clutch"`
	Handbrake float64 `json:"handbrake"`
	Steering  float64 `json:"steering"`
}

type Velocity struct {
	LocalX      float64 `json:"localX"`
	LocalY      float64 `json:"localY"`
	LocalZ      float64 `json:"localZ"`
	AngularX    float64 `json:"angularX"`
	AngularY    float64 `json:"angularY"`
	AngularZ    float64 `json:"angularZ"`
	AccelLocalX float64 `json:"accelLocalX"`
	AccelLocalY float64 `json:"accelLocalY"`
	AccelLocalZ float64 `json:"accelLocalZ"`
}

type Tire struct {
	RPS                    float64 `json:"rps"`
	Camber                 float64 `json:"camber"`
	SlipRatio              float64 `json:"slipRatio"`
	SlipAngle              float64 `json:"slipAngle"`
	RadiusUnloaded         float64 `json:"radiusUnloaded"`
	LoadVertical           float64 `json:"loadVertical"`
	ForceLat               float64 `json:"forceLat"`
	ForceLong              float64 `json:"forceLong"`
	TemperatureInner       float64 `json:"temperatureInner"`
	TemperatureTread       float64 `json:"temperatureTread"`
	SuspensionVelocity     float64 `json:"suspensionVelocity"`
	SuspensionDisplacement float64 `json:"suspensionDisplacement"`
	SuspensionDispNorm     float64 `json:"suspensionDispNorm"`
	PositionVertical       float64 `json:"positionVertical"`
	SurfaceType            byte    `json:"surfaceType"`
}

type CarFull struct {
	Assists     Assists  `json:"assists"`
	Chassis     Chassis  `json:"chassis"`
	Driveline   Driveline `json:"driveline"`
	Engine      Engine   `json:"engine"`
	Input       Input    `json:"input"`
	Orientation Motion   `json:"orientation"` // position + quaternion + extents (speed unused)
	Velocity    Velocity `json:"velocity"`
	Tires       [TireCount]Tire `json:"tires"`
}

type InputExtended struct {
	Flags    byte    `json:"flags"`
	FFBForce float64 `json:"ffbForce"`
}

// PacketMain is the local-player + session datagram (type 0).
type PacketMain struct {
	Header         Header         `json:"header"`
	MarshalFlags   uint16         `json:"marshalFlags"`
	Leaderboard    Leaderboard    `json:"leaderboard"`
	Timing         Timing         `json:"timing"`
	Sectors        TimingSectors  `json:"sectors"`
	Info           Info           `json:"info"`
	Damage         Damage         `json:"damage"`
	Car            CarFull        `json:"car"`
	Session        Session        `json:"session"`
	PlayerStatus   uint16         `json:"playerStatusFlags"`
	InputExtended  InputExtended  `json:"inputExtended"`
}

type PacketParticipantsLeaderboard struct {
	Header     Header
	Visibility byte
	Rows       [ParticipantsMax]Leaderboard
}

type PacketParticipantsTiming struct {
	Header     Header
	Visibility byte
	Rows       [ParticipantsMax]Timing
}

type PacketParticipantsSectors struct {
	Header     Header
	Visibility byte
	Rows       [ParticipantsMax]TimingSectors
}

type PacketParticipantsMotion struct {
	Header     Header
	Visibility byte
	Rows       [ParticipantsMax]Motion
}

type PacketParticipantsInfo struct {
	Header     Header
	Visibility byte
	Rows       [ParticipantsMax]Info
}

type PacketParticipantsDamage struct {
	Header     Header
	Visibility byte
	Rows       [ParticipantsMax]Damage
}
