package main

import (
	"io"
	"os"

	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/collector"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	initLogSettings()

	config.LoadDevice(config.C.DeviceID)
	log.Info("Device loaded.")

	mntbackend.Init()
	log.Info("MNT backend connected.")

	aws.Init()
	log.Info("AWS connected.")

	log.Info("Start collecting data.")
	collector.Run()
	log.Info("Program finished.")
}

func initLogSettings() {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "logfile.log",
		MaxSize:    2,
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
