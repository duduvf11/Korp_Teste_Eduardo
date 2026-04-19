$ErrorActionPreference = "Stop"

$legacyPowerShellDir = Join-Path $env:SystemRoot "System32\WindowsPowerShell\v1.0"
if (-not ($env:Path -split ';' | Where-Object { $_ -eq $legacyPowerShellDir })) {
    $env:Path = "$env:Path;$legacyPowerShellDir"
}

$airExe = Join-Path $env:USERPROFILE "go\bin\air.exe"
if (-not (Test-Path $airExe)) {
    Write-Error "Air nao encontrado em $airExe. Rode: go install github.com/air-verse/air@latest"
}

Set-Location (Join-Path $PSScriptRoot "..\estoque")
& $airExe -c .air.toml
