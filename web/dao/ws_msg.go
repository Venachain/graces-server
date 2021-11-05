package dao

import (
	"PlatONE-Graces/db"
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sirupsen/logrus"
)

const (
	collectionNameWSMessage = "ws_msg"
)

var (
	DefaultWSMsgDao IWSMsgDao
)

func init() {
	DefaultWSMsgDao = newWSMsgDao()
}

func newWSMsgDao() *wsMsgDao {
	return &wsMsgDao{
		DB: db.DefaultDB,
	}
}

type wsMsgDao struct {
	*db.DB
}

func (d *wsMsgDao) InsertWSMsg(msg model.WSMsg) error {
	collection := d.Db.Collection(collectionNameWSMessage)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.InsertOne(ctx, msg)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	logrus.Debugf("insert: %+v", msg)
	return nil
}

func (d *wsMsgDao) UpdateWSMsg(filter interface{}, update interface{}) error {
	collection := d.Db.Collection(collectionNameWSMessage)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	logrus.Debugf("filter: %+v, update: %+v", filter, update)
	return nil
}

func (d *wsMsgDao) UpdateWSMsgHash(msgID string, topic string, msgHash string) error {
	id, err := primitive.ObjectIDFromHex(msgID)
	if err != nil {
		return exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"_id":              id,
		"extra.topic.name": topic,
	}
	update := bson.M{
		"$set": bson.M{
			"hash": msgHash,
		},
	}
	return d.UpdateWSMsg(filter, update)
}

func (d *wsMsgDao) WSMsg(filter interface{}) (*model.WSMsg, error) {
	collection := d.Db.Collection(collectionNameWSMessage)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	var msg model.WSMsg
	err := collection.FindOne(ctx, filter).Decode(&msg)
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	logrus.Debugf("filter: %+v, result: %+v", filter, msg)
	return &msg, nil
}
