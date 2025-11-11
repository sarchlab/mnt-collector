#!/bin/bash

source venv/bin/activate

for title in backprop pathfinder lud gaussian heartwall cfd nn kmeans bfs lavaMD "b+tree"; do
    echo "Running: python schedule/schedule.py --collect etc/rodinia/${title}.yaml --no-profile --no-trace"
    python schedule/schedule.py --collect etc/rodinia/${title}.yaml --no-profile --no-trace || true
done