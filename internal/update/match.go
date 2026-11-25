package update

import (
	"strings"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

// MatchPortable picks a bare Windows exe or portable zip for in-place swap.
// Skips setup/installer packages and checksum sidecars.
func MatchPortable(req updater.CheckRequest, assets []github.ReleaseAsset) int {
	plat := strings.ToLower(req.Platform)
	arch := strings.ToLower(req.Arch)

	bestZip := -1
	for i, a := range assets {
		name := strings.ToLower(a.Name)
		if skipAsset(name) || isSetupOrInstaller(name) {
			continue
		}
		if plat != "" && !strings.Contains(name, plat) {
			continue
		}
		if arch != "" && !archOK(name, arch) {
			continue
		}
		// Prefer a bare .exe (updater swap target) over a zip.
		if strings.HasSuffix(name, ".exe") {
			return i
		}
		if bestZip < 0 && (strings.HasSuffix(name, ".zip") || strings.Contains(name, "portable")) {
			bestZip = i
		}
	}
	return bestZip
}

// MatchSetup picks the NSIS/setup installer for installed builds (launch, don't swap).
func MatchSetup(req updater.CheckRequest, assets []github.ReleaseAsset) int {
	plat := strings.ToLower(req.Platform)
	arch := strings.ToLower(req.Arch)

	for i, a := range assets {
		name := strings.ToLower(a.Name)
		if skipAsset(name) || !isSetupOrInstaller(name) {
			continue
		}
		if plat != "" && !strings.Contains(name, plat) {
			continue
		}
		if arch != "" && !archOK(name, arch) {
			continue
		}
		return i
	}
	return -1
}

func isSetupOrInstaller(name string) bool {
	return strings.Contains(name, "-setup.") ||
		strings.Contains(name, "_setup.") ||
		strings.Contains(name, "-installer.") ||
		strings.Contains(name, "_installer.") ||
		name == "installer.exe" ||
		name == "setup.exe"
}

func skipAsset(name string) bool {
	if strings.HasSuffix(name, ".sig") || strings.HasSuffix(name, ".asc") {
		return true
	}
	for _, ext := range []string{".sha256", ".sha512", ".sums", ".checksum", ".checksums"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	for _, token := range []string{"checksums", "sha256sums", "sha512sums"} {
		if name == token || strings.HasPrefix(name, token+".") || strings.Contains(name, token+".") {
			return true
		}
	}
	return false
}

func archOK(name, arch string) bool {
	if strings.Contains(name, arch) {
		return true
	}
	if arch == "amd64" && (strings.Contains(name, "x86_64") || strings.Contains(name, "x64")) {
		return true
	}
	if arch == "arm64" && strings.Contains(name, "aarch64") {
		return true
	}
	return false
}
