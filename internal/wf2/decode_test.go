package wf2

import (
	"encoding/binary"
	"wreckscope/internal/racefeed"
	"math"
	"testing"
	"time"
)

func putf(b []byte, o int, x float32) { binary.LittleEndian.PutUint32(b[o:], math.Float32bits(x)) }
func TestMotionAndLeaderboard(t *testing.T) {
	e := racefeed.New()
	m := make([]byte, 15+36*38)
	copy(m, "pkro")
	m[4] = 4
	m[5] = 32
	putf(m, 15, 123)
	putf(m, 19, 4)
	putf(m, 23, -22)
	putf(m, 15+34, 12)
	if !Decode(m, e) {
		t.Fatal("motion decode failed")
	}
	lb := make([]byte, 1231)
	copy(lb, "pkro")
	lb[4] = 1
	lb[5] = 32
	lb[15] = 2
	lb[15+4] = 3
	lb[15+5] = 79
	if !Decode(lb, e) {
		t.Fatal("leaderboard decode failed")
	}
	info := make([]byte, 8195)
	copy(info, "pkro")
	info[4] = 5
	info[5] = 32
	copy(info[15+160:], []byte("Test Racer"))
	if !Decode(info, e) {
		t.Fatal("info decode failed")
	}
	cars := e.Snapshot().Cars
	if len(cars) != 1 {
		t.Fatalf("got %d cars", len(cars))
	}
	if cars[0].X != 123 || cars[0].Z != -22 || cars[0].Speed < 43 || cars[0].Place != 3 || cars[0].Name != "Test Racer" {
		t.Fatalf("bad car: %+v", cars[0])
	}
}
func TestInvalidPacketDoesNotChangeRaceState(t *testing.T) {
	e := racefeed.New()
	for _, tc := range []struct {
		header string
		kind   byte
		size   int
	}{
		{"pkro", 4, 50},
		{"pkro", 4, 1384},
		{"nope", 4, 1383},
		{"pkro", 7, 1383},
		{"pkro", 5, 15 + 36*213},
	} {
		b := make([]byte, tc.size)
		copy(b, tc.header)
		b[4], b[5] = tc.kind, 32
		if Decode(b, e) {
			t.Fatalf("accepted invalid packet: %+v", tc)
		}
	}
	if snapshot := e.Snapshot(); snapshot.PacketCount != 0 || snapshot.SessionRacing || snapshot.LastPacket != "" {
		t.Fatalf("invalid packets updated diagnostics: %+v", snapshot)
	}
}

func TestMainDriverMustMatchRoster(t *testing.T) {
	e := racefeed.New()
	main := make([]byte, 1218)
	copy(main, "pkro")
	copy(main[276:300], "Local driver")
	motion := make([]byte, 1383)
	copy(motion, "pkro")
	motion[4] = 4
	putf(motion, 15, 10)
	info := make([]byte, 8195)
	copy(info, "pkro")
	info[4] = 5
	copy(info[175:199], "Another driver")
	Decode(main, e)
	Decode(motion, e)
	Decode(info, e)
	if got := e.Snapshot().LocalSlot; got != nil {
		t.Fatalf("mismatched driver identified as local: %d", *got)
	}
	clear(info[175:199])
	copy(info[175:199], "Local driver")
	Decode(info, e)
	if got := e.Snapshot().LocalSlot; got == nil || *got != 0 {
		t.Fatal("local driver in slot 0 not identified")
	}
}

func TestRaceDetailsAndStandingsWithoutMotion(t *testing.T) {
	now := time.Unix(100, 0)
	e := racefeed.NewWithClock(func() time.Time { return now })
	main := make([]byte, 1218)
	copy(main, "pkro")
	main[5] = 32
	binary.LittleEndian.PutUint32(main[10:], 12345)
	copy(main[882:946], "track02_1_rev")
	copy(main[946:1042], "Savolax Sandpit")
	putf(main, 1042, 2261.2)
	binary.LittleEndian.PutUint16(main[1046:], 5)
	main[1050], main[1051], main[1062] = 7, 6, 1
	board := make([]byte, 1231)
	copy(board, "pkro")
	board[4] = 1
	board[15], board[19], board[20] = 2, 1, 87
	binary.LittleEndian.PutUint32(board[35:], 1200)
	timing := make([]byte, 1231)
	copy(timing, "pkro")
	timing[4] = 2
	binary.LittleEndian.PutUint32(timing[15:], 10000)
	binary.LittleEndian.PutUint32(timing[27:], 60000)
	for _, packet := range [][]byte{main, board, timing} {
		binary.LittleEndian.PutUint32(packet[10:], 12345)
		if !Decode(packet, e) {
			t.Fatal("valid packet rejected")
		}
	}
	s := e.Snapshot()
	if s.Race == nil || s.Race.TrackID != "track02_1_rev" || s.Race.Clock != 12345 || s.Race.Grid != 7 || s.Race.Laps != 5 {
		t.Fatalf("unexpected race: %+v", s.Race)
	}
	if len(s.Cars) != 0 || len(s.Standings) != 1 {
		t.Fatalf("standings missing without motion: %+v", s)
	}
	row := s.Standings[0]
	if row.Health != 87 || row.Gap == nil || *row.Gap != 1200 || row.Current == nil || *row.Current != 10000 || row.Best == nil || *row.Best != 60000 || row.Last != nil {
		t.Fatalf("unexpected standing: %+v", row)
	}
	now = now.Add(3 * time.Second)
	s = e.Snapshot()
	if s.Race != nil || len(s.Standings) != 0 {
		t.Fatal("stale race data is still visible")
	}
}
