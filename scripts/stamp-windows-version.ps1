# Stamp Windows packaging metadata from a single version string.
# Keeps info.json (exe Details), wails.exe.manifest, and NSIS INFO_PRODUCTVERSION
# in sync with build/config.yml / package VERSION. Safe to re-run.
param(
	[Parameter(Mandatory = $true)]
	[string]$Version,
	[string]$Root = ""
)

$ErrorActionPreference = "Stop"

if (-not $Root) {
	$Root = Split-Path -Parent $PSScriptRoot
}

if ($Version -notmatch '^\d+\.\d+\.\d+') {
	throw "Version must look like X.Y.Z (got '$Version')"
}

# Windows VERSIONINFO / assemblyIdentity want four parts.
$parts = $Version.Split('.')
while ($parts.Count -lt 4) { $parts += '0' }
$fourPart = ($parts[0..3] -join '.')

$infoPath = Join-Path $Root "build\windows\info.json"
$manifestPath = Join-Path $Root "build\windows\wails.exe.manifest"
$nshPath = Join-Path $Root "build\windows\nsis\wails_tools.nsh"

if (-not (Test-Path $infoPath)) { throw "missing $infoPath" }
if (-not (Test-Path $manifestPath)) { throw "missing $manifestPath" }
if (-not (Test-Path $nshPath)) { throw "missing $nshPath" }

# Rewrite info.json with stable formatting (0409 = en-US; Explorer ignores neutral 0000).
$infoText = @"
{
	"fixed": {
		"file_version": "$fourPart",
		"product_version": "$fourPart"
	},
	"info": {
		"0409": {
			"FileVersion": "$Version",
			"ProductVersion": "$Version",
			"CompanyName": "WreckScope",
			"FileDescription": "WreckScope",
			"LegalCopyright": "Copyright Monkie",
			"ProductName": "WreckScope",
			"Comments": "WreckScope dashboard"
		}
	}
}
"@
[System.IO.File]::WriteAllText($infoPath, ($infoText.TrimEnd() + "`n"), (New-Object System.Text.UTF8Encoding $false))

$manifest = Get-Content $manifestPath -Raw -Encoding UTF8
# Only the app identity — never Common-Controls (must stay 6.0.0.0).
$manifest2 = [regex]::Replace(
	$manifest,
	'(<assemblyIdentity\b[^>]*\bname="com\.wreckscope\.app"[^>]*\bversion=")[^"]+(")',
	('${1}' + $fourPart + '${2}'),
	1
)
if ($manifest2 -eq $manifest -and $manifest -notmatch ('name="com\.wreckscope\.app"[^>]*version="' + [regex]::Escape($fourPart) + '"')) {
	throw "failed to stamp app assemblyIdentity version in $manifestPath"
}
# Guard against accidental Common-Controls corruption from older stamp runs.
$manifest2 = [regex]::Replace(
	$manifest2,
	'(<assemblyIdentity\b[^>]*\bname="Microsoft\.Windows\.Common-Controls"[^>]*\bversion=")[^"]+(")',
	'${1}6.0.0.0${2}',
	1
)
[System.IO.File]::WriteAllText($manifestPath, $manifest2, (New-Object System.Text.UTF8Encoding $false))

$nsh = Get-Content $nshPath -Raw -Encoding UTF8
$nsh2 = [regex]::Replace(
	$nsh,
	'(!ifndef INFO_PRODUCTVERSION\s*\r?\n\s*!define INFO_PRODUCTVERSION ")[^"]+(")',
	('${1}' + $Version + '${2}'),
	1
)
if ($nsh2 -eq $nsh -and $nsh -notmatch [regex]::Escape("!define INFO_PRODUCTVERSION `"$Version`"")) {
	throw "failed to stamp INFO_PRODUCTVERSION in $nshPath"
}
[System.IO.File]::WriteAllText($nshPath, $nsh2, (New-Object System.Text.UTF8Encoding $false))

Write-Host "Stamped Windows version $Version ($fourPart) into info.json, manifest, wails_tools.nsh"
