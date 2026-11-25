package racefeed

import (
	"math"
	"sort"
	"time"
)

// RaceDetails: Session + PacketMain status. JSON includes UI aliases and Session names.
type RaceDetails struct {
	TrackID           string   `json:"trackId"`
	Track             string   `json:"track"` // UI alias of trackName
	TrackName         string   `json:"trackName,omitempty"`
	Distance          *float64 `json:"distance,omitempty"` // UI alias of trackLength
	TrackLength       *float64 `json:"trackLength,omitempty"`
	Laps              int      `json:"laps"`
	EventLength       int      `json:"eventLength,omitempty"`
	Mode              byte     `json:"mode"` // UI alias of gameMode
	GameMode          byte     `json:"gameMode,omitempty"`
	Damage            byte     `json:"damage"` // UI alias of damageMode
	DamageMode        byte     `json:"damageMode,omitempty"`
	Grid              byte     `json:"grid"` // UI alias of gridSize
	GridSize          byte     `json:"gridSize,omitempty"`
	Remaining         byte     `json:"remaining"` // UI alias of gridSizeRemaining
	GridSizeRemaining byte     `json:"gridSizeRemaining,omitempty"`
	SectorCount       byte     `json:"sectorCount,omitempty"`
	SectorFract1      float64  `json:"sectorFract1,omitempty"`
	SectorFract2      float64  `json:"sectorFract2,omitempty"`
	SessionStatus     byte     `json:"sessionStatus,omitempty"`
	Status            byte     `json:"status,omitempty"` // Session.status
	Flags             int      `json:"flags"`          // marshalFlagsPlayer
	GameStatus        byte     `json:"gameStatus,omitempty"`
	SessionMS         int32    `json:"sessionMs,omitempty"`
	Clock             int32    `json:"clock"` // raceTime
}

// BoardStats is the small UI leaderboard subset; full row is Leaderboard.
type BoardStats struct {
	TrackStatus byte   `json:"trackStatus"`
	Frags       int    `json:"frags"`
	Wrecks      int    `json:"wrecks"`
	Assists     int    `json:"assists,omitempty"`
	Score       int32  `json:"score,omitempty"`
	Points      int32  `json:"points,omitempty"`
	LapTiming   uint16 `json:"lapTiming,omitempty"`
}

type Standing struct {
	*BoardStats
	Leaderboard *ParticipantLeaderboard   `json:"leaderboard,omitempty"`
	Timing      *ParticipantTiming        `json:"timing,omitempty"`
	Sectors     *ParticipantTimingSectors `json:"sectors,omitempty"`
	Info        *ParticipantInfo          `json:"info,omitempty"`
	Damage      *ParticipantDamage        `json:"damage,omitempty"`

	// Flat UI aliases.
	Progress    *float64    `json:"progress,omitempty"`
	BestSectors *[3]*uint32 `json:"bestSectors,omitempty"`
	sectorsAt   time.Time
	Slot        int     `json:"slot"`
	Name        string  `json:"name"`
	CarID       string  `json:"carId"`
	CarName     string  `json:"carName,omitempty"`
	Status      int     `json:"status"`
	Place       int     `json:"place"`
	Lap         int     `json:"lap"`
	Health      int     `json:"health"`
	Gap         *int32  `json:"gap,omitempty"`
	Current     *uint32 `json:"current,omitempty"`
	Last        *uint32 `json:"last,omitempty"`
	Best        *uint32 `json:"best,omitempty"`
	boardAt     time.Time
	timingAt    time.Time
	damageAt    time.Time
	infoAt      time.Time
	bot         bool
}

func (e *Engine) UpdateRace(details RaceDetails) {
	e.mu.Lock()
	defer e.mu.Unlock()
	details.Clock = e.raceMS
	if details.TrackName == "" {
		details.TrackName = details.Track
	}
	if details.Track == "" {
		details.Track = details.TrackName
	}
	if details.TrackLength == nil {
		details.TrackLength = details.Distance
	}
	if details.Distance == nil {
		details.Distance = details.TrackLength
	}
	if details.GameMode == 0 && details.Mode != 0 {
		details.GameMode = details.Mode
	}
	if details.Mode == 0 {
		details.Mode = details.GameMode
	}
	if details.DamageMode == 0 && details.Damage != 0 {
		details.DamageMode = details.Damage
	}
	if details.Damage == 0 {
		details.Damage = details.DamageMode
	}
	if details.GridSize == 0 {
		details.GridSize = details.Grid
	}
	if details.Grid == 0 {
		details.Grid = details.GridSize
	}
	if details.GridSizeRemaining == 0 {
		details.GridSizeRemaining = details.Remaining
	}
	if details.Remaining == 0 {
		details.Remaining = details.GridSizeRemaining
	}
	if details.Status == 0 {
		details.Status = details.SessionStatus
	}
	if details.SessionStatus == 0 {
		details.SessionStatus = details.Status
	}
	e.raceDetails = &details
}

