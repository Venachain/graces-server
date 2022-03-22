package model

import (
	"fmt"
	"graces/exterr"
	"graces/util"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chain struct {
	ID          primitive.ObjectID     `json:"id" bson:"_id"`
	Name        string                 `json:"name" bson:"name"`
	Username    string                 `json:"username" bson:"username"`
	IP          string                 `json:"ip" bson:"ip"`
	RPCPort     uint64                 `json:"rpc_port" bson:"rpc_port"`
	P2PPort     uint64                 `json:"p2p_port" bson:"p2p_port"`
	WSPort      uint64                 `json:"ws_port" bson:"ws_port"`
	Desc        string                 `json:"desc" bson:"desc"`
	ChainConfig map[string]interface{} `json:"chain_config" bson:"chain_config"`
	CreateTime  int64                  `json:"create_time" bson:"create_time"`
	UpdateTime  int64                  `json:"update_time" bson:"update_time"`
	DeleteTime  int64                  `json:"delete_time" bson:"delete_time"`
}

type ChainDTO struct {
	// 主键ID
	ID string `json:"id" swaggerignore:"true"`
	// 链名称
	Name string `json:"name" binding:"required,min=1,max=100"`
	// 在服务端部署链时所需的服务端操作用户
	Username string `json:"username" swaggerignore:"true"`
	// 链第一个节点的 IP 地址
	IP string `json:"ip" binding:"required,min=7,max=15"`
	// 链第一个节点的 rpc 端口号
	RPCPort uint64 `json:"rpc_port" binding:"required,min=0,max=65535"`
	// 链第一个节点的 P2P 端口号
	P2PPort uint64 `json:"p2p_port" binding:"required,min=0,max=65535"`
	// 链第一个节点的 websocket 端口号
	WSPort uint64 `json:"ws_port" binding:"required,min=0,max=65535"`
	// 链的描述信息
	Desc string `json:"desc" binding:"max=200"`
	// 链第一个节点的配置信息
	ChainConfig map[string]interface{} `json:"chain_config"`
}

type ChainQueryCondition struct {
	PageDTO
	SortDTO
	// 主键ID
	ID string `json:"id"`
	// 链名称
	Name string `json:"name" binding:"max=20"`
	// 链第一个节点的 IP 地址
	IP string `json:"ip" binding:"max=15"`
	// 链第一个节点的 rpc 端口号
	RPCPort uint64 `json:"rpc_port"`
	// 链第一个节点的 p2p 端口号
	P2PPort int64 `json:"p2p_port"`
	// 链第一个节点的 websocket 端口号
	WSPort uint64 `json:"ws_port"`
}

type ChainVO struct {
	// 主键ID
	ID string `json:"id"`
	// 链名称
	Name string `json:"name"`
	// 在服务端部署链时所需的服务端操作用户
	Username string `json:"username"`
	// 链第一个节点的 IP 地址
	IP string `json:"ip"`
	// 链第一个节点的 rpc 端口号
	RPCPort uint64 `json:"rpc_port"`
	// 链第一个节点的 p2p 端口号
	P2PPort uint64 `json:"p2p_port"`
	// 链第一个节点的 websocket 端口号
	WSPort uint64 `json:"ws_port"`
	// 链的描述信息
	Desc string `json:"desc"`
	// 链第一个节点的配置信息
	ChainConfig map[string]interface{} `json:"chain_config"`
	// 创建时间
	CreateTime string `json:"create_time"`
	// 最后一次更新时间
	UpdateTime string `json:"update_time"`
	// 删除时间
	DeleteTime string `json:"delete_time"`
}

func (chain *Chain) ValueOfDTO(dto ChainDTO) error {
	if err := util.SimpleCopyProperties(chain, dto); err != nil {
		logrus.Errorln(err)
		return exterr.ErrConvert
	}
	chain.ID = primitive.NewObjectID()
	chain.CreateTime = time.Now().Unix()
	return nil
}

func (chain *Chain) ToVO() (*ChainVO, error) {
	var vo ChainVO
	if err := util.SimpleCopyProperties(&vo, chain); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	if vo.ChainConfig != nil {
		if _, ok := vo.ChainConfig["node"].(map[string]interface{}); ok {
			vo.ChainConfig["node"].(map[string]interface{})["passphrase"] = "******"
		}
	}
	vo.ID = chain.ID.Hex()
	vo.CreateTime = util.Timestamp2TimeStr(chain.CreateTime)
	vo.UpdateTime = util.Timestamp2TimeStr(chain.UpdateTime)
	vo.DeleteTime = util.Timestamp2TimeStr(chain.DeleteTime)
	return &vo, nil
}

func (chain *Chain) GetChainEndPoint() string {
	return fmt.Sprintf("%v:%v", chain.IP, chain.RPCPort)
}
