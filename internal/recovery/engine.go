// Package recovery classifies stuck/incident cars from timed observations.
package recovery

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ScoreFactor struct {
	Label  string `json:"label"`
	Points int    `json:"points"`
}

type Car struct {
	Extents       *[3]uint16    `json:"extents,omitempty"`
	Orientation   *[4]float64   `json:"orientation,omitempty"`
	ScoreFactors  []ScoreFactor `json:"scoreFactors,omitempty"`
	Slot          int           `json:"slot"`
	Name          string        `json:"name"`
	CarID         string        `json:"carId,omitempty"`
	X             float64       `json:"x"`
	Y             float64       `json:"y"`
	Z             float64       `json:"z"`
	Speed         float64       `json:"speed"`
	Place         int           `json:"place"`
	Lap           int           `json:"lap"`
	Health        int           `json:"health"`
	State         string        `json:"state"`
	Stopped       float64       `json:"stopped"`
	Reason        string        `json:"reason"`
	Confidence    string        `json:"confidence"`
	Score         int           `json:"score"`
	Updated       time.Time     `json:"updated"`
	Active        bool          `json:"active"`
	IsBot         bool          `json:"isBot"`
	Status        int           `json:"status"`
	Tilt          float64       `json:"tilt"`
	Inverted      bool          `json:"inverted"`
	IncidentID    string        `json:"incidentId,omitempty"`
	IncidentSince time.Time     `json:"incidentSince,omitempty"`
}
type Settings struct {
	CautionSeconds float64 `json:"cautionSeconds"`
	AlertSeconds   float64 `json:"alertSeconds"`
	SpeedThreshold float64 `json:"speedThreshold"`
	HideBots       bool    `json:"hideBots"`
}
type sample struct {
	at   time.Time
	x, z float64
}
type tracked struct {
	Car
	samples         []sample
	slowSince       time.Time
	lastPosition    sample
	lastMove        time.Time
	hasDriven       bool
	lastDamage      time.Time
	previousHealth  int
	lastBoard       time.Time
	lastSpeed       float64
	crashAt         time.Time
	tiltSince       time.Time
	invertedSince   time.Time
	candidateSince  time.Time
	incidentSince   time.Time
	incidentID      string
	persistentState string
	clearSince      time.Time
	clearPosition   sample
	incidentReason  string
	speedHistory    []float64
}

// Engine tracks one race. New engine per session, or Reset. Times must be chronological.
type Engine struct {
	mu       sync.Mutex
	cars     map[int]*tracked
	settings Settings
}

func New() *Engine {
	return &Engine{cars: map[int]*tracked{}, settings: Settings{CautionSeconds: 1, AlertSeconds: 3.5, SpeedThreshold: 3}}
}

func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	clear(e.cars)
}

func (e *Engine) Settings(s Settings) Settings {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !finite(s.CautionSeconds) || s.CautionSeconds < .5 {
		s.CautionSeconds = .5
	}
	if s.CautionSeconds > 60 {
		s.CautionSeconds = 60
	}
	if !finite(s.AlertSeconds) || s.AlertSeconds < s.CautionSeconds+.5 {
		s.AlertSeconds = s.CautionSeconds + .5
	}
	if s.AlertSeconds > 180 {
		s.AlertSeconds = 180
	}
	if !finite(s.SpeedThreshold) || s.SpeedThreshold < .5 {
		s.SpeedThreshold = .5
	}
	if s.SpeedThreshold > 20 {
		s.SpeedThreshold = 20
	}
	e.settings = s
	return s
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (e *Engine) UpdateMotion(now time.Time, slot int, x, y, z, speed float64) {
	if !finite(x) || !finite(y) || !finite(z) || !finite(speed) || math.Abs(x) > 1e7 || math.Abs(z) > 1e7 {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	p := e.get(slot)
	if p.Updated.IsZero() {
		p.lastMove = now
		p.lastPosition = sample{now, x, z}
	}
	if len(p.samples) > 0 {
		prev := p.samples[len(p.samples)-1]
		// Teleport / reset / big jump: not driving progress.
		jump := math.Hypot(x-prev.x, z-prev.z)
		if jump > 100 {
			p.samples = nil
			p.lastMove = now
			p.lastPosition = sample{now, x, z}
		} else if math.Hypot(x-p.lastPosition.x, z-p.lastPosition.z) >= .65 {
			// Small steps accumulate so packet rate isn't the move gate.
			p.lastMove = now
			p.lastPosition = sample{now, x, z}
			if speed*3.6 >= e.settings.SpeedThreshold {
				p.hasDriven = true
			}
		}
	}
	// Short rolling window; drops jitter noise.
	p.samples = append(p.samples, sample{now, x, z})
	cutoff := now.Add(-10 * time.Second)
	for len(p.samples) > 2 && p.samples[1].at.Before(cutoff) {
		p.samples = p.samples[1:]
	}
	p.X = x
	p.Y = y
	p.Z = z
	nextSpeed := math.Max(0, speed*3.6)
	if !p.Updated.IsZero() && now.Sub(p.Updated) < 500*time.Millisecond && p.Speed-nextSpeed >= 45 && p.Speed >= 55 {
		p.crashAt = now
	}
	p.lastSpeed = p.Speed
	p.Speed = nextSpeed
	if p.Speed < e.settings.SpeedThreshold {
		if p.slowSince.IsZero() {
			p.slowSince = now
		}
	} else {
		p.slowSince = time.Time{}
	}
	p.Updated = now
	p.Active = true
}
func (e *Engine) UpdateInfo(slot int, name, carID string, bot bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	p := e.get(slot)
	if name != "" {
		p.Name = name
	}
	p.CarID = carID
	p.IsBot = bot
}
func (e *Engine) UpdateBoard(now time.Time, slot, status, place, lap, health int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	p := e.get(slot)
	if !p.lastBoard.IsZero() && health >= 0 && health < p.previousHealth-5 {
		p.lastDamage = now
	}
	p.previousHealth = health
	p.lastBoard = now
	p.Status = status
	p.Place = place
	p.Lap = lap
	p.Health = health
}
func (e *Engine) get(slot int) *tracked {
	p := e.cars[slot]
	if p == nil {
		p = &tracked{Car: Car{Slot: slot, Name: "Car #" + strconv.Itoa(slot+1), Health: 100}}
		e.cars[slot] = p
	}
	return p
}

func (e *Engine) UpdateExtents(slot int, extents [3]uint16) {
	if extents[0] == 0 || extents[1] == 0 || extents[2] == 0 {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	p := e.get(slot)
	if p.Extents == nil || *p.Extents != extents {
		p.Extents = &extents
	}
}

// UpdateOrientation: quaternion (x,y,z,w). Bad values ignored (no false alerts).
func (e *Engine) UpdateOrientation(now time.Time, slot int, qx, qy, qz, qw float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	p := e.get(slot)
	n := math.Sqrt(qx*qx + qy*qy + qz*qz + qw*qw)
	if !finite(n) || n < .5 || n > 1.5 {
		return
	}
	qx /= n
	qy /= n
	qz /= n
	qw /= n
	p.Orientation = &[4]float64{qx, qy, qz, qw}
	// Up component of R*(0,1,0), independent of yaw.
	up := 1 - 2*(qx*qx+qz*qz)
	up = math.Max(-1, math.Min(1, up))
	p.Tilt = math.Acos(up) * 180 / math.Pi
	p.Inverted = up < -.35
	if p.Inverted {
		if p.invertedSince.IsZero() {
			p.invertedSince = now
		}
	} else {
		p.invertedSince = time.Time{}
	}
	if up < .40 {
		if p.tiltSince.IsZero() {
			p.tiltSince = now
		}
	} else {
		p.tiltSince = time.Time{}
	}
}

// classify: position + recent crash/damage + orientation. Speed alone isn't an emergency.
func classify(p *tracked, now time.Time, s Settings, racing bool) (string, float64, string, string, int) {
	p.ScoreFactors = nil
	if !racing || p.Status != 2 || p.Health == 0 {
		return "moving", 0, "", "", 0
	}
	if p.lastMove.IsZero() || p.Updated.IsZero() || now.Sub(p.Updated) > 800*time.Millisecond || len(p.samples) == 0 {
		return "moving", 0, "", "", 0
	}
	age := now.Sub(p.lastMove).Seconds()
	if age < 0 {
		age = 0
	}
	base := p.samples[0]
	cutoff := now.Add(-2 * time.Second)
	for _, pt := range p.samples {
		if pt.at.After(cutoff) {
			break
		}
		base = pt
	}
	displacement := math.Hypot(p.X-base.x, p.Z-base.z)
	slow := p.Speed < s.SpeedThreshold
	noProgress := age >= .35 && displacement < 1.5
	invertedSettled := p.Inverted && !p.invertedSince.IsZero() && now.Sub(p.invertedSince) > 650*time.Millisecond && p.Speed < 25
	tiltedSettled := !p.tiltSince.IsZero() && now.Sub(p.tiltSince) > 1200*time.Millisecond && age > .8
	crashRecent := !p.crashAt.IsZero() && now.Sub(p.crashAt) < 6*time.Second
	damageRecent := !p.lastDamage.IsZero() && now.Sub(p.lastDamage) < 7*time.Second
	if !p.hasDriven && !invertedSettled && !(crashRecent && damageRecent) {
		return "moving", 0, "", "", 0
	}
	// Speed with almost no displacement (wheelspin / collision / telemetry fight).
	motionMismatch := noProgress && p.Speed >= s.SpeedThreshold+5
	if !noProgress && !invertedSettled && !tiltedSettled && !slow {
		return "moving", age, "", "", 0
	}
	score := 0
	reasons := []string{}
	add := func(points int, label string) {
		score += points
		p.ScoreFactors = append(p.ScoreFactors, ScoreFactor{Label: label, Points: points})
	}
	if noProgress {
		add(30, "No meaningful movement")
		reasons = append(reasons, "No meaningful movement")
	}
	if slow {
		add(10, "Low speed")
	}
	if age > 1 {
		add(10, "Stopped over 1 second")
	}
	if age > 2.5 {
		add(15, "Stopped over 2.5 seconds")
	}
	if age > 5 {
		add(10, "Stopped over 5 seconds")
	}
	if crashRecent {
		add(15, "Recent sudden deceleration")
		reasons = append(reasons, "Recent sudden deceleration")
	}
	if damageRecent {
		add(15, "Recent damage")
		reasons = append(reasons, "Recent damage")
	}
	if motionMismatch {
		add(20, "Speed without forward displacement")
		reasons = append(reasons, "Speed without forward displacement")
	}
	if tiltedSettled {
		add(20, "Unusual settled orientation")
		reasons = append(reasons, "Unusual settled orientation")
	}
	if invertedSettled {
		add(70, "Vehicle overturned")
		reasons = append(reasons, "Vehicle overturned")
	}
	score = int(math.Min(100, float64(score)))
	confidence := "low"
	if score >= 75 {
		confidence = "high"
	} else if score >= 45 {
		confidence = "medium"
	}
	reason := strings.Join(reasons, "; ")
	if reason == "" {
		reason = "Vehicle slowing"
	}
	if invertedSettled {
		return "alert", age, reason, confidence, score
	}
	if noProgress && age >= s.AlertSeconds && score >= 65 {
		return "alert", age, reason, confidence, score
	}
	if noProgress && (crashRecent || damageRecent || motionMismatch) && age >= 1.5 && score >= 65 {
		return "alert", age, reason, confidence, score
	}
	cautionAt := s.CautionSeconds
	if (crashRecent || damageRecent) && noProgress {
		cautionAt = math.Min(cautionAt, .65)
	}
	if score >= 45 && age >= cautionAt {
		return "caution", age, reason, confidence, score
	}
	if slow || noProgress {
		return "slow", age, reason, confidence, score
	}
	return "moving", age, "", "", 0
}

// classifyPersistent holds an incident through a brief nudge; sustained move clears it.
func classifyPersistent(p *tracked, now time.Time, s Settings, racing bool) (string, float64, string, string, int) {
	state, age, reason, conf, score := classify(p, now, s, racing)
	if !racing || p.Status != 2 || p.Health == 0 || now.Sub(p.Updated) > 800*time.Millisecond {
		p.persistentState = ""
		p.incidentID = ""
		p.incidentSince = time.Time{}
		p.clearSince = time.Time{}
		p.candidateSince = time.Time{}
		return state, age, reason, conf, score
	}
	if state == "alert" || state == "caution" {
		if p.candidateSince.IsZero() {
			p.candidateSince = now
		}
		if p.persistentState == "alert" && state == "caution" {
			state = "alert"
		}
		if p.incidentID == "" {
			p.incidentSince = now
			p.incidentID = fmt.Sprintf("%d-%d", p.Slot, now.UnixNano())
		}
		p.persistentState = state
		p.incidentReason = reason
		p.clearSince = time.Time{}
	} else if p.persistentState == "alert" || p.persistentState == "caution" {
		progressing := state == "moving" && !p.Inverted && p.Speed >= s.SpeedThreshold &&
			now.Sub(p.lastMove) <= 500*time.Millisecond
		if p.persistentState == "alert" && !progressing {
			p.clearSince = time.Time{}
		} else if p.clearSince.IsZero() {
			p.clearSince = now
			p.clearPosition = sample{x: p.X, z: p.Z}
		}
		confirmed := !p.clearSince.IsZero() && now.Sub(p.clearSince) >= 2*time.Second &&
			(p.persistentState == "caution" || math.Hypot(p.X-p.clearPosition.x, p.Z-p.clearPosition.z) >= 2)
		if !confirmed {
			state = p.persistentState
			reason = p.incidentReason
			if !p.clearSince.IsZero() {
				reason = "Checking sustained recovery"
			}
			conf = "medium"
		} else {
			p.persistentState = ""
			p.incidentID = ""
			p.incidentSince = time.Time{}
			p.candidateSince = time.Time{}
			p.clearSince = time.Time{}
		}
	}
	return state, age, reason, conf, score
}

// Evaluate returns an ordered car snapshot. Call on packets and as time advances.
func (e *Engine) Evaluate(now time.Time, racing bool) []Car {
	e.mu.Lock()
	defer e.mu.Unlock()
	cars := []Car{}
	for _, p := range e.cars {
		if p.Updated.IsZero() || now.Sub(p.Updated) > 2*time.Second {
			continue
		}
		if e.settings.HideBots && p.IsBot {
			continue
		}
		c := p.Car
		c.State, c.Stopped, c.Reason, c.Confidence, c.Score = classifyPersistent(p, now, e.settings, racing)
		c.ScoreFactors = p.ScoreFactors
		c.IncidentID = p.incidentID
		c.IncidentSince = p.incidentSince
		cars = append(cars, c)
	}
	sort.Slice(cars, func(i, j int) bool {
		a, b := cars[i], cars[j]
		if a.Place != b.Place {
			if a.Place == 0 {
				return false
			}
			if b.Place == 0 {
				return true
			}
			return a.Place < b.Place
		}
		return a.Slot < b.Slot
	})
	return cars
}
