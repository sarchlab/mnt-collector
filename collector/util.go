package collector

import (
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

func runCommandWithTimer(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		log.WithError(err).Error("Failed to start command")
		return err
	}
	log.WithFields(log.Fields{
		"cmd":  cmd.Path,
		"args": cmd.Args,
		"env":  cmd.Env,
	}).Debug("Command started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\rTime elapsed: %.0f seconds", elapsed)
		case err := <-done:
			if err != nil {
				log.WithError(err).Error("Command finished with error")
				return err
			}
			fmt.Println("\nCommand finished successfully.")
			log.WithField("elapsed", time.Since(startTime)).Info("Command finished")
			return nil
		}
	}
}
