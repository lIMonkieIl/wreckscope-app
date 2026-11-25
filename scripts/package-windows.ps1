# Build Windows NSIS installer (default), portable zip, and bare updater exe.
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

# Bare exe for in-app updater (portable swap). Do not use *-setup* for binary replace.
$bareOut = Join-Path $outDir "WreckScope-windows-$Arch.exe"
Copy-Item $exe $bareOut -Force
Write-Host "Updater:   $bareOut"

$stage = Join-Path $outDir "portable-stage"
if (Test-Path $stage) { Remove-Item $stage -Recurse -Force }
New-Item -ItemType Directory -Force -Path $stage | Out-Null
Copy-Item $exe (Join-Path $stage "WreckScope.exe")
# Portable mode: settings.json beside the exe with portable: true (no sibling marker file).
# utf8NoBOM — PowerShell's "utf8" writes a BOM that breaks Go json.Unmarshal / portable detect.
$portableSettings = @'
{
  "portable": true,
  "setupCompleted": false,
  "theme": "wreckscope",
  "mode": "system",
  "view": "recovery",
  "speedUnit": "km/h",
  "cautionSeconds": 1,
  "alertSeconds": 3.5,
  "speedThreshold": 3
}
'@
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText((Join-Path $stage "settings.json"), $portableSettings.Trim() + "`n", $utf8NoBom)
$zip = Join-Path $outDir "WreckScope-$Version-windows-$Arch-portable.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path (Join-Path $stage "*") -DestinationPath $zip -Force
Remove-Item $stage -Recurse -Force
Write-Host "Portable:  $zip"

# Optional sidecar for GitHub updater ChecksumAsset: SHA256SUMS
$sumsPath = Join-Path $outDir "SHA256SUMS"
$lines = @()
foreach ($f in @($installerOut, $bareOut, $zip)) {
	$hash = (Get-FileHash -Algorithm SHA256 -Path $f).Hash.ToLowerInvariant()
	$lines += "$hash  $(Split-Path $f -Leaf)"
}
[System.IO.File]::WriteAllText($sumsPath, ($lines -join "`n") + "`n", $utf8NoBom)
Write-Host "Checksums: $sumsPath"
Write-Host "Done."
