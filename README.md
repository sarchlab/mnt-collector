### Feat
- [x] use mnt-backend's request type
- [ ] DeviceSetGpuLockedClocks
- [ ] DeviceSetPersistenceMode
- [ ] profile to the same file

### Environment Check List
- gpu
- nsight compute
- lib/tracer_tool.so
- lib/post-traces-processing

### Run
```
Data Collector for the MGPUSim NVIDIA Trace Project.
[Commands]
traces
profiles
simulations
[Flags]
-collect to specify the collection settings file
-secret to specify the secret tokens file
[Example]
./mnt-collector simulations --collect etc/simulations.yaml 

Usage:
  mnt-collector [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  profiles    Use Nvidia system to profile the cases and upload the data to database & cloud.
  simulations Use the given simulator to run traces and upload the data to database.
  traces      Use Nvbit to generate traces and upload the data to database & cloud.
  delete      Connect to both mongodb and s3 to delete specified entries together.

Flags:
      --collect string   yaml file that store collection settings (default is etc/collects.yaml) (default "ect/collects.yaml")
  -h, --help             help for mnt-collector
      --secret string    yaml file that store secret tokens (default is etc/secrets.yaml) (default "etc/secrets.yaml")
      --machine string       machine name filter, for delete only (required)
      --cuda-version string  CUDA version filter, for delete only (required)
      --suite string         suite name filter, for delete only (optional, default is "all")
      --benchmark string     benchmark title filter, for delete only (optional, default is "all")
  -t, --toggle           Help message for toggle

Use "mnt-collector [command] --help" for more information about a command.
```


### Schedule [Run profiles + traces + simulations together]
To run `mnt-collector` with profiles, traces, and simulations together, use the Python script `schedule/schedule.py`.

#### Steps

1. **Create a Configuration File**  
   Create a new YAML file under the path `etc/{suite}/{benchmark}.yaml`. For example, `etc/polybench/atax.yaml`. Below is an example of the content:

   ```yaml
   # Environment
   device-id: 1
   exclusive-mode: false

   # Simulation Config
   upload-to-server: true
   trace-collect: 
     enable: false
   profile-collect:
     enable: true
   repeat-times: 3

   # Benchmark Details
   cases:
     - title: atax
       suite: polybench
       directory: /home/exu03/workspace/GPU_Benchmarks/polybench/ATAX/
       command: /home/exu03/workspace/GPU_Benchmarks/polybench/ATAX/atax.exe
       args:
       - size: 32
       - size: 64
   ```

2. **Remove Existing Trace ID File**  
   Ensure there is no file at the path `traceid/{suite}-{benchmark}.txt`. For example, `traceid/polybench-atax.txt`. If it exists, delete it.

3. **Run the Schedule Script**  
   Activate your Python environment and execute the script with the configuration file:
   ```shell
   python schedule/schedule.py --collect etc/{suite}/{benchmark}.yaml
   # Example: python schedule/schedule.py --collect etc/polybench/atax.yaml
   ```

4. **Check Progress Logs**  
   Monitor the progress of profiles, traces, and simulations in the log file `etc/{suite}/{benchmark}-log.csv`. For example, `etc/polybench/atax-log.csv`.

### Delete
The `delete` command allows you to remove specific entries from both MongoDB and S3 storage based on the provided filters.

#### Usage
```shell
mnt-collector delete --machine <machine_name> --cuda-version <cuda_version> [--suite <suite_name>] [--benchmark <benchmark_title>]
```