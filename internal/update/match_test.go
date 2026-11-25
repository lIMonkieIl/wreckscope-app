package update

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

func TestMatchSetupPicksSetup(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "WreckScope-windows-amd64.exe"},
		{Name: "WreckScope-0.2.0-windows-amd64-setup.exe"},
		{Name: "SHA256SUMS"},
	}
	got := MatchSetup(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, assets)
	if got != 1 {
		t.Fatalf("want setup index 1, got %d", got)
	}
}

func TestMatchSetupSkipsBareExe(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "WreckScope-windows-amd64.exe"},
		{Name: "SHA256SUMS"},
	}
	got := MatchSetup(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, assets)
	if got != -1 {
		t.Fatalf("bare exe must not match setup, got %d", got)
	}
}
