package wf2

import (
	"encoding/binary"
	"testing"
	"wreckscope/internal/racefeed"
)

func TestPacketSizesMatchLayout(t *testing.T) {
	want := [...]int{1218, 1231, 1231, 1375, 1383, 8195, 1027}
	if PacketSizes != want {
		t.Fatalf("PacketSizes=%v want %v", PacketSizes, want)
	}
	// Derived from header + visibility + strides + reserved tails.
	if 15+ParticipantsMax*StrideLeaderboard+64 != PacketParticipantsLeaderboardSize {
		t.Fatal("leaderboard size mismatch")
	}
	if 15+ParticipantsMax*StrideTiming+64 != PacketParticipantsTimingSize {
		t.Fatal("timing size mismatch")
	}
	if 15+ParticipantsMax*StrideSectors+64 != PacketParticipantsSectorsSize {
		t.Fatal("sectors size mismatch")
	}
	if 15+ParticipantsMax*StrideMotion != PacketParticipantsMotionSize {
		t.Fatal("motion size mismatch")
	}
	if 15+ParticipantsMax*StrideInfo+512 != PacketParticipantsInfoSize {
		t.Fatal("info size mismatch")
	}
	if 15+ParticipantsMax*StrideDamage+256 != PacketParticipantsDamageSize {
		t.Fatal("damage size mismatch")
	}
}

func TestMainPacketOffsets(t *testing.T) {
	if MainMarshalFlags != 14 || MainPlayerBoard != 16 || MainPlayerTiming != 48 ||
		MainPlayerSector != 80 || MainPlayerInfo != 116 || MainPlayerDamage != 329 ||
		MainCarPlayer != 350 || MainSession != 882 || MainPlayerStatus != 1090 || MainInputExt != 1092 {
		t.Fatalf("main offsets drifted")
	}
	if MainSessionEnd-MainSession != 208 {
		t.Fatalf("session block size=%d", MainSessionEnd-MainSession)
	}
	if CarChassisOff != 8 || CarDrivelineOff != 60 || CarEngineOff != 84 ||
		CarInputOff != 140 || CarOrientOff != 160 || CarVelocityOff != 194 || CarTiresOff != 230 {
		t.Fatal("CarFull sub-offsets drifted")
	}
	if CarTiresOff+TireCount*TireSize+14 != 532 {
		t.Fatalf("CarFull size=%d", CarTiresOff+TireCount*TireSize+14)
	}
	if MainCarPlayer+532 != MainSession {
		t.Fatal("carPlayer does not abut session")
	}
}

