package rpc

import (
	"github.com/Venachain/Venachain/ethclient"
	"github.com/Venachain/Venachain/rpc"
)

// Client 链 RPC 连接客户端
type Client struct {
	ethClient   *ethclient.Client
	rpcClient   *rpc.Client
	passphrase  string
	keyfilePath string
}

// MsgCaller 链 RPC 消息调用器
type MsgCaller struct {
	*Client
}

// TxParams 交易相关参数
type TxParams struct {
	From     string `json:"from"`         // the address used to send the transaction
	To       string `json:"to,omitempty"` // the address receives the transactions
	Gas      string `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Value    string `json:"value,omitempty"`
	Data     string `json:"data,omitempty"`
}

// ContractParams 合约相关参数
type ContractParams struct {
	ContractAddr string `json:"contract_addr"`
	Method       string `json:"method,omitempty"`
	Interpreter  string `json:"interpreter,omitempty"`
	AbiMethods   []byte `json:"-"`
	/// Data         contractData `json:"data"`
	Data interface{} `json:"data"`
}
