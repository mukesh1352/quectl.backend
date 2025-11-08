#!/usr/bin/env bash
set -e

echo " Running queuectl end-to-end demo test..."
echo "-------------------------------------------"

# Clean slate
rm -f queue.db || true

#  Enqueue some jobs
echo "[1/5] Enqueueing jobs..."
go run main.go enqueue '{"command":"echo Job 1 completed"}'
go run main.go enqueue '{"command":"false"}'
go run main.go enqueue '{"command":"sleep 1 && echo Job 3 done"}'

echo
echo "[2/5] Current jobs (should all be pending):"
go run main.go list

echo
echo "[3/5] Starting 2 workers for 10 seconds..."
timeout 10 go run main.go worker start --count 2 || true

echo
echo "[4/5] Checking job status:"
go run main.go status

echo
echo "[5/5] Dead Letter Queue:"
go run main.go dlq

echo
echo " End-to-end test complete!"
