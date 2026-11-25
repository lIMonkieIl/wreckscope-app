package wf2

// Packet sizes and offsets correspond to GSIReader.Wreckfest2 packet layout,
// MIT-licensed layout documentation by JonathanSoszka/wf2-telemetry.
import (
	"math"
	"wreckscope/internal/racefeed"
)

// Decode applies a raw Pino UDP datagram to the racefeed engine.
// Returns false only for structurally invalid packets (bad signature/size/type).
func Decode(b []byte, e *racefeed.Engine) bool {
	if len(b) < HeaderSize || string(b[:4]) != Signature {
		return false
	}
	typ := int(b[4])
	if typ >= len(PacketSizes) || len(b) != PacketSizes[typ] {
		return false
	}
	h := parseHeader(b)
	racing := (h.StatusFlags&32) != 0 && (h.StatusFlags&2) == 0
	if !e.MarkPacket(typ, racing, h.SessionMS, h.RaceMS) {
		return true // Valid datagram, outside the current main packet's time window.
	}
	switch typ {
	case 0:
		main, ok := ParseMain(b)
		if !ok {
			return false
		}
		applyMain(e, main)
		return true
	case 1:
		p, ok := ParseParticipantsLeaderboard(b)
		if !ok {
			return false
		}
		e.UpdateVisibility(p.Visibility)
		for i := 0; i < ParticipantsMax; i++ {
			row := p.Rows[i]
			full := toLeaderboard(row)
			e.UpdateBoardRow(i, int(row.Status), int(row.Position), int(row.LapCurrent), int(row.Health), row.DeltaLeader, boardStats(row), &full)
		}
		return true
	case 2:
		p, ok := ParseParticipantsTiming(b)
		if !ok {
			return false
		}
		e.UpdateVisibility(p.Visibility)
		for i := 0; i < ParticipantsMax; i++ {
			e.UpdateTimingRow(i, toTiming(p.Rows[i]))
		}
		return true
	case 3:
		p, ok := ParseParticipantsSectors(b)
		if !ok {
			return false
		}
		e.UpdateVisibility(p.Visibility)
		for i := 0; i < ParticipantsMax; i++ {
			e.UpdateSectorsRow(i, toSectors(p.Rows[i]))
		}
		return true
	case 4:
		p, ok := ParseParticipantsMotion(b)
		if !ok {
			return false
		}
		e.UpdateVisibility(p.Visibility)
		for i := 0; i < ParticipantsMax; i++ {
			m := p.Rows[i]
			if math.Abs(m.X)+math.Abs(m.Z) < .000001 && m.Speed == 0 {
				continue
			}
			e.UpdateMotion(i, m.X, m.Y, m.Z, m.Speed)
			e.UpdateExtents(i, m.Extents)
			e.UpdateOrientation(i, m.QX, m.QY, m.QZ, m.QW)
		}
		return true
	case 5:
		p, ok := ParseParticipantsInfo(b)
		if !ok {
			return false
		}
		e.UpdateVisibility(p.Visibility)
		for i := 0; i < ParticipantsMax; i++ {
			info := p.Rows[i]
			e.UpdateInfoFull(i, toInfo(info), info.Bot)
		}
		return true
	case 6:
		p, ok := ParseParticipantsDamage(b)
		if !ok {
			return false
		}
		e.UpdateVisibility(p.Visibility)
		for i := 0; i < ParticipantsMax; i++ {
			e.UpdateDamage(i, toDamage(p.Rows[i]))
		}
		return true
	}
	return false
}

