package model

import (
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const simCollection = "simulations"

type DBSim struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	SimKey       `bson:",inline"`
	PredictCycle float64 `json:"predict_cycle" bson:"predict_cycle"`
}

type SimKey struct {
	ExpID   primitive.ObjectID `json:"exp_id" bson:"exp_id"`
	TraceID primitive.ObjectID `json:"trace_id" bson:"trace_id"`
}

func InsertSim(client *mongo.Client, sim DBSim) (DBSim, error) {
	col := client.Database(config.C.DB.Database).Collection(simCollection)
	sim.ID = primitive.NewObjectID()

	_, err := col.InsertOne(ctx, sim)
	return sim, err
}

func FindSimBySimKey(client *mongo.Client, simKey SimKey) (DBSim, error) {
	col := client.Database(config.C.DB.Database).Collection(simCollection)

	simFound := DBSim{}
	err := col.FindOne(ctx, simKey).Decode(&simFound)
	return simFound, err
}

func UpdateSim(client *mongo.Client, idObj primitive.ObjectID, sim DBSim) (DBSim, error) {
	col := client.Database(config.C.DB.Database).Collection(simCollection)
	filter := bson.M{"_id": idObj}
	sim.ID = idObj
	update := bson.M{"$set": sim}

	res, err := col.UpdateOne(ctx, filter, update)
	if res.MatchedCount == 0 {
		return DBSim{}, mongo.ErrNoDocuments
	}
	return sim, err
}
