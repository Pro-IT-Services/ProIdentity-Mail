param(
    [switch]$SkipVulnCheck
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

Write-Host "== Go tests =="
go test ./...

Write-Host "== Module inventory =="
go list -m all

if (-not $SkipVulnCheck) {
    Write-Host "== govulncheck =="
    $govulncheck = Get-Command govulncheck -ErrorAction SilentlyContinue
    if ($govulncheck) {
        & $govulncheck.Source ./...
    } else {
        go run golang.org/x/vuln/cmd/govulncheck@latest ./...
    }
}
