# Build Windows NSIS installer (default) and portable zip.
# Requires: Go, pnpm, Wails CLI, NSIS (makensis on PATH or under Program Files).
param(
	[string]$Version = "",
	[string]$Platform = "windows/amd64"
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
	$wails = Get-Content (Join-Path $root "wails.json") -Raw | ConvertFrom-Json
	$Version = [string]$wails.info.productVersion
}
if (-not $Version) { $Version = "0.0.0" }

$outDir = Join-Path $root "dist\windows"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

Write-Host "Building Wails binary + NSIS ($Platform)..."
# Prefer the makensis Wails will invoke via PATH for this process.
$env:Path = (Split-Path $makensis) + ";" + $env:Path
wails build -platform $Platform -nsis -webview2 download
if ($LASTEXITCODE -ne 0) { throw "wails build failed ($LASTEXITCODE)" }

$binDir = Join-Path $root "build\bin"
$exe = Join-Path $binDir "WreckScope.exe"
if (-not (Test-Path $exe)) { throw "missing $exe" }

$installer = Get-ChildItem $binDir -Filter "*-installer.exe" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
if (-not $installer) { throw "NSIS installer not found under build\bin" }

$installerOut = Join-Path $outDir "WreckScope-$Version-windows-amd64-setup.exe"
Copy-Item $installer.FullName $installerOut -Force
Write-Host "Installer: $installerOut"

$stage = Join-Path $outDir "portable-stage"
if (Test-Path $stage) { Remove-Item $stage -Recurse -Force }
New-Item -ItemType Directory -Force -Path $stage | Out-Null
Copy-Item $exe (Join-Path $stage "WreckScope.exe")
# Marker forces settings.json beside the exe.
New-Item -ItemType File -Path (Join-Path $stage "portable") -Force | Out-Null
$zip = Join-Path $outDir "WreckScope-$Version-windows-amd64-portable.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path (Join-Path $stage "*") -DestinationPath $zip -Force
Remove-Item $stage -Recurse -Force
Write-Host "Portable:  $zip"
Write-Host "Done."
