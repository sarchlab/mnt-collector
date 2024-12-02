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

Usage:
  mnt-collector [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  profiles    Use Nvidia system to profile the cases and upload the data to database & cloud.
  simulations Use the given simulator to run traces and upload the data to database.
  traces      Use Nvbit to generate traces and upload the data to database & cloud.
```