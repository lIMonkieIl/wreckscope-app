package wf2

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"wreckscope/internal/racefeed"
)

// TestRaceCapture replays a .jsonl UDP capture (set WF2_RACE_CAPTURE). No wall-clock sleep.
func TestRaceCapture(t *testing.T) {
	path := os.Getenv("WF2_RACE_CAPTURE")
	if path == "" {
		t.Skip("set WF2_RACE_CAPTURE to a race capture recording")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(io.LimitReader(f, info.Size()))
	sc.Buffer(make([]byte, 4096), 256*1024)
	var first, now time.Time
	engine := racefeed.NewWithClock(func() time.Time { return now })
	alerts := make(map[string]bool)
	var gridAlerts, lateAlerts int
	var crashSeen bool
	for sc.Scan() {
		var row Capture
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		payload, err := base64.StdEncoding.DecodeString(row.UDP)
		if err != nil {
			t.Fatal(err)
		}
		now = time.Unix(0, row.At)
		if first.IsZero() {
			first = now
		}
		if !Decode(payload, engine) || payload[4] != 4 {
			continue
		}
		elapsed := now.Sub(first)
		for _, car := range engine.Snapshot().Cars {
			if car.State != "alert" || alerts[car.IncidentID] {
				continue
			}
			alerts[car.IncidentID] = true
			if elapsed >= 180*time.Second && elapsed < 186*time.Second {
				gridAlerts++
			}
			if elapsed >= 451*time.Second {
				lateAlerts++
			}
			if car.Slot == 23 && elapsed >= 199*time.Second && elapsed <= 202*time.Second &&
				strings.Contains(car.Reason, "Recent damage") && car.Tilt > 80 {
				crashSeen = true
				t.Logf("crash-like incident at %.3fs, tilt %.1f degrees, health %d: %s", elapsed.Seconds(), car.Tilt, car.Health, car.Reason)
			}
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("distinct alerts=%d grid=%d late=%d", len(alerts), gridAlerts, lateAlerts)
	if gridAlerts != 0 || lateAlerts != 0 || !crashSeen {
		t.Fatalf("unexpected replay: grid=%d late=%d crash seen=%t", gridAlerts, lateAlerts, crashSeen)
	}
}
