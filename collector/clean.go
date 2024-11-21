package collector

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func DeleteLocalFiles(filepaths []string) error {
	var lastErr error
	for _, path := range filepaths {
		if err := os.Remove(path); err != nil {
			log.WithField("file", path).WithError(err).Error("Failed to remove temporary file")
			lastErr = err
		}
	}
	return lastErr
}

func DeleteLocalDir(dirpath string) error {
	err := os.RemoveAll(dirpath)
	if err != nil {
		log.WithField("directory", dirpath).WithError(err).Error("Failed to remove temporary directory")
	}
	return err
}
