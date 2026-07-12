#!/bin/bash

source venv/bin/activate

# nn kmeans backprop bfs gaussian lavaMD pathfinder lud b+tree cfd heartwall

for title in nn kmeans backprop bfs gaussian lavaMD pathfinder lud b+tree cfd heartwall; do
    echo "Running: python schedule/schedule.py --collect etc/rodinia/${title}.yaml"
    python schedule/schedule.py --collect etc/rodinia/${title}.yaml || true
done