#!/bin/sh
set -e

mkdir -p /app/uploads
chown -R appuser:appuser /app/uploads 2>/dev/null || true
chmod 777 /app/uploads 2>/dev/null || true

exec su-exec appuser "$@"
