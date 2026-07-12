package collector

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sarchlab/mnt-collector/config"
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/model"
	"github.com/sarchlab/mnt-collector/mntbackend"
	log "github.com/sirupsen/logrus"
)

/*
NCU-based profiling (mode = ncu)

Steps per case:
1. Run ncu with --csv to a temp file (repeat config.C.RepeatTimes).
2. Parse CSV rows:
   - Sum all rows where Section="GPU Speed Of Light Throughput" AND MetricName="Elapsed Cycles".
   - Collect Duration rows (MetricName="Duration") + their units.
   - Compute runCycles (sum elapsed cycles for this run).
   - Compute runTimeNs (sum durations converted to ns).
3. After repeats:
   avgCycles = mean(runCycles)
   avgFreqHz = totalCyclesAcrossAllRuns / totalDurationNsAcrossAllRuns
   avgNanoSec = avgCycles / avgFreqHz * 1e9
   freqMHz (reported) = avgFreqHz / 1e6
4. Populate ProfileData (extended with Cycle + ProfileType) and call uploadProfileToDB.

Constants below define the strings we match so they are single-sourced.
*/

const (
	ncuSectionThroughput = "GPU Speed Of Light Throughput"
	ncuMetricElapsed     = "Elapsed Cycles"
	ncuMetricDuration    = "Duration"

	profilesNcuDir   = "./tmp/mnt-collector/profiles_ncu/"
	ncuProfileType   = "ncu"
	defaultNcuBinary = "/usr/local/cuda-12.5/bin/ncu"
)

func RunProfileCollectionNCU() {
	caseSettings := generateCaseSettings(config.C.Cases)
	repeatTimes := int(config.C.RepeatTimes)

	if err := os.MkdirAll(profilesNcuDir, 0o755); err != nil {
		log.WithError(err).Error("Failed to create ncu profiles dir")
		return
	}

	for _, c := range caseSettings {
		log.WithFields(log.Fields{
			"Title":   c.Title,
			"Suite":   c.Suite,
			"Command": c.Command,
			"Params":  c.ParamStr,
		}).Info("Start NCU profile collection")

		var runCyclesList []float64
		var runDurNsList []float64

		for i := 0; i < repeatTimes; i++ {
			log.WithField("Repeat", i).Info("NCU profiling run")
			cycles, durNs, err := profileOnceNCU(c)
			if err != nil {
				log.WithError(err).Error("NCU profiling failed; abort case")
				break
			}
			runCyclesList = append(runCyclesList, cycles)
			runDurNsList = append(runDurNsList, durNs)
		}

		if len(runCyclesList) != repeatTimes {
			log.Error("NCU profile not completed for all repeats; skipping aggregation")
			continue
		}

		var totalCycles, totalDurNs float64
		for i := range runCyclesList {
			totalCycles += runCyclesList[i]
			totalDurNs += runDurNsList[i]
		}
		avgCycles := totalCycles / float64(repeatTimes)
		avgFreqHz := totalCycles / (totalDurNs / 1e9) // cycles / seconds
		freqMHz := avgFreqHz / 1e6
		if freqMHz < 100 || freqMHz > 3000 {
			log.WithFields(log.Fields{
				"freqMHz": freqMHz,
			}).Panic("Derived average SM frequency (MHz) out of expected range (100-3000)")
		}
		avgNanoSec := (avgCycles / avgFreqHz) * 1e9

		data := ProfileData{
			AvgNanoSec:   avgNanoSec,
			Frequency:    uint32(freqMHz), // store MHz truncated (original struct was uint32)
			MaxFrequency: config.MaxFrequency(),
			Cycle:        avgCycles,
			ProfileType:  ncuProfileType,
		}

		// Print a debug map before upload for troubleshooting
		log.WithFields(log.Fields{
			"suite":        c.Suite,
			"benchmark":    c.Title,
			"param":        c.param,
			"avg_cycles":   avgCycles,
			"avg_freq_mhz": freqMHz,
			"avg_nano_sec": avgNanoSec,
			"repeats":      repeatTimes,
			"profile_type": ncuProfileType,
		}).Info("NCU aggregated profile result")

		if config.C.UploadToServer {
			log.Info("Uploading NCU profile result")
			uploadProfileToDBNCU(c, data, int32(repeatTimes))
		} else {
			log.Info("Upload disabled; skipping")
		}
	}
}

func resolveNCUPath() string {
	// 2. Preferred fixed install location
	if st, err := os.Stat(defaultNcuBinary); err == nil && st.Mode()&0o111 != 0 {
		return defaultNcuBinary
	}
	// 3. PATH lookup
	if lp, err := exec.LookPath("ncu"); err == nil {
		return lp
	}
	log.Warn("ncu not found; ensure CUDA Nsight Compute is installed and PATH set")
	return "ncu" // will likely fail; surfaced for logging
}

