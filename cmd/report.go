package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate instruction and trace-size reports for a machine/cuda-version and suite.",
	Long:  `Download matched trace folders, parse .traceg files to count instructions and sizes, then write CSV reports to reports/`,
	Run: func(cmd *cobra.Command, args []string) {
		machine, _ := cmd.Flags().GetString("machine")
		cudaVersion, _ := cmd.Flags().GetString("cuda-version")
		suite, _ := cmd.Flags().GetString("suite")

		if machine == "" || cudaVersion == "" {
			log.Fatalf("Both --machine and --cuda-version are required")
		}
		if suite == "" {
			log.Fatalf("--suite must be specified (e.g., rodinia)")
		}

		if err := runReport(machine, cudaVersion, suite); err != nil {
			log.Fatalf("report failed: %v", err)
		}
	},
}

func init() {
	// flags local to report (use same names as delete command)
	reportCmd.Flags().String("machine", "", "machine name (required)")
	reportCmd.Flags().String("cuda-version", "", "cuda version (required)")
	reportCmd.Flags().String("suite", "", "suite name (required)")
	rootCmd.AddCommand(reportCmd)
}

func runReport(machine, cudaVersion, suite string) error {
	initLogSettings()

	// load secrets and init mongo
	secrets, err := loadSecrets("etc/secrets.yaml")
	if err != nil {
		return fmt.Errorf("loadSecrets: %w", err)
	}
	client, db, ctx, err := connectToMongoDB(secrets)
	if err != nil {
		return fmt.Errorf("connectToMongoDB: %w", err)
	}
	defer client.Disconnect(ctx)
	defer ctx.Done()

	envID, err := findEnvironment(ctx, db, machine, cudaVersion)
	if err != nil {
		return fmt.Errorf("findEnvironment: %w", err)
	}

	// query traces collection to build map s3Path -> benchmark
	traceCol := db.Collection("traces")
	filter := bson.M{"env_id": envID, "suite": suite}
	cur, err := traceCol.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("query traces: %w", err)
	}
	defer cur.Close(ctx)

	type traceEntry struct {
		S3Path    string
		Benchmark string
	}
	traceEntries := make([]traceEntry, 0)
	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			log.WithError(err).Warn("failed to decode trace doc")
			continue
		}
		s3p, _ := doc["s3_path"].(string)
		bench, _ := doc["benchmark"].(string)
		if s3p != "" && bench != "" {
			traceEntries = append(traceEntries, traceEntry{S3Path: s3p, Benchmark: bench})
		}
	}

	if len(traceEntries) == 0 {
		log.Infof("No traces found for env=%v suite=%s", envID, suite)
		return nil
	}

	// Group trace entries by benchmark for per-benchmark progress bars
	benchToEntries := make(map[string][]traceEntry)
	for _, te := range traceEntries {
		benchToEntries[te.Benchmark] = append(benchToEntries[te.Benchmark], te)
	}

	// init S3 client
	mntBucket = aws.String(secrets.AWS.Bucket)
	mntClient = s3.New(s3.Options{
		Region: secrets.AWS.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			secrets.AWS.AccessKeyID, secrets.AWS.SecretAccessKey, "")),
	})

	// tmp download root
	tmpRoot := filepath.Join("tmp", "report")
	_ = os.RemoveAll(tmpRoot)
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return err
	}

	// ensure reports directories
	if err := os.MkdirAll("reports", 0o755); err != nil {
		return err
	}
	benchReportsDir := filepath.Join("reports", "benchmarks")
	if err := os.MkdirAll(benchReportsDir, 0o755); err != nil {
		return err
	}

	// accumulators
	globalInstrCounts := make(map[string]int64)
	benchTraceCount := make(map[string]int64)
	benchTraceSizeBytes := make(map[string]int64)

	// iterate benchmarks in sorted order and show a progress bar per benchmark
	benches := make([]string, 0, len(benchToEntries))
	for b := range benchToEntries {
		benches = append(benches, b)
	}
	sort.Strings(benches)

	for _, bench := range benches {
		entries := benchToEntries[bench]
		total := len(entries)

		// per-benchmark instruction counts (accumulate for this benchmark only)
		benchInstrCounts := make(map[string]int64)

		// create progress bar
		bar := progressbar.NewOptions(total,
			progressbar.OptionSetDescription(fmt.Sprintf("benchmark: %s", bench)),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(60),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionOnCompletion(func() { fmt.Fprintln(os.Stdout) }),
		)

		for idx, te := range entries {
			// treat each s3 folder as one trace (increment trace count per benchmark once)
			benchTraceCount[bench]++

			// create local folder for this trace
			localFolder := filepath.Join(tmpRoot, strings.ReplaceAll(te.S3Path, "/", "_"))
			_ = os.MkdirAll(localFolder, 0o755)

			// prepare prefix and INFO key
			prefix := te.S3Path
			if !strings.HasSuffix(prefix, "/") {
				prefix = prefix + "/"
			}
			infoKey := strings.TrimSuffix(prefix, "/") + "/INFO"

			// download INFO first to extract Params
			paramsStr := ""
			getOut, err := mntClient.GetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: mntBucket,
				Key:    aws.String(infoKey),
			})
			if err == nil {
				infoBytes, _ := io.ReadAll(getOut.Body)
				getOut.Body.Close()
				lines := strings.Split(string(infoBytes), "\n")
				for _, l := range lines {
					l = strings.TrimSpace(l)
					if strings.HasPrefix(l, "Params:") {
						paramsStr = strings.TrimSpace(strings.TrimPrefix(l, "Params:"))
						break
					}
				}
			} else {
				// no INFO or download failed; ignore and continue
				log.WithError(err).Debugf("failed to download INFO %s", infoKey)
			}

			// print a short status line including params and index/total (avoids using non-existent ChangeDescription)
			if paramsStr != "" {
				fmt.Fprintf(os.Stdout, "Processing %s %d/%d %s\n", bench, idx+1, total, paramsStr)
			} else {
				fmt.Fprintf(os.Stdout, "Processing %s %d/%d\n", bench, idx+1, total)
			}

			// list objects under prefix and download .traceg files
			p := s3.NewListObjectsV2Paginator(mntClient, &s3.ListObjectsV2Input{
				Bucket: mntBucket,
				Prefix: aws.String(prefix),
			})

			var tracegDownloadedTotalBytes int64 = 0
			for p.HasMorePages() {
				page, err := p.NextPage(context.TODO())
				if err != nil {
					log.WithError(err).Warnf("list objects failed for %s", te.S3Path)
					break
				}
				for _, obj := range page.Contents {
					if strings.HasSuffix(*obj.Key, ".traceg") {
						// download object
						getOut, err := mntClient.GetObject(context.TODO(), &s3.GetObjectInput{
							Bucket: mntBucket,
							Key:    obj.Key,
						})
						if err != nil {
							log.WithError(err).Warnf("GetObject %s failed", *obj.Key)
							continue
						}
						localName := filepath.Join(localFolder, filepath.Base(*obj.Key))
						dst, err := os.Create(localName)
						if err != nil {
							getOut.Body.Close()
							log.WithError(err).Warn("create local file failed")
							continue
						}
						n, err := io.Copy(dst, getOut.Body)
						getOut.Body.Close()
						dst.Close()
						if err != nil {
							log.WithError(err).Warn("copy local file failed")
							continue
						}
						tracegDownloadedTotalBytes += n

						// parse localName
						mcounts, err := ParseTraceFile(localName)
						if err != nil {
							log.WithError(err).Warnf("failed to parse %s", localName)
							continue
						}
						for k, v := range mcounts {
							globalInstrCounts[k] += v
							benchInstrCounts[k] += v
						}
					}
				}
			}

			benchTraceSizeBytes[bench] += tracegDownloadedTotalBytes

			// advance progress bar for this benchmark (one trace processed)
			_ = bar.Add(1)

			// remove local folder for this trace immediately to save disk
			_ = os.RemoveAll(localFolder)
		}

		// finish bar explicitly (ensures newline)
		_ = bar.Finish()

		// write per-benchmark instruction report immediately after benchmark finished
		benchReportPath := filepath.Join(benchReportsDir, fmt.Sprintf("%s-%s_instruction_report.csv", suite, bench))
		bf, err := os.Create(benchReportPath)
		if err != nil {
			log.WithError(err).Warnf("failed to create bench report %s", benchReportPath)
		} else {
			bw := csv.NewWriter(bf)
			_ = bw.Write([]string{"instruction", "count"})
			// sort benchInstrCounts by count desc
			type kvb struct {
				K string
				V int64
			}
			bkvs := make([]kvb, 0, len(benchInstrCounts))
			for k, v := range benchInstrCounts {
				bkvs = append(bkvs, kvb{k, v})
			}
			sort.Slice(bkvs, func(i, j int) bool { return bkvs[i].V > bkvs[j].V })
			for _, e := range bkvs {
				_ = bw.Write([]string{e.K, strconv.FormatInt(e.V, 10)})
			}
			bw.Flush()
			bf.Close()
			log.Infof("Wrote %s", benchReportPath)
		}
	}

	// instruction report CSV: instruction,count (desc) for whole suite
	instrFile := filepath.Join("reports", fmt.Sprintf("%s_instruction_report.csv", suite))
	f, err := os.Create(instrFile)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{"instruction", "count"})
	// sort by count desc
	type kv struct {
		K string
		V int64
	}
	kvs := make([]kv, 0, len(globalInstrCounts))
	for k, v := range globalInstrCounts {
		kvs = append(kvs, kv{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].V > kvs[j].V })
	for _, e := range kvs {
		_ = w.Write([]string{e.K, strconv.FormatInt(e.V, 10)})
	}
	w.Flush()
	f.Close()
	log.Infof("Wrote %s", instrFile)

	// trace size CSV: benchmark,trace count,total trace size mb (benchmark asc)
	sizeFile := filepath.Join("reports", fmt.Sprintf("%s_trace_size.csv", suite))
	f2, err := os.Create(sizeFile)
	if err != nil {
		return err
	}
	w2 := csv.NewWriter(f2)
	_ = w2.Write([]string{"benchmark", "trace count", "total trace size mb"})
	// build bench list sorted
	benches = make([]string, 0, len(benchTraceCount))
	for b := range benchTraceCount {
		benches = append(benches, b)
	}
	sort.Strings(benches)
	for _, b := range benches {
		count := benchTraceCount[b]
		sizeMB := float64(benchTraceSizeBytes[b]) / (1024.0 * 1024.0)
		// use integer MB as example (round)
		_ = w2.Write([]string{b, strconv.FormatInt(count, 10), strconv.FormatInt(int64(sizeMB+0.5), 10)})
	}
	w2.Flush()
	f2.Close()
	log.Infof("Wrote %s", sizeFile)

	// cleanup tmp root
	_ = os.RemoveAll(tmpRoot)
	log.Infof("Report generation finished.")
	return nil
}
