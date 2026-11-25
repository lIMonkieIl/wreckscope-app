# Verify Windows package artifacts carry the expected version everywhere that matters.
param(
	[Parameter(Mandatory = $true)]
	[string]$Version,
	[string]$Root = "",
	[string]$Arch = "amd64"
)

$ErrorActionPreference = "Stop"

if (-not $Root) {
	$Root = Split-Path -Parent $PSScriptRoot
}

$exe = Join-Path $Root "bin\WreckScope.exe"
$setup = Join-Path $Root "dist\windows\WreckScope-$Version-windows-$Arch-setup.exe"
$installerBin = Join-Path $Root "bin\WreckScope-$Arch-installer.exe"

$failed = @()

function Check-FileVersion([string]$Path, [string]$Label) {
	if (-not (Test-Path $Path)) {
		$script:failed += "missing $Label ($Path)"
		return
	}
	$vi = [System.Diagnostics.FileVersionInfo]::GetVersionInfo((Resolve-Path $Path))
	foreach ($prop in @('FileVersion', 'ProductVersion')) {
		$got = $vi.$prop
		if ($got -ne $Version -and $got -ne "$Version.0") {
			$script:failed += "$Label $prop='$got' want '$Version'"
		}
	}
	Write-Host "OK $Label FileVersion=$($vi.FileVersion) ProductVersion=$($vi.ProductVersion)"
}

function Check-AsciiContains([string]$Path, [string]$Needle, [string]$Label) {
	if (-not (Test-Path $Path)) {
		$script:failed += "missing $Label ($Path)"
		return
	}
	$bytes = [IO.File]::ReadAllBytes((Resolve-Path $Path))
	$ascii = [Text.Encoding]::ASCII.GetString($bytes)
	$unicode = [Text.Encoding]::Unicode.GetString($bytes)
	if (($ascii -notlike "*$Needle*") -and ($unicode -notlike "*$Needle*")) {
		$script:failed += "$Label missing string '$Needle'"
	} else {
		Write-Host "OK $Label contains '$Needle'"
	}
}

Check-FileVersion $exe "WreckScope.exe"
Check-FileVersion $setup "setup.exe"
if (Test-Path $installerBin) {
	Check-FileVersion $installerBin "bin-installer.exe"
}

Check-AsciiContains $exe $Version "WreckScope.exe"
Check-AsciiContains $setup $Version "setup.exe"

# Stale 0.1.0 must not be the only Windows resource version when shipping a newer release.
if ($Version -ne '0.1.0') {
	foreach ($path in @($exe, $setup)) {
		if (-not (Test-Path $path)) { continue }
		$vi = [System.Diagnostics.FileVersionInfo]::GetVersionInfo((Resolve-Path $path))
		if ($vi.ProductVersion -eq '0.1.0' -or $vi.FileVersion -eq '0.1.0') {
			$script:failed += "$(Split-Path $path -Leaf) still reports 0.1.0 resources"
		}
	}
}

if ($failed.Count -gt 0) {
	Write-Host "VERIFY FAILED:"
	$failed | ForEach-Object { Write-Host " - $_" }
	exit 1
}

Write-Host "VERIFY OK for version $Version"
