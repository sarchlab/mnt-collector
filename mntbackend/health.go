package mntbackend

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

var ErrorNotHealthy = errors.New("mnt backend is not healthy")

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
		log.Printf("mnt backend is not healthy: %s", resp.Status)
		return ErrorNotHealthy
	}

	return nil
}
