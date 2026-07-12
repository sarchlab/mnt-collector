#!/bin/bash

source venv/bin/activate

for title in pointerchasing-l1 pointerchasing-l2 pointerchasing-dram; do
    timestamp1=$(TZ='America/New_York' date '+%Y-%m-%d %H:%M:%S %Z')
    echo "[${timestamp1}] Start: ${title}" >> simtune-pointerchasing-simulation-only.log
    echo "Running: python schedule/schedule.py --collect etc/simtune/${title}.yaml --no-profile --no-trace"
    python schedule/schedule.py --collect etc/simtune/${title}.yaml --no-profile --no-trace || true
    timestamp2=$(TZ='America/New_York' date '+%Y-%m-%d %H:%M:%S %Z')
    echo "[${timestamp2}] Finished: ${title}" >> simtune-pointerchasing-simulation-only.log
done

echo "All pointerchasing benchmarks completed." >> simtune-pointerchasing-simulation-only.log