func applyMain(e *racefeed.Engine, main PacketMain) {
	session := main.Session
	var distance *float64
	if session.TrackLength > 0 && !math.IsInf(session.TrackLength, 0) {
		length := session.TrackLength
		distance = &length
	}
	e.UpdateRace(racefeed.RaceDetails{
		TrackID:           session.TrackID,
		Track:             session.TrackName,
		TrackName:         session.TrackName,
		Distance:          distance,
		TrackLength:       distance,
		Laps:              int(session.Laps),
		EventLength:       int(session.EventLength),
		Grid:              session.GridSize,
		GridSize:          session.GridSize,
		Remaining:         session.GridRemaining,
		GridSizeRemaining: session.GridRemaining,
		SectorCount:       session.SectorCount,
		SectorFract1:      session.SectorFract1,
		SectorFract2:      session.SectorFract2,
		Mode:              session.GameMode,
		GameMode:          session.GameMode,
		Damage:            session.DamageMode,
		DamageMode:        session.DamageMode,
		SessionStatus:     session.Status,
		Status:            session.Status,
		Flags:             int(main.MarshalFlags),
		GameStatus:        main.Header.StatusFlags,
		SessionMS:         main.Header.SessionMS,
	})
	e.UpdateLocalDriver(int(main.Info.ParticipantIndex), main.Info.PlayerName)
	e.UpdatePlayer(racefeed.PlayerTelemetry{
		MarshalFlagsPlayer:             main.MarshalFlags,
		ParticipantPlayerLeaderboard:   toLeaderboard(main.Leaderboard),
		ParticipantPlayerTiming:        toTiming(main.Timing),
		ParticipantPlayerTimingSectors: toSectors(main.Sectors),
		ParticipantPlayerInfo:          toInfo(main.Info),
		ParticipantPlayerDamage:        toDamage(main.Damage),
		CarPlayer:                      toCarFull(main.Car),
		PlayerStatusFlags:              main.PlayerStatus,
		InputExtended:                  racefeed.InputExtended{Flags: main.InputExtended.Flags, FFBForce: main.InputExtended.FFBForce},
	})
}

func boardStats(row Leaderboard) *racefeed.BoardStats {
	return &racefeed.BoardStats{
		TrackStatus: row.TrackStatus,
		Wrecks:      int(row.Wrecks),
		Frags:       int(row.Frags),
		Assists:     int(row.Assists),
		Score:       row.Score,
		Points:      row.Points,
		LapTiming:   row.LapTiming,
	}
}

func toLeaderboard(row Leaderboard) racefeed.ParticipantLeaderboard {
	return racefeed.ParticipantLeaderboard{
		Status: row.Status, TrackStatus: row.TrackStatus, LapCurrent: row.LapCurrent,
		Position: row.Position, Health: row.Health, Wrecks: row.Wrecks, Frags: row.Frags,
		Assists: row.Assists, Score: row.Score, Points: row.Points,
		DeltaLeader: row.DeltaLeader, LapTiming: row.LapTiming,
	}
}

func toTiming(t Timing) racefeed.ParticipantTiming {
	return racefeed.ParticipantTiming{
		LapTimeCurrent: t.LapTimeCurrent, LapTimePenaltyCurrent: t.LapTimePenaltyCurrent,
		LapTimeLast: t.LapTimeLast, LapTimeBest: t.LapTimeBest, LapBest: t.LapBest,
		DeltaAhead: t.DeltaAhead, DeltaBehind: t.DeltaBehind, LapProgress: t.LapProgress,
	}
}

func toSectors(s TimingSectors) racefeed.ParticipantTimingSectors {
	return racefeed.ParticipantTimingSectors{
		SectorTimeCurrentLap1: s.SectorTimeCurrentLap1, SectorTimeCurrentLap2: s.SectorTimeCurrentLap2,
		SectorTimeLastLap1: s.SectorTimeLastLap1, SectorTimeLastLap2: s.SectorTimeLastLap2,
		SectorTimeBestLap1: s.SectorTimeBestLap1, SectorTimeBestLap2: s.SectorTimeBestLap2,
		SectorTimeBest1: s.SectorTimeBest1, SectorTimeBest2: s.SectorTimeBest2, SectorTimeBest3: s.SectorTimeBest3,
	}
}

func toInfo(info Info) racefeed.ParticipantInfo {
	return racefeed.ParticipantInfo{
		CarID: info.CarID, CarName: info.CarName, PlayerName: info.PlayerName,
		ParticipantIndex: info.ParticipantIndex,
		LastNormalTrackStatusTime: info.LastNormalTrackStatusTime,
		LastCollisionTime:         info.LastCollisionTime,
		LastResetTime:             info.LastResetTime,
	}
}

