package model

// AccountDTO 账户数据传输DTO
type AccountDTO struct {
	// 链ID
	ChainID string `json:"chain_id" binding:"min=0,max=50"`
	// 节点ID
	NodeID string `json:"node_id" binding:"min=0,max=50"`
}

// LockAccountDTO 锁定账户DTO
type LockAccountDTO struct {
	AccountDTO
	// 账户
	Account string `json:"account" binding:"required,min=0,max=70"`
}

// UnlockAccountDTO 解锁账户DTO
type UnlockAccountDTO struct {
	LockAccountDTO
	// 账户密码
	Password string `json:"password" binding:"required,min=0,max=50"`
	// 解锁持续时间，单位：秒。如果为 0，则该值被默认设置为 300
	Duration uint64 `json:"duration" binding:"min=0"`
}

type AccountVO struct {
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 所属节点ID
	NodeID string `json:"node_id"`
	// 账户地址
	Account string `json:"account"`
}
