package model

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/util"

	"github.com/sirupsen/logrus"
)

type ChainDataSyncInfo struct {
	// 链ID
	ChainID string `json:"chain_id"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime int64 `json:"start_time"`
	// 预计完成时间
	EstimateCompleteTime int64 `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
	// 区块同步信息
	BlockDataSyncInfo *BlockDataSyncInfo `json:"block_sync_info"`
	// CNS同步信息
	CNSDataSyncInfo *CNSDataSyncInfo `json:"snc_data_sync_info"`
	// 节点同步信息
	NodeDataSyncInfo *NodeDataSyncInfo `json:"node_data_sync_info"`
	// 交易统计信息同步
	//TxStatsInfo NodeDataSyncInfo `json:"node_data_sync_info"`
}

type ChainDataSyncInfoVO struct {
	// 链ID
	ChainID string `json:"chain_id"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime string `json:"start_time"`
	// 预计完成时间
	EstimateCompleteTime string `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
	// 节点同步信息
	NodeDataSyncInfoVO *NodeDataSyncInfoVO `bson:"node_data_sync_info_vo"`
	// 区块同步信息
	BlockDataSyncInfoVO *BlockDataSyncInfoVO `json:"block_data_sync_info"`
	// CNS同步信息
	CNSDataSyncInfoVO *CNSDataSyncInfoVO `json:"cns_data_sync_info_vo"`
}

type BlockDataSyncInfo struct {
	// 链上最新区块的块高
	LatestHeight uint64 `json:"latest_height"`
	// 当前已经同步到的块高
	CurrentHeight uint64 `json:"current_height"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime int64 `json:"start_time"`
	// 同步每个区块的平均耗时（单位：ms）
	BlockSyncTimeAvg int64 `json:"block_sync_time_avg"`
	// 预计完成时间
	EstimateCompleteTime int64 `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
}

type BlockDataSyncInfoVO struct {
	// 链上最新区块的块高
	LatestHeight uint64 `json:"latest_height"`
	// 当前已经同步到的块高
	CurrentHeight uint64 `json:"current_height"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime string `json:"start_time"`
	// 同步每个区块的平均耗时（单位：ms）
	BlockSyncTimeAvg int64 `json:"block_sync_time_avg"`
	// 预计完成时间
	EstimateCompleteTime string `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
}

// CNSDataSyncInfo CNS数据同步信息
type CNSDataSyncInfo struct {
	// 总的 CNS合约映射信息 数量
	Size int `json:"size"`
	// 当前已经同步到的下标
	Index int `json:"index"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime int64 `json:"start_time"`
	// 同步每个 CNS合约映射信息 的平均耗时（单位：ms）
	SyncTimeAvg int64 `json:"sync_time_avg"`
	// 预计完成时间
	EstimateCompleteTime int64 `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
}

type CNSDataSyncInfoVO struct {
	// 总的 CNS合约映射信息 数量
	Size int `json:"size"`
	// 当前已经同步到的下标
	Index int `json:"index"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime string `json:"start_time"`
	// 同步每个 CNS合约映射信息 的平均耗时（单位：ms）
	CNSSyncTimeAvg int64 `json:"block_sync_time_avg"`
	// 预计完成时间
	EstimateCompleteTime string `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
}

// NodeDataSyncInfo 节点数据同步信息
type NodeDataSyncInfo struct {
	// 总的 节点 数量
	Size int `json:"size"`
	// 当前已经同步到的下标
	Index int `json:"index"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime int64 `json:"start_time"`
	// 同步每个节点的平均耗时（单位：ms）
	SyncTimeAvg int64 `json:"sync_time_avg"`
	// 预计完成时间
	EstimateCompleteTime int64 `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
}

type NodeDataSyncInfoVO struct {
	// 总的 节点 数量
	Size int `json:"size"`
	// 当前已经同步到的下标
	Index int `json:"index"`
	// 数据同步状态：同步中（syncing）、同步出错（error）、同步成功（success）
	Status string `json:"status"`
	// 开始时间
	StartTime string `json:"start_time"`
	// 同步每个节点的平均耗时（单位：ms）
	SyncTimeAvg int64 `json:"sync_time_avg"`
	// 预计完成时间
	EstimateCompleteTime string `json:"estimate_complete_time"`
	// 错误信息
	ErrMsg string `json:"err_msg"`
}

// SyncErrMsg 数据同步错误信息
type SyncErrMsg struct {
	// 链ID
	ChainID string
	// 错误类型
	ErrType int
	// 错误信息
	Err error
}

func (info *ChainDataSyncInfo) ToVO() (*ChainDataSyncInfoVO, error) {
	var vo ChainDataSyncInfoVO
	err := util.SimpleCopyProperties(&vo, info)
	if err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.StartTime = util.Timestamp2TimeStr(info.StartTime)
	vo.EstimateCompleteTime = util.Timestamp2TimeStr(info.EstimateCompleteTime)
	nodeDataSyncInfoVO, err := info.NodeDataSyncInfo.ToVO()
	if err != nil {
		return nil, err
	}
	vo.NodeDataSyncInfoVO = nodeDataSyncInfoVO

	cnsDataSyncInfoVO, err := info.CNSDataSyncInfo.ToVO()
	if err != nil {
		return nil, err
	}
	vo.CNSDataSyncInfoVO = cnsDataSyncInfoVO

	blockDataSyncInfoVO, err := info.BlockDataSyncInfo.ToVO()
	if err != nil {
		return nil, err
	}
	vo.BlockDataSyncInfoVO = blockDataSyncInfoVO

	return &vo, nil
}

func (info *BlockDataSyncInfo) ToVO() (*BlockDataSyncInfoVO, error) {
	var vo BlockDataSyncInfoVO
	err := util.SimpleCopyProperties(&vo, info)
	if err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.StartTime = util.Timestamp2TimeStr(info.StartTime)
	vo.EstimateCompleteTime = util.Timestamp2TimeStr(info.EstimateCompleteTime)
	return &vo, nil
}

func (info *CNSDataSyncInfo) ToVO() (*CNSDataSyncInfoVO, error) {
	var vo CNSDataSyncInfoVO
	err := util.SimpleCopyProperties(&vo, info)
	if err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.StartTime = util.Timestamp2TimeStr(info.StartTime)
	vo.EstimateCompleteTime = util.Timestamp2TimeStr(info.EstimateCompleteTime)
	return &vo, nil
}

func (info *NodeDataSyncInfo) ToVO() (*NodeDataSyncInfoVO, error) {
	var vo NodeDataSyncInfoVO
	err := util.SimpleCopyProperties(&vo, info)
	if err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.StartTime = util.Timestamp2TimeStr(info.StartTime)
	vo.EstimateCompleteTime = util.Timestamp2TimeStr(info.EstimateCompleteTime)
	return &vo, nil
}
