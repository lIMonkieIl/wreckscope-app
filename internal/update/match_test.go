package update

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

func TestMatchPortablePrefersBareExe(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "WreckScope-0.2.0-windows-amd64-setup.exe"},
		{Name: "WreckScope-0.2.0-windows-amd64-portable.zip"},
		{Name: "WreckScope-windows-amd64.exe"},
		{Name: "SHA256SUMS"},
	}
	got := MatchPortable(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, assets)
	if got != 2 {
		t.Fatalf("want bare exe index 2, got %d", got)
	}
}

func TestMatchPortableFallsBackToZip(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "WreckScope-0.2.0-windows-amd64-setup.exe"},
		{Name: "WreckScope-0.2.0-windows-amd64-portable.zip"},
	}
	got := MatchPortable(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, assets)
	if got != 1 {
		t.Fatalf("want portable zip index 1, got %d", got)
	}
}

func TestMatchPortableSkipsSetup(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "WreckScope-0.2.0-windows-amd64-setup.exe"},
	}
	got := MatchPortable(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, assets)
	if got != -1 {
		t.Fatalf("setup must not match for portable, got %d", got)
	}
}

func TestMatchSetupPrefersSetup(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "WreckScope-windows-amd64.exe"},
		{Name: "WreckScope-0.2.0-windows-amd64-setup.exe"},
		{Name: "WreckScope-0.2.0-windows-amd64-portable.zip"},
	}
	got := MatchSetup(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, assets)
	if got != 1 {
		t.Fatalf("want setup index 1, got %d", got)
	}
}
