package model

import (
	"graces/exterr"
	"graces/util"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CNS struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	ChainID primitive.ObjectID `json:"chain_id" bson:"chain_id"`
	Name    string             `json:"name" bson:"name"`
	Version string             `json:"version" bson:"version"`
	Address string             `json:"address" bson:"address"`
}

type CNSQueryCondition struct {
	PageDTO
	SortDTO
	// 主键ID
	ID string `json:"id"`
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 合约别名
	Name string `json:"name"`
	// 合约版本号
	Version string `json:"version"`
	// 合约地址
	Address string `json:"address"`
}

// CNSRedirectDTO CNS重定向DTO
type CNSRedirectDTO struct {
	// 链ID
	ChainID string `json:"chain_id" binding:"required"`
	// 合约别名
	Name string `json:"name" binding:"required"`
	// 合约版本号
	Version string `json:"version" binding:"required"`
}

// CNSRegisterDTO CNS注册DTO
type CNSRegisterDTO struct {
	CNSRedirectDTO
	// CNS合约地址
	Address string `json:"address" binding:"required"`
}

type CNSVO struct {
	// CNS ID
	ID string `json:"id"`
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 合约别名
	Name string `json:"name"`
	// 合约版本号
	Version string `json:"version"`
	// 合约地址
	Address string `json:"address"`
}

func (cns *CNS) ToVO() (*CNSVO, error) {
	var vo CNSVO
	if err := util.SimpleCopyProperties(&vo, cns); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.ID = cns.ID.Hex()
	vo.ChainID = cns.ChainID.Hex()
	return &vo, nil
}
