package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var C *SimSetting
var SC *SecretConfig

var (
	projectRoot string
	hostName    string
	cudaVersion string
	nsysVersion string
)

var (
	tracerToolSo        string
	tracerToolProcessor string
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

	loadSimSetting()
	loadSecretConfig()

	tracerToolSo = filepath.Join(projectRoot, "lib/tracer_tool.so")
	tracerToolProcessor = filepath.Join(projectRoot, "lib/post-traces-processing")
	fileMustExist(tracerToolSo)
	fileMustExist(tracerToolProcessor)
}

func loadSimSetting() {
	file := "etc/simsetting.yaml"
	if _, flag := os.LookupEnv("SIM_SETTING_FILE"); flag {
		file = os.Getenv("SIM_SETTING_FILE")
	}
	log.WithField("file", file).Info("Loading sim setting")
	C.load(file)
}

func loadSecretConfig() {
	file := "etc/secret.yaml"
	if _, flag := os.LookupEnv("SECRET_FILE"); flag {
		file = os.Getenv("SECRET_FILE")
	}
	log.WithField("file", file).Info("Loading secret config")
	SC.load(file)
}

func getProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		log.WithError(err).Panic("could not get current working directory")
	}

	var root string
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			root = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Panic("could not find project root")
		}
		dir = parent
	}
	return root
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		log.WithError(err).Warn("could not get hostname")
	}
	return hostName
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

func getNsysVersion() string {
	cmd := exec.Command("nsys", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithError(err).Panic("could not get Nsight Systems version")
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "NVIDIA Nsight Systems") {
			version := strings.Split(line, " ")[3]
			return version
		}
	}

	log.Panic("could not find Nsight Systems version")
	return ""
}

func fileMustExist(file string) {
	_, err := os.Stat(file)
	if err != nil {
		log.WithField("file", file).WithError(err).Panic("file does not exist")
	}
}
