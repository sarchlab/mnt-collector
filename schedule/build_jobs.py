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
            "repeat-times": 3,
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
    }
}

PARAM_DICT = {"blockDimX": [8, 16, 32], "size": [32, 48, 64, 96, 128, 192, 256]}

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