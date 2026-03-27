#!/usr/bin/env sh
# Wipes the persisted Docker SQLite volume, then imports ./mybackup.mbz from scratch.
# For normal re-imports without deleting the volume, use ./scripts/load-sample.sh

set -e
cd "$(dirname "$0")/.."

if [ ! -f mybackup.mbz ]; then
  echo "Missing ./mybackup.mbz — place your Moodle backup in the project root." >&2
  exit 1
fi

echo "Stopping containers and removing timo-data volume (old SQLite + WAL removed)..."
docker compose down -v

echo "Rebuilding image and starting stack (empty /data volume)..."
docker compose up -d --build

echo "Importing backup into fresh database..."
docker compose run --rm \
  -v "$(pwd)/mybackup.mbz:/tmp/mybackup.mbz:ro" \
  timo import /tmp/mybackup.mbz \
  -dsn 'file:/data/timo.db?cache=shared&_fk=1'

echo "Done. Open http://localhost:8080"
