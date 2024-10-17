package collector

import (
	"C"
	"fmt"
	"log"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)
import "os"

var (
	device             nvml.Device
	deviceName         string
	computerCapability int
)

func DeviceName() string {
	return deviceName
}

func ComputeCapability() int {
	return computerCapability
}

func LoadDevice(deviceID int) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()

	var err nvml.Return
	device, err = nvml.DeviceGetHandleByIndex(deviceID)
	if err != nvml.SUCCESS {
		log.Panicf("Failed to get handle for device %d: %v", deviceID, nvml.ErrorString(err))
	}

	deviceName, err = device.GetName()
	if err != nvml.SUCCESS {
		log.Panicf("Failed to get name of device %d: %v", deviceID, nvml.ErrorString(err))
	}
	fmt.Printf("Loading device %d: %s\n", deviceID, deviceName)

	deviceMustIdle(device)
	setExclusiveMode(device)
	setVisibleDevices(deviceID)
	fmt.Printf("Device %d is ready\n", deviceID)
}

func deviceMustIdle(device nvml.Device) {
	utilization, err := device.GetUtilizationRates()
	if err != nvml.SUCCESS {
		log.Panicf("Failed to get utilization rates of device: %v", nvml.ErrorString(err))
	}
	if utilization.Gpu != 0 {
		log.Panic("Device is not idle")
	}
}

func setExclusiveMode(device nvml.Device) {
	mode, err := device.GetComputeMode()
	if err != nvml.SUCCESS {
		log.Panicf("Failed to get compute mode of device: %v", nvml.ErrorString(err))
	}
	if mode != nvml.COMPUTEMODE_EXCLUSIVE_PROCESS {
		err = device.SetComputeMode(nvml.COMPUTEMODE_EXCLUSIVE_PROCESS)
		if err != nvml.SUCCESS {
			log.Panicf("Failed to set compute mode of device: %v", nvml.ErrorString(err))
		}
	}
}

// fix: check if this is correct
func setVisibleDevices(deviceID int) {
	err := os.Setenv("CUDA_VISIBLE_DEVICES", fmt.Sprintf("%d", deviceID))
	if err != nil {
		log.Panic(err)
	}
}
