package collector

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/sarchlab/mnt-collector/config"
)

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
		log.WithField("error", nvml.ErrorString(ret)).Fatal("Unable to initialize NVML")
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.WithField("error", nvml.ErrorString(ret)).Fatal("Unable to shutdown NVML")
		}
	}()

	var err nvml.Return
	device, err = nvml.DeviceGetHandleByIndex(deviceID)
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get handle for device")
	}

	deviceName, err = device.GetName()
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get name of device")
	}
	log.WithFields(log.Fields{
		"deviceID": deviceID,
		"device":   deviceName,
	}).Info("Device loading")

	deviceMustIdle(device)
	setVisibleDevices(deviceID)
	if config.C.Root {
		setExclusiveMode(device)
		log.Info("Exclusive mode set")
	} else {
		log.Warn("Not running under exclusive mode")
	}
	log.WithFields(log.Fields{
		"deviceID": deviceID,
		"device":   deviceName,
	}).Info("Device is ready")
}

func deviceMustIdle(device nvml.Device) {
	utilization, err := device.GetUtilizationRates()
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get utilization rates of device")
	}
	if utilization.Gpu != 0 {
		log.Panic("Device is not idle")
	}
}

func setExclusiveMode(device nvml.Device) {
	mode, err := device.GetComputeMode()
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get compute mode of device")
	}
	if mode != nvml.COMPUTEMODE_EXCLUSIVE_PROCESS {
		err = device.SetComputeMode(nvml.COMPUTEMODE_EXCLUSIVE_PROCESS)
		if err != nvml.SUCCESS {
			log.WithField("error", nvml.ErrorString(err)).Panic("Failed to set compute mode of device")
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
