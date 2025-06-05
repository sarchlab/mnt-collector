package model

import (
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const traceCollection = "traces"

type DBTrace struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	CaseKey `bson:",inline"`
	S3Path  string `bson:"s3_path" json:"s3_path"`
	Size    string `bson:"size" json:"size"`
}

func InsertTrace(client *mongo.Client, trace DBTrace) (DBTrace, error) {
	col := client.Database(config.C.DB.Database).Collection(traceCollection)
	trace.ID = primitive.NewObjectID()

	_, err := col.InsertOne(ctx, trace)
	return trace, err
}

func GetAllTraces(client *mongo.Client) ([]DBTrace, error) {
	col := client.Database(config.C.DB.Database).Collection(traceCollection)
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	traces := []DBTrace{}
	err = cursor.All(ctx, &traces)
	return traces, err
}

func FindTraceByCaseKey(client *mongo.Client, key CaseKey) (DBTrace, error) {
	col := client.Database(config.C.DB.Database).Collection(traceCollection)

	var traceFound DBTrace
	err := col.FindOne(ctx, key).Decode(&traceFound)
	return traceFound, err
}

func FindTraceByID(client *mongo.Client, idObj primitive.ObjectID) (DBTrace, error) {
	col := client.Database(config.C.DB.Database).Collection(traceCollection)
	filter := bson.M{"_id": idObj}

	var trace DBTrace
	err := col.FindOne(ctx, filter).Decode(&trace)
	return trace, err
}

func UpdateTrace(client *mongo.Client, idObj primitive.ObjectID, trace DBTrace) (DBTrace, error) {
	col := client.Database(config.C.DB.Database).Collection(traceCollection)
	filter := bson.M{"_id": idObj}
	trace.ID = idObj
	update := bson.M{"$set": trace}

	res, err := col.UpdateOne(ctx, filter, update)
	if res.MatchedCount == 0 {
		return DBTrace{}, mongo.ErrNoDocuments
	}
	return trace, err
}

func GetTracesByEnvID(client *mongo.Client, envIDObj primitive.ObjectID) ([]DBTrace, error) {
	col := client.Database(config.C.DB.Database).Collection(traceCollection)
	filter := bson.M{"env_id": envIDObj}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	traces := []DBTrace{}
	err = cursor.All(ctx, &traces)
	return traces, err
}
