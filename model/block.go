package model

import (
	"graces/exterr"
	"graces/util"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Block struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	ChainID    primitive.ObjectID `json:"chain_id" bson:"chain_id"`
	Hash       string             `json:"hash" bson:"hash"`
	Height     uint64             `json:"height" bson:"height"`
	Timestamp  int64              `json:"timestamp" bson:"timestamp"`
	TxAmount   uint64             `json:"tx_amount" bson:"tx_amount"`
	Proposer   string             `json:"proposer" bson:"proposer"`
	GasUsed    uint64             `json:"gas_used" bson:"gas_used"`
	GasLimit   uint64             `json:"gas_limit" bson:"gas_limit"`
	ParentHash string             `json:"parent_hash" bson:"parent_hash"`
	ExtraData  string             `json:"extra_data" bson:"extra_data"`
	Size       string             `json:"size" bson:"size"`
	Head       *BLockHead         `json:"head" bson:"head"`
}

type BLockHead struct {
	ParentHash       string `json:"parent_hash" bson:"parent_hash"`
	Miner            string `json:"miner" bson:"miner"`
	StateRoot        string `json:"state_root" bson:"state_root"`
	TransactionsRoot string `json:"transactions_root" bson:"transactions_root"`
	ReceiptsRoot     string `json:"receipts_root" bson:"receipts_root"`
	LogsBloom        string `json:"logs_bloom" bson:"logs_bloom"`
	Height           uint64 `json:"height" bson:"height"`
	GasLimit         uint64 `json:"gas_limit" bson:"gas_limit"`
	GasUsed          uint64 `json:"gas_used" bson:"gas_used"`
	Timestamp        int64  `json:"timestamp" bson:"timestamp"`
	ExtraData        string `json:"extra_data" bson:"extra_data"`
	MixHash          string `json:"mix_hash" bson:"mix_hash"`
	Nonce            uint64 `json:"nonce" bson:"nonce"`
	Hash             string `json:"hash" bson:"hash"`
}

type BlockByHashDTO struct {
	// 所属链ID
	ChainID string `json:"chain_id" binding:"min=0,max=50"`
	// 区块哈希
	Hash string `json:"hash" binding:"required,min=1,max=70"`
}

type BlockQueryCondition struct {
	PageDTO
	SortDTO
	// 主键ID
	ID string `json:"id" binding:"min=0,max=50"`
	// 所属链ID
	ChainID string `json:"chain_id" binding:"min=0,max=50"`
	// 挖出该区块的矿工地址
	Proposer string `json:"proposer" binding:"min=0,max=70"`
	// 区块哈希
	Hash string `json:"hash" binding:"min=0,max=70"`
	// 区块高度
	Height uint64 `json:"height" binding:"min=0"`
	// 起始时间
	TimeStart int64 `json:"time_start"`
	// 终止时间
	TimeEnd int64 `json:"time_end"`
}

type BlockVO struct {
	// 主键ID
	ID string `json:"id"`
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 区块哈希
	Hash string `json:"hash"`
	// 区块高度
	Height uint64 `json:"height"`
	// 区块生成时间
	Timestamp string `json:"timestamp"`
	// 区块内部交易数量
	TxAmount uint64 `json:"tx_amount"`
	// 挖出该区块的矿工地址
	Proposer string `json:"proposer"`
	// 区块内部所有交易的 Gas 使用总量
	GasUsed uint64 `json:"gas_used"`
	// 区块内部所有交易的 Gas 限制量
	GasLimit uint64 `json:"gas_limit"`
	// 上一个区块的哈希
	ParentHash string `json:"parent_hash"`
	// 区块的额外信息
	ExtraData string `json:"extra_data"`
	// 区块大小
	Size string `json:"size"`
	// 区块头信息
	Head *BLockHeadVO `json:"head"`
}

type BLockHeadVO struct {
	// 上一个区块的哈希
	ParentHash string `json:"parent_hash" bson:"parent_hash"`
	// 挖出该区块的矿工地址
	Miner string `json:"miner" bson:"miner"`
	// merkle 状态树的根哈希
	StateRoot string `json:"state_root" bson:"state_root"`
	// merkle 交易树的根哈希
	TransactionsRoot string `json:"transactions_root" bson:"transactions_root"`
	// merkle 收据树的根哈希
	ReceiptsRoot string `json:"receipts_root" bson:"receipts_root"`
	// LogsBloom
	LogsBloom string `json:"logs_bloom" bson:"logs_bloom"`
	// 区块高度
	Height uint64 `json:"height" bson:"height"`
	// 区块内部所有交易的 Gas 限制量
	GasLimit uint64 `json:"gas_limit" bson:"gas_limit"`
	// 区块内部所有交易的 Gas 使用总量
	GasUsed uint64 `json:"gas_used" bson:"gas_used"`
	// 区块生成时间
	Timestamp string `json:"timestamp" bson:"timestamp"`
	// 区块的额外信息
	ExtraData string `json:"extra_data" bson:"extra_data"`
	// 混合哈希
	MixHash string `json:"mix_hash" bson:"mix_hash"`
	// 区块 POW 随机数
	Nonce uint64 `json:"nonce" bson:"nonce"`
	// 区块哈希
	Hash string `json:"hash" bson:"hash"`
}

func (block *Block) ToVO() (*BlockVO, error) {
	var vo BlockVO
	if err := util.SimpleCopyProperties(&vo, block); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.ID = block.ID.Hex()
	vo.ChainID = block.ChainID.Hex()
	vo.Timestamp = util.Timestamp2TimeStr(block.Timestamp / 1e3)
	headVO, err := block.Head.ToVO()
	if err != nil {
		return nil, err
	}
	vo.Head = headVO
	return &vo, err
}

func (head *BLockHead) ToVO() (*BLockHeadVO, error) {
	var vo BLockHeadVO
	if err := util.SimpleCopyProperties(&vo, head); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.Timestamp = util.Timestamp2TimeStr(head.Timestamp / 1e3)
	return &vo, nil
}
