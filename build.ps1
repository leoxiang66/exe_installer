<#!
.SYNOPSIS
  Cross/Windows build helper for the installer stub + packager.

.PARAMETER Arch
  Target architecture (amd64|386|arm64). Default: amd64.

.PARAMETER Mode
  build       -> only build stub.exe
  package     -> build stub.exe + run main.go to produce final bundled exe
  clean       -> remove generated artifacts

.PARAMETER ManifestMethod
  auto (default) | mt | windres | rsrc | none
  Choose how to embed the administrator manifest.
  auto: pick first available in order: existing .syso, mt, windres, rsrc.

.PARAMETER Verbose
  Show extra logs.

.EXAMPLE
  # Build + package using auto method
  ./build.ps1 -Mode package

.EXAMPLE
  # Force windres
  ./build.ps1 -Mode build -ManifestMethod windres

.EXAMPLE
  # Clean
  ./build.ps1 -Mode clean

.NOTES
  Requires Go installed. For windres you need mingw-w64; for mt you need Windows SDK; for rsrc the Go tool will install it if missing.
!#>
[CmdletBinding()]
param(
  [ValidateSet('amd64','386','arm64')]
  [string]$Arch = 'amd64',
  [ValidateSet('build','package','clean')]
  [string]$Mode = 'package',
  [ValidateSet('auto','mt','windres','rsrc','none')]
  [string]$ManifestMethod = 'auto'
)

# Use built-in common parameter -Verbose supplied by [CmdletBinding()]
$VerboseEnabled = $PSBoundParameters.ContainsKey('Verbose') -or $VerbosePreference -eq 'Continue'

$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $Root

$StubDir = Join-Path $Root 'installer/stub'
$StubExe = Join-Path $Root 'stub.exe'
$Syso = Join-Path $StubDir 'stub_windows.syso'
$Manifest = Join-Path $StubDir 'stub.manifest'
$RCFile = Join-Path $StubDir 'stub.rc'

function Log { param([string]$m) if($VerboseEnabled){ Write-Host "[+] $m" -ForegroundColor Cyan } }

function Clean {
  Log 'Cleaning artifacts'
  Remove-Item $StubExe -ErrorAction SilentlyContinue
  Remove-Item (Join-Path $StubDir 'stub_windows.syso') -ErrorAction SilentlyContinue
  Get-ChildItem *.exe -Exclude stub.exe -ErrorAction SilentlyContinue | Remove-Item -ErrorAction SilentlyContinue
  Write-Host 'Clean complete.'
}

if($Mode -eq 'clean'){ Clean; return }

if(!(Test-Path $Manifest)) { throw "Manifest not found: $Manifest" }

function EnsureRsrc {
  if(-not (Get-Command rsrc -ErrorAction SilentlyContinue)) {
    Log 'Installing rsrc tool...'
    go install github.com/akavel/rsrc@latest
    $env:PATH = "$([System.Environment]::GetEnvironmentVariable('GOPATH'))/bin;" + $env:PATH
  }
}

function EmbedManifestAuto {
  if(Test-Path $Syso){ Log '.syso already present (skip embed)'; return 'syso' }
  if(Get-Command mt.exe -ErrorAction SilentlyContinue){ EmbedManifestMt; return 'mt' }
  if(Get-Command windres -ErrorAction SilentlyContinue){ EmbedManifestWindres; return 'windres' }
  EnsureRsrc; EmbedManifestRsrc; return 'rsrc'
}
function EmbedManifestMt {
  if(-not (Get-Command mt.exe -ErrorAction SilentlyContinue)) { throw 'mt.exe not found' }
  Log 'Embedding manifest via mt.exe (post-build injection)'
  & mt.exe -manifest $Manifest -outputresource:$StubExe';#1'
}
function EmbedManifestWindres {
  if(-not (Get-Command windres -ErrorAction SilentlyContinue)) { throw 'windres not found' }
  Log 'Generating .syso via windres'
  & windres $RCFile -O coff -o $Syso
}
function EmbedManifestRsrc {
  EnsureRsrc
  Log 'Generating .syso via rsrc'
  & rsrc -manifest $Manifest -o $Syso
}

# Step 1: (optional) remove previous stub.exe so mt injection is deterministic
if(Test-Path $StubExe){ Remove-Item $StubExe }

# Step 2: if method chooses .syso path, create before build (except mt/none)
$chosen = $ManifestMethod
switch($ManifestMethod){
  'auto' { $chosen = EmbedManifestAuto }
  'windres' { EmbedManifestWindres }
  'rsrc' { EmbedManifestRsrc }
  'mt' { # mt runs after build
  }
  'none' { Log 'Skipping manifest embedding' }
}

# Step 3: build stub
$env:GOOS = 'windows'
$env:GOARCH = $Arch
Log "Building stub (GOARCH=$Arch, method=$chosen)"
go build -o $StubExe ./installer/stub

if($ManifestMethod -eq 'mt' -or ($ManifestMethod -eq 'auto' -and $chosen -eq 'mt')){
  EmbedManifestMt
}

if($Mode -eq 'build'){
  Write-Host 'Stub build complete.'
  return
}

# Step 4: Run packager to create final installer (setup.exe)
Log 'Running packager (go run main.go)'
go run ./main.go

Write-Host 'Done.'
