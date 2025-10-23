#!/bin/bash

source venv/bin/activate

# 2dconv 2mm 3dconv 3mm atax bicg gemm gesummv mvt syrk

for title in 2dconv 2mm 3dconv 3mm atax bicg gemm gesummv mvt syrk; do
    echo "Running: python schedule/schedule.py --collect etc/polybench/${title}.yaml --no-profile --no-simulation"
    python schedule/schedule.py --collect etc/polybench/${title}.yaml --no-profile --no-simulation || true
done

for title in fastwalshtransform mergesort scalarprod scan-long scan-short sortingnetworks-bitonic sortingnetworks-oddeven transpose vectoradd; do
    echo "Running: python schedule/schedule.py --collect etc/cuda-sdk/${title}.yaml --no-profile --no-simulation"
    python schedule/schedule.py --collect etc/cuda-sdk/${title}.yaml --no-profile --no-simulation || true
done