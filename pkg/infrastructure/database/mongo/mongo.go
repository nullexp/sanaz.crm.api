package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Host string

	Port int

	Username string

	Password string

	Name string

	Driver string

	IgnorePermissionDenied bool
}

func (config *Config) DSN() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d",

		config.Username, config.Password, config.Host, config.Port)
}

func (config *Config) RawDSN() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d",

		config.Username, config.Password, config.Host, config.Port)
}

type MongoDatabase struct {
	Database *mongo.Database

	DBConfig Config
}

func NewMongoDatabase(DBConfig Config) *MongoDatabase {
	md := MongoDatabase{}

	md.DBConfig = DBConfig

	return &md
}

func (md *MongoDatabase) GetMongo() *mongo.Database {
	return md.Database
}

func (md *MongoDatabase) open() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	// Use the SetServerAPIOptions() method to set the Stable API version to 1

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(md.DBConfig.DSN()).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	md.Database = client.Database(md.DBConfig.Name)

	return nil
}

func (md *MongoDatabase) Open() error {
	return md.open()
}

// TODO: ??

func (md *MongoDatabase) OpenRaw() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	// Use the SetServerAPIOptions() method to set the Stable API version to 1

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(md.DBConfig.DSN()).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	md.Database = client.Database(md.DBConfig.Name)

	return nil
}
