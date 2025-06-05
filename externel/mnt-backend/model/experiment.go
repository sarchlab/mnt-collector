package model

import (
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const expCollection = "experiments"

type DBExp struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	ExpKey  `bson:",inline"`
	Message string `json:"message" bson:"message"`
}

type ExpKey struct {
	Version string `json:"version" bson:"version"`
}

func InsertExp(client *mongo.Client, exp DBExp) (DBExp, error) {
	col := client.Database(config.C.DB.Database).Collection(expCollection)
	exp.ID = primitive.NewObjectID()

	_, err := col.InsertOne(ctx, exp)
	return exp, err
}

func GetAllExps(client *mongo.Client) ([]DBExp, error) {
	col := client.Database(config.C.DB.Database).Collection(expCollection)
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var exps []DBExp
	err = cursor.All(ctx, &exps)
	return exps, err
}

func FindExpByExpKey(client *mongo.Client, key ExpKey) (DBExp, error) {
	col := client.Database(config.C.DB.Database).Collection(expCollection)

	var expFound DBExp
	err := col.FindOne(ctx, key).Decode(&expFound)
	return expFound, err
}

func FindExpByID(client *mongo.Client, idObj primitive.ObjectID) (DBExp, error) {
	col := client.Database(config.C.DB.Database).Collection(expCollection)
	filter := bson.M{"_id": idObj}

	var exp DBExp
	err := col.FindOne(ctx, filter).Decode(&exp)
	return exp, err
}
