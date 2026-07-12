import os
import itertools
import yaml
import argparse
import sys

def build_jobs(job_dict, param_dict, suite, title, output_folder_path):
    if suite == "None":
        raise ValueError("suite must be specified (not 'None').")
    # Helper to get all titles to process
    if title == "ALL":
        titles = list(job_dict[suite].keys())
    elif isinstance(title, list):
        titles = title
    else:
        titles = [title]

    n_job_counts = len(titles)
    for idx, t in enumerate(titles):
        command = job_dict[suite][t]
        directory = os.path.dirname(command) + "/"
        # Build all param combinations
        keys = list(param_dict.keys())
        values = [param_dict[k] for k in keys]
        args_list = []
        for combo in itertools.product(*values):
            arg = {k: v for k, v in zip(keys, combo)}
            args_list.append(arg)
        # Prepare YAML data
        yaml_data = {
            "device-id": 0,
            "exclusive-mode": False,
            "upload-to-server": True,
            "trace-collect": {"enable": False},
            "profile-collect": {"enable": True},
            "repeat-times": 1,
            "cases": [
                {
                    "title": t,
                    "suite": suite,
                    "directory": directory,
                    "command": command,
                    "args": args_list
                }
            ]
        }
        # Use template_str if needed (for now, we just dump yaml_data)
        out_dir = os.path.join(output_folder_path, suite)
        os.makedirs(out_dir, exist_ok=True)
        out_path = os.path.join(out_dir, f"{t}.yaml")
        with open(out_path, "w") as f:
            yaml.dump(yaml_data, f, sort_keys=False)
        print(f"[{idx+1:02d}/{n_job_counts:02d}] {suite}/{t}: saved to {out_path}. {len(args_list)} arg settings.")

# Example usage (uncomment to test)
JOB_DICT = {
    "polybench": {
        "2dconv": "/home/enze/workspace/GPU_Benchmarks/polybench/2DCONV/2DConvolution.exe",
        "2mm": "/home/enze/workspace/GPU_Benchmarks/polybench/2MM/2mm.exe",
        "3dconv": "/home/enze/workspace/GPU_Benchmarks/polybench/3DCONV/3DConvolution.exe",
        "3mm": "/home/enze/workspace/GPU_Benchmarks/polybench/3MM/3mm.exe",
        "atax": "/home/enze/workspace/GPU_Benchmarks/polybench/ATAX/atax.exe",
        "bicg": "/home/enze/workspace/GPU_Benchmarks/polybench/BICG/bicg.exe",
        "gemm": "/home/enze/workspace/GPU_Benchmarks/polybench/GEMM/gemm.exe",
        "gesummv": "/home/enze/workspace/GPU_Benchmarks/polybench/GESUMMV/gesummv.exe",
        "mvt": "/home/enze/workspace/GPU_Benchmarks/polybench/MVT/mvt.exe",
        "syrk": "/home/enze/workspace/GPU_Benchmarks/polybench/SYRK/syrk.exe",
    },
    "rodinia": {
        "b+tree": "/home/enze/workspace/GPU_Benchmarks/rodinia/b+tree/b+tree.exe",
        "lavamd": "/home/enze/workspace/GPU_Benchmarks/rodinia/lavaMD/lavaMD.exe",
        "backprop": "/home/enze/workspace/GPU_Benchmarks/rodinia/backprop/backprop.exe",
        "kmeans": "/home/enze/workspace/GPU_Benchmarks/rodinia/kmeans/kmeans.exe",
        "bfs": "/home/enze/workspace/GPU_Benchmarks/rodinia/bfs/bfs.exe",
        "nn": "/home/enze/workspace/GPU_Benchmarks/rodinia/nn/nn.exe",
        "gaussian": "/home/enze/workspace/GPU_Benchmarks/rodinia/gaussian/gaussian.exe",
        "lavaMD": "/home/enze/workspace/GPU_Benchmarks/rodinia/lavaMD/lavaMD.exe",
        "b+tree": "/home/enze/workspace/GPU_Benchmarks/rodinia/b+tree/b+tree.exe",
        "heartwall": "/home/enze/workspace/GPU_Benchmarks/rodinia/heartwall/heartwall.exe",
        "hotspot": "/home/enze/workspace/GPU_Benchmarks/rodinia/hotspot/hotspot.exe",
        "cfd": "/home/enze/workspace/GPU_Benchmarks/rodinia/cfd/cfd.exe",
        "pathfinder": "/home/enze/workspace/GPU_Benchmarks/rodinia/pathfinder/pathfinder.exe",
        "lud": "/home/enze/workspace/GPU_Benchmarks/rodinia/lud/lud.exe",
    },
    "gpu-benches": {
        "cuda-memcpy": "/home/enze/workspace/GPU_Benchmarks/gpu-benches/cuda-memcpy/cuda-memcpy.exe",
        "gpu-cache": "/home/enze/workspace/GPU_Benchmarks/gpu-benches/gpu-cache/gpu-cache.exe",
        "gpu-small-kernels": "/home/enze/workspace/GPU_Benchmarks/gpu-benches/gpu-small-kernels/gpu-small-kernels.exe",
        "gpu-l2-cache": "/home/enze/workspace/GPU_Benchmarks/gpu-benches/gpu-l2-cache/gpu-l2-cache.exe",
        "gpu-stream": "/home/enze/workspace/GPU_Benchmarks/gpu-benches/gpu-stream/gpu-stream.exe",
        "gpu-l2-stream": "/home/enze/workspace/GPU_Benchmarks/gpu-benches/gpu-l2-stream/gpu-l2-stream.exe",
        "gpu-strides": "/home/enze/workspace/GPU_Benchmarks/gpu-benches/gpu-strides/gpu-strides.exe",
    },
    "simtune" : {
        "emptykernel": "/home/enze/workspace/GPU_Benchmarks/simtune/emptykernel/emptykernel.exe",
        "ffmakernel": "/home/enze/workspace/GPU_Benchmarks/simtune/ffmakernel/ffmakernel.exe",
        "ffmakernel-large": "/home/enze/workspace/GPU_Benchmarks/simtune/ffmakernel-large/ffmakernel-large.exe",
        "ffmakernel-large-large": "/home/enze/workspace/GPU_Benchmarks/simtune/ffmakernel-large-large/ffmakernel-large-large.exe",
        "pointerchasing-l1": "/home/enze/workspace/GPU_Benchmarks/simtune/pointerchasing-l1/pointerchasing-l1.exe",
        "pointerchasing-l2": "/home/enze/workspace/GPU_Benchmarks/simtune/pointerchasing-l2/pointerchasing-l2.exe",
        "pointerchasing-dram": "/home/enze/workspace/GPU_Benchmarks/simtune/pointerchasing-dram/pointerchasing-dram.exe",
        "ldgekernel": "/home/enze/workspace/GPU_Benchmarks/simtune/ldgekernel/ldgekernel.exe",
    }
}

# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32, 48, 64, 96, 128, 192, 256, 384, 512]}
# PARAM_DICT = {"size": [524288], "order": [256, 512], "k": [1250, 2500, 5000, 10000, 20000, 40000]}
# PARAM_DICT = {"boxes1d": [10, 20, 30, 40], "n": [50, 100]}
# PARAM_DICT = {"thread": [16] + [32*i for i in range(1, 33)], "block": [2**i for i in range(17)]} # 5, 10
# PARAM_DICT = {"thread": [16, 32, 64, 128], "size": [256 * i for i in [3, 5, 6, 7, 9, 10, 11, 12, 13, 14, 15]]} # 5, 10
# PARAM_DICT = {"thread": [16], "size": [2**i for i in range(0,1)]} # 5, 10
# PARAM_DICT = {"block": [2, 4, 8], "size": [2**i for i in range(6, 13)]} # 5, 10
# PARAM_DICT = {"block": [256], "size": [100000 * (2 ** i) for i in range(0, 1)]} # 5, 10
# PARAM_DICT = {"block": [128, 256], "size": [10000 * (2 ** i) for i in range(0, 7)]} # 5, 10
# PARAM_DICT = {"block": [32], "size": [16]} # 5, 10
# PARAM_DICT = {"block": [32, 64, 128], "size": [2**i for i in range(3, 11)]} # 5, 10
# PARAM_DICT = {"block": [64, 128,256], "size": [10000, 15000, 20000, 30000, 40000, 60000, 80000, 120000, 160000]} # 5, 10

# 2mm:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32*i for i in range(1, 9)]} # 5, 10
# 3mm:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32*i for i in range(1, 9)]} # 5, 10
# 3dconv:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [16*i for i in range(1, 9)]}
# 2dconv:
# PARAM_DICT = {"block": [8, 16, 32], "size": [32, 48, 64, 96, 128, 192, 256, 384, 512]} # 5, 10
# atax:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32*i for i in range(1, 9)]} # 5, 10
# bicg:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32*i for i in range(1, 9)]}
# mvt:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32*i for i in range(1, 9)]}
# gemm:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32*i for i in range(1, 9)]}
# syrk:
# PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32*i for i in range(1, 9)]}
# gesummv:
# PARAM_DICT = {"block": [8, 16, 32], "size": [32, 48, 64, 96, 128, 192, 256, 384, 512]} # 5, 10