func TestParseMainCarAndSession(t *testing.T) {
	b := make([]byte, PacketMainSize)
	copy(b, Signature)
	b[4] = 0
	b[5] = 32 // in race
	binary.LittleEndian.PutUint16(b[MainMarshalFlags:], 0x0001) // green
	binary.LittleEndian.PutUint32(b[10:], 5000)

	// Player info name + index
	copy(b[MainPlayerInfo+160:], []byte("Driver"))
	b[MainPlayerInfo+184] = 3

	// Car: gear 4, rpm 4500, throttle 0.5, position
	b[MainCarPlayer+CarDrivelineOff+1] = 4
	binary.LittleEndian.PutUint32(b[MainCarPlayer+CarEngineOff+1:], 4500)
	putf(b, MainCarPlayer+CarInputOff, 0.5)
	putf(b, MainCarPlayer+CarOrientOff, 10)
	putf(b, MainCarPlayer+CarOrientOff+8, -20)

	// Session
	copy(b[MainSession:], []byte("track99"))
	copy(b[MainSession+64:], []byte("Test Track"))
	putf(b, MainSession+160, 1200)
	binary.LittleEndian.PutUint16(b[MainSession+164:], 8)
	b[MainSession+168], b[MainSession+169] = 12, 11
	b[MainSession+170] = 3
	putf(b, MainSession+171, 0.33)
	putf(b, MainSession+175, 0.66)
	b[MainSession+179], b[MainSession+180], b[MainSession+181] = 0, 1, 3

	// player status + ffb
	binary.LittleEndian.PutUint16(b[MainPlayerStatus:], 5)
	b[MainInputExt] = 1
	putf(b, MainInputExt+1, 0.25)

	main, ok := ParseMain(b)
	if !ok {
		t.Fatal("ParseMain failed")
	}
	if main.Info.PlayerName != "Driver" || main.Info.ParticipantIndex != 3 {
		t.Fatalf("info=%+v", main.Info)
	}
	if main.Car.Driveline.Gear != 4 || main.Car.Engine.RPM != 4500 || main.Car.Input.Throttle < 0.49 {
		t.Fatalf("car=%+v", main.Car)
	}
	if main.Car.Orientation.X != 10 || main.Car.Orientation.Z != -20 {
		t.Fatalf("orient=%+v", main.Car.Orientation)
	}
	if main.Session.TrackID != "track99" || main.Session.Laps != 8 || main.Session.SectorCount != 3 ||
		main.Session.Status != 3 || main.Session.DamageMode != 1 {
		t.Fatalf("session=%+v", main.Session)
	}
	if main.PlayerStatus != 5 || main.InputExtended.Flags != 1 || main.InputExtended.FFBForce < 0.24 {
		t.Fatalf("status/ffb=%+v", main)
	}

	e := racefeed.New()
	if !Decode(b, e) {
		t.Fatal("Decode main failed")
	}
	s := e.Snapshot()
	if s.Player == nil || s.Player.CarPlayer.Engine.RPM != 4500 || s.Player.ParticipantPlayerInfo.ParticipantIndex != 3 {
		t.Fatalf("player snapshot=%+v", s.Player)
	}
	if s.Race == nil || s.Race.TrackName != "Test Track" || s.Race.GridSize != 12 || s.Race.SectorCount != 3 {
		t.Fatalf("race=%+v", s.Race)
	}
}

