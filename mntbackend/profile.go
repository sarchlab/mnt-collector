package mntbackend

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProfileRequest struct {
	EnvID     primitive.ObjectID `json:"env_id"`
	Suite     string             `json:"suite"`
	Benchmark string             `json:"benchmark"`
	Param1    string             `json:"param1"`
	Param2    string             `json:"param2"`

	RepeatTimes int32   `json:"repeat_times"`
	AvgCycles   float64 `json:"avg_cycles"`
}
