package model

// Result 响应结果结构体
type Result struct {
	// HTTP 响应状态码
	Code int `json:"code"`
	// 响应提示信息
	Msg string `json:"msg"`
	// 响应数据
	Data interface{} `json:"data"`
	// 请求资源路径
	Request string `json:"request"`
}

type JsonRpcStruct struct {
	Jsonrpc string   `json:"jsonrpc" bson:"jsonrpc"`
	Method  string   `json:"method" bson:"method"`
	Params  []string `json:"params" bson:"params"`
	Id      int      `json:"id" bson:"id"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc" bson:"jsonrpc"`
	Id      int         `json:"id" bson:"id"`
	Result  interface{} `json:"Result" bson:"id"`
}

type NodeSyncReq struct {
	// 节点 IP 地址
	Ip string `json:"ip" bson:"ip"`
	// 节点端口号
	Port uint64 `json:"port" bson:"port"`
}

type DeployContract struct {
	Interpreter string `json:"interpreter" bson:"interpreter"`
}

type DeployRequest struct {
	CodePath    string `json:"codepath"`
	AbiPath     string `json:"abipath"`
	Interpreter string `json:"interpreter"`
	Name        string `json:"name"`
}

type DeployInfo struct {
	CodeBytes string
	AbiBytes  string
	Params    string `json:"params"`
}

type DeployRequestBody struct {
	Tx       TxStruct       `json:"tx" bson:"tx"`
	Rpc      RpcStruct      `json:"rpc" bson:"rpc"`
	Contract DeployContract `json:"contract" bson:"contract"`
}

type TxStruct struct {
	From string `json:"from" bson:"from"`
}

type RpcStruct struct {
	Endpoint   string `json:"endPoint" bson:"endPoint"`
	Passphrase string `json:"passphrase,omitempty" bson:"passphrase"`
}
