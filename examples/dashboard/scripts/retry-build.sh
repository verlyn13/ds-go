#!/usr/bin/env bash
set -euo pipefail

CMD=${1:-"bun run build"}
TRIES=${TRIES:-3}
DELAY=${DELAY:-2}

for i in $(seq 1 "$TRIES"); do
  echo "Attempt $i/$TRIES: $CMD"
  if eval "$CMD"; then
    exit 0
  fi
  echo "Build failed. Retrying in ${DELAY}s…" >&2
  sleep "$DELAY"
done

echo "Build failed after $TRIES attempts." >&2
exit 1