func toDamage(d Damage) racefeed.ParticipantDamage {
	return racefeed.ParticipantDamage{
		DamageStates: append([]byte(nil), d.Raw[:]...),
		Parts:        append([]byte(nil), d.Parts[:]...),
	}
}

func toCarFull(c CarFull) racefeed.CarFull {
	var tires [4]racefeed.CarTire
	for i := range c.Tires {
		t := c.Tires[i]
		tires[i] = racefeed.CarTire{
			RPS: t.RPS, Camber: t.Camber, SlipRatio: t.SlipRatio, SlipAngle: t.SlipAngle,
			RadiusUnloaded: t.RadiusUnloaded, LoadVertical: t.LoadVertical,
			ForceLat: t.ForceLat, ForceLong: t.ForceLong,
			TemperatureInner: t.TemperatureInner, TemperatureTread: t.TemperatureTread,
			SuspensionVelocity: t.SuspensionVelocity, SuspensionDisplacement: t.SuspensionDisplacement,
			SuspensionDispNorm: t.SuspensionDispNorm, PositionVertical: t.PositionVertical,
			SurfaceType: t.SurfaceType,
		}
	}
	o := c.Orientation
	v := c.Velocity
	return racefeed.CarFull{
		Assists: racefeed.CarAssists{
			Flags: c.Assists.Flags, AssistGearbox: c.Assists.AssistGearbox,
			LevelABS: c.Assists.LevelABS, LevelTCS: c.Assists.LevelTCS, LevelESC: c.Assists.LevelESC,
		},
		Chassis: racefeed.CarChassis{
			TrackWidth: c.Chassis.TrackWidth, WheelBase: c.Chassis.WheelBase,
			SteeringWheelLockToLock: c.Chassis.SteeringWheelLockToLock,
			SteeringLock: c.Chassis.SteeringLock, CornerWeights: c.Chassis.CornerWeights,
		},
		Driveline: racefeed.CarDriveline{
			Type: c.Driveline.Type, Gear: c.Driveline.Gear, GearMax: c.Driveline.GearMax, Speed: c.Driveline.Speed,
		},
		Engine: racefeed.CarEngine{
			Flags: c.Engine.Flags, RPM: c.Engine.RPM, RPMMax: c.Engine.RPMMax,
			RPMRedline: c.Engine.RPMRedline, RPMIdle: c.Engine.RPMIdle,
			Torque: c.Engine.Torque, Power: c.Engine.Power,
			TempBlock: c.Engine.TempBlock, TempWater: c.Engine.TempWater,
			PressureManifold: c.Engine.PressureManifold, PressureOil: c.Engine.PressureOil,
		},
		Input: racefeed.CarInput{
			Throttle: c.Input.Throttle, Brake: c.Input.Brake, Clutch: c.Input.Clutch,
			Handbrake: c.Input.Handbrake, Steering: c.Input.Steering,
		},
		Orientation: racefeed.CarMotionOrientation{
			PositionX: o.X, PositionY: o.Y, PositionZ: o.Z,
			OrientationQuaternionX: o.QX, OrientationQuaternionY: o.QY,
			OrientationQuaternionZ: o.QZ, OrientationQuaternionW: o.QW,
			ExtentsX: o.Extents[0], ExtentsY: o.Extents[1], ExtentsZ: o.Extents[2],
		},
		Velocity: racefeed.CarMotionVelocity{
			VelocityLocalX: v.LocalX, VelocityLocalY: v.LocalY, VelocityLocalZ: v.LocalZ,
			AngularVelocityX: v.AngularX, AngularVelocityY: v.AngularY, AngularVelocityZ: v.AngularZ,
			AccelerationLocalX: v.AccelLocalX, AccelerationLocalY: v.AccelLocalY, AccelerationLocalZ: v.AccelLocalZ,
		},
		Tires: tires,
	}
}
