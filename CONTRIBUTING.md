# Contributing

Thank you for your interest in Timo. This document explains how to report issues, propose changes, and get a pull request merged smoothly.

## Code of conduct

All participants are expected to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## How to contribute

1. **Issues first (for larger changes)** — open an issue to describe the problem or feature. That avoids duplicate work and lets maintainers agree on direction.
2. **Fork and branch** — create a branch from `main` with a short, descriptive name (e.g. `fix-import-multichoice`, `docs-readme-docker`).
3. **Keep changes focused** — one logical change per pull request is easier to review.
4. **Tests** — add or update tests when you change behavior. Run `go test ./...` locally (see [docs/development.md](docs/development.md)).
5. **Commit messages** — use clear, imperative summaries (e.g. `Fix slot ordering when sections are empty`).

## Development setup

See [docs/development.md](docs/development.md) for Go version, **CGO**, build commands, and Docker.

## Pull request checklist

- [ ] `go test ./...` passes.
- [ ] New or updated tests cover the change where applicable.
- [ ] Documentation (README or `docs/`) updated if user-visible behavior or setup changed.
- [ ] No secrets, personal backup files, or large binary fixtures committed (see `.gitignore`).

## Go module path

The module is currently `timo` (short path). If you maintain a fork under your own GitHub account and want a canonical import path, you may rename the module to `github.com/yourname/timo` in `go.mod` and update imports — that is a deliberate refactor; do it in a dedicated change if needed.

## Security

Do not report security issues in public issues. See [SECURITY.md](SECURITY.md).
