package collector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

const (
	tracesDir   = "/tmp/mnt-collector/traces/"
	tracesDirS3 = "traces/"
)

func runTraceCollect(cases []Case) {
	err := os.MkdirAll(tracesDir, 0755)
	if err != nil {
		log.WithError(err).Error("Failed to create trace directory")
	}

	for _, c := range cases {
		log.WithFields(log.Fields{
			"Title":   c.Title,
			"Suite":   c.Suite,
			"Command": c.Command,
			"Params":  c.ParamStr,
		}).Info("Start trace collection")

		traceDir, err := generateTrace(c)
		if err != nil {
			log.WithError(err).Error("Failed to generate trace")
			continue
		}

		log.Info("Start processing trace")
		processTrace(traceDir)

		if config.C.UploadToServer {
			log.Info("Start uploading to server")

			s3Path := storeTraceToS3(traceDir)
			log.WithField("s3Path", s3Path).Info("Trace stored to S3")

			uploadTraceToDB(c, s3Path)
		} else {
			log.Info("Skip uploading to server")
		}
	}
}

func generateTrace(c Case) (string, error) {
	args := strings.Split(c.ParamStr, " ")
	dir, err := os.MkdirTemp(tracesDir, "trace-*")
	if err != nil {
		log.WithError(err).Error("Failed to create trace directory")
		return "", err
	}

	getCmd := func() *exec.Cmd {
		cmd := exec.Command(c.Command, args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("LD_PRELOAD=%s", config.TracerToolSo()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CUDA_VISIBLE_DEVICES=%d", config.C.DeviceID))
		cmd.Env = append(cmd.Env, "USER_DEFINED_FOLDERS=1")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TRACES_FOLDER=%s", dir))
		return cmd
	}

	log.Info("Start generating trace")
	for err := runGPUCmdWithTimer(getCmd()); err != nil; {
		if err == ErrorInterrupted {
			log.Warn("Interrupted, retry trace collection")
			waitTillDeviceIdle()
			err = runGPUCmdWithTimer(getCmd())
		} else {
			log.WithError(err).Error("Failed to generate trace")
			return "", err
		}
	}

	return dir, nil
}

func uploadTraceToDB(c Case, s3Path string) {
	req := mntbackend.TraceRequest{
		EnvID:     mntbackend.EnvID(),
		Suite:     c.Suite,
		Benchmark: c.Title,
		Param:     c.param,
		S3Path:    s3Path,
	}
	traceID, err := mntbackend.UploadTrace(req)
	if err != nil {
		log.WithFields(log.Fields{
			"Case":   c,
			"S3Path": s3Path,
		}).WithError(err).Error("Failed to upload trace")
	} else {
		log.WithField("TraceID", traceID).Info("Trace uploaded")
	}
}

func storeTraceToS3(traceDir string) string {
	base, err := filepath.Rel(tracesDir, traceDir)
	if err != nil {
		log.WithError(err).Error("Failed to get relative path")
		return ""
	}

	objectPath := filepath.Join(tracesDirS3, base)
	aws.UploadDirectoryAsObjects(objectPath, traceDir)

	return objectPath
}

func processTrace(dir string) {
	cmd := exec.Command(config.TracerToolProcessor(), fmt.Sprintf("%s/kernelslist", dir))
	if err := runNormalCmdWithTimer(cmd); err != nil {
		log.WithError(err).Error("Failed to run traces processor")
		return
	}

	log.Info("Deleting trace files")
	cmd = exec.Command("bash", "-c", fmt.Sprintf("rm %s/*.trace", dir))
	err := cmd.Run()
	if err != nil {
		log.WithField("dir", dir).WithError(err).Error("Failed to remove trace files")
		return
	}

	cmd = exec.Command("rm", fmt.Sprintf("%s/kernelslist", dir))
	err = cmd.Run()
	if err != nil {
		log.WithError(err).Error("Failed to remove kernelslist")
		return
	}
}
