package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sarchlab/mnt-collector/aws"
	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/model"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

const (
	traceRootLocal  = "./tmp/mnt-collector/traces/"
	traceRootRemote = "traces/"
)

func RunTraceCollection() {
	caseSettings := generateCaseSettings(config.C.Cases)

	err := os.MkdirAll(traceRootLocal, 0755)
	if err != nil {
		log.WithError(err).Error("Failed to create trace directory")
	}

	for _, c := range caseSettings {
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

		log.WithField("trace-dir", traceDir).Info("Start processing trace")
		processTrace(traceDir)
		createInfoFile(traceDir, c)

		if config.C.UploadToServer {
			traceSize := getDirSize(traceDir)
			log.Info("Start uploading to server")

			s3Path := storeTraceToS3(traceDir)
			log.WithField("s3Path", s3Path).Info("Trace stored to S3")

			// cleanOldTraceData(c)
			uploadTraceToDB(c, s3Path, traceSize)
		} else {
			log.Info("Skip uploading to server")
		}
	}
}

func generateTrace(c CaseSetting) (string, error) {
	args := strings.Split(c.ParamStr, " ")
	dir, err := os.MkdirTemp(traceRootLocal, "trace-*")
	if err != nil {
		log.WithError(err).Error("Failed to create trace directory")
		return "", err
	}

	getCmd := func() *exec.Cmd {
		cmd := exec.Command(c.Command, args...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("LD_PRELOAD=%s", config.TracerToolSo()))
		// cmd.Env = append(cmd.Env, fmt.Sprintf("LD_PRELOAD=%s", config.TracerToolSo()))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CUDA_VISIBLE_DEVICES=%d", config.C.DeviceID))
		cmd.Env = append(cmd.Env, "USER_DEFINED_FOLDERS=1")
		cmd.Env = append(cmd.Env, fmt.Sprintf("TRACES_FOLDER=%s", dir))
		fmt.Printf("cmd.Env: %v\n", cmd.Env)
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

func cleanOldTraceData(c CaseSetting) {
	key := model.CaseKey{
		EnvID:     mntbackend.EnvID(),
		Suite:     c.Suite,
		Benchmark: c.Title,
		Param:     c.param,
	}

	oldTrace, err := mntbackend.FindTrace(key)
	if mntbackend.IsObjectNotFound(err) {
		log.Info("Old trace not exist, skip deletion")
		return
	} else if err != nil {
		log.WithError(err).Error("Failed to find old trace")
		return
	}

	exist, err := aws.CheckObjectExist(oldTrace.S3Path)
	if err != nil {
		log.WithError(err).Error("Failed to check old trace")
		return
	}
	if exist {
		err = aws.DeleteObjectDirectory(oldTrace.S3Path)
		if err != nil {
			log.WithError(err).Error("Failed to delete old trace")
		} else {
			log.WithField("s3", oldTrace.S3Path).Info("Old trace deleted")
		}
	} else {
		log.Info("Remote old trace not exist, skip deletion")
	}

	reletivePath, err := filepath.Rel(traceRootRemote, oldTrace.S3Path)
	if err != nil {
		log.WithError(err).Error("Failed to get relative path")
		return
	}
	localDir := filepath.Join(traceRootLocal, reletivePath)
	err = DeleteLocalDir(localDir)
	if err != nil {
		log.WithError(err).Error("Failed to delete local trace")
	} else {
		log.WithField("localPath", localDir).Info("Local trace deleted")
	}
}

func uploadTraceToDB(c CaseSetting, s3Path string, size string) {
	req := model.DBTrace{
		CaseKey: model.CaseKey{
			EnvID:     mntbackend.EnvID(),
			Suite:     c.Suite,
			Benchmark: c.Title,
			Param:     c.param,
		},
		S3Path: s3Path,
		Size:   size,
	}
	traceID, err := mntbackend.UpdOrUplTrace(req)
	if err != nil {
		log.WithFields(log.Fields{
			"Case":   c,
			"S3Path": s3Path,
			"Size":   size,
		}).WithError(err).Error("Failed to upload trace")
	} else {
		log.WithField("TraceID", traceID).Info("Trace uploaded")

		// Build the param string nicely
		paramList := strings.Split(c.ParamStr, " ")
		formattedParams := []string{}
		for i := 0; i < len(paramList)-1; i += 2 {
			key := paramList[i]
			value := paramList[i+1]
			formattedParams = append(formattedParams, fmt.Sprintf("%s: %s", key, value))
		}
		paramStr := strings.Join(formattedParams, ", ")

		// Get the plain string of traceID
		traceIDStr := strings.TrimPrefix(fmt.Sprintf("%v", traceID), "ObjectID(\"")
		traceIDStr = strings.TrimSuffix(traceIDStr, "\")")

		// Create the log line
		logLine := fmt.Sprintf("- %s # %s / %s / %s\n", traceIDStr, c.Suite, c.Title, paramStr)

		// Build the file path: traceid/<Suite>-<Title>.txt
		err := os.MkdirAll("traceid", os.ModePerm) // Make sure the folder exists
		if err != nil {
			log.WithError(err).Error("Failed to create traceid directory")
			return
		}
		fileName := fmt.Sprintf("traceid/%s-%s.txt", c.Suite, c.Title)

		exists := false
		if file, err := os.Open(fileName); err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				if scanner.Text() == strings.TrimSuffix(logLine, "\n") {
					exists = true
					break
				}
			}
			file.Close()
		} else if !os.IsNotExist(err) {
			log.WithError(err).Errorf("Failed to open %s for reading", fileName)
			return
		}

		if exists {
			// No need to write the line again
			log.Infof("No need to write the line %s again", logLine)
			return
		}

		// Append the line since it's new
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.WithError(err).Errorf("Failed to open %s", fileName)
			return
		}
		defer f.Close()

		if _, err := f.WriteString(logLine); err != nil {
			log.WithError(err).Errorf("Failed to write to %s", fileName)
		}
	}
}

