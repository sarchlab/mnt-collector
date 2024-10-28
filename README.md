### Todo List
- [ ] use nsight-compute to collect gold data at local, upload it to mnt-backend
- [ ] use nvbit-tracer to generate traces, upload it to s3 and mnt-backend

### Feat
- [ ] use mnt-backend's request type
- [x] os env to set config `export SECRET_FILE` `export SIM_SETTING_FILE`
    - [x] DeviceExclusiveMode (need root permission)
- [ ] DeviceSetGpuLockedClocks
- [ ] DeviceSetPersistenceMode
- [x] logrus
- [x] fix the param part 
    - [x] sync with backend
- [ ] profile to the same file

### Environment Check List
- gpu
- nsight compute
- lib/tracer_tool.so
- lib/post-traces-processing

### Config

```
export SECRET_FILE=etc/secret-local.yaml
export SIM_SETTING_FILE=etc/simsetting-local.yaml
```