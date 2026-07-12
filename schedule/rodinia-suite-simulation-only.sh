#!/bin/bash

source venv/bin/activate

for title in kmeans backprop pathfinder lud gaussian heartwall cfd nn bfs lavaMD "b+tree"; do
    echo "Running: python schedule/schedule.py --collect etc/rodinia/${title}.yaml --no-profile --no-trace"
    python schedule/schedule.py --collect etc/rodinia/${title}.yaml --no-profile --no-trace || true
    timestamp=$(TZ='America/New_York' date '+%Y-%m-%d %H:%M:%S %Z')
    echo "[${timestamp}] Finished: ${title}" >> rodinia-simulation-only.log
done