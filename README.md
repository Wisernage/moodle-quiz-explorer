# Moodle Quiz Explorer

**Import Moodle quiz backups and browse them locally.** Moodle Quiz Explorer reads standard Moodle course backups (`.mbz` as ZIP or gzip-compressed tar), stores quiz metadata and questions in **SQLite**, and serves a small **HTTP UI** plus a **JSON API** for listing quizzes and viewing question payloads.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://github.com/Wisernage/moodle-quiz-explorer/actions/workflows/ci.yml/badge.svg)](https://github.com/Wisernage/moodle-quiz-explorer/actions/workflows/ci.yml)

> Formerly Timo. The `timo` command, Go module, database filename and Docker service/volume names remain compatible with existing installations.

---

## Table of contents

- [Features](#features)
- [Quick start](#quick-start)
- [CLI reference](#cli-reference)
- [HTTP API](#http-api)
- [Docker](#docker)
- [Documentation](#documentation)
- [Question types](#question-types)
- [Limitations](#limitations)
- [Contributing](#contributing)
- [License](#license)

---

## Features

- **Moodle `.mbz` import** — detects **ZIP** (common) or **tar.gz** (Moodle `tgz_packer`) archives.
- **SQLite storage** via [GORM](https://gorm.io/); schema auto-migrated on startup.
- **Quiz-centric model** — quizzes, slots, sections, feedbacks, grades, attempts (as present in backup), and question bank entries with JSON payloads.
- **Web UI** — list quizzes; open a quiz to see slots, stems, and parsed payloads (HTML stems rendered; unsupported types are clearly marked).
- **JSON API** — `Accept: application/json` or `?format=json` for list and quiz detail.
- **Practice attempts** — `POST` to record lightweight practice submissions (separate from Moodle-native attempt rows).

---

## Quick start

**Requirements:** [Go 1.22+](https://go.dev/dl/), a C compiler and **CGO enabled** (SQLite driver `mattn/go-sqlite3`).

```bash
git clone https://github.com/Wisernage/moodle-quiz-explorer.git
cd moodle-quiz-explorer

# Build (CGO must be on; on Windows use a dev environment that provides gcc, e.g. MSYS2 or TDM-GCC)
go build -o timo ./cmd/timo

# Import a Moodle backup (replaces existing quiz data in the DB)
./timo import /path/to/course.mbz

# Serve UI + API
./timo serve -addr :8080
```

Open [http://localhost:8080](http://localhost:8080).

**Custom database path:**

```bash
./timo import -dsn 'file:./data/app.db?cache=shared&_fk=1' course.mbz
./timo serve -dsn 'file:./data/app.db?cache=shared&_fk=1' -addr :8080
```

**Already extracted backup directory** (e.g. for debugging):

```bash
./timo import-dir /path/to/extracted/backup
```

---

## CLI reference

| Command | Purpose |
|--------|---------|
| `timo import [-dsn …] <file.mbz>` | Extract and import a backup file |
| `timo import-dir [-dsn …] <dir>` | Import from an extracted backup tree |
| `timo serve [-addr :8080] [-dsn …]` | Run HTTP server |

Default DSN: `file:timo.db?cache=shared&_fk=1` (creates parent directories when needed).

---

## HTTP API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | HTML quiz list |
| `GET` | `/health` | JSON `{"status":"ok"}` |
| `GET` | `/quizzes` | Quiz list (HTML or JSON) |
| `GET` | `/quizzes/{id}` | Quiz with relations (HTML or JSON) |
| `POST` | `/quizzes/{id}/attempts` | Body: `{"answers": {…}}` — stores a practice attempt |

**JSON:** send `Accept: application/json` or append `?format=json`.

---

## Docker

```bash
docker compose up --build
```

- Server listens on **8080**; database persists in a named volume (`timo-data`).
- Mount course backups read-only under `./imports` (see `docker-compose.yml`).
- Import example (after placing `mybackup.mbz` in the project root — **not** committed):

  - **Linux/macOS:** `./scripts/load-sample.sh` (expects `./mybackup.mbz`)
  - **Windows:** `./scripts/load-sample.ps1`

Or run the importer manually:

```bash
docker compose run --rm timo import /imports/your-course.mbz \
  -dsn 'file:/data/timo.db?cache=shared&_fk=1'
```

**Tests in Docker:**

```bash
docker compose --profile test run --rm test
```

See [docs/development.md](docs/development.md) for CGO and build details.

---

## Documentation

| Document | Contents |
|----------|----------|
| [docs/development.md](docs/development.md) | Building, testing, Docker, CGO |
| [docs/architecture.md](docs/architecture.md) | Packages, import pipeline, data model |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |
| [SECURITY.md](SECURITY.md) | Reporting vulnerabilities |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community standards |

---

## Question types

Structured JSON payloads are built for common types including **multichoice**, **truefalse**, **shortanswer**, **essay**, **match**, and **description**. Other types are still imported with metadata and **raw XML** preserved; the UI marks them as non-interactive in v1.

---

## Limitations

- Not a full Moodle replacement: grading rules, question behaviours, and plugins beyond what the backup exposes are only approximated.
- **Security:** bind to localhost or protect with a reverse proxy; do not expose raw to the internet without authentication and TLS.
- Backups may contain **sensitive data** — handle `.mbz` files like production data.

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) and the [Code of Conduct](CODE_OF_CONDUCT.md).

---

## License

[MIT License](LICENSE)