func storeTraceToS3(traceDir string) string {
	base, err := filepath.Rel(traceRootLocal, traceDir)
	if err != nil {
		log.WithError(err).Error("Failed to get relative path")
		return ""
	}

	objectPath := filepath.Join(traceRootRemote, base)
	aws.UploadDirectoryAsObjects(objectPath, traceDir)

	return objectPath
}

func moveTracesToDir(dir string) error {
	tracesDir := filepath.Join(dir, "traces")
	fmt.Printf("in moveTracesToDir, tracesDir: %s\n", tracesDir)

	// Check if traces directory exists
	info, err := os.Stat(tracesDir)
	if os.IsNotExist(err) || !info.IsDir() {
		return nil // Nothing to do
	} else if err != nil {
		return err
	}
	fmt.Printf("in moveTracesToDir, tracesDir exists: %s\n", tracesDir)

	// Read files directly under tracesDir
	entries, err := os.ReadDir(tracesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		srcPath := filepath.Join(tracesDir, entry.Name())
		dstPath := filepath.Join(dir, entry.Name())

		// Open source file
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Create destination file
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// Copy contents
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}
	}

	// Remove the traces directory after copying
	err = os.RemoveAll(tracesDir)
	return err
}

func processTrace(dir string) {
	// dir = filepath.Join(dir, "traces")
	moveTracesToDir(dir)
	// cmd := exec.Command(config.TracerToolProcessor(), fmt.Sprintf("%s/kernelslist", dir))
	pattern := filepath.Join(dir, "kernelslist*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.WithError(err).Error("Failed to search for kernelslist files")
		return
	}
	if len(matches) == 0 {
		log.Warnf("No kernelslist file found in: %s", dir)
		return
	}
	if len(matches) > 1 {
		log.Warnf("More than 1 kernelslist files found in: %s", dir)
	}

	kernelslistPath := matches[0] // use the first match

	file, err := os.Open(kernelslistPath)
	if err != nil {
		log.WithError(err).Errorf("Failed to open file: %s", kernelslistPath)
		return
	}
	defer file.Close()

	// Check if unxz is available
	if _, err := exec.LookPath("unxz"); err != nil {
		log.Warn("unxz command not found, skipping decompression")
	}

	var processedLines []string
	var unxzCount int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "kernel-") {
			fullPath := filepath.Join(dir, trimmed)

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				log.Errorf("File %s is recorded in %s, but not found in path %s", trimmed, filepath.Base(kernelslistPath), fullPath)
				processedLines = append(processedLines, trimmed)
				continue
			}

			if strings.HasSuffix(trimmed, ".xz") {
				cmd := exec.Command("unxz", fullPath)
				if err := cmd.Run(); err != nil {
					log.WithError(err).Errorf("Failed to unxz file: %s", fullPath)
				} else {
					unxzCount++
				}
				trimmed = strings.TrimSuffix(trimmed, ".xz")
			}
		}

		processedLines = append(processedLines, trimmed)
	}

	if err := scanner.Err(); err != nil {
		log.WithError(err).Errorf("Failed to read file: %s", kernelslistPath)
		return
	}

	outputPath := filepath.Join(dir, "kernelslist_processed")
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.WithError(err).Errorf("Failed to create processed kernelslist file: %s", outputPath)
		return
	}
	defer outFile.Close()

	for _, line := range processedLines {
		_, _ = outFile.WriteString(line + "\n")
	}

	log.Infof("Finished unxz for %d files and created a new kernelslist_processed file in path %s", unxzCount, outputPath)

	cmd := exec.Command(config.TracerToolProcessor(), outputPath) // kernelslistPath

	if err := runNormalCmdWithTimer(cmd); err != nil {
		log.WithError(err).Error("Failed to run traces processor")
		return
	}

	log.Info("Deleting original trace files")
	cmd = exec.Command("bash", "-c", fmt.Sprintf("rm %s/*.trace", dir))
	err = cmd.Run()
	if err != nil {
		log.WithField("dir", dir).WithError(err).Error("Failed to remove .trace files")
		return
	}

	cmd = exec.Command("bash", "-c", fmt.Sprintf("rm %s/*.xz", dir))
	err = cmd.Run()
	if err != nil {
		log.WithField("dir", dir).WithError(err).Error("Failed to remove .xz files")
	}

	// cmd = exec.Command("rm", fmt.Sprintf("%s/kernelslist", dir))
	// err = cmd.Run()
	// if err != nil {
	// 	log.WithError(err).Error("Failed to remove kernelslist")
	// 	return
	// }
	removeTraceRelatedFiles(dir)
}

