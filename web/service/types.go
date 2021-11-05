package service

import "PlatONE-Graces/model"

type IWebsocketService interface {
	// Manager 获取 Websocket Manager 信息
	Manager() model.WSManagerVO

	// Group 获取指定 Websocket 组详细信息
	// name 组名称
	Group(name string) (model.WSGroupVO, error)

	// Groups 获取所有 Websocket 组详细信息
	Groups() ([]model.WSGroupVO, error)

	// Send 向指定组的指定 Websocket 客户端发送信息
	Send(dto model.WSMessageDTO) error

	// SendGroup 向指定组的所有 Websocket 客户端广播信息
	SendGroup(dto model.WSGroupMessageDTO) error

	// SendAll 向所有组的所有 Websocket 客户端广播信息
	SendAll(dto model.WSBroadCastMessageDTO)

	// Dial 作为客户端向其他服务端拨号
	Dial(dto model.WSDialDTO) (model.WSClientVO, error)

	// ClientSend 让作为客户端的 Websocket 向其所连接的服务端发送消息
	ClientSend(dto model.WSMessageDTO) error

	// SaveMessage 保存消息记录
	SaveMessage(dto model.WSMsgDTO) error
	// UpdateMessageHash 更新消息记录的订阅哈希
	UpdateMessageHash(msgID string, msgHash string) error
}

type IChainService interface {
	IsExist(chainDTO model.ChainDTO) bool
	InsertChain(chainDTO model.ChainDTO) error
	ChainByID(id string) (*model.ChainVO, error)
	ChainByName(name string) (*model.ChainVO, error)
	ChainByAddress(ip string, port int64) (*model.ChainVO, error)
	Chains(condition model.ChainQueryCondition) ([]*model.ChainVO, error)
	Count(condition model.ChainQueryCondition) (int64, error)
	GetSysConfigString(id string, funcName string) (string, error)
	SetSysConfigString(id string, funcName string, funcParams interface{}) (string, error)
}

type IBlockService interface {
	BlockByID(id string) (*model.BlockVO, error)
	BlockByHash(chainID string, hash string) (*model.BlockVO, error)
	Blocks(condition model.BlockQueryCondition) ([]*model.BlockVO, error)
	Count(condition model.BlockQueryCondition) (int64, error)
	ChainStats(chainID string) (model.StatsVO, error)
}

type ITXService interface {
	TXByID(id string) (*model.TXVO, error)
	TXByHash(chainID string, hash string) (*model.TXVO, error)
	TXs(condition model.TXQueryCondition) ([]*model.TXVO, error)
	TXsForContractCall(condition model.TXQueryCondition) ([]*model.TXVO, error)
	Count(condition model.TXQueryCondition) (int64, error)
	TXShow(txdata *model.TX) (*model.TXVO, error)
	History(chainid string) (*[]model.TxStatsVO, error)
}

type INodeService interface {
	NodeSyncServer(node *model.NodeSyncReq) (*model.SyncNodeResult, error)
	NodeByID(id string) (*model.NodeVO, error)
	Nodes(condition model.NodeQueryCondition) ([]*model.NodeVO, error)
	Count(condition model.NodeQueryCondition) (int64, error)
}

type IContractService interface {
	// OpenFireWall 开启合约防火墙
	OpenFireWall(request model.FireWall) (string, error)
	// CloseFireWall 关闭合约防火墙
	CloseFireWall(request model.FireWall) (string, error)
	// FireWallStatus 获取防火墙状态
	FireWallStatus(request model.FireWall) (string, error)
	// Contracts 查询合约信息
	Contracts(condition model.ContractQueryCondition) ([]*model.ContractVO, error)
	// Count 统计合约信息
	Count(condition model.ContractQueryCondition) (int64, error)
	// ContractByAddress 通过合约地址查询合约信息
	ContractByAddress(chainID string, address string) (*model.ContractVO, error)
	// ParseContractCallResult 解析合约调用返回值
	ParseContractCallResult(chainID string, callResults []interface{}) ([]*model.ContractCallResult, error)
	// ShowContract 展示合约内容
	ShowContract(input string) ([]map[string]interface{}, []byte,error)
}

type ICNSService interface {
	CNSByID(id string) (*model.CNSVO, error)
	CNS(chainID string, name, address, version string) (*model.CNSVO, error)
	CNSs(condition model.CNSQueryCondition) ([]*model.CNSVO, error)
	Count(condition model.CNSQueryCondition) (int64, error)
	Register(dto model.CNSRegisterDTO) (*model.ContractCallResult, error)
	Redirect(dto model.CNSRedirectDTO) (*model.ContractCallResult, error)
}

type IAccountService interface {
	LockAccount(dto model.LockAccountDTO) (bool, error)
	UnlockAccount(dto model.UnlockAccountDTO) (bool, error)
	FirstAccount(chainID string) (string, error)
	ListAccounts(dto model.AccountDTO) ([]*model.AccountVO, error)
}
