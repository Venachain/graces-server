package dao

import (
	"PlatONE-Graces/db"
	"PlatONE-Graces/model"
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
)

const (
	collectionNameNode = "nodes"
)

var (
	DefaultNodeDao INodeDao
)

func init() {
	DefaultNodeDao = newNodeDao()
}

func newNodeDao() *nodeDao {
	return &nodeDao{
		db.DefaultDB,
	}
}

type nodeDao struct {
	*db.DB
}

func (d *nodeDao) InsertNode(node model.Node) error {
	collection := d.Db.Collection(collectionNameNode)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.InsertOne(ctx, node)
	if err != nil {
		logrus.Errorln(err)
		return err
	}

	logrus.Debug("Insert node success：%+v", node)
	return nil
}

func (d *nodeDao) Node(filter interface{}) (*model.Node, error) {
	collection := d.Db.Collection(collectionNameNode)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	var node model.Node
	err := collection.FindOne(ctx, filter).Decode(&node)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("filter: %+v, result: %+v", filter, node)
	return &node, nil
}

func (d *nodeDao) Nodes(filter interface{}, findOps *options.FindOptions) ([]*model.Node, error) {
	collection := d.Db.Collection(collectionNameNode)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	cursor, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := make([]*model.Node, 0)
	if err = cursor.All(ctx, &results); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, results)
	return results, nil
}

func (d *nodeDao) Count(filter interface{}, countOps *options.CountOptions) (int64, error) {
	collection := d.Db.Collection(collectionNameNode)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	count, err := collection.CountDocuments(ctx, filter, countOps)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (d *nodeDao) Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error {
	collection := d.Db.Collection(collectionNameNode)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	_, err := collection.UpdateOne(ctx, filter, update, updateOps)
	if err != nil {
		return err
	}
	logrus.Debugf("update node success, filter：%v, update: %v, updateOps: %v", filter, update, updateOps)
	return nil
}
