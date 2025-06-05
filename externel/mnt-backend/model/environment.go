package model

import (
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const envCollection = "environments"

type DBEnv struct {
	ID     primitive.ObjectID `bson:"_id" json:"id"`
	EnvKey `bson:",inline"`
}

type EnvKey struct {
	GPU         string `bson:"gpu" json:"gpu"`
	Machine     string `bson:"machine" json:"machine"`
	CUDAVersion string `bson:"cuda_version" json:"cuda_version"`
}

func InsertEnv(client *mongo.Client, env DBEnv) (DBEnv, error) {
	col := client.Database(config.C.DB.Database).Collection(envCollection)
	env.ID = primitive.NewObjectID()

	_, err := col.InsertOne(ctx, env)
	return env, err
}

func GetAllEnvs(client *mongo.Client) ([]DBEnv, error) {
	col := client.Database(config.C.DB.Database).Collection(envCollection)
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	envs := []DBEnv{}
	err = cursor.All(ctx, &envs)
	return envs, err
}

func FindEnvByEnvKey(client *mongo.Client, key EnvKey) (DBEnv, error) {
	col := client.Database(config.C.DB.Database).Collection(envCollection)

	var envFound DBEnv
	err := col.FindOne(ctx, key).Decode(&envFound)
	return envFound, err
}

func FindEnvByID(client *mongo.Client, idObj primitive.ObjectID) (DBEnv, error) {
	col := client.Database(config.C.DB.Database).Collection(envCollection)
	filter := bson.M{"_id": idObj}

	var env DBEnv
	err := col.FindOne(ctx, filter).Decode(&env)
	return env, err
}
