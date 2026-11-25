package racefeed

import (
	"reflect"
	"testing"
)

func TestToggleMedicAssignNoCap(t *testing.T) {
	e := New()
	if got := e.Medics(); len(got) != 0 {
		t.Fatalf("start medics=%v", got)
	}
	if got := e.ToggleMedic(4); !reflect.DeepEqual(got, []int{4}) {
		t.Fatalf("assign 4: %v", got)
	}
	if got := e.ToggleMedic(9); !reflect.DeepEqual(got, []int{4, 9}) {
		t.Fatalf("assign 9: %v", got)
	}
	if got := e.ToggleMedic(4); !reflect.DeepEqual(got, []int{9}) {
		t.Fatalf("unmark 4: %v", got)
	}
	if got := e.ToggleMedic(-1); !reflect.DeepEqual(got, []int{9}) {
		t.Fatalf("invalid: %v", got)
	}
	if !reflect.DeepEqual(e.Snapshot().Medics, []int{9}) {
		t.Fatalf("snapshot=%v", e.Snapshot().Medics)
	}
	e.SetMode("demo")
	if len(e.Medics()) != 0 {
		t.Fatalf("cleared on mode: %v", e.Medics())
	}
}

func TestToggleMedicClearsOnRaceEpoch(t *testing.T) {
	e := New()
	_ = e.MarkPacket(0, true, 1000, 5000)
	e.ToggleMedic(2)
	_ = e.MarkPacket(0, true, 100, 100)
	if got := e.Medics(); len(got) != 0 {
		t.Fatalf("cleared on race epoch: %v", got)
	}
}
