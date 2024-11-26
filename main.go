package main

import (
	"io"
	"os"

	"github.com/sarchlab/mnt-collector/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	initLogSettings()

	cmd.Execute()
}

func initLogSettings() {
	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	multiWriter := io.MultiWriter(file, os.Stdout)

	log.SetOutput(multiWriter)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}
