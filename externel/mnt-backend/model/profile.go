package model

import (
	"github.com/sarchlab/mnt-collector/externel/mnt-backend/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const profileCollection = "profiles"

type DBProf struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	CaseKey      `bson:",inline"`
	RepeatTimes  int32   `json:"repeat_times" bson:"repeat_times"`
	AvgNanoSec   float64 `json:"avg_nano_sec" bson:"avg_nano_sec"`
	Frequency    uint32  `json:"frequency" bson:"frequency"`
	MaxFrequency uint32  `json:"max_frequency" bson:"max_frequency"`
}

func InsertProf(client *mongo.Client, prof DBProf) (DBProf, error) {
	col := client.Database(config.C.DB.Database).Collection(profileCollection)
	prof.ID = primitive.NewObjectID()

	_, err := col.InsertOne(ctx, prof)
	return prof, err
}

func GetAllProfs(client *mongo.Client) ([]DBProf, error) {
	col := client.Database(config.C.DB.Database).Collection(profileCollection)
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	profiles := []DBProf{}
	err = cursor.All(ctx, &profiles)
	return profiles, err
}

func FindProfByCaseKey(client *mongo.Client, key CaseKey) (DBProf, error) {
	col := client.Database(config.C.DB.Database).Collection(profileCollection)

	var profileFound DBProf
	err := col.FindOne(ctx, key).Decode(&profileFound)
	return profileFound, err
}

func UpdateProf(client *mongo.Client, idObj primitive.ObjectID, prof DBProf) (DBProf, error) {
	col := client.Database(config.C.DB.Database).Collection(profileCollection)
	filter := bson.M{"_id": idObj}
	prof.ID = idObj
	update := bson.M{"$set": prof}

	res, err := col.UpdateOne(ctx, filter, update)
	if res.MatchedCount == 0 {
		return DBProf{}, mongo.ErrNoDocuments
	}
	return prof, err
}

func GetProfsByEnvID(client *mongo.Client, envIDObj primitive.ObjectID) ([]DBProf, error) {
	col := client.Database(config.C.DB.Database).Collection(profileCollection)
	filter := bson.M{"env_id": envIDObj}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	profiles := []DBProf{}
	err = cursor.All(ctx, &profiles)
	return profiles, err
}
