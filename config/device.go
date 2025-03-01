package config

import (
	"fmt"
	"strings"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

var (
	device            nvml.Device
	deviceName        string
	computeCapability string
	frequency         uint32
	maxFrequency      uint32
	cudaVersion 	  string
)

func DeviceName() string {
	if deviceName == "" {
		log.Panic("deviceName is not initialized")
	}
	return deviceName
}

func ComputeCapability() string {
	if computeCapability == "" {
		log.Panic("computeCapability is not initialized")
	}
	return computeCapability
}

func Frequency() uint32 {
	if frequency == 0 {
		log.Panic("frequency is not initialized")
	}
	return frequency
}

func MaxFrequency() uint32 {
	if maxFrequency == 0 {
		log.Panic("maxFrequency is not initialized")
	}
	return maxFrequency
}

func Device() nvml.Device {
	if device == nil {
		log.Panic("device is not initialized")
	}
	return device
}

func CudaVersion() string {
	if cudaVersion == "" {
		log.Panic("cudaVersion is not initialized")
	}
	return cudaVersion
}

func LoadDevice(deviceID int) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(ret)).Fatal("Unable to initialize NVML")
	}

	var err nvml.Return
	device, err = nvml.DeviceGetHandleByIndex(deviceID)
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get handle for device")
	}


	cudaVersion = getCudaVersion()
	initDeviceName(device)
	initComputeCapability(device)
	initFrequency(device)

	deviceMustIdle(device)
	if C.ExclusiveMode {
		setExclusiveMode(device)
		log.Info("Exclusive mode set")
	} else {
		log.Warn("Not running under exclusive mode")
	}
	log.WithField("deviceID", deviceID).Info("Device is ready")
}

func ShutdownDevice() {
	ret := nvml.Shutdown()
	if ret != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(ret)).Fatal("Unable to shutdown NVML")
	}
}

func initDeviceName(device nvml.Device) {
	var err nvml.Return
	deviceName, err = device.GetName()
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get name of device")
	}
	log.WithField("deviceName", deviceName).Info("Device loading")
}

func initComputeCapability(device nvml.Device) {
	v1, v2, err := device.GetCudaComputeCapability()
	computeCapability = fmt.Sprintf("%d.%d", v1, v2)
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get compute capability of device")
	}
	log.WithField("computeCapability", computeCapability).Info("Device loading")
}

func initFrequency(device nvml.Device) {
	var err nvml.Return
	frequency, err = device.GetClockInfo(nvml.CLOCK_SM)
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get current frequency of device")
	}
	maxFrequency, err = device.GetMaxClockInfo(nvml.CLOCK_SM)
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get max frequency of device")
	}
	log.WithFields(log.Fields{
		"frequency":    frequency,
		"maxFrequency": maxFrequency,
	}).Info("Device loading")
}

func deviceMustIdle(device nvml.Device) {
	utilization, err := device.GetUtilizationRates()
	if err != nvml.SUCCESS {
		log.WithField("error", nvml.ErrorString(err)).Panic("Failed to get utilization rates of device")
	}
	if utilization.Gpu != 0 {
		log.Warn("Device is not idle")
		// log.Panic("Device is not idle")
	}
}

func getCudaVersion() string {
	cmd := exec.Command("nvcc", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithError(err).Panic("could not get CUDA version")
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "release") {
			version := strings.Split(line, ",")[1]
			return strings.TrimSpace(version)
		}
	}

	log.Panic("could not find CUDA version")
	return ""
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
