#!/bin/bash

source venv/bin/activate

# 2dconv 2mm 3dconv 3mm atax bicg gemm gesummv mvt syrk

for title in 3dconv 3mm atax bicg gemm gesummv mvt syrk; do
    echo "Running: python schedule/schedule.py --collect etc/polybench/${title}.yaml"
    python schedule/schedule.py --collect etc/polybench/${title}.yaml
done