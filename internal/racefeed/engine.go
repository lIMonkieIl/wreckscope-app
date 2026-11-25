// Package racefeed turns recovery detections into dashboard snapshots.
package racefeed

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	detector "wreckscope/internal/recovery"
)

type Car = detector.Car

// Settings: UI + detection (AppData when installed; beside exe when portable).
// TelemetryConfigPath is the game config.json path (game file stays separate).
// SetupCompleted is first-run only, not "telemetry already enabled".
type Settings struct {
	DrawTrackOutline bool `json:"drawTrackOutline"`
	detector.Settings
	SpeedUnit           string `json:"speedUnit"`
	MuteAlertSound      bool   `json:"muteAlertSound"`
	Theme               string `json:"theme"`
	Mode                string `json:"mode"` // system | light | dark
	View                string `json:"view"` // race | recovery
	TelemetryConfigPath string `json:"telemetryConfigPath,omitempty"`
	SetupCompleted      bool   `json:"setupCompleted"`
}

type Snapshot struct {
	Race          *RaceDetails     `json:"race,omitempty"`
	Player        *PlayerTelemetry `json:"player,omitempty"`
	Standings     []Standing       `json:"standings"`
	Mode          string           `json:"mode"`
	Connection    string           `json:"connection"`
	LastPacket    string           `json:"lastPacket"`
	PacketCount   uint64           `json:"packetCount"`
	PacketTypes   map[int]uint64   `json:"packetTypes"`
	Cars          []Car            `json:"cars"`
	Settings      Settings         `json:"settings"`
	SessionRacing bool             `json:"sessionRacing"`
	RaceEpoch     uint64           `json:"raceEpoch"`
	LocalSlot     *int             `json:"localSlot,omitempty"`
	Medics        []int            `json:"medics"` // player slots assigned as medics (no cap)
	Visibility    *byte            `json:"participantVisibility,omitempty"`
	Message       string           `json:"message"`
}
type Engine struct {
	raceDetails *RaceDetails
	standings   [36]Standing
	player         *PlayerTelemetry
	playerAt       time.Time
	visibility     byte
	hasVisibility  bool
	clock          func() time.Time
	mu          sync.Mutex
	detector    *detector.Engine
	settings    Settings
	packetCount uint64
	packetTypes map[int]uint64
	lastPacket  time.Time
	mainAt      time.Time
	clockAt     time.Time
	sessionMS   int32
	raceMS      int32
	raceEpoch   uint64
	localSlot   int
	localName   string
	racing      bool
	mode        string
	message     string
	medics      map[int]struct{}
}

func New() *Engine {
	return NewWithClock(time.Now)
}

func defaultSettings() Settings {
	return Settings{
		Settings:  detector.Settings{CautionSeconds: 1.0, AlertSeconds: 3.5, SpeedThreshold: 3.0},
		SpeedUnit: "km/h",
		Theme:     "wreckscope",
		Mode:      "dark",
		View:      "recovery",
	}
}

// NewWithClock: for capture replay (clock from packet timestamps).
func NewWithClock(clock func() time.Time) *Engine {
	return &Engine{
		clock:       clock,
		detector:    detector.New(),
		packetTypes: map[int]uint64{},
		settings:    defaultSettings(),
		mode:        "live",
		medics:      map[int]struct{}{},
	}
}

func (e *Engine) medicSlotsLocked() []int {
	if len(e.medics) == 0 {
		return []int{}
	}
	slots := make([]int, 0, len(e.medics))
	for slot := range e.medics {
		slots = append(slots, slot)
	}
	sort.Ints(slots)
	return slots
}

func (e *Engine) clearMedicsLocked() {
	if len(e.medics) == 0 {
		return
	}
	clear(e.medics)
}

// ToggleMedic assigns or clears a medic. No cap; bad slots ignored.
func (e *Engine) ToggleMedic(slot int) []int {
	e.mu.Lock()
	defer e.mu.Unlock()
	if slot < 0 || slot >= 36 {
		return e.medicSlotsLocked()
	}
	if e.medics == nil {
		e.medics = map[int]struct{}{}
	}
	if _, marked := e.medics[slot]; marked {
		delete(e.medics, slot)
	} else {
		e.medics[slot] = struct{}{}
	}
	return e.medicSlotsLocked()
}

func (e *Engine) Medics() []int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.medicSlotsLocked()
}

func normalizeAppearance(theme, mode, fallbackTheme, fallbackMode string) (string, string) {
	if theme == "" {
		theme = fallbackTheme
	}
	if theme == "" {
		theme = "wreckscope"
	}
	switch mode {
	case "system", "light", "dark":
	default:
		mode = fallbackMode
	}
	if mode == "" {
		mode = "dark"
	}
	return theme, mode
}

func normalizeView(view, fallback string) string {
	switch view {
	case "race", "recovery":
		return view
	}
	if fallback == "race" || fallback == "recovery" {
		return fallback
	}
	return "recovery"
}

// ApplySettings updates settings only after persist succeeds (when set).
func (e *Engine) ApplySettings(s Settings, persist func(Settings) error) (Settings, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if s.SpeedUnit != "mph" && s.SpeedUnit != "km/h" {
		s.SpeedUnit = e.settings.SpeedUnit
	}
	s.Theme, s.Mode = normalizeAppearance(s.Theme, s.Mode, e.settings.Theme, e.settings.Mode)
	s.View = normalizeView(s.View, e.settings.View)
	s.Settings = e.detector.Settings(s.Settings)
	if persist != nil {
		if err := persist(s); err != nil {
			e.detector.Settings(e.settings.Settings)
			return e.settings, err
		}
	}
	e.settings = s
	return e.settings, nil
}

func (e *Engine) Settings() Settings {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.settings
}

func (e *Engine) UpdateMotion(slot int, x, y, z, speed float64) {
	e.detector.UpdateMotion(e.clock(), slot, x, y, z, speed)
}

func (e *Engine) UpdateBoard(slot, status, place, lap, health int, gap int32, stats *BoardStats) {
	e.UpdateBoardRow(slot, status, place, lap, health, gap, stats, nil)
}

func (e *Engine) UpdateBoardRow(slot, status, place, lap, health int, gap int32, stats *BoardStats, full *ParticipantLeaderboard) {
	e.updateStanding(slot, status, place, lap, health, gap, stats, full)
	e.detector.UpdateBoard(e.clock(), slot, status, place, lap, health)
}

func (e *Engine) UpdateExtents(slot int, extents [3]uint16) {
	e.detector.UpdateExtents(slot, extents)
}

func (e *Engine) UpdateOrientation(slot int, x, y, z, w float64) {
	e.detector.UpdateOrientation(e.clock(), slot, x, y, z, w)
}

func (e *Engine) UpdateInfo(slot int, name, carID, carName string, bot bool) {
	e.detector.UpdateInfo(slot, name, carID, bot)
	e.mu.Lock()
	row := &e.standings[slot]
	row.Name, row.CarID, row.CarName, row.bot = name, carID, carName, bot
	e.mu.Unlock()
}

func (e *Engine) UpdateLocalDriver(slot int, name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if slot >= 36 || name == "" {
		return
	}
	e.localSlot, e.localName = slot, name
}

// MarkPacket advances clocks and only admits participant data for the same race.
func (e *Engine) MarkPacket(typ int, racing bool, sessionMS, raceMS int32) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.packetCount++
	e.packetTypes[typ]++
	now := e.clock()
	e.lastPacket = now
	if typ == 0 {
		if !e.mainAt.IsZero() && (int64(sessionMS)+250 < int64(e.sessionMS) || int64(raceMS)+1000 < int64(e.raceMS)) {
			e.detector.Reset()
			e.raceDetails = nil
			e.standings = [36]Standing{}
			e.player = nil
			e.playerAt = time.Time{}
			e.raceEpoch++
			e.localName = ""
			e.clearMedicsLocked()
		}
		if e.mainAt.IsZero() || raceMS != e.raceMS {
			e.clockAt = now
		}
		e.mainAt, e.sessionMS, e.raceMS = now, sessionMS, raceMS
		e.racing = racing && raceMS > 0
		return true
	}
	if e.mainAt.IsZero() {
		return true // Participant data can arrive before the first main packet.
	}
	maxAge := 750 * time.Millisecond
	if typ == 5 {
		maxAge = 2500 * time.Millisecond
	}
	return now.Sub(e.mainAt) <= maxAge &&
		math.Abs(float64(sessionMS)-float64(e.sessionMS)) <= float64(maxAge.Milliseconds()) &&
		math.Abs(float64(raceMS)-float64(e.raceMS)) <= float64(maxAge.Milliseconds())
}
func (e *Engine) Mode() string { e.mu.Lock(); defer e.mu.Unlock(); return e.mode }
func (e *Engine) SetMode(m string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mode = m
	e.detector.Reset()
	e.raceDetails = nil
	e.standings = [36]Standing{}
	e.player = nil
	e.playerAt = time.Time{}
	e.visibility = 0
	e.hasVisibility = false
	e.packetCount = 0
	e.packetTypes = map[int]uint64{}
	e.lastPacket = time.Time{}
	e.mainAt = time.Time{}
	e.clockAt = time.Time{}
	e.racing = false
	e.raceEpoch++
	e.localName = ""
	e.clearMedicsLocked()
}
func (e *Engine) Snapshot() Snapshot {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.clock()
	racing := e.mode == "demo" || (e.racing && now.Sub(e.clockAt) <= 750*time.Millisecond)
	lastPacket := ""
	if !e.lastPacket.IsZero() {
		lastPacket = e.lastPacket.UTC().Format(time.RFC3339Nano)
	}
	s := Snapshot{Mode: e.mode, PacketCount: e.packetCount, PacketTypes: map[int]uint64{}, LastPacket: lastPacket, Settings: e.settings, SessionRacing: racing, Cars: []Car{}, Medics: e.medicSlotsLocked(), Message: e.message}
	for k, v := range e.packetTypes {
		s.PacketTypes[k] = v
	}
	if e.mode == "demo" {
		s.Connection = "demo"
	} else if e.lastPacket.IsZero() {
		s.Connection = "waiting"
	} else if now.Sub(e.lastPacket) > 3*time.Second {
		s.Connection = "stale"
	} else {
		s.Connection = "live"
	}
	s.Cars = e.detector.Evaluate(now, racing)
	if now.Sub(e.mainAt) <= 2*time.Second {
		s.Race = e.raceDetails
	}
	s.Standings = e.currentStandings(now)
	s.RaceEpoch = e.raceEpoch
	if e.player != nil && now.Sub(e.playerAt) <= 2*time.Second {
		cp := *e.player
		s.Player = &cp
	}
	if e.hasVisibility {
		v := e.visibility
		s.Visibility = &v
	}
	if e.mode == "live" && e.localName != "" && now.Sub(e.mainAt) <= 750*time.Millisecond {
		for _, car := range s.Cars {
			if car.Slot == e.localSlot && car.Name == e.localName && !car.IsBot {
				slot := car.Slot
				s.LocalSlot = &slot
				break
			}
		}
	}
	return s
}
func (e *Engine) DemoTick(t float64) {
	e.mu.Lock()
	e.racing = true
	e.lastPacket = e.clock()
	e.mainAt = e.lastPacket
	e.raceDetails = &RaceDetails{
		TrackID: "track07_5", Track: "Scrapyard Figure 8",
		Grid: 13, Remaining: 13, Mode: 0, Damage: 1, Clock: int32(t * 1000),
	}
	e.packetCount++
	e.packetTypes[4]++
	e.mu.Unlock()
	for i := 0; i < 13; i++ {
		phase := t*.14 + float64(i)*.46
		speed := 24.0
		if i == 4 && t > 8 {
			phase = 8*.14 + float64(i)*.46
			speed = 0
		}
		if i == 9 && math.Mod(t, 48) > 20 && math.Mod(t, 48) < 36 {
			speed = 0
			phase = (t-math.Mod(t, 48)+20)*.14 + float64(i)*.46
		}
		x := 220*math.Cos(phase) + 36*math.Cos(2*phase)
		z := 125 * math.Sin(phase)
		e.UpdateMotion(i, x, 0, z, speed/3.6)
		if i == 4 && t > 8 {
			e.UpdateOrientation(i, 1, 0, 0, 0)
		} else {
			e.UpdateOrientation(i, 0, 0, 0, 1)
		}
		e.UpdateInfo(i, []string{"Recovery Driver", "SmashKing", "RoadRunner", "BangerBoy", "LostWheel", "Redline", "KartMan", "Bishop", "Sideways", "TowNeeded", "Wrecker", "CrashTest", "HillClimber"}[i], fmt.Sprintf("car%02d", i+1), "", false)
		e.UpdateBoard(i, 2, i+1, int(t/40)+1, 90, -1, nil)
	}
}
