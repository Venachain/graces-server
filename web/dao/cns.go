package dao

import (
	"PlatONE-Graces/db"
	"PlatONE-Graces/model"
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
)

const (
	collectionNameCNS = "cns"
)

var (
	DefaultCNSDao ICNSDao
)

func init() {
	DefaultCNSDao = newCNSDao()
}

func newCNSDao() ICNSDao {
	return &cnsDao{db.DefaultDB}
}

type cnsDao struct {
	*db.DB
}

func (d *cnsDao) InsertCNS(cns model.CNS) error {
	collection := d.Db.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.InsertOne(ctx, cns)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	logrus.Debugf("insert: %+v", cns)
	return nil
}

func (d *cnsDao) CNS(filter interface{}) (*model.CNS, error) {
	collection := d.Db.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	var cns model.CNS
	err := collection.FindOne(ctx, filter).Decode(&cns)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("filter: %+v, result: %+v", filter, cns)
	return &cns, nil
}

func (d *cnsDao) CNSs(filter interface{}, findOps *options.FindOptions) ([]*model.CNS, error) {
	collection := d.Db.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	cursor, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := make([]*model.CNS, 0)
	if err = cursor.All(ctx, &results); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, results)
	return results, nil
}

func (d *cnsDao) Count(filter interface{}, countOps *options.CountOptions) (int64, error) {
	collection := d.Db.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	count, err := collection.CountDocuments(ctx, filter, countOps)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (d *cnsDao) Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error {
	collection := d.Db.Collection(collectionNameCNS)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.UpdateOne(ctx, filter, update, updateOps)
	if err != nil {
		return err
	}
	logrus.Debugf("filter: %+v, update: %+v", filter, update)
	return nil
}
