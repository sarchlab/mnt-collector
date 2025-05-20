package collector

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/sarchlab/mnt-collector/config"
	log "github.com/sirupsen/logrus"
)

var ErrorInterrupted = errors.New("interrupted by other process")

func runNormalCmdWithTimer(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		log.WithError(err).Error("Failed to start command")
		return err
	}

	log.WithFields(log.Fields{
		"cmd":  cmd.Path,
		"args": cmd.Args,
		//"env":  cmd.Env,
	}).Debug("Command started in runNormalCmdWithTimer")

	oneSecondTicker := time.NewTicker(1 * time.Second)
	defer oneSecondTicker.Stop()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	startTime := time.Now()

	for {
		select {
		case <-oneSecondTicker.C:
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

func runGPUCmdWithTimer(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		log.WithError(err).Error("Failed to start command")
		return err
	}

	log.WithFields(log.Fields{
		"cmd":  cmd.Path,
		"args": cmd.Args,
		//"env":  cmd.Env,
	}).Debug("Command started in runGPUCmdWithTimer")

	oneSecondTicker := time.NewTicker(1 * time.Second)
	defer oneSecondTicker.Stop()
	threeSecondsTicker := time.NewTicker(3 * time.Second)
	defer threeSecondsTicker.Stop()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	startTime := time.Now()

	for {
		select {
		case <-threeSecondsTicker.C:
			processes, err := nvml.DeviceGetComputeRunningProcesses(config.Device())
			if err != nvml.SUCCESS {
				log.WithError(err).Error("Failed to get running processes")
				return err
			}
			if len(processes) > 1 {
				log.Warn("Other process detected, interrupting")
				if err := cmd.Process.Kill(); err != nil {
					log.WithError(err).Error("Failed to kill process")
				}
				return ErrorInterrupted
			}
		case <-oneSecondTicker.C:
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

func waitTillDeviceIdle() {
	startTime := time.Now()
	for {
		processes, err := nvml.DeviceGetComputeRunningProcesses(config.Device())
		if err != nvml.SUCCESS {
			log.WithError(err).Panic("Failed to get running processes")
			return
		}
		if len(processes) == 0 {
			log.Info("Device is idle")
			return
		}
		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("\rTime elapsed: %.0f seconds", elapsed)
		time.Sleep(5 * time.Second)
	}
}
