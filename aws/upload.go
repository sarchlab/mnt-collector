package aws

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const partSize = 512 * 1024 * 1024 // 0.5GB

func multiplePartUpload(object string, filepath string) {
	key := aws.String(object)
	file, err := os.Open(filepath)
	if err != nil {
		log.WithError(err).Panic("Failed to open file")
	}
	defer file.Close()

	log.WithField("object", object).Info("Creating multipart upload")
	output, err := mntClient.CreateMultipartUpload(context.TODO(), &s3.CreateMultipartUploadInput{
		Bucket: mntBucket,
		Key:    key,
	})
	if err != nil {
		log.WithError(err).Panic("Failed to create multipart upload")
	}

	log.WithField("uploadID", *output.UploadId).Info("Uploading parts")
	parts := uploadParts(output.UploadId, key, file)

	_, err = mntClient.CompleteMultipartUpload(context.TODO(), &s3.CompleteMultipartUploadInput{
		Bucket:   mntBucket,
		UploadId: output.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	if err != nil {
		log.WithError(err).Error("Failed to complete multipart upload")
	}

	log.WithField("object", object).Info("File uploaded")
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
			log.WithError(err).Panic("Failed to read file")
		}

		output, err := mntClient.UploadPart(context.TODO(), &s3.UploadPartInput{
			Bucket:     mntBucket,
			Key:        key,
			UploadId:   uploadID,
			PartNumber: aws.Int32(partNumber),
			Body:       bytes.NewReader(data[:bytesRead]),
		})
		if err != nil {
			log.WithError(err).Panic("Failed to upload part")
		}

		parts = append(parts, types.CompletedPart{
			ETag:       output.ETag,
			PartNumber: aws.Int32(partNumber),
		})

		log.WithFields(log.Fields{
			"part":   partNumber,
			"etag":   *output.ETag,
			"upload": uploadID,
		}).Info("Part uploaded")
		partNumber++
	}

	return parts
}

func uploadFileAsObject(object string, filepath string) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		log.WithError(err).Panic("Failed to get file info")
	}

	if fileInfo.Size() > partSize {
		log.WithFields(log.Fields{
			"filesize": fileInfo.Size(),
			"partsize": partSize,
		}).Info("Using multipart upload")
		multiplePartUpload(object, filepath)
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.WithError(err).Panic("Failed to open file")
	}
	defer file.Close()

	log.WithField("object", object).Info("Uploading file")
	_, err = mntClient.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: mntBucket,
		Key:    aws.String(object),
		Body:   file,
	})
	if err != nil {
		log.WithError(err).Panic("Failed to upload file")
	}

	log.WithFields(log.Fields{
		"object": object,
		"size":   fileInfo.Size(),
	}).Info("File uploaded")
}

func UploadDirectoryAsObjects(objectDir string, dirpath string) {
	err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		base, err := filepath.Rel(dirpath, path)
		if err != nil {
			log.WithError(err).Panic("Failed to get relative path")
		}

		objectPath := filepath.Join(objectDir, base)
		uploadFileAsObject(objectPath, path)

		return nil
	})

	if err != nil {
		log.WithError(err).Panic("Failed to walk directory")
	}
}