func TestParseTimingSectorsLeaderboardDamage(t *testing.T) {
	// Timing
	tb := make([]byte, PacketParticipantsTimingSize)
	copy(tb, Signature)
	tb[4] = 2
	tb[5] = 32
	binary.LittleEndian.PutUint32(tb[15:], 1111)
	binary.LittleEndian.PutUint32(tb[15+4:], 50) // penalty
	binary.LittleEndian.PutUint32(tb[15+8:], 2222)
	binary.LittleEndian.PutUint32(tb[15+12:], 3333)
	tb[15+16] = 2 // lapBest
	binary.LittleEndian.PutUint32(tb[15+17:], 0xffffff9c) // int32(-100)
	binary.LittleEndian.PutUint32(tb[15+21:], 200)
	putf(tb, 15+25, 0.42)
	pt, ok := ParseParticipantsTiming(tb)
	if !ok || pt.Rows[0].LapTimeCurrent != 1111 || pt.Rows[0].LapTimePenaltyCurrent != 50 ||
		pt.Rows[0].LapBest != 2 || pt.Rows[0].DeltaAhead != -100 || pt.Rows[0].LapProgress < 0.41 {
		t.Fatalf("timing=%+v", pt.Rows[0])
	}

	// Sectors: all nine fields
	sb := make([]byte, PacketParticipantsSectorsSize)
	copy(sb, Signature)
	sb[4] = 3
	sb[5] = 32
	for i := 0; i < 9; i++ {
		binary.LittleEndian.PutUint32(sb[15+i*4:], uint32(1000+i))
	}
	ps, ok := ParseParticipantsSectors(sb)
	if !ok || ps.Rows[0].SectorTimeCurrentLap1 != 1000 || ps.Rows[0].SectorTimeBest3 != 1008 {
		t.Fatalf("sectors=%+v", ps.Rows[0])
	}

	// Leaderboard extras: assists/score/points/lapTiming
	lb := make([]byte, PacketParticipantsLeaderboardSize)
	copy(lb, Signature)
	lb[4] = 1
	lb[5] = 32
	lb[15] = 2
	lb[15+4] = 1
	lb[15+5] = 90
	binary.LittleEndian.PutUint16(lb[15+10:], 7)
	binary.LittleEndian.PutUint32(lb[15+12:], 123)
	binary.LittleEndian.PutUint32(lb[15+16:], 45)
	binary.LittleEndian.PutUint16(lb[15+24:], 9)
	pl, ok := ParseParticipantsLeaderboard(lb)
	if !ok || pl.Rows[0].Assists != 7 || pl.Rows[0].Score != 123 || pl.Rows[0].Points != 45 || pl.Rows[0].LapTiming != 9 {
		t.Fatalf("leaderboard=%+v", pl.Rows[0])
	}

	// Damage packed bits: part 0 = state 3, part 1 = state 2 → first byte 0b010011 = 0x13
	db := make([]byte, PacketParticipantsDamageSize)
	copy(db, Signature)
	db[4] = 6
	db[5] = 32
	db[15] = 0x13
	pd, ok := ParseParticipantsDamage(db)
	if !ok || pd.Rows[0].Parts[0] != 3 || pd.Rows[0].Parts[1] != 2 {
		t.Fatalf("damage parts=%v raw=%v", pd.Rows[0].Parts[:4], pd.Rows[0].Raw[:2])
	}

	e := racefeed.New()
	for _, pkt := range [][]byte{lb, tb, sb, db} {
		binary.LittleEndian.PutUint32(pkt[10:], 1)
		if !Decode(pkt, e) {
			t.Fatal("participant decode failed")
		}
	}
	s := e.Snapshot()
	if len(s.Standings) != 1 {
		t.Fatalf("standings=%d", len(s.Standings))
	}
	row := s.Standings[0]
	if row.Leaderboard == nil || row.Leaderboard.Assists != 7 || row.Leaderboard.LapTiming != 9 {
		t.Fatalf("standing leaderboard=%+v", row.Leaderboard)
	}
	if row.Timing == nil || row.Timing.LapTimePenaltyCurrent != 50 || row.Timing.DeltaBehind != 200 {
		t.Fatalf("standing timing=%+v", row.Timing)
	}
	if row.Sectors == nil || row.Sectors.SectorTimeBest3 != 1008 || row.Sectors.SectorTimeCurrentLap1 != 1000 {
		t.Fatalf("standing sectors=%+v", row.Sectors)
	}
	if row.Damage == nil || row.Damage.Parts[0] != 3 || len(row.Damage.DamageStates) != 21 {
		t.Fatalf("standing damage=%+v", row.Damage)
	}
	if s.Visibility == nil || *s.Visibility != 0 {
		t.Fatalf("visibility=%v", s.Visibility)
	}
}

func TestUnpackDamageSpanningBytes(t *testing.T) {
	// Part 2 starts at bit 6 → spans byte0 and byte1. Set state 5 (terminal) = 0b101 at bits 6-8.
	raw := make([]byte, DamageBytesPerDriver)
	raw[0] = 0b01000000 // bit6 = 1 (low bit of 101... wait we need bits 6,7 of byte0 and bit0 of byte1)
	// bit 6,7,8 for value 5 (101): bit6=1, bit7=0, bit8=1
	raw[0] = (1 << 6) // bit6
	// bit7 stays 0
	raw[1] = 1 // bit8
	d := UnpackDamage(raw)
	if d.Parts[2] != 5 {
		t.Fatalf("part2=%d want 5 (raw %02x %02x)", d.Parts[2], raw[0], raw[1])
	}
}

func TestInfoExtrasAndBot(t *testing.T) {
	b := make([]byte, PacketParticipantsInfoSize)
	copy(b, Signature)
	b[4] = 5
	b[5] = 32
	copy(b[15:], []byte("car:ai_01"))
	copy(b[15+64:], []byte("Botmobile"))
	copy(b[15+160:], []byte("*BOT 1"))
	b[15+184] = 0
	binary.LittleEndian.PutUint32(b[15+189:], 777)
	info, ok := ParseParticipantsInfo(b)
	if !ok || !info.Rows[0].Bot || info.Rows[0].LastCollisionTime != 777 {
		t.Fatalf("info=%+v", info.Rows[0])
	}
}
