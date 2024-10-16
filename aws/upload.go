package aws

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const partSize = 512 * 1024 * 1024 // 0.5GB

func MultiplePartUpload(object string, filepath string) {
	key := aws.String(object)
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed to open file, %v", err)
	}
	defer file.Close()

	output, err := mntClient.CreateMultipartUpload(context.TODO(), &s3.CreateMultipartUploadInput{
		Bucket: mntBucket,
		Key:    key,
	})
	if err != nil {
		log.Fatalf("Failed to create multipart upload, %v", err)
	}

	log.Printf("UploadID: %s\n", *output.UploadId)
	parts := uploadParts(output.UploadId, key, file)

	_, err = mntClient.CompleteMultipartUpload(context.TODO(), &s3.CompleteMultipartUploadInput{
		Bucket:   mntBucket,
		UploadId: output.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		log.Fatalf("Failed to complete multipart upload, %v", err)
	}

	log.Printf("Uploaded %s as %s\n", filepath, object)
}

func uploadParts(uploadID *string, key *string, file *os.File) []types.CompletedPart {
	parts := []types.CompletedPart{}
	partNumber := int32(1)
	for {
		data := make([]byte, partSize)
		bytesRead, err := file.Read(data)

		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Failed to read file, %v", err)
		}

		output, err := mntClient.UploadPart(context.TODO(), &s3.UploadPartInput{
			Bucket:     mntBucket,
			Key:        key,
			UploadId:   uploadID,
			PartNumber: aws.Int32(partNumber),
			Body:       bytes.NewReader(data[:bytesRead]),
		})
		if err != nil {
			log.Fatalf("Failed to upload part, %v", err)
		}

		parts = append(parts, types.CompletedPart{
			ETag:       output.ETag,
			PartNumber: aws.Int32(partNumber),
		})

		log.Printf("Uploaded part %d of upload %d", partNumber, uploadID)
		partNumber++
	}

	return parts
}

func UploadFileAsObject(object string, filepath string) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		log.Fatalf("Failed to get file info, %v", err)
	}

	if fileInfo.Size() > partSize {
		log.Printf("File size is greater than %d bytes, using multipart upload\n", partSize)
		MultiplePartUpload(object, filepath)
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed to open file, %v", err)
	}
	defer file.Close()

	_, err = mntClient.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: mntBucket,
		Key:    aws.String(object),
		Body:   file,
	})
	if err != nil {
		log.Fatalf("Failed to upload file, %v", err)
	}

	log.Printf("Uploaded %s as %s\n", filepath, object)
}

func UploadDirectoryAsObjects(objectDir string, dirpath string) {
	err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		objectPath := filepath.Join(objectDir, filepath.Base(path))
		log.Printf("Uploading %s as %s\n", path, objectPath)
		UploadFileAsObject(objectPath, path)

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk directory, %v", err)
	}
}
