package dao

import (
	"context"
	"time"

	"graces/db"
	"graces/model"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionNameTX = "txs"
)

var (
	DefaultTXDao ITXDao
)

func init() {
	DefaultTXDao = newTXDao()
}

func newTXDao() *txDao {
	return &txDao{db.DefaultDB}
}

type txDao struct {
	*db.DB
}

func (d *txDao) InsertTX(tx model.TX) error {
	collection := d.Db.Collection(collectionNameTX)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.InsertOne(ctx, tx)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	logrus.Debugf("insert: %+v", tx)
	return nil
}

func (d *txDao) TX(filter interface{}) (*model.TX, error) {
	collection := d.Db.Collection(collectionNameTX)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	var tx model.TX
	err := collection.FindOne(ctx, filter).Decode(&tx)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("filter: %+v, result: %+v", filter, tx)
	return &tx, nil
}

func (d *txDao) TXs(filter interface{}, findOps *options.FindOptions) ([]*model.TX, error) {
	collection := d.Db.Collection(collectionNameTX)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	cursor, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := make([]*model.TX, 0)
	if err = cursor.All(ctx, &results); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, results)
	return results, nil
}

func (d *txDao) Count(filter interface{}, countOps *options.CountOptions) (int64, error) {
	collection := d.Db.Collection(collectionNameTX)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	count, err := collection.CountDocuments(ctx, filter, countOps)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (d *txDao) Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error {
	collection := d.Db.Collection(collectionNameTX)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.UpdateOne(ctx, filter, update, updateOps)
	if err != nil {
		return err
	}
	logrus.Debugf("filter: %+v, update: %+v", filter, update)
	return nil
}

// ??????????????????????????????
func (d *txDao) TXByDate(timestamp int64) (int64, error) {
	now := time.Now().AddDate(0, 0, -1)
	y, m, day := now.Date()
	start := time.Date(y, m, day, 0, 0, 0, 0, time.Local)
	end := time.Date(y, m, day, 23, 59, 59, 0, time.Local)

	collection := d.Db.Collection(collectionNameTX)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	filter := bson.M{}
	filter["timestamp"] = bson.M{"$gte": start, "$lte": end}
	amount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logrus.Errorln(err)
		return 0, err
	}
	return amount, nil
}
