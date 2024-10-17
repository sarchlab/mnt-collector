package config

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var C *SimSetting
var SC *SecretConfig

func LoadSimSetting(file string) {
	C.load(file)
}

func LoadSecretConfig(file string) {
	SC.load(file)
}

var (
	projectRoot string
	hostName    string
	cudaVersion string
	nsysVersion string
)

func ProjectRoot() string {
	return projectRoot
}

func HostName() string {
	return hostName
}

func CudaVersion() string {
	return cudaVersion
}

func NsysVersion() string {
	return nsysVersion
}

var (
	tracerToolSo        string
	tracerToolProcessor string
)

func TracerToolSo() string {
	return tracerToolSo
}

func TracerToolProcessor() string {
	return tracerToolProcessor
}

func init() {
	projectRoot = getProjectRoot()
	hostName = getHostName()
	cudaVersion = getCudaVersion()
	nsysVersion = getNsysVersion()

	C = &SimSetting{}
	SC = &SecretConfig{}

	tracerToolSo = filepath.Join(projectRoot, "lib/tracer_tool.so")
	tracerToolProcessor = filepath.Join(projectRoot, "lib/post-traces-processing")
	fileMustExist(tracerToolSo)
	fileMustExist(tracerToolProcessor)
}

func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	var root string
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			root = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Panic("config: could not find project root")
		}
		dir = parent
	}
	return root
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		log.Panic(err)
	}
	return hostName
}

func getCudaVersion() string {
	cmd := exec.Command("nvcc", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(err)
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "release") {
			version := strings.Split(line, ",")[1]
			return strings.TrimSpace(version)
		}
	}

	log.Panic("config: could not find CUDA version")
	return ""
}

func getNsysVersion() string {
	cmd := exec.Command("nsys", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(err)
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "NVIDIA Nsight Systems") {
			version := strings.Split(line, " ")[3]
			return version
		}
	}

	log.Panic("config: could not find Nsight Systems version")
	return ""
}

func fileMustExist(file string) {
	_, err := os.Stat(file)
	if err != nil {
		log.Panic(err)
	}
}
