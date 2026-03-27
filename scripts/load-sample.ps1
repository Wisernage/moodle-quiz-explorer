# Imports ./mybackup.mbz (project root) into the same DB volume the server uses.
# Each import replaces all quiz/question rows in SQLite (full reset of app data).
# To also delete the persisted DB file and WAL (clean slate): .\scripts\reset-and-import.ps1
# Start the stack first: docker compose up -d
# Requires: mybackup.mbz in the project root (next to docker-compose.yml)

$ErrorActionPreference = "Stop"
$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $root

$mbz = Join-Path $root "mybackup.mbz"
if (-not (Test-Path -LiteralPath $mbz)) {
    Write-Error "Missing file: $mbz - place your Moodle backup there."
}
$len = (Get-Item -LiteralPath $mbz).Length
if ($len -lt 10) {
    Write-Error "mybackup.mbz is too small ($len bytes) to be a valid Moodle backup."
}

# Mount outside /imports - Docker Desktop cannot file-bind under /imports (compose volume).
docker compose run --rm `
    -v "${mbz}:/tmp/mybackup.mbz:ro" `
    timo import /tmp/mybackup.mbz `
    -dsn "file:/data/timo.db?cache=shared&_fk=1"

Write-Host "Done. Refresh http://localhost:8080"
