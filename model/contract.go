package model

import (
	"graces/exterr"
	"graces/util"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contract struct {
	// 主键ID
	ID primitive.ObjectID `bson:"_id" json:"id"`
	// 所属链ID
	ChainID primitive.ObjectID `bson:"chain_id" json:"chain_id"`
	// 合约地址
	Address string `bson:"address" json:"address"`
	// 合约创建人地址
	Creator string `bson:"creator" json:"creator"`
	// 部署合约时的交易哈希
	TxHash string `bson:"tx_hash" json:"tx_hash"`
	// 合约内容（对应的是 tx 的 input 数据）
	Content string `bson:"content" json:"content"`
	// 部署时间
	Timestamp int64 `bson:"timestamp" json:"timestamp"`
}

type ContractVO struct {
	// 主键ID
	ID string `bson:"_id" json:"id"`
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 合约地址
	Address string `json:"address"`
	// CNS名称
	CNS []*CNSVO `json:"cns"`
	// 合约创建人地址
	Creator string `json:"creator"`
	// 部署合约时的交易哈希
	TxHash string `json:"tx_hash"`
	// 合约内容
	Content interface{} `json:"content"`
	// 部署时间
	Timestamp string `json:"timestamp"`
}

// ContractQueryCondition 合约查询条件
type ContractQueryCondition struct {
	PageDTO
	SortDTO
	// 部署合约的交易ID
	ID string `json:"id" binding:"min=0,max=50"`
	// 所属链ID
	ChainID string `json:"chain_id" binding:"min=0,max=50"`
	// 合约地址
	Address string `json:"address" binding:"min=0,max=70"`
	// CNS名称
	CNSName string `json:"name" binding:"min=0,max=100"`
	// 合约创建人地址
	Creator string `json:"creator" binding:"min=0,max=70"`
	// 部署合约时的交易哈希
	TxHash string `json:"tx_hash" binding:"min=0,max=70"`
	// 起始时间
	TimeStart int64 `json:"time_start"`
	// 终止时间
	TimeEnd int64 `json:"time_end"`
}

// ContractByAddressDTO 通过合约地址查询合约数据
type ContractByAddressDTO struct {
	// 所属链ID
	ChainID string `json:"chain_id" binding:"min=0,max=50"`
	// 合约地址
	ContractAddress string `json:"contract_address" binding:"required,min=1,max=70"`
}

// ContractCallResult 合约调用返回结果
type ContractCallResult struct {
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 合约调用状态
	Status string `json:"status"`
	// 合约调用日志
	Logs []string `json:"logs"`
	// 合约调用产生的交易所在的区块高度
	BlockNumber uint64 `json:"block_number"`
	// 合约调用的 Gas 消耗
	GasUsed uint64 `json:"gas_used"`
	// 合约调用产生的交易的发起人
	From string `json:"from"`
	// 合约调用产生的交易的交易目标
	To string `json:"to"`
	// 合约调用产生的交易的交易哈希
	TxHash string `json:"tx_hash"`
	// 合约调用产生的错误信息
	ErrMsg string `json:"err_msg"`
}

func (contract *Contract) ToVO() (*ContractVO, error) {
	var vo ContractVO
	if err := util.SimpleCopyProperties(&vo, contract); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.ID = contract.ID.Hex()
	vo.ChainID = contract.ChainID.Hex()
	vo.Timestamp = util.Timestamp2TimeStr(contract.Timestamp / 1e3)
	return &vo, nil
}
