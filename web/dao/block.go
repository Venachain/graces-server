package dao

import (
	"context"
	"errors"

	"graces/db"
	"graces/model"

	"github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionNameBlock = "blocks"
)

var (
	DefaultBlockDao IBlockDao
)

func init() {
	DefaultBlockDao = newBlockDao()
}

func newBlockDao() *blockDao {
	return &blockDao{db.DefaultDB}
}

type blockDao struct {
	*db.DB
}

func (d *blockDao) Block(filter interface{}) (*model.Block, error) {
	collection := d.Db.Collection(collectionNameBlock)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	var block model.Block
	err := collection.FindOne(ctx, filter).Decode(&block)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("filter: %+v, result: %+v", filter, block)
	return &block, nil
}

func (d *blockDao) Blocks(filter interface{}, findOps *options.FindOptions) ([]*model.Block, error) {
	collection := d.Db.Collection(collectionNameBlock)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	cursor, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := make([]*model.Block, 0)
	if err = cursor.All(ctx, &results); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, results)
	return results, nil
}

func (d *blockDao) Count(filter interface{}, countOps *options.CountOptions) (int64, error) {
	collection := d.Db.Collection(collectionNameBlock)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	count, err := collection.CountDocuments(ctx, filter, countOps)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (d *blockDao) InsertBlock(block model.Block) error {
	collection := d.Db.Collection(collectionNameBlock)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.InsertOne(ctx, block)
	if err != nil {
		return err
	}
	logrus.Debugf("insert: %+v", block)
	return nil
}

func (d *blockDao) Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error {
	collection := d.Db.Collection(collectionNameBlock)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.UpdateOne(ctx, filter, update, updateOps)
	if err != nil {
		return err
	}
	logrus.Debugf("filter: %+v, update: %+v", filter, update)
	return nil
}

func (d *blockDao) LatestBlock(chainID primitive.ObjectID) (*model.Block, error) {
	collection := d.Db.Collection(collectionNameBlock)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	pipeline := mongo.Pipeline{
		bson.D{
			{
				"$group", bson.D{
					{"_id", "$chain_id"},
					{"max_height", bson.D{{"$max", "$height"}}},
				}},
		},
		bson.D{
			{"$match", bson.D{
				{"_id", chainID},
			}},
		},
	}
	opts := options.Aggregate()
	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("document is nil")
	}
	filter := bson.M{
		"chain_id": result[0]["_id"],
		"height":   result[0]["max_height"],
	}
	return d.Block(filter)
}
