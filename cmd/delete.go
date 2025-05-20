/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/transport/http"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

var mntClient *s3.Client
var mntBucket *string

// deleteCmd represents the delete entries from s3 and mongodb commands
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Use the given simulator to run traces and upload the data to database.",
	Long: `Use the given simulator to run traces and upload the data to database.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Retrieve the flag values
		machine, _ := cmd.Flags().GetString("machine")
		cudaVersion, _ := cmd.Flags().GetString("cuda-version")
		suite, _ := cmd.Flags().GetString("suite")
		benchmark, _ := cmd.Flags().GetString("benchmark")

		// Log the retrieved values (optional)
		log.Infof("Machine: %s, CUDA Version: %s, Suite: %s, Benchmark: %s", machine, cudaVersion, suite, benchmark)

		delete(machine, cudaVersion, suite, benchmark)
	},
}

// Structs for parsing secrets.yaml
type Secrets struct {
	AWS struct {
		Bucket          string `yaml:"bucket"`
		Region          string `yaml:"region"`
		AccessKeyID     string `yaml:"access-key-id"`
		SecretAccessKey string `yaml:"secret-access-key"`
	} `yaml:"s3"`

	MongoDB struct {
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	} `yaml:"mongodb"`
}

func loadSecrets(path string) (*Secrets, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var secrets Secrets
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&secrets)
	if err != nil {
		return nil, err
	}
	return &secrets, nil
}

func connectToMongoDB(secrets *Secrets) (*mongo.Client, *mongo.Database, context.Context, error) {
	uri := fmt.Sprintf("mongodb+srv://%s:%s@%s",
		secrets.MongoDB.User,
		secrets.MongoDB.Password,
		secrets.MongoDB.Host)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		// Ensure cancel is called to avoid context leaks
		if ctx.Err() != nil {
			cancel()
		}
	}()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		cancel() // Explicitly cancel if connection fails
		return nil, nil, nil, err
	}

	db := client.Database(secrets.MongoDB.Database)
	return client, db, ctx, nil
}

func findEnvironment(ctx context.Context, db *mongo.Database, machine, cudaVersion string) (interface{}, error) {
	envCol := db.Collection("environments")
	filter := bson.M{"machine": machine, "cuda_version": cudaVersion}

	var matchedEnvs []bson.M
	cursor, err := envCol.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query error on environments: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &matchedEnvs); err != nil {
		return nil, fmt.Errorf("cursor error on environments: %w", err)
	}

	if len(matchedEnvs) != 1 {
		return nil, fmt.Errorf("machine %s cuda_version %s not found (or ambiguous)", machine, cudaVersion)
	}

	return matchedEnvs[0]["_id"], nil
}

func countProfilesAndTraces(ctx context.Context, db *mongo.Database, envID interface{}, suite, benchmark string) ([]interface{}, map[string]struct{}, error) {
	profileCol := db.Collection("profiles")
	traceCol := db.Collection("traces")

	// Build the filter for suite and benchmark
	filter := bson.M{"env_id": envID}
	if suite != "all" {
		filter["suite"] = suite
	}
	if benchmark != "all" {
		filter["benchmark"] = benchmark
	}

	totalProfiles, _ := profileCol.CountDocuments(ctx, bson.M{})
	profileCursor, err := profileCol.Find(ctx, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying profiles: %w", err)
	}
	defer profileCursor.Close(ctx)

	totalTraces, _ := traceCol.CountDocuments(ctx, bson.M{})
	traceCursor, err := traceCol.Find(ctx, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("error querying traces: %w", err)
	}
	defer traceCursor.Close(ctx)

	profileMap := make(map[string]int)
	traceMap := make(map[string]int)

	profileCount := 0
	traceCount := 0

	traceIDs := []interface{}{}
	s3PathSet := make(map[string]struct{})

	for profileCursor.Next(ctx) {
		var doc bson.M
		profileCursor.Decode(&doc)
		key := fmt.Sprintf("%v/%v", doc["suite"], doc["benchmark"])
		profileMap[key]++
		profileCount++
	}

	for traceCursor.Next(ctx) {
		var doc bson.M
		traceCursor.Decode(&doc)
		key := fmt.Sprintf("%v/%v", doc["suite"], doc["benchmark"])
		traceMap[key]++
		traceIDs = append(traceIDs, doc["_id"])
		if s3Path, ok := doc["s3_path"].(string); ok {
			s3PathSet[s3Path] = struct{}{}
		}
		traceCount++
	}

	log.Infof("[Mongo Profiles] %d / %d items found in mnt.profiles", profileCount, totalProfiles)
	for k, v := range profileMap {
		log.Infof("  ## %d from %s", v, k)
	}

	log.Infof("[Mongo Traces] %d / %d items found in mnt.traces", traceCount, totalTraces)
	for k, v := range traceMap {
		log.Infof("  ## %d from %s", v, k)
	}

	return traceIDs, s3PathSet, nil
}

func countSimulations(ctx context.Context, db *mongo.Database, traceIDs []interface{}) error {
	simCol := db.Collection("simulations")
	totalSims, _ := simCol.CountDocuments(ctx, bson.M{})
	simCount, _ := simCol.CountDocuments(ctx, bson.M{"trace_id": bson.M{"$in": traceIDs}})

	log.Infof("[Mongo Simulations] %d / %d items found in mnt.simulations", simCount, totalSims)
	return nil
}

func deleteS3Folders(s3PathSet map[string]struct{}) error {
	deletedCount := 0

	for s3Path := range s3PathSet {
		// List all objects under the folder prefix
		listObjectsInput := &s3.ListObjectsV2Input{
			Bucket: mntBucket,
			Prefix: aws.String(s3Path + "/"), // Ensure we target the folder prefix
		}

		paginator := s3.NewListObjectsV2Paginator(mntClient, listObjectsInput)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(context.TODO())
			if err != nil {
				log.WithError(err).Errorf("Failed to list objects for folder: %s", s3Path)
				return err
			}

			// Delete all objects in the current page
			for _, object := range page.Contents {
				_, err := mntClient.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
					Bucket: mntBucket,
					Key:    object.Key,
				})
				if err != nil {
					log.WithError(err).Errorf("Failed to delete object: %s", *object.Key)
					return err
				}
				log.Infof("Deleted object: %s", *object.Key)
			}
		}

		// Log folder deletion
		deletedCount++
		log.Infof("Deleted folder: %s", s3Path)
	}

	log.Infof("[AWS Traces] %d trace folders deleted successfully", deletedCount)
	return nil
}

func checkS3Paths(s3PathSet map[string]struct{}, secrets *Secrets) error {
	mntBucket = aws.String(secrets.AWS.Bucket)
	mntClient = s3.New(s3.Options{
		Region: secrets.AWS.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			secrets.AWS.AccessKeyID, secrets.AWS.SecretAccessKey, "")),
	})

	existCount := 0
	for s3Path := range s3PathSet {
		s3InfoPath := filepath.Join(s3Path, "INFO")
		exist, err := checkObjectExist(s3InfoPath)
		if err != nil {
			log.WithError(err).Errorf("Failed to process finding old trace %s", s3Path)
			return err
		}
		if exist {
			existCount++
		} else {
			log.Errorf("S3 folders not found: %s", s3Path)
		}
	}
	log.Infof("[AWS Traces] %d traces found (INFO file detected) out of %d paths recorded in mnt.traces", existCount, len(s3PathSet))
	return nil
}

func promptForDeletion() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want delete all items from these tables? (Y/N): ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if strings.ToLower(text) == "y" {
		fmt.Printf("%s: Get Yes\n", text)
		return true
	}
	fmt.Printf("%s: Get No\n", text)
	log.Infof("Deletion aborted by user.")
	return false
}

func checkObjectExist(objectKey string) (bool, error) {
	_, err := mntClient.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: mntBucket,
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var respErr *http.ResponseError
		if errors.As(err, &respErr) && respErr.Response != nil && respErr.Response.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func initLogSettings() {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "logfile.log",
		MaxSize:    2,
		MaxBackups: 3,
		MaxAge:     30,
	}
	multiWriter := io.MultiWriter(lumberjackLogger, os.Stdout)

	log.SetOutput(multiWriter)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

func deleteSimulations(ctx context.Context, db *mongo.Database, traceIDs []interface{}) error {
	simCol := db.Collection("simulations")

	// Delete simulations where trace_id is in the list of traceIDs
	deleteResult, err := simCol.DeleteMany(ctx, bson.M{"trace_id": bson.M{"$in": traceIDs}})
	if err != nil {
		log.WithError(err).Error("Failed to delete simulations")
		return err
	}

	log.Infof("[Mongo Simulations] %d simulations deleted successfully", deleteResult.DeletedCount)
	return nil
}

func deleteProfilesAndTraces(ctx context.Context, db *mongo.Database, envID interface{}, traceIDs []interface{}, suite, benchmark string) error {
	// Build the filter for suite and benchmark
	filter := bson.M{"env_id": envID}
	if suite != "all" {
		filter["suite"] = suite
	}
	if benchmark != "all" {
		filter["benchmark"] = benchmark
	}
	// Delete profiles
	profileCol := db.Collection("profiles")
	profileDeleteResult, err := profileCol.DeleteMany(ctx, filter)
	if err != nil {
		log.WithError(err).Error("Failed to delete profiles")
		return err
	}
	log.Infof("[Mongo Profiles] %d profiles deleted successfully", profileDeleteResult.DeletedCount)

	// Delete traces
	traceCol := db.Collection("traces")
	traceDeleteResult, err := traceCol.DeleteMany(ctx, filter)
	if err != nil {
		log.WithError(err).Error("Failed to delete traces")
		return err
	}
	log.Infof("[Mongo Traces] %d traces deleted successfully", traceDeleteResult.DeletedCount)
	return nil
}

func deleteEnvironment(ctx context.Context, db *mongo.Database, envID interface{}) error {
	envCol := db.Collection("environments")

	// Delete the environment document where _id matches envID
	deleteResult, err := envCol.DeleteOne(ctx, bson.M{"_id": envID})
	if err != nil {
		log.WithError(err).Error("Failed to delete environment")
		return err
	}

	if deleteResult.DeletedCount == 0 {
		log.Warnf("No environment found with ID: %v", envID)
	} else {
		log.Infof("[Mongo Environment] Environment with ID %v deleted successfully", envID)
	}

	return nil
}

func delete(machine, cudaVersion, suite, benchmark string) {
	initLogSettings()

	// machine := flag.String("machine", "", "The machine name (required)")
	// cudaVersion := flag.String("cuda-version", "", "The CUDA version (required)")
	// suite := flag.String("suite", "all", "The suite name (optional, default: all)")
	// benchmark := flag.String("benchmark", "all", "The benchmark name (optional, default: all)")
	// flag.Parse()
	if machine == "" || cudaVersion == "" {
		log.Fatalf("Both --machine and --cuda-version are required parameters for the `delete` operation")
	}
	log.Infof("Optional parameters: suite=%s, title=%s", suite, benchmark)

	secrets, err := loadSecrets("etc/secrets.yaml")
	if err != nil {
		log.Errorf("Failed to load secrets: %v", err)
	}

	// Step 1: Connect to MongoDB
	client, db, ctx, err := connectToMongoDB(secrets)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v. Did you check if `mongodb:` exists in your secrets.yaml?", err)
	}
	defer client.Disconnect(ctx) // Disconnect explicitly in main
	defer ctx.Done()

	// Step 2: Find environment
	envID, err := findEnvironment(ctx, db, machine, cudaVersion)
	if err != nil {
		log.Fatalf("Error finding environment: %v", err)
	}

	// Step 3: Count and group Profiles and Traces
	traceIDs, s3PathSet, err := countProfilesAndTraces(ctx, db, envID, suite, benchmark)
	if err != nil {
		log.Fatalf("Error counting profiles and traces: %v", err)
	}

	// Step 4: Count simulations
	if err := countSimulations(ctx, db, traceIDs); err != nil {
		log.Fatalf("Error counting simulations: %v", err)
	}

	// Step 5: Check S3 folders
	if err := checkS3Paths(s3PathSet, secrets); err != nil {
		log.Fatalf("Error checking S3 paths: %v", err)
	}

	// Step 6: Prompt for deletion
	if !promptForDeletion() {
		return
	}

	// Step 7: Delete S3 folders
	if err := deleteS3Folders(s3PathSet); err != nil {
		log.Fatalf("Error deleting S3 folders: %v", err)
	}

	//Step 8: Delete MongoDB Simulations
	if err := deleteSimulations(ctx, db, traceIDs); err != nil {
		log.Fatalf("Error deleting MongoDB simulations: %v", err)
	}

	//Step 9: Delete MongoDB Profiles and Traces
	if err := deleteProfilesAndTraces(ctx, db, envID, traceIDs, suite, benchmark); err != nil {
		log.Fatalf("Error deleting MongoDB profiles and traces: %v", err)
	}

	// Step 10: Delete MongoDB Environment (only if suite and benchmark are "all")
	if suite == "all" && benchmark == "all" {
		if err := deleteEnvironment(ctx, db, envID); err != nil {
			log.Fatalf("Error deleting MongoDB environment: %v", err)
		}
	} else {
		log.Infof("Skipping environment deletion because suite=%s and benchmark=%s", suite, benchmark)
	}
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
