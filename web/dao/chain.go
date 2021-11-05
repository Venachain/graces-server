package dao

import (
	"PlatONE-Graces/db"
	"PlatONE-Graces/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
)

const (
	collectionNameChains = "chains"
	defaultTimeout       = 30 * time.Second
)

var (
	DefaultChainDao IChainDao
)

func init() {
	DefaultChainDao = newChainDao(db.DefaultDB)
}

type chainDao struct {
	*db.DB
}

func newChainDao(db *db.DB) *chainDao {
	return &chainDao{db}
}

func (d *chainDao) InsertChain(chain model.Chain) error {
	collection := d.Db.Collection(collectionNameChains)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.InsertOne(ctx, chain)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	return nil
}

func (d *chainDao) Chain(filter interface{}) (*model.Chain, error) {
	collection := d.Db.Collection(collectionNameChains)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	var c model.Chain
	err := collection.FindOne(ctx, filter).Decode(&c)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, c)
	return &c, nil
}

func (d *chainDao) Chains(filter interface{}, findOps *options.FindOptions) ([]*model.Chain, error) {
	collection := d.Db.Collection(collectionNameChains)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	cursor, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := make([]*model.Chain, 0)
	if err := cursor.All(ctx, &results); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, results)
	return results, nil
}

func (d *chainDao) Count(filter interface{}, countOps *options.CountOptions) (int64, error) {
	collection := d.Db.Collection(collectionNameChains)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	count, err := collection.CountDocuments(ctx, filter, countOps)
	if err != nil {
		return 0, err
	}
	return count, nil
}
