package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	projectRoot string
	hostName    string
)

func ProjectRoot() string {
	if projectRoot == "" {
		log.Panic("projectRoot is not initialized")
	}
	return projectRoot
}

func HostName() string {
	if hostName == "" {
		log.Panic("hostName is not initialized")
	}
	return hostName
}

func prepareBasicEnvirons() {
	projectRoot = getProjectRoot()
	hostName = getHostName()
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