func removeTraceRelatedFiles(dir string) {
	patterns := []string{
		"kernelslist_*",
		"kernel*.trace",
		"stats_*",
	}

	for _, p := range patterns {
		fullPattern := filepath.Join(dir, p)
		files, err := filepath.Glob(fullPattern)
		if err != nil {
			log.WithError(err).Errorf("Failed to find files matching pattern: %s", p)
			continue
		}

		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				log.WithError(err).Errorf("Failed to remove file: %s", file)
			} else {
				log.Infof("Successfully removed file: %s", file)
			}
		}
	}
}

func createInfoFile(dir string, c CaseSetting) {
	infoFile := filepath.Join(dir, "INFO")
	file, err := os.Create(infoFile)
	if err != nil {
		log.WithError(err).Error("Failed to create info file")
		return
	}
	defer file.Close()

	infoContent := fmt.Sprintf(
		`Title: %s
Suite: %s
Command: %s
Params: %s
DeviceID: %d
HostName: %s
DeviceName: %s
CudaVersion: %s
`,
		c.Title, c.Suite, c.Command, c.ParamStr,
		config.C.DeviceID, config.HostName(),
		config.DeviceName(), config.CudaVersion())

	_, err = file.WriteString(infoContent)
	if err != nil {
		log.WithError(err).Error("Failed to write info to file")
		return
	}
}

func getDirSize(dir string) string {
	cmd := exec.Command("du", "-sh", dir)
	out, err := cmd.Output()
	if err != nil {
		log.WithError(err).Error("Failed to get directory size")
		return ""
	}

	return strings.Split(string(out), "\t")[0]
}
