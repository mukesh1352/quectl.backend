#!/usr/bin/env bash
set -euo pipefail

echo ""
echo "=========================================="
echo "  Running Queuectl End-to-End Demo Test"
echo "=========================================="
echo ""

# --- Clean slate ---
if [ -f queue.db ]; then
  echo "[Cleanup] Removing existing queue.db..."
  rm -f queue.db
fi

# --- Step 1: Enqueue Jobs ---
echo ""
echo "[1/5] Enqueueing jobs..."
go run main.go enqueue '{"command":"echo Job 1 completed"}'
go run main.go enqueue '{"command":"false"}'
go run main.go enqueue '{"command":"sleep 1 && echo Job 3 done"}'
sleep 1

# --- Step 2: List Jobs ---
echo ""
echo "[2/5] Current jobs (should all be pending):"
go run main.go list
sleep 1

# --- Step 3: Start Workers (cross-platform compatible) ---
echo ""
echo "[3/5] Starting 2 workers for 10 seconds..."
if command -v timeout >/dev/null 2>&1; then
  timeout 10s go run main.go worker start --count 2 --timeout 5s || true
elif command -v gtimeout >/dev/null 2>&1; then
  # macOS (with GNU coreutils)
  gtimeout 10s go run main.go worker start --count 2 --timeout 5s || true
else
  # Fallback: manual background process with kill
  go run main.go worker start --count 2 --timeout 5s &
  WORKER_PID=$!
  sleep 10
  echo "[Info] Stopping worker after 10s..."
  kill "$WORKER_PID" 2>/dev/null || true
fi
sleep 1

# --- Step 4: Queue Status ---
echo ""
echo "[4/5] Checking job status:"
go run main.go status
sleep 1

# --- Step 5: Dead Letter Queue ---
echo ""
echo "[5/5] Dead Letter Queue:"
go run main.go dlq

echo ""
echo "=========================================="
echo " End-to-End Demo Test Complete!"
echo "=========================================="
