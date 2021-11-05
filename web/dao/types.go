package dao

import (
	"PlatONE-Graces/model"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChainDao interface {
	InsertChain(chain model.Chain) error
	Chain(filter interface{}) (*model.Chain, error)
	Chains(filter interface{}, findOps *options.FindOptions) ([]*model.Chain, error)
	Count(filter interface{}, countOps *options.CountOptions) (int64, error)
}

type IWSMsgDao interface {
	WSMsg(filter interface{}) (*model.WSMsg, error)
	InsertWSMsg(msg model.WSMsg) error
	UpdateWSMsg(filter interface{}, update interface{}) error
	UpdateWSMsgHash(msgID string, topic string, msgHash string) error
}

type IBlockDao interface {
	Block(filter interface{}) (*model.Block, error)
	Blocks(filter interface{}, findOps *options.FindOptions) ([]*model.Block, error)
	Count(filter interface{}, countOps *options.CountOptions) (int64, error)
	LatestBlock(chainID primitive.ObjectID) (*model.Block, error)
	InsertBlock(block model.Block) error
	Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error
}

type ITXDao interface {
	InsertTX(tx model.TX) error
	TX(filter interface{}) (*model.TX, error)
	Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error
	TXs(filter interface{}, findOps *options.FindOptions) ([]*model.TX, error)
	Count(filter interface{}, countOps *options.CountOptions) (int64, error)
}

type INodeDao interface {
	Node(filter interface{}) (*model.Node, error)
	InsertNode(node model.Node) error
	Nodes(filter interface{}, findOps *options.FindOptions) ([]*model.Node, error)
	Count(filter interface{}, countOps *options.CountOptions) (int64, error)
	Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error
}

type ICNSDao interface {
	InsertCNS(cns model.CNS) error
	CNS(filter interface{}) (*model.CNS, error)
	CNSs(filter interface{}, findOps *options.FindOptions) ([]*model.CNS, error)
	Count(filter interface{}, countOps *options.CountOptions) (int64, error)
	Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error
}

type IContractDao interface {
	InsertContract(contract model.Contract) error
	Contract(filter interface{}) (*model.Contract, error)
	Contracts(filter interface{}, findOps *options.FindOptions) ([]*model.Contract, error)
	Count(filter interface{}, countOps *options.CountOptions) (int64, error)
	Update(filter interface{}, update interface{}, updateOps *options.UpdateOptions) error
}
