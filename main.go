package main

import (
	"io"
	"os"

	"github.com/sarchlab/mnt-collector/cmd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	initLogSettings()

	cmd.Execute()
}

func initLogSettings() {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "logfile.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     30,
	}
	multiWriter := io.MultiWriter(lumberjackLogger, os.Stdout)

	log.SetOutput(multiWriter)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}
