#!/bin/bash

source venv/bin/activate

for title in 2dconv 2mm 3dconv 3mm atax bicg gemm gesummv mvt syrk; do
    echo "Running: python schedule/schedule.py --collect etc/polybench/${title}.yaml --no-profile --no-trace"
    python schedule/schedule.py --collect etc/polybench/${title}.yaml --no-profile --no-trace || true
    timestamp=$(TZ='America/New_York' date '+%Y-%m-%d %H:%M:%S %Z')
    echo "[${timestamp}] Finished: ${title}" >> polybench-simulation-only.log
done