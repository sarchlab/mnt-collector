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

Flags:
      --collect string   yaml file that store collection settings (default is etc/collects.yaml) (default "ect/collects.yaml")
  -h, --help             help for mnt-collector
      --secret string    yaml file that store secret tokens (default is etc/secrets.yaml) (default "etc/secrets.yaml")
  -t, --toggle           Help message for toggle

Use "mnt-collector [command] --help" for more information about a command.
```


### Schedule
To run mnt-collector with profiles, traces and simulations together, you may use python script schedule/schedule.py

#### Example

- [Step 1] Create a new yaml file under path `etc/{suite}/{bemnchmark}.yaml`, e.g., `etc/polybench/atax.yaml`. Example for the content: 

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

- [Step 2] Check currently there is no file with path `traceid/{suite}-{benchmark}.txt`, e.g., `traceid/polybench-atax.txt`. Remove it if exists.

- [Step 3] Activate your python environment and then run:
```shell
python schedule/schedule.py --collect etc/{suite}/{benchmark}.yaml
# Example: python schedule/schedule.py --collect etc/polybench/atax.yaml
```

- [Step 4] You may check the progress logs of the profiles, traces and simulations processes in file `etc/{suite}/{bemnchmark}-log.csv`, e.g., `etc/polybench/atax-log.csv`.