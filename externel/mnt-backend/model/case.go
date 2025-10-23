package model

import (
	"context"

	"github.com/sarchlab/mnt-collector/externel/mnt-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CaseKey struct {
	EnvID     primitive.ObjectID `json:"env_id" bson:"env_id"`
	Suite     string             `json:"suite" bson:"suite"`
	Benchmark string             `json:"benchmark" bson:"benchmark"`
	Param     Param              `json:"param" bson:"param"`
}
type Param struct {
	Size        string `json:"size,omitempty" bson:"size,omitempty" yaml:"size,omitempty"`
	Sizemult    string `json:"sizemult,omitempty" bson:"sizemult,omitempty" yaml:"sizemult,omitempty"`
	VectorN     string `json:"vectorN,omitempty" bson:"vectorN,omitempty" yaml:"vectorN,omitempty"`
	ElementN    string `json:"elementN,omitempty" bson:"elementN,omitempty" yaml:"elementN,omitempty"`
	Log2Data    string `json:"log2data,omitempty" bson:"log2data,omitempty" yaml:"log2data,omitempty"`
	Log2Kernel  string `json:"log2kernel,omitempty" bson:"log2kernel,omitempty" yaml:"log2kernel,omitempty"`
	ArrayLength string `json:"arrayLength,omitempty" bson:"arrayLength,omitempty" yaml:"arrayLength,omitempty"`
	DimX        string `json:"dimX,omitempty" bson:"dimX,omitempty" yaml:"dimX,omitempty"`
	DimY        string `json:"dimY,omitempty" bson:"dimY,omitempty" yaml:"dimY,omitempty"`
	DimZ        string `json:"dimZ,omitempty" bson:"dimZ,omitempty" yaml:"dimZ,omitempty"`
	KernelID    string `json:"kernelID,omitempty" bson:"kernelID,omitempty" yaml:"kernelID,omitempty"`
	BlockDimX   string `json:"blockDimX,omitempty" bson:"blockDimX,omitempty" yaml:"blockDimX,omitempty"`
	BlockDimY   string `json:"blockDimY,omitempty" bson:"blockDimY,omitempty" yaml:"blockDimY,omitempty"`
	M           string `json:"m,omitempty" bson:"m,omitempty" yaml:"m,omitempty"`
	N           string `json:"n,omitempty" bson:"n,omitempty" yaml:"n,omitempty"`
	I           string `json:"i,omitempty" bson:"i,omitempty" yaml:"i,omitempty"`
	J           string `json:"j,omitempty" bson:"j,omitempty" yaml:"j,omitempty"`
	K           string `json:"k,omitempty" bson:"k,omitempty" yaml:"k,omitempty"`
	Ni          string `json:"ni,omitempty" bson:"ni,omitempty" yaml:"ni,omitempty"`
	Nj          string `json:"nj,omitempty" bson:"nj,omitempty" yaml:"nj,omitempty"`
	Nk          string `json:"nk,omitempty" bson:"nk,omitempty" yaml:"nk,omitempty"`
	Alpha       string `json:"alpha,omitempty" bson:"alpha,omitempty" yaml:"alpha,omitempty"`
	Beta        string `json:"beta,omitempty" bson:"beta,omitempty" yaml:"beta,omitempty"`
	Gamma       string `json:"gamma,omitempty" bson:"gamma,omitempty" yaml:"gamma,omitempty"`
	Order       string `json:"order,omitempty" bson:"order,omitempty" yaml:"order,omitempty"`
	D           string `json:"d,omitempty" bson:"d,omitempty" yaml:"d,omitempty"`
	L           string `json:"l,omitempty" bson:"l,omitempty" yaml:"l,omitempty"`
	Boxes1D     string `json:"boxes1d,omitempty" bson:"boxes1d,omitempty" yaml:"boxes1d,omitempty"`
	R           string `json:"r,omitempty" bson:"r,omitempty" yaml:"r,omitempty"`
	Lat         string `json:"lat,omitempty" bson:"lat,omitempty" yaml:"lat,omitempty"`
	Lng         string `json:"lng,omitempty" bson:"lng,omitempty" yaml:"lng,omitempty"`
	Thread      string `json:"thread,omitempty" bson:"thread,omitempty" yaml:"thread,omitempty"`
	Block       string `json:"block,omitempty" bson:"block,omitempty" yaml:"block,omitempty"`
	BlockSize   string `json:"blockSize,omitempty" bson:"blockSize,omitempty" yaml:"blockSize,omitempty"`
	Clusters    string `json:"clusters,omitempty" bson:"clusters,omitempty" yaml:"clusters,omitempty"`
	Features    string `json:"features,omitempty" bson:"features,omitempty" yaml:"features,omitempty"`
	Iteration   string `json:"iteration,omitempty" bson:"iteration,omitempty" yaml:"iteration,omitempty"`
	Iters       string `json:"iters,omitempty" bson:"iters,omitempty" yaml:"iters,omitempty"`
	Degree      string `json:"degree,omitempty" bson:"degree,omitempty" yaml:"degree,omitempty"`
}

func GetAllCases(client *mongo.Client) ([]CaseKey, error) {
	collection := client.Database(config.C.DB.Database).Collection("profiles")

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$unionWith", Value: bson.D{
				{Key: "coll", Value: "traces"},
				{Key: "pipeline", Value: bson.A{}},
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "env_id", Value: "$env_id"},
					{Key: "suite", Value: "$suite"},
					{Key: "benchmark", Value: "$benchmark"},
					{Key: "param", Value: "$param"},
				}},
			}},
		},
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "env_id", Value: "$_id.env_id"},
				{Key: "suite", Value: "$_id.suite"},
				{Key: "benchmark", Value: "$_id.benchmark"},
				{Key: "param", Value: "$_id.param"},
			}},
		},
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []CaseKey
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func GetCasesByEnvID(client *mongo.Client, envID primitive.ObjectID) ([]CaseKey, error) {
	collection := client.Database("your_database").Collection("profiles")

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "env_id", Value: envID},
			}},
		},
		bson.D{
			{Key: "$unionWith", Value: bson.D{
				{Key: "coll", Value: "traces"},
				{Key: "pipeline", Value: bson.A{
					bson.D{
						{Key: "$match", Value: bson.D{
							{Key: "env_id", Value: envID},
						}},
					},
				}},
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "env_id", Value: "$env_id"},
					{Key: "suite", Value: "$suite"},
					{Key: "benchmark", Value: "$benchmark"},
					{Key: "param", Value: "$param"},
				}},
			}},
		},
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "env_id", Value: "$_id.env_id"},
				{Key: "suite", Value: "$_id.suite"},
				{Key: "benchmark", Value: "$_id.benchmark"},
				{Key: "param", Value: "$_id.param"},
			}},
		},
	}

	cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []CaseKey
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
