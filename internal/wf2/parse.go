package wf2

import (
	"encoding/binary"
	"math"
	"strings"
	"unicode/utf8"
)

func u16(b []byte, o int) int {
	return int(binary.LittleEndian.Uint16(b[o:]))
}
func u16n(b []byte, o int) uint16 {
	return binary.LittleEndian.Uint16(b[o:])
}
func i32(b []byte, o int) int32 {
	return int32(binary.LittleEndian.Uint32(b[o:]))
}
func u32(b []byte, o int) uint32 {
	return binary.LittleEndian.Uint32(b[o:])
}
func f32(b []byte, o int) float64 {
	return float64(math.Float32frombits(binary.LittleEndian.Uint32(b[o:])))
}
func str(b []byte) string {
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		b = b[:i]
	}
	if !utf8.Valid(b) {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func parseHeader(b []byte) Header {
	return Header{
		Signature:   string(b[:4]),
		Type:        b[4],
		StatusFlags: b[5],
		SessionMS:   i32(b, 6),
		RaceMS:      i32(b, 10),
	}
}

func parseLeaderboard(b []byte, o int) Leaderboard {
	return Leaderboard{
		Status:      b[o],
		TrackStatus: b[o+1],
		LapCurrent:  u16n(b, o+2),
		Position:    b[o+4],
		Health:      b[o+5],
		Wrecks:      u16n(b, o+6),
		Frags:       u16n(b, o+8),
		Assists:     u16n(b, o+10),
		Score:       i32(b, o+12),
		Points:      i32(b, o+16),
		DeltaLeader: i32(b, o+20),
		LapTiming:   u16n(b, o+24),
	}
}

func parseTiming(b []byte, o int) Timing {
	return Timing{
		LapTimeCurrent:        u32(b, o),
		LapTimePenaltyCurrent: u32(b, o+4),
		LapTimeLast:           u32(b, o+8),
		LapTimeBest:           u32(b, o+12),
		LapBest:               b[o+16],
		DeltaAhead:            i32(b, o+17),
		DeltaBehind:           i32(b, o+21),
		LapProgress:           f32(b, o+25),
	}
}

func parseSectors(b []byte, o int) TimingSectors {
	return TimingSectors{
		SectorTimeCurrentLap1: u32(b, o),
		SectorTimeCurrentLap2: u32(b, o+4),
		SectorTimeLastLap1:    u32(b, o+8),
		SectorTimeLastLap2:    u32(b, o+12),
		SectorTimeBestLap1:    u32(b, o+16),
		SectorTimeBestLap2:    u32(b, o+20),
		SectorTimeBest1:       u32(b, o+24),
		SectorTimeBest2:       u32(b, o+28),
		SectorTimeBest3:       u32(b, o+32),
	}
}

func parseOrientation(b []byte, o int) Motion {
	m := Motion{
		X:       f32(b, o),
		Y:       f32(b, o+4),
		Z:       f32(b, o+8),
		QX:      f32(b, o+12),
		QY:      f32(b, o+16),
		QZ:      f32(b, o+20),
		QW:      f32(b, o+24),
		Extents: [3]uint16{u16n(b, o+28), u16n(b, o+30), u16n(b, o+32)},
	}
	m.Orientation = [4]float64{m.QX, m.QY, m.QZ, m.QW}
	return m
}

func parseMotion(b []byte, o int) Motion {
	m := parseOrientation(b, o)
	m.Speed = f32(b, o+34)
	return m
}

func parseInfo(b []byte, o int) Info {
	carID := str(b[o : o+64])
	name := str(b[o+160 : o+184])
	return Info{
		CarID:                     carID,
		CarName:                   str(b[o+64 : o+160]),
		PlayerName:                name,
		ParticipantIndex:          b[o+184],
		LastNormalTrackStatusTime: i32(b, o+185),
		LastCollisionTime:         i32(b, o+189),
		LastResetTime:             i32(b, o+193),
		Bot:                       strings.HasPrefix(name, "*BOT") && strings.Contains(carID, ":ai_"),
	}
}

// UnpackDamage expands the 21-byte 3-bit packed damageStates into per-part states.
func UnpackDamage(raw []byte) Damage {
	var d Damage
	copy(d.Raw[:], raw)
	for part := 0; part < DamagePartCount; part++ {
		bit := part * DamageBitsPerPart
		byteIdx := bit / 8
		shift := bit % 8
		var v byte
		if shift <= 5 {
			v = (raw[byteIdx] >> shift) & 0x7
		} else {
			// Spans two bytes.
			lo := raw[byteIdx] >> shift
			hi := raw[byteIdx+1] << (8 - shift)
			v = (lo | hi) & 0x7
		}
		d.Parts[part] = v
	}
	return d
}

func parseDamage(b []byte, o int) Damage {
	return UnpackDamage(b[o : o+DamageBytesPerDriver])
}

func parseAssists(b []byte, o int) Assists {
	return Assists{
		Flags:         b[o],
		AssistGearbox: b[o+1],
		LevelABS:      b[o+2],
		LevelTCS:      b[o+3],
		LevelESC:      b[o+4],
	}
}

func parseChassis(b []byte, o int) Chassis {
	return Chassis{
		TrackWidth:              [2]float64{f32(b, o), f32(b, o+4)},
		WheelBase:               f32(b, o+8),
		SteeringWheelLockToLock: i32(b, o+12),
		SteeringLock:            f32(b, o+16),
		CornerWeights:           [4]float64{f32(b, o+20), f32(b, o+24), f32(b, o+28), f32(b, o+32)},
	}
}

func parseDriveline(b []byte, o int) Driveline {
	return Driveline{
		Type:    b[o],
		Gear:    b[o+1],
		GearMax: b[o+2],
		Speed:   f32(b, o+3),
	}
}

func parseEngine(b []byte, o int) Engine {
	return Engine{
		Flags:            b[o],
		RPM:              i32(b, o+1),
		RPMMax:           i32(b, o+5),
		RPMRedline:       i32(b, o+9),
		RPMIdle:          i32(b, o+13),
		Torque:           f32(b, o+17),
		Power:            f32(b, o+21),
		TempBlock:        f32(b, o+25),
		TempWater:        f32(b, o+29),
		PressureManifold: f32(b, o+33),
		PressureOil:      f32(b, o+37),
	}
}

func parseInput(b []byte, o int) Input {
	return Input{
		Throttle:  f32(b, o),
		Brake:     f32(b, o+4),
		Clutch:    f32(b, o+8),
		Handbrake: f32(b, o+12),
		Steering:  f32(b, o+16),
	}
}

func parseVelocity(b []byte, o int) Velocity {
	return Velocity{
		LocalX:      f32(b, o),
		LocalY:      f32(b, o+4),
		LocalZ:      f32(b, o+8),
		AngularX:    f32(b, o+12),
		AngularY:    f32(b, o+16),
		AngularZ:    f32(b, o+20),
		AccelLocalX: f32(b, o+24),
		AccelLocalY: f32(b, o+28),
		AccelLocalZ: f32(b, o+32),
	}
}

func parseTire(b []byte, o int) Tire {
	return Tire{
		RPS:                    f32(b, o),
		Camber:                 f32(b, o+4),
		SlipRatio:              f32(b, o+8),
		SlipAngle:              f32(b, o+12),
		RadiusUnloaded:         f32(b, o+16),
		LoadVertical:           f32(b, o+20),
		ForceLat:               f32(b, o+24),
		ForceLong:              f32(b, o+28),
		TemperatureInner:       f32(b, o+32),
		TemperatureTread:       f32(b, o+36),
		SuspensionVelocity:     f32(b, o+40),
		SuspensionDisplacement: f32(b, o+44),
		SuspensionDispNorm:     f32(b, o+48),
		PositionVertical:       f32(b, o+52),
		SurfaceType:            b[o+56],
	}
}

func parseCarFull(b []byte, o int) CarFull {
	var tires [TireCount]Tire
	for i := 0; i < TireCount; i++ {
		tires[i] = parseTire(b, o+CarTiresOff+i*TireSize)
	}
	return CarFull{
		Assists:     parseAssists(b, o+CarAssistsOff),
		Chassis:     parseChassis(b, o+CarChassisOff),
		Driveline:   parseDriveline(b, o+CarDrivelineOff),
		Engine:      parseEngine(b, o+CarEngineOff),
		Input:       parseInput(b, o+CarInputOff),
		Orientation: parseOrientation(b, o+CarOrientOff),
		Velocity:    parseVelocity(b, o+CarVelocityOff),
		Tires:       tires,
	}
}

func parseSession(b []byte, o int) Session {
	return Session{
		TrackID:       str(b[o : o+64]),
		TrackName:     str(b[o+64 : o+160]),
		TrackLength:   f32(b, o+160),
		Laps:          int16(u16n(b, o+164)),
		EventLength:   int16(u16n(b, o+166)),
		GridSize:      b[o+168],
		GridRemaining: b[o+169],
		SectorCount:   b[o+170],
		SectorFract1:  f32(b, o+171),
		SectorFract2:  f32(b, o+175),
		GameMode:      b[o+179],
		DamageMode:    b[o+180],
		Status:        b[o+181],
	}
}

func parseInputExtended(b []byte, o int) InputExtended {
	return InputExtended{
		Flags:    b[o],
		FFBForce: f32(b, o+1),
	}
}

// ParseMain decodes a type-0 PacketMain datagram.
func ParseMain(b []byte) (PacketMain, bool) {
	if len(b) != PacketMainSize || string(b[:4]) != Signature || b[4] != 0 {
		return PacketMain{}, false
	}
	return PacketMain{
		Header:        parseHeader(b),
		MarshalFlags:  u16n(b, MainMarshalFlags),
		Leaderboard:   parseLeaderboard(b, MainPlayerBoard),
		Timing:        parseTiming(b, MainPlayerTiming),
		Sectors:       parseSectors(b, MainPlayerSector),
		Info:          parseInfo(b, MainPlayerInfo),
		Damage:        parseDamage(b, MainPlayerDamage),
		Car:           parseCarFull(b, MainCarPlayer),
		Session:       parseSession(b, MainSession),
		PlayerStatus:  u16n(b, MainPlayerStatus),
		InputExtended: parseInputExtended(b, MainInputExt),
	}, true
}

func ParseParticipantsLeaderboard(b []byte) (PacketParticipantsLeaderboard, bool) {
	if len(b) != PacketParticipantsLeaderboardSize || string(b[:4]) != Signature || b[4] != 1 {
		return PacketParticipantsLeaderboard{}, false
	}
	p := PacketParticipantsLeaderboard{Header: parseHeader(b), Visibility: b[14]}
	for i := 0; i < ParticipantsMax; i++ {
		p.Rows[i] = parseLeaderboard(b, 15+i*StrideLeaderboard)
	}
	return p, true
}

func ParseParticipantsTiming(b []byte) (PacketParticipantsTiming, bool) {
	if len(b) != PacketParticipantsTimingSize || string(b[:4]) != Signature || b[4] != 2 {
		return PacketParticipantsTiming{}, false
	}
	p := PacketParticipantsTiming{Header: parseHeader(b), Visibility: b[14]}
	for i := 0; i < ParticipantsMax; i++ {
		p.Rows[i] = parseTiming(b, 15+i*StrideTiming)
	}
	return p, true
}

func ParseParticipantsSectors(b []byte) (PacketParticipantsSectors, bool) {
	if len(b) != PacketParticipantsSectorsSize || string(b[:4]) != Signature || b[4] != 3 {
		return PacketParticipantsSectors{}, false
	}
	p := PacketParticipantsSectors{Header: parseHeader(b), Visibility: b[14]}
	for i := 0; i < ParticipantsMax; i++ {
		p.Rows[i] = parseSectors(b, 15+i*StrideSectors)
	}
	return p, true
}

func ParseParticipantsMotion(b []byte) (PacketParticipantsMotion, bool) {
	if len(b) != PacketParticipantsMotionSize || string(b[:4]) != Signature || b[4] != 4 {
		return PacketParticipantsMotion{}, false
	}
	p := PacketParticipantsMotion{Header: parseHeader(b), Visibility: b[14]}
	for i := 0; i < ParticipantsMax; i++ {
		p.Rows[i] = parseMotion(b, 15+i*StrideMotion)
	}
	return p, true
}

func ParseParticipantsInfo(b []byte) (PacketParticipantsInfo, bool) {
	if len(b) != PacketParticipantsInfoSize || string(b[:4]) != Signature || b[4] != 5 {
		return PacketParticipantsInfo{}, false
	}
	p := PacketParticipantsInfo{Header: parseHeader(b), Visibility: b[14]}
	for i := 0; i < ParticipantsMax; i++ {
		p.Rows[i] = parseInfo(b, 15+i*StrideInfo)
	}
	return p, true
}

func ParseParticipantsDamage(b []byte) (PacketParticipantsDamage, bool) {
	if len(b) != PacketParticipantsDamageSize || string(b[:4]) != Signature || b[4] != 6 {
		return PacketParticipantsDamage{}, false
	}
	p := PacketParticipantsDamage{Header: parseHeader(b), Visibility: b[14]}
	for i := 0; i < ParticipantsMax; i++ {
		p.Rows[i] = parseDamage(b, 15+i*StrideDamage)
	}
	return p, true
}
