package collector

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func MannualCleanTraceDirectory() {
	
}

func DeleteTempProfileFiles(filenames []string) {
	for _, filename := range filenames {
		file := filepath.Join(profilesDir, filename)
		if err := os.Remove(file); err != nil {
			log.WithField("file", file).WithError(err).Error("Failed to remove temporary file")
		}
	}
}

func DeleteTempTraceDirectory(dirname string) {
	dir := filepath.Join(tracesDir, dirname)
	if err := os.RemoveAll(dir); err != nil {
		log.WithField("directory", dir).WithError(err).Error("Failed to remove temporary directory")
	}
}
