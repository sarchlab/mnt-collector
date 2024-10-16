package mntbackend

import "go.mongodb.org/mongo-driver/bson/primitive"

type TraceRequest struct {
	EnvID     primitive.ObjectID `json:"env_id"`
	Suite     string             `json:"suite"`
	Benchmark string             `json:"benchmark"`
	Param1    string             `json:"param1"`
	Param2    string             `json:"param2"`

	S3Path string `json:"s3_path"`
}
