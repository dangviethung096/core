package core

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDBInfo struct {
	Host     string
	Port     int32
	Username string
	Password string
	Database string
	Timeout  int64
}

// Build connection string from mongo db info
func buildMongoDbConnectionString(info mongoDBInfo) string {
	var connectionString string
	if info.Username == BLANK && info.Password == BLANK {
		connectionString = fmt.Sprintf("mongodb://%s:%d/%s", info.Host, info.Port, info.Database)
	} else {
		connectionString = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", info.Username, info.Password, info.Host, info.Port, info.Database)
	}
	return connectionString
}

// Mongodb session struct
type mongoSession struct {
	client *mongo.Client
}

// Connect to mongo database and return session
func openMongoDBConnection(info mongoDBInfo) mongoSession {
	// Build mongo db connection string
	connectionString := buildMongoDbConnectionString(info)

	// Set client options
	clientOptions := options.Client().ApplyURI(connectionString)

	// Connect to MongoDB
	client, err := mongo.Connect(coreContext, clientOptions)
	if err != nil {
		LoggerInstance.Fatal("Error when connect to MongoDB: %v", err)
	}

	// Check the connection
	err = client.Ping(coreContext, nil)
	if err != nil {
		LoggerInstance.Fatal("Error when ping to MongoDB: %v", err)
	}

	LoggerInstance.Info("Connect to MongoDB success")

	return mongoSession{client: client}
}