# kmeans:
# PARAM_DICT = {"clusters": [16, 24, 32], "size": [64*i for i in range(2, 15)]} # 5, 10
# backprop:
# PARAM_DICT = {"block": [32, 64, 128], "size": [64*i for i in range(2, 12)]} # 5, 10
# bfs:
# PARAM_DICT = {"degree": [3, 4, 5], "size": [2**i for i in range(3, 11)]} # 5, 10
# gaussian:
# PARAM_DICT = {"block": [32, 64, 128], "size": [128*i for i in range(1, 9)]} # 5, 10
# lavaMD:
# PARAM_DICT = {"boxes1d": [256]} # 5, 10
# PARAM_DICT = {"boxes1d": [1*i for i in range(1, 11)]} # 5, 10
# PARAM_DICT = {"boxes1d": [2]} # 5, 10
# b+tree:
# PARAM_DICT = {"block": [128, 256], "size": [1000*i for i in range(1,7)]} 
# heartwall:
# PARAM_DICT = {"block": [128, 256], "size": [32, 48, 64, 96, 128, 192, 256, 384, 512, 768, 1024]} 
# hotspot: [failed to run]
# PARAM_DICT = {"block": [128], "size": [2048]}
# PARAM_DICT = {"block": [128, 256], "size": [32, 48, 64, 96, 128, 192, 256, 384, 512, 768, 1024]}
# cfd:
# PARAM_DICT = {"block": [128], "size": [1024]}
# PARAM_DICT = {"block": [64, 128], "size": [64] + [64*i for i in range(2, 11)]}
# pathfinder:
# PARAM_DICT = {"block": [64, 128], "size": [4*i for i in range(3, 12)]}
# PARAM_DICT = {"block": [128, 256], "size": [64] + [128*i for i in range(1, 9)]}
# lud:
# PARAM_DICT = {"size": [8] + [16*i for i in range(1, 9)]}
# PARAM_DICT = {"block": [128, 256], "size": [64] + [128*i for i in range(1, 9)]}

# ffmakernel:
# PARAM_DICT = {"thread": [16,32,64] + [128*i for i in range(1, 9)], "block": [4 ** i for i in range(0, 9)]} # 5, 10
# PARAM_DICT = {"thread": [32], "block": [1], "iters": [10, 20, 30, 40]} # 5, 10
# PARAM_DICT = {"thread": [32, 64, 128], "block": [1, 2, 3], "iters": [10, 20, 30, 40]} # 5, 10
# PARAM_DICT = {"thread": [32], "block": [1], "iters": [10]} # 5, 10
# PARAM_DICT = {"thread": [64, 128], "block": [32, 64, 128, 256, 384, 512, 640, 768, 896, 1024], "iters": [20, 40]} # 5, 10
# PARAM_DICT = {"thread": [16] + [32*i for i in range(1, 33)], "block": [2**i for i in range(17)]} # 5, 10
# ffmakernel-large:
# PARAM_DICT = {"thread": [32, 64, 128], "block": [100, 200, 300], "iters": [10, 20, 30, 40]} # 5, 10
# ffmakernel-large-large:
# PARAM_DICT = {"thread": [128], "block": [200*i for i in range(1, 31)], "iters": [10, 20]} # 5, 10
# PARAM_DICT = {"thread": [1024], "block": [200*i for i in range(1, 31)], "iters": [1]} # 5, 10
# emptykernel:
# PARAM_DICT = {"thread": [32*i for i in range(8, 33)], "block": [2**i for i in range(10, 17)]} # 5, 10


# ldgekernel:
PARAM_DICT = {"size": [8192, 524288, 8388608], "iters": [4, 6, 8, 10, 12, 14, 16]}

def add_middle(orginial_list):
    new_list = []
    n = len(orginial_list)
    for i in range(n - 1):
        new_list.append(orginial_list[i])
        mid_value = (orginial_list[i] + orginial_list[i + 1]) // 2
        if mid_value not in orginial_list:
            new_list.append(mid_value)
    new_list.append(orginial_list[-1])
    return new_list

# pointerchasing-l1:
# PARAM_DICT = {"block": [1], "thread": [32], "size": add_middle([512, 1024, 2048, 4096, 8192, 16384]), "iters": [10000]}
# pointerchasing-l2:
# PARAM_DICT = {"block": [1], "thread": [32], "size": add_middle([1 << 16, 1 << 17, 1 << 18, 1 << 19, 1 << 20, 1 << 21]), "iters": [10000]}
# pointerchasing-dram:
# PARAM_DICT = {"block": [1], "thread": [32], "size": add_middle([1 << 24, 1 << 25, 1 << 26, 1 << 27, 1 << 28, 1 << 29]), "iters": [10000]}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Build YAML job files for benchmarks.")
    parser.add_argument("--suite", type=str, default="None", help="Suite name (e.g., polybench)")
    parser.add_argument("--title", type=str, default="ALL", help="Benchmark title (e.g., 2dconv or ALL)")
    parser.add_argument("--output-folder", type=str, default="./etc", help="Output folder path")
    args = parser.parse_args()
    
    try:
        build_jobs(JOB_DICT, PARAM_DICT, args.suite, args.title, args.output_folder)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)