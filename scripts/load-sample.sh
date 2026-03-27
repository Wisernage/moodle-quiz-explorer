#!/usr/bin/env sh
# Imports ./mybackup.mbz (project root) into the same DB volume the server uses.
# Each import replaces all quiz/question rows. For a wiped volume + fresh file: ./scripts/reset-and-import.sh
# Start the stack first: docker compose up -d

set -e
cd "$(dirname "$0")/.."

if [ ! -f mybackup.mbz ]; then
  echo "Missing ./mybackup.mbz — place your Moodle backup in the project root." >&2
  exit 1
fi

# Do not mount under /imports — that path is already a compose volume (breaks on Docker Desktop).
docker compose run --rm \
  -v "$(pwd)/mybackup.mbz:/tmp/mybackup.mbz:ro" \
  timo import /tmp/mybackup.mbz \
  -dsn 'file:/data/timo.db?cache=shared&_fk=1'

echo "Done. Refresh http://localhost:8080"
