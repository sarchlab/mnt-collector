### Todo List
- [ ] use nsight-compute to collect gold data at local, upload it to mnt-backend
- [ ] use nvbit-tracer to generate traces, upload it to s3 and mnt-backend

### feat
- [ ] use mnt-backend's request type
- [x] os env to set config `export SECRET_FILE` `export SIM_SETTING_FILE`
    - [x] DeviceExclusiveMode (need root permission)
- [ ] DeviceSetGpuLockedClocks
- [ ] DeviceSetPersistenceMode
- [x] logrus

### Environment Check List
- gpu
- nsight compute
- lib/tracer_tool.so
- lib/post-traces-processing