# Build Windows NSIS installer + SHA256SUMS for GitHub Releases.
# Requires: Go, pnpm, Wails CLI v3 (wails3), NSIS (makensis on PATH or under Program Files).
param(
	[string]$Version = "",
	[string]$Arch = "amd64"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

function Find-MakeNsis {
	$cmd = Get-Command makensis -ErrorAction SilentlyContinue
	if ($cmd) { return $cmd.Source }
	$candidates = @(
		"${env:ProgramFiles(x86)}\NSIS\makensis.exe",
		"${env:ProgramFiles}\NSIS\makensis.exe"
	)
	foreach ($path in $candidates) {
		if (Test-Path $path) { return $path }
	}
	return $null
}

$makensis = Find-MakeNsis
if (-not $makensis) {
	throw "makensis not found. Install NSIS (https://nsis.sourceforge.io) and retry."
}
Write-Host "makensis: $makensis"

if (-not $Version) {
	$cfg = Get-Content (Join-Path $root "build\config.yml") -Raw
	if ($cfg -match '(?m)^\s*version:\s*"([^"]+)"') {
		$Version = $Matches[1]
	}
}
if (-not $Version) { $Version = "0.0.0" }

$outDir = Join-Path $root "dist\windows"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
# Drop leftover portable smoke extract folders from older packages.
Get-ChildItem $outDir -Directory -ErrorAction SilentlyContinue |
	Where-Object { $_.Name -like '*portable*' } |
	Remove-Item -Recurse -Force

Write-Host "Building Wails binary + NSIS (windows/$Arch) VERSION=$Version..."
$env:Path = (Split-Path $makensis) + ";" + $env:Path
# wails3/task may write harmless stderr (e.g. missing uname on Windows); only exit code matters.
$prevEap = $ErrorActionPreference
$ErrorActionPreference = "Continue"
& wails3 task windows:package "ARCH=$Arch" "VERSION=$Version"
$packExit = $LASTEXITCODE
$ErrorActionPreference = $prevEap
if ($packExit -ne 0) { throw "wails3 package failed ($packExit)" }

$binDir = Join-Path $root "bin"
$exe = Join-Path $binDir "WreckScope.exe"
if (-not (Test-Path $exe)) { throw "missing $exe" }

$installer = Get-ChildItem $binDir -Filter "*-installer.exe" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
if (-not $installer) { throw "NSIS installer not found under bin" }

$installerOut = Join-Path $outDir "WreckScope-$Version-windows-$Arch-setup.exe"
Copy-Item $installer.FullName $installerOut -Force
Write-Host "Installer: $installerOut"

# Sidecar for GitHub updater ChecksumAsset: SHA256SUMS
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
$sumsPath = Join-Path $outDir "SHA256SUMS"
$hash = (Get-FileHash -Algorithm SHA256 -Path $installerOut).Hash.ToLowerInvariant()
[System.IO.File]::WriteAllText($sumsPath, "$hash  $(Split-Path $installerOut -Leaf)`n", $utf8NoBom)
Write-Host "Checksums: $sumsPath"
Write-Host "Done."