func (e *Engine) UpdateTiming(slot int, current, last, best uint32, progress float64) {
	e.UpdateTimingRow(slot, ParticipantTiming{
		LapTimeCurrent: current,
		LapTimeLast:    last,
		LapTimeBest:    best,
		LapProgress:    progress,
	})
}

func (e *Engine) UpdateTimingRow(slot int, t ParticipantTiming) {
	e.mu.Lock()
	defer e.mu.Unlock()
	row := &e.standings[slot]
	cp := t
	row.Timing = &cp
	row.Current, row.Last, row.Best = validTime(t.LapTimeCurrent), validTime(t.LapTimeLast), validTime(t.LapTimeBest)
	row.Progress = nil
	if t.LapProgress >= 0 && t.LapProgress <= 1 {
		p := t.LapProgress
		row.Progress = &p
	}
	row.timingAt = e.clock()
}

func (e *Engine) UpdateSectors(slot int, first, second, third uint32) {
	e.UpdateSectorsRow(slot, ParticipantTimingSectors{
		SectorTimeBest1: first,
		SectorTimeBest2: second,
		SectorTimeBest3: third,
	})
}

func (e *Engine) UpdateSectorsRow(slot int, s ParticipantTimingSectors) {
	e.mu.Lock()
	defer e.mu.Unlock()
	row := &e.standings[slot]
	cp := s
	row.Sectors = &cp
	row.BestSectors = &[3]*uint32{validTime(s.SectorTimeBest1), validTime(s.SectorTimeBest2), validTime(s.SectorTimeBest3)}
	row.sectorsAt = e.clock()
}

func (e *Engine) UpdateInfoFull(slot int, info ParticipantInfo, bot bool) {
	e.UpdateInfo(slot, info.PlayerName, info.CarID, info.CarName, bot)
	e.mu.Lock()
	defer e.mu.Unlock()
	cp := info
	e.standings[slot].Info = &cp
	e.standings[slot].infoAt = e.clock()
}

func (e *Engine) UpdateDamage(slot int, d ParticipantDamage) {
	e.mu.Lock()
	defer e.mu.Unlock()
	cp := ParticipantDamage{
		DamageStates: append([]byte(nil), d.DamageStates...),
		Parts:        append([]byte(nil), d.Parts...),
	}
	e.standings[slot].Damage = &cp
	e.standings[slot].damageAt = e.clock()
}

func (e *Engine) UpdateVisibility(v byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.visibility = v
	e.hasVisibility = true
}

func (e *Engine) UpdatePlayer(p PlayerTelemetry) {
	e.mu.Lock()
	defer e.mu.Unlock()
	cp := p
	if len(cp.ParticipantPlayerDamage.DamageStates) > 0 {
		cp.ParticipantPlayerDamage.DamageStates = append([]byte(nil), cp.ParticipantPlayerDamage.DamageStates...)
	}
	if len(cp.ParticipantPlayerDamage.Parts) > 0 {
		cp.ParticipantPlayerDamage.Parts = append([]byte(nil), cp.ParticipantPlayerDamage.Parts...)
	}
	e.player = &cp
	e.playerAt = e.clock()
}

func validTime(ms uint32) *uint32 {
	if ms == 0 || ms >= 1<<31 {
		return nil
	}
	return &ms
}

func (e *Engine) updateStanding(slot, status, place, lap, health int, gap int32, stats *BoardStats, full *ParticipantLeaderboard) {
	e.mu.Lock()
	defer e.mu.Unlock()
	row := &e.standings[slot]
	row.Slot, row.Status, row.Place, row.Lap, row.Health = slot, status, place, lap, health
	row.BoardStats = stats
	if full != nil {
		cp := *full
		row.Leaderboard = &cp
	}
	row.Gap = nil
	if gap >= 0 && gap < math.MaxInt32 {
		row.Gap = &gap
	}
	row.boardAt = e.clock()
}

func (e *Engine) currentStandings(now time.Time) []Standing {
	rows := make([]Standing, 0, len(e.standings))
	for _, row := range e.standings {
		if row.Status < 1 || row.Status > 8 || now.Sub(row.boardAt) > 2*time.Second || (e.settings.HideBots && row.bot) {
			continue
		}
		if now.Sub(row.timingAt) > 2*time.Second {
			row.Current, row.Last, row.Best, row.Progress, row.Timing = nil, nil, nil, nil, nil
		}
		if now.Sub(row.sectorsAt) > 2*time.Second {
			row.BestSectors, row.Sectors = nil, nil
		}
		if now.Sub(row.damageAt) > 2*time.Second {
			row.Damage = nil
		}
		if now.Sub(row.infoAt) > 2*time.Second {
			row.Info = nil
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Place == 0 {
			return false
		}
		return rows[j].Place == 0 || rows[i].Place < rows[j].Place
	})
	return rows
}
