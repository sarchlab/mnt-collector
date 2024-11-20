package mntbackend

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func checkHealth() error {
	url := fmt.Sprintf("%s/health", URLBase)
	client := http.Client{
		Timeout: 5 * time.Second, // Set a timeout for the request
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status", resp.Status).Warn("mnt backend is not healthy")
		return ErrorNotHealthy
	}

	return nil
}
