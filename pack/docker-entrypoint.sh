#!/bin/sh
# Orange TV container entrypoint: start backend in background, nginx in foreground.
set -e

CONFIG_FILE="${CONFIG_FILE:-/app/configs/config.prod.yaml}"

echo "[entrypoint] starting orange-tv with config: ${CONFIG_FILE}"
/app/orange-tv serve -c "${CONFIG_FILE}" &
BACKEND_PID=$!

# Wait briefly and verify backend is still running (catches immediate crashes
# like config errors or missing migrations before nginx takes over).
sleep 2
if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo "[entrypoint] ERROR: backend process exited during startup, container stopping"
    wait "$BACKEND_PID"
    exit 1
fi

echo "[entrypoint] starting nginx"
exec nginx -g 'daemon off;'
