# Wipes the persisted Docker SQLite volume, then imports ./mybackup.mbz from scratch.
# Use this when you want a completely new database file (not only row replacement).
# For normal re-imports without deleting the volume, use .\scripts\load-sample.ps1

$ErrorActionPreference = "Stop"
$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $root

$mbz = Join-Path $root "mybackup.mbz"
if (-not (Test-Path -LiteralPath $mbz)) {
    Write-Error "Missing file: $mbz - place your Moodle backup in the project root."
}
$len = (Get-Item -LiteralPath $mbz).Length
if ($len -lt 10) {
    Write-Error "mybackup.mbz is too small ($len bytes) to be a valid Moodle backup."
}

Write-Host "Stopping containers and removing timo-data volume (old SQLite + WAL removed)..."
docker compose down -v

Write-Host "Starting stack (empty /data volume)..."
docker compose up -d

Write-Host "Importing backup into fresh database..."
docker compose run --rm `
    -v "${mbz}:/tmp/mybackup.mbz:ro" `
    timo import /tmp/mybackup.mbz `
    -dsn "file:/data/timo.db?cache=shared&_fk=1"

Write-Host "Done. Open http://localhost:8080"
