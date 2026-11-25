package recovery

import (
	"testing"
	"time"
)

func TestClassifyQuickWarningWithDamage(t *testing.T) {
	now := time.Unix(100, 0)
	p := &tracked{hasDriven: true, Car: Car{Speed: 0, Status: 2, Health: 100, Updated: now, X: 10, Z: 20}, lastMove: now.Add(-900 * time.Millisecond), lastDamage: now.Add(-200 * time.Millisecond), samples: []sample{{at: now.Add(-1500 * time.Millisecond), x: 10, z: 20}, {at: now, x: 10, z: 20}}}
	state, _, _, _, score := classify(p, now, Settings{CautionSeconds: 2, AlertSeconds: 5, SpeedThreshold: 3}, true)
	if state != "caution" || score < 45 {
		t.Fatalf("wanted caution for damaged nonmoving car; state=%s score=%d", state, score)
	}
}
func TestClassifyExcludedAndStale(t *testing.T) {
	now := time.Unix(100, 0)
	p := &tracked{hasDriven: true, Car: Car{Speed: 0, Status: 8, Updated: now}, lastMove: now.Add(-10 * time.Second), samples: []sample{{at: now, x: 0, z: 0}}}
	state, _, _, _, _ := classify(p, now, Settings{1, 3, 3, false}, true)
	if state != "moving" {
		t.Fatal(state)
	}
}
func TestClassifyAlertAfterThreshold(t *testing.T) {
	now := time.Unix(100, 0)
	p := &tracked{hasDriven: true, Car: Car{Speed: 0, Status: 2, Health: 100, Updated: now, X: 2, Z: 2}, lastMove: now.Add(-4 * time.Second), samples: []sample{{at: now.Add(-2 * time.Second), x: 2, z: 2}, {at: now, x: 2, z: 2}}}
	state, _, _, _, _ := classify(p, now, Settings{1, 3.5, 3, false}, true)
	if state != "alert" {
		t.Fatal(state)
	}
}

func TestInvertedCarTriggersEarly(t *testing.T) {
	e := New()
	now := time.Unix(100, 0)
	e.UpdateMotion(now, 0, 5, 0, 5, 0)
	e.UpdateOrientation(now, 0, 1, 0, 0, 0) // 180 degree roll
	e.mu.Lock()
	p := e.cars[0]
	p.invertedSince = now.Add(-900 * time.Millisecond)
	p.lastMove = now.Add(-900 * time.Millisecond)
	p.samples = []sample{{at: now.Add(-1 * time.Second), x: 5, z: 5}, {at: now, x: 5, z: 5}}
	p.Status = 2
	state, _, reason, _, _ := classify(p, now, e.settings, true)
	e.mu.Unlock()
	if state != "alert" || reason == "" {
		t.Fatalf("wanted early inverted alert, got %s %q", state, reason)
	}
}
func TestAlertDoesNotFlickerAfterBriefMovement(t *testing.T) {
	now := time.Unix(100, 0)
	p := &tracked{hasDriven: true, Car: Car{Speed: 0, Status: 2, Health: 100, Updated: now, X: 2, Z: 2}, lastMove: now.Add(-5 * time.Second), samples: []sample{{at: now.Add(-3 * time.Second), x: 2, z: 2}, {at: now, x: 2, z: 2}}}
	settings := Settings{1, 3.5, 3, false}
	s, _, _, _, _ := classifyPersistent(p, now, settings, true)
	if s != "alert" {
		t.Fatal(s)
	}
	id := p.incidentID
	p.lastMove = now
	p.Speed = 90
	p.samples = []sample{{at: now, x: 2, z: 2}, {at: now.Add(100 * time.Millisecond), x: 15, z: 15}}
	p.X = 15
	p.Z = 15
	p.Updated = now.Add(100 * time.Millisecond)
	s, _, _, _, _ = classifyPersistent(p, now.Add(100*time.Millisecond), settings, true)
	if s != "alert" || p.incidentID != id {
		t.Fatalf("alert flickered: %s %s", s, p.incidentID)
	}
	for step := 2; step <= 24; step++ {
		at := now.Add(time.Duration(step) * 100 * time.Millisecond)
		p.Updated = at
		p.lastMove = at
		p.X += 1
		p.samples = append(p.samples, sample{at: at, x: p.X, z: p.Z})
		s, _, _, _, _ = classifyPersistent(p, at, settings, true)
	}

	if s != "moving" || p.incidentID != "" {
		t.Fatalf("wanted resolved; got %s", s)
	}
}

func TestInactiveCarsCannotAlert(t *testing.T) {
	now := time.Unix(100, 0)
	for _, car := range []Car{{Status: 3, Health: 100}, {Status: 4, Health: 100}, {Status: 7, Health: 100}, {Status: 8, Health: 100}, {Status: 2, Health: 0}} {
		car.Updated = now
		p := &tracked{Car: car, hasDriven: true, lastMove: now.Add(-10 * time.Second), samples: []sample{{at: now}}}
		if state, _, _, _, _ := classify(p, now, Settings{1, 3.5, 3, false}, true); state == "alert" || state == "caution" {
			t.Errorf("inactive car with status %d and health %d produced %s", car.Status, car.Health, state)
		}
	}
}

func TestAlertStaysActiveWhileCarIsStillSlow(t *testing.T) {
	now := time.Unix(100, 0)
	p := &tracked{hasDriven: true, Car: Car{Status: 2, Health: 100, Updated: now}, lastMove: now.Add(-5 * time.Second), samples: []sample{{at: now}}}
	settings := Settings{1, 3.5, 3, false}
	state, _, _, _, _ := classifyPersistent(p, now, settings, true)
	if state != "alert" {
		t.Fatal(state)
	}
	id := p.incidentID
	for i := 1; i <= 30; i++ {
		at := now.Add(time.Duration(i) * 100 * time.Millisecond)
		p.Updated = at
		p.X += .1
		p.Speed = 1
		p.samples = append(p.samples, sample{at: at, x: p.X})
		state, _, _, _, _ = classifyPersistent(p, at, settings, true)
		if state != "alert" || p.incidentID != id {
			t.Fatalf("incident cleared during slow movement at step %d", i)
		}
	}
}
