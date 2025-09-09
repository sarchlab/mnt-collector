#!/bin/bash

source venv/bin/activate

# fastwalshtransform mergesort scalarprod scan-long scan-short sortingnetworks-bitonic sortingnetworks-oddeven transpose vectoradd

for title in fastwalshtransform mergesort scalarprod scan-long scan-short sortingnetworks-bitonic sortingnetworks-oddeven transpose vectoradd; do
    echo "Running: python schedule/schedule.py --collect etc/cuda-sdk/${title}.yaml"
    python schedule/schedule.py --collect etc/cuda-sdk/${title}.yaml || true
done