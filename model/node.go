package model

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/util"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sirupsen/logrus"
)

type Node struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	ChainID    primitive.ObjectID `json:"chain_id" bson:"chain_id"`
	Name       string             `json:"name" bson:"name"`
	PublicKey  string             `json:"public_key" bson:"public_key"`
	Desc       string             `json:"desc" bson:"desc"`
	InternalIP string             `json:"internal_ip" bson:"internal_ip"`
	ExternalIP string             `json:"external_ip" bson:"external_ip"`
	RPCPort    int                `json:"rpc_port" bson:"rpc_port"`
	P2PPort    int                `json:"p2p_port" bson:"p2p_port"`
	Type       int                `json:"type" bson:"type"`
	Status     int                `json:"status" bson:"status"`
	Owner      string             `json:"owner" bson:"owner"`
}

type NodeDTO struct {
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 节点名称
	Name string `json:"name"`
	// 节点公钥
	PublicKey string `json:"public_key"`
	// 节点描述
	Desc string `json:"desc"`
	// 节点内网 IP
	InternalIP string `json:"internal_ip"`
	// 节点公网 IP
	ExternalIP string `json:"external_ip"`
	// 节点 RPC 端口号
	RPCPort int `json:"rpc_port"`
	// 节点 P2P 端口号
	P2PPort int `json:"p2p_port"`
}

type NodeQueryCondition struct {
	PageDTO
	SortDTO
	// 主键ID
	ID string `json:"id"`
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 节点名称
	Name string `json:"name"`
	// 内网 IP 地址
	InternalIP string `json:"internal_ip"`
	// 公网 IP 地址
	ExternalIP string `json:"external_ip"`
}

type NodeVO struct {
	// 节点 id
	ID string `json:"id"`
	// 所属链 id
	ChainID string `json:"chain_id"`
	// 节点名称
	Name string `json:"name"`
	// 节点公钥
	PublicKey string `json:"public_key"`
	// 节点描述
	Desc string `json:"desc"`
	// 节点内网 IP
	InternalIP string `json:"internal_ip"`
	// 节点公网 IP
	ExternalIP string `json:"external_ip"`
	// 节点 RPC 端口号
	RPCPort int `json:"rpc_port"`
	// 节点 P2P 端口号
	P2PPort int `json:"p2p_port"`
	// 节点类型
	Type int `json:"type"`
	// 节点状态
	Status int `json:"status"`
	// 节点部署人
	Owner string `json:"owner"`
	// 节点是否在线
	IsAlive bool `json:"is_alive"`
	// 节点当前最新区块的块高
	Blocknumber uint32 `json:"blocknumber"`
}

func (node *Node) ValueOfDTO(dto NodeDTO) error {
	if err := util.SimpleCopyProperties(node, dto); err != nil {
		logrus.Errorln(err)
		return exterr.ErrConvert
	}
	node.ID = primitive.NewObjectID()
	cid, err := primitive.ObjectIDFromHex(dto.ChainID)
	if err != nil {
		return err
	}
	node.ChainID = cid
	return nil
}

func (node *Node) ToVO() (*NodeVO, error) {
	var vo NodeVO
	if err := util.SimpleCopyProperties(&vo, node); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.ID = node.ID.Hex()
	vo.ChainID = node.ChainID.Hex()
	// isAlive 现在在请求获取节点最新区块高度之后来判断
	//vo.IsAlive = node.NodeIsAlive()
	return &vo, nil
}

// NodeIsAlive 验证节点是否存活
//func (node *Node) NodeIsAlive() bool {
//	address := fmt.Sprintf("%s:%d", node.InternalIP, node.P2PPort)
//	conn, err := net.Dial("tcp", address)
//	if nil != err {
//		return false
//	}
//	defer func() {
//		if conn != nil {
//			err = conn.Close()
//			if err != nil {
//				return
//			}
//		}
//	}()
//
//	timeout := time.Second * 5
//	err = conn.SetWriteDeadline(time.Now().Add(timeout))
//	if nil != err {
//		return false
//	}
//
//	_, err = conn.Write([]byte("ping"))
//	if nil != err {
//		return false
//	}
//	return true
//}
