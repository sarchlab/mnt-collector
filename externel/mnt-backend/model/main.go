package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/sarchlab/mnt-collector/externel/mnt-backend/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.TODO()
var Client *mongo.Client

func Connect() {
	var err error

	// uri := fmt.Sprintf("mongodb://%s:%s@%s%s/%s",
	// 	config.C.DB.User, config.C.DB.Password, config.C.DB.Host, config.C.DB.Port, config.C.DB.Database)
	var uri string
	if strings.Contains(config.C.DB.Host, ".net") {
		// Use the MongoDB+SRV connection string for cloud-hosted MongoDB
		uri = fmt.Sprintf("mongodb+srv://%s:%s@%s/%s",
			config.C.DB.User, config.C.DB.Password, config.C.DB.Host, config.C.DB.Database)
	} else {
		// Use the standard MongoDB connection string for local MongoDB
		uri = fmt.Sprintf("mongodb://%s:%s@%s%s/%s",
			config.C.DB.User, config.C.DB.Password, config.C.DB.Host, config.C.DB.Port, config.C.DB.Database)
	}

	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	err = Client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}
}

func DisConnect() {
	Client.Disconnect(ctx)
}