// profileOnceNCU executes one NCU run, parses CSV, returns (totalElapsedCycles, totalDurationNs)
func profileOnceNCU(c CaseSetting) (float64, float64, error) {
	tmpFile, err := os.CreateTemp(profilesNcuDir, "ncu-*.csv")
	if err != nil {
		return 0, 0, fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	// Build command line
	param := strings.Split(strings.TrimSpace(c.ParamStr), " ")
	if len(param) == 1 && param[0] == "" {
		param = []string{}
	}

	// You may want to make ncu path configurable; hard-code common path fallback.
	ncuPath := resolveNCUPath()
	log.WithField("ncuPath", ncuPath).Debug("Resolved ncu path")
	// if p := config.C.NCUPath; p != "" { // optional: add NCUPath in config if desired
	// 	ncuPath = p
	// }

	args := []string{
		"--set=full",
		"--csv",
		"--target-processes=all",
		c.Command,
	}
	args = append(args, param...)

	cmd := exec.Command(ncuPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("CUDA_VISIBLE_DEVICES=%d", config.C.DeviceID))

	// NOTE about sudo: prefer granting user access to CUPTI / GPU devices instead of invoking sudo here.
	// If absolutely required, wrap as: exec.Command("sudo", append([]string{ncuPath}, args...)...)
	// We log a warning if not root.
	if os.Geteuid() != 0 {
		log.Warn("Running ncu without sudo. If metrics are missing, rerun binary with sudo or adjust permissions.")
	}

	outFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, 0, fmt.Errorf("open tmp csv: %w", err)
	}
	defer outFile.Close()
	cmd.Stdout = outFile
	cmd.Stderr = outFile

	log.WithField("csv", tmpPath).Debug("Executing ncu")
	if err := runGPUCmdWithTimer(cmd); err != nil {
		return 0, 0, fmt.Errorf("run ncu: %w", err)
	}

	// Parse the CSV file we just produced
	cycles, durNs, err := parseNcuCSV(tmpPath)
	if err != nil {
		return 0, 0, fmt.Errorf("parse csv: %w", err)
	}

	// Remove file after use to save space
	_ = os.Remove(tmpPath)

	return cycles, durNs, nil
}

func parseNcuCSV(path string) (float64, float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	// We first need to skip any preamble lines until we reach a line beginning with "ID","Process ID",...
	// We'll scan line-by-line until header found, then feed remainder into csv.Reader for structured parsing.
	var headerLine string
	var preamble []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "\"ID\",\"Process ID\"") {
			headerLine = line
			break
		}
		preamble = append(preamble, line)
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}
	if headerLine == "" {
		return 0, 0, errors.New("CSV header not found")
	}

	var csvContent strings.Builder
	csvContent.WriteString(headerLine + "\n")
	for scanner.Scan() {
		csvContent.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	r := csv.NewReader(strings.NewReader(csvContent.String()))
	r.FieldsPerRecord = -1

	// Read header
	_, err = r.Read()
	if err != nil {
		return 0, 0, fmt.Errorf("read header: %w", err)
	}

	var totalCycles float64
	var totalDurNs float64
	var lastSection string

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, fmt.Errorf("csv read: %w", err)
		}
		if len(rec) < 15 {
			continue
		}
		// Columns based on sample:
		// 11 Section Name, 12 Metric Name, 13 Metric Unit, 14 Metric Value
		section := strings.Trim(rec[11], "\"")
		if section == "" {
			section = lastSection // continue same section
		} else {
			lastSection = section
		}
		metricName := strings.Trim(rec[12], "\"")
		unit := strings.Trim(rec[13], "\"")
		valueStr := strings.Trim(rec[14], "\"")
		valueStr = strings.ReplaceAll(valueStr, ",", "")
		if valueStr == "" {
			continue
		}
		val, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		if section == ncuSectionThroughput {
			switch metricName {
			case ncuMetricElapsed:
				// Already in cycles
				totalCycles += val
			case ncuMetricDuration:
				ns := convertDurationToNs(val, unit)
				totalDurNs += ns
			}
		}
	}

	return totalCycles, totalDurNs, nil
}

func convertDurationToNs(v float64, unit string) float64 {
	switch strings.ToLower(unit) {
	case "ns":
		return v
	case "us":
		return v * 1e3
	case "ms":
		return v * 1e6
	case "s", "sec":
		return v * 1e9
	default:
		log.WithField("unit", unit).Warn("Unknown duration unit; assuming microseconds")
		return v * 1e3
	}
}

func uploadProfileToDBNCU(c CaseSetting, data ProfileData, repeatTimes int32) {
	req := model.DBProf{
		CaseKeyProfile: model.CaseKeyProfile{
			EnvID:       mntbackend.EnvID(),
			Suite:       c.Suite,
			Benchmark:   c.Title,
			Param:       c.param,
			ProfileType: data.ProfileType,
		},
		RepeatTimes:  repeatTimes,
		AvgNanoSec:   data.AvgNanoSec,
		Frequency:    data.Frequency,
		MaxFrequency: data.MaxFrequency,
		Cycle:        data.Cycle,
		// ProfileType:  data.ProfileType,
		// Add fields below only if model.DBProf supports them; otherwise they are just logged.
		// Cycle:       data.Cycle,
		// ProfileType: data.ProfileType,
	}
	// Debug full payload (including added fields)
	log.WithFields(log.Fields{
		"Case":        c,
		"RepeatTimes": repeatTimes,
		"AvgNanoSec":  data.AvgNanoSec,
		"Frequency":   data.Frequency,
		"MaxFreq":     data.MaxFrequency,
		"Cycle":       data.Cycle,
		"ProfileType": data.ProfileType,
	}).Info("Uploading profile")
	profileID, err := mntbackend.UpdOrUplProfile(req)
	if err != nil {
		log.WithFields(log.Fields{
			"Case": c,
			"Data": data,
		}).WithError(err).Error("Failed to upload profile")
	} else {
		log.WithField("ProfileID", profileID).Info("Profile uploaded")
	}
}

// ...existing code...
