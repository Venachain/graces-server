package db

import (
	"context"
	"time"

	"graces/config"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	DefaultDB *DB
	ctx       context.Context
)

func init() {
	uri := options.Client().ApplyURI(config.Config.DBConf.Uri())
	ctx, _ = context.WithTimeout(context.Background(), config.Config.DBConf.Timeout*time.Second)
	clientConnect, err := mongo.Connect(ctx, uri)
	if err != nil {
		panic(err)
	}
	DefaultDB = &DB{
		client: clientConnect,
		Db:     clientConnect.Database(config.Config.DBConf.DBName),
	}
	if err = DefaultDB.Ping(); err != nil {
		logrus.Fatalln("failed to connection DB：", err)
	}
	logrus.Info("db successfully connected and pinged.")
}

type DB struct {
	client *mongo.Client
	Db     *mongo.Database
}

func (db *DB) Collection(name string) *mongo.Collection {
	return db.Db.Collection(name)
}

func (db *DB) Ping() error {
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	if err := db.client.Ping(ctx, readpref.Primary()); nil != err {
		return err
	}
	return nil
}

func BuildOptionsByQuery(pageIndex, pageSize int64) *options.FindOptions {
	findOps := options.Find()
	findOps.SetSkip((pageIndex - 1) * pageSize)
	findOps.SetLimit(pageSize)

	return findOps
}
