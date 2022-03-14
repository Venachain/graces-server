package dao

import (
	"context"

	"graces/db"
	"graces/model"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionNameContract = "contracts"
)

var (
	DefaultContractDao IContractDao
)

func init() {
	DefaultContractDao = newContractDao()
}

func newContractDao() IContractDao {
	return &contractDao{db.DefaultDB}
}

type contractDao struct {
	*db.DB
}

func (d *contractDao) InsertContract(contract model.Contract) error {
	collection := d.Db.Collection(collectionNameContract)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.InsertOne(ctx, contract)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	logrus.Debugf("insert: %+v", contract)
	return nil
}

func (d *contractDao) Contract(filter interface{}) (*model.Contract, error) {
	collection := d.Db.Collection(collectionNameContract)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	var contract model.Contract
	err := collection.FindOne(ctx, filter).Decode(&contract)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("filter: %+v, result: %+v", filter, contract)
	return &contract, nil
}
func (d *contractDao) Contracts(filter interface{}, findOps *options.FindOptions) ([]*model.Contract, error) {
	collection := d.Db.Collection(collectionNameContract)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	cursor, err := collection.Find(ctx, filter, findOps)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	results := make([]*model.Contract, 0)
	if err = cursor.All(ctx, &results); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, results)
	return results, nil
}

func (d *contractDao) Count(filter interface{}, countOps *options.CountOptions) (int64, error) {
	collection := d.Db.Collection(collectionNameContract)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	count, err := collection.CountDocuments(ctx, filter, countOps)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (d *contractDao) Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error {
	collection := d.Db.Collection(collectionNameContract)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.UpdateOne(ctx, filter, update, updateOps)
	if err != nil {
		return err
	}
	logrus.Debugf("filter: %+v, update: %+v", filter, update)
	return nil
}
