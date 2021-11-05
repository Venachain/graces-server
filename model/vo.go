package model

type SystemConfigVO struct {
	// 所属链ID
	ChainID string `json:"chainID, omitempty" bson:"chainID"`
	// 设置区块 Gas 限制
	BlockGasLimit string `json:"blockGasLimit, omitempty" bson:"blockGasLimit"`
	// 设置交易 Gas 限制
	TxGasLimit string `json:"txGasLimit, omitempty" bson:"txGasLimit"`
	// 设置是否使用指定合约提供的 Gas，需与 gasContractName 共同使用才能起作用
	IsUseGas string `json:"isUseGas, omitempty" bson:"isUseGas"`
	// 设置是否允许部署合约
	IsApproveDeployedContract string `json:"isApproveDeployedContract, omitempty" bson:"isApproveDploeyedContract"`
	// 设置是否检查合约部署权限
	IsCheckDeployPermission string `json:"isCheckDeployPermission, omitempty" bson:"isCheckDeployPermission"`
	// 设置是否允许出空块
	IsProduceEmptyBlock string `json:"isProduceEmptyBlock, omitempty" bson:"isProduceEmptyBlock"`
	// 设置交易所消耗的 Gas 由指定的合约名称来提供
	GasContractName string `json:"gasContractName, omitempty" bson:"gasContractName"`
}

type SyncNodeResult struct {
	// 节点当前最新区块的块高
	BlockNumber uint32 `json:"blocknumber"`
	// 是否处于挖矿状态
	IsMining bool `json:"ismining"`
	// Gas 价格
	GasPrice uint32 `json:"gasprice"`
	// 节点交易池的 pending 交易的哈希列表
	PendingTx []string `json:"pendingtx"`
	// 节点交易池 pending 交易的数量
	PendingNumber int `json:"pendingnumber"`
}

type FireWall struct {
	// 所属链ID
	Chainid string `json:"chainid"`
	// 合约地址
	ContractAddress string `json:"contractAddress"`
}

type StatsVO struct {
	LatestBlock   uint64 `json:"latest_height" bson:"latest_height"`
	TotalTx       int64  `json:"total_tx" bson:"total_tx"`
	TotalContract int64  `json:"total_contract" bson:"total_contract"`
	TotalNode     int64  `json:"total_node" bson:"total_node"`
}

type TxStatsVO struct {
	Date     string `json:"date" bson:"date"`
	TxAmount int    `json:"tx_amount" bson:"tx_amount"`
}
