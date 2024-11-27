package config

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var (
	tracerToolSo        string
	tracerToolProcessor string
)

func TracerToolSo() string {
	if tracerToolSo == "" {
		log.Panic("tracerToolSo is not initialized")
	}
	return tracerToolSo
}

func TracerToolProcessor() string {
	if tracerToolProcessor == "" {
		log.Panic("tracerToolProcessor is not initialized")
	}
	return tracerToolProcessor
}

func PrepareTracesEnvirons() {
	tracerToolSo = filepath.Join(projectRoot, "lib/tracer_tool.so")
	tracerToolProcessor = filepath.Join(projectRoot, "lib/post-traces-processing")
	fileMustExist(tracerToolSo)
	fileMustExist(tracerToolProcessor)
}
