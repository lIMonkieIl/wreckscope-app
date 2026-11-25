package racefeed

// Packet-shaped types (JonathanSoszka/wf2-telemetry layout.json / Wreckfest2Structs.cs).
// Flat aliases on Standing / RaceDetails stay for the UI.

type ParticipantLeaderboard struct {
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
	LapTiming   uint16 `json:"lapTiming"`
}

type ParticipantTiming struct {
	LapTimeCurrent        uint32  `json:"lapTimeCurrent"`
	LapTimePenaltyCurrent uint32  `json:"lapTimePenaltyCurrent"`
	LapTimeLast           uint32  `json:"lapTimeLast"`
	LapTimeBest           uint32  `json:"lapTimeBest"`
	LapBest               byte    `json:"lapBest"`
	DeltaAhead            int32   `json:"deltaAhead"`
	DeltaBehind           int32   `json:"deltaBehind"`
	LapProgress           float64 `json:"lapProgress"`
}

type ParticipantTimingSectors struct {
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

type ParticipantInfo struct {
	CarID                     string `json:"carId"`
	CarName                   string `json:"carName"`
	PlayerName                string `json:"playerName"`
	ParticipantIndex          byte   `json:"participantIndex"`
	LastNormalTrackStatusTime int32  `json:"lastNormalTrackStatusTime"`
	LastCollisionTime         int32  `json:"lastCollisionTime"`
	LastResetTime             int32  `json:"lastResetTime"`
}

type ParticipantDamage struct {
	DamageStates []byte `json:"damageStates"`          // raw 21-byte packed field
	Parts        []byte `json:"parts,omitempty"`       // unpacked 3-bit states (PART_COUNT)
}

type CarAssists struct {
	Flags         byte `json:"flags"`
	AssistGearbox byte `json:"assistGearbox"`
	LevelABS      byte `json:"levelAbs"`
	LevelTCS      byte `json:"levelTcs"`
	LevelESC      byte `json:"levelEsc"`
}

type CarChassis struct {
	TrackWidth              [2]float64 `json:"trackWidth"`
	WheelBase               float64    `json:"wheelBase"`
	SteeringWheelLockToLock int32      `json:"steeringWheelLockToLock"`
	SteeringLock            float64    `json:"steeringLock"`
	CornerWeights           [4]float64 `json:"cornerWeights"`
}

type CarDriveline struct {
	Type    byte    `json:"type"`
	Gear    byte    `json:"gear"`
	GearMax byte    `json:"gearMax"`
	Speed   float64 `json:"speed"`
}

type CarEngine struct {
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

type CarInput struct {
	Throttle  float64 `json:"throttle"`
	Brake     float64 `json:"brake"`
	Clutch    float64 `json:"clutch"`
	Handbrake float64 `json:"handbrake"`
	Steering  float64 `json:"steering"`
}

type CarMotionOrientation struct {
	PositionX              float64  `json:"positionX"`
	PositionY              float64  `json:"positionY"`
	PositionZ              float64  `json:"positionZ"`
	OrientationQuaternionX float64  `json:"orientationQuaternionX"`
	OrientationQuaternionY float64  `json:"orientationQuaternionY"`
	OrientationQuaternionZ float64  `json:"orientationQuaternionZ"`
	OrientationQuaternionW float64  `json:"orientationQuaternionW"`
	ExtentsX               uint16   `json:"extentsX"`
	ExtentsY               uint16   `json:"extentsY"`
	ExtentsZ               uint16   `json:"extentsZ"`
}

type CarMotionVelocity struct {
	VelocityLocalX     float64 `json:"velocityLocalX"`
	VelocityLocalY     float64 `json:"velocityLocalY"`
	VelocityLocalZ     float64 `json:"velocityLocalZ"`
	AngularVelocityX   float64 `json:"angularVelocityX"`
	AngularVelocityY   float64 `json:"angularVelocityY"`
	AngularVelocityZ   float64 `json:"angularVelocityZ"`
	AccelerationLocalX float64 `json:"accelerationLocalX"`
	AccelerationLocalY float64 `json:"accelerationLocalY"`
	AccelerationLocalZ float64 `json:"accelerationLocalZ"`
}

type CarTire struct {
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
	Assists     CarAssists           `json:"assists"`
	Chassis     CarChassis           `json:"chassis"`
	Driveline   CarDriveline         `json:"driveline"`
	Engine      CarEngine            `json:"engine"`
	Input       CarInput             `json:"input"`
	Orientation CarMotionOrientation `json:"orientation"`
	Velocity    CarMotionVelocity    `json:"velocity"`
	Tires       [4]CarTire           `json:"tires"`
}

type InputExtended struct {
	Flags    byte    `json:"flags"`
	FFBForce float64 `json:"ffbForce"`
}

// PlayerTelemetry mirrors PacketMain's player-side fields (session lives on Race).
type PlayerTelemetry struct {
	MarshalFlagsPlayer             uint16                    `json:"marshalFlagsPlayer"`
	ParticipantPlayerLeaderboard   ParticipantLeaderboard    `json:"participantPlayerLeaderboard"`
	ParticipantPlayerTiming        ParticipantTiming         `json:"participantPlayerTiming"`
	ParticipantPlayerTimingSectors ParticipantTimingSectors  `json:"participantPlayerTimingSectors"`
	ParticipantPlayerInfo          ParticipantInfo           `json:"participantPlayerInfo"`
	ParticipantPlayerDamage        ParticipantDamage         `json:"participantPlayerDamage"`
	CarPlayer                      CarFull                   `json:"carPlayer"`
	PlayerStatusFlags              uint16                    `json:"playerStatusFlags"`
	InputExtended                  InputExtended             `json:"inputExtended"`
}
