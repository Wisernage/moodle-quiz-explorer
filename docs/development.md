# Development

## Prerequisites

- **Go 1.22+**
- **CGO enabled** — the project uses `github.com/mattn/go-sqlite3`, which requires CGO and a C toolchain.

### C toolchain hints

- **Linux:** `gcc` (e.g. `build-essential` on Debian/Ubuntu).
- **macOS:** Xcode Command Line Tools (`xcode-select --install`).
- **Windows:** install a GCC distribution used by Go (e.g. [MSYS2](https://www.msys2.org/) with `pacman -S mingw-w64-x86_64-gcc`) and ensure `gcc` is on `PATH`, or build inside **WSL** / **Docker** where GCC is available.

Verify:

```bash
go env CGO_ENABLED   # should be 1 when building
```

## Build

From the repository root:

```bash
go build -o timo ./cmd/timo
```

Release-style binary (smaller, no paths in binary):

```bash
CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o timo ./cmd/timo
```

On Windows PowerShell:

```powershell
$env:CGO_ENABLED=1; go build -trimpath -ldflags="-s -w" -o timo.exe ./cmd/timo
```

## Test

```bash
go test ./... -count=1
```

Run with coverage:

```bash
go test ./... -count=1 -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Docker

- **Runtime image:** `docker compose up --build` uses the `runtime` stage (Debian slim + SQLite runtime libs).
- **Tests:** `docker compose --profile test run --rm test` mounts the repo and runs `go test ./...` with CGO in the build image.

See the comments in `docker-compose.yml` for import examples and volume layout.

## Project layout

| Path | Role |
|------|------|
| `cmd/timo` | CLI entry (`import`, `import-dir`, `serve`) |
| `internal/api` | HTTP handlers and embedded HTML templates |
| `internal/db` | SQLite open + AutoMigrate |
| `internal/importer` | MBZ extraction and backup XML ingestion |
| `internal/moodle` | Moodle XML parsing and question payload helpers |
| `internal/model` | GORM models |
| `testdata/moodle/minimal` | Minimal backup fixture for tests |
