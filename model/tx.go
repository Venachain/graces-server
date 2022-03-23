package model

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"graces/exterr"
	"graces/util"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 	默认值设置在 core/vm/sc_param_manager.go:51
// 为了不改动 Venachain，故在此写死
var TxGasLimitConst = "1500000000"

type TX struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	ChainID   primitive.ObjectID `json:"chain_id" bson:"chain_id"`
	BlockID   primitive.ObjectID `json:"block_id" bson:"block_id"`
	Hash      string             `json:"hash" bson:"hash"`
	Height    uint64             `json:"height" bson:"height"`
	Timestamp int64              `json:"timestamp" bson:"timestamp"`
	From      string             `json:"from" bson:"from"`
	To        string             `json:"to" bson:"to"`
	GasLimit  uint64             `json:"gas_limit" bson:"gas_limit"`
	GasPrice  uint64             `json:"gas_price" bson:"gas_price"`
	Nonce     string             `json:"nonce" bson:"nonce"`
	Input     string             `json:"input" bson:"input"`
	Value     uint64             `json:"value" bson:"value"`
	Receipt   *Receipt           `json:"receipt" bson:"receipt"`
}

type Receipt struct {
	ContractAddress string `json:"contract_address" bson:"contract_address"`
	Status          uint64 `json:"status" bson:"status"`
	Event           string `json:"event" bson:"event"`
	GasUsed         uint64 `json:"gas_used" bson:"gas_used"`
}

type TXQueryCondition struct {
	PageDTO
	SortDTO
	// 主键ID
	ID string `json:"id" binding:"min=0,max=50"`
	// 所属链ID
	ChainID string `json:"chain_id" binding:"min=0,max=50"`
	// 所属区块ID
	BlockID string `json:"block_id" binding:"min=0,max=50"`
	// 交易哈希
	Hash string `json:"hash" binding:"min=0,max=70"`
	// 所属区块高度
	Height uint64 `json:"height" binding:"min=0"`
	// 起始时间
	TimeStart int64 `json:"time_start"`
	// 终止时间
	TimeEnd int64 `json:"time_end"`
	// 交易参与人哈希
	ParticipantHash string `json:"participant_hash" binding:"min=0,max=70"`
	// 交易状态
	Status uint64 `json:"status" binding:"min=0,max=1"`
	// 合约地址
	ContractAddress string `json:"contract_address" binding:"min=0,max=70"`
}

type TXByHashDTO struct {
	// 所属链ID
	ChainID string `json:"chain_id" binding:"min=0,max=50"`
	// 交易哈希
	Hash string `json:"hash" binding:"required, min=1,max=70"`
}

type Txdata struct {
	ChainID string `json:"chainID"`
	Txhash  string `json:"txhash"`
	To      string `json:"to"`
	Input   string `json:"input"`
	Value   int    `json:"value"`
}

type TXVO struct {
	// 主键ID
	ID string `json:"id"`
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 所属区块ID
	BlockID string `json:"block_id"`
	// 交易哈希
	Hash string `json:"hash"`
	// 所属区块高度
	Height uint64 `json:"height"`
	// 交易发起时间
	Timestamp string `json:"timestamp"`
	// 交易发起人地址
	From string `json:"from"`
	// 交易目标地址
	To string `json:"to"`
	// Gas 限制
	GasLimit uint64 `json:"gas_limit"`
	// Gas 价格
	GasPrice uint64 `json:"gas_price"`
	// 随机数
	Nonce string `json:"nonce"`
	// input 数据
	Input string `json:"input"`
	// 交易数额
	Value uint64 `json:"value"`
	// 收据信息
	Receipt *ReceiptVO `json:"receipt"`
	// 是否是合约调用
	IsToContract bool `json:"is_to_contract"`
	// 交易执行的操作： 合约部署、合约调用或者转账
	Action string `json:"action"`
	// 交易的详细信息
	Detail *TxDetail `json:"detail"`
	// 合约的wasm
	Wasm []byte `json:"wasm"`
}

type TxDetail struct {
	// 显示合约相关的内容
	Contract string `json:"contract"`
	// 交易类型
	Txtype int64 `json:"txtype"`
	// 如果是合约调用，显示调用的合约方法
	Method string `json:"method"`
	// 如果是合约调用，显示调用的合约参数
	Params []interface{} `json:"params"`
	// 显示其它的信息
	Extra interface{} `json:"extra"`
}

type ReceiptVO struct {
	// 合约地址
	ContractAddress string `json:"contract_address"`
	// 交易状态
	Status uint64 `json:"status"`
	// 交易状态名称
	StatusName string `json:"status_name"`
	// 事件信息
	Event string `json:"event"`
	// Gas 使用量
	GasUsed uint64 `json:"gas_used"`
}

func (tx *TX) ToVO() (*TXVO, error) {
	var vo TXVO
	if err := util.SimpleCopyProperties(&vo, tx); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	vo.ID = tx.ID.Hex()
	vo.ChainID = tx.ChainID.Hex()
	vo.BlockID = tx.BlockID.Hex()
	vo.Timestamp = util.Timestamp2TimeStr(tx.Timestamp / 1e3)
	receiptVO, err := tx.Receipt.ToVO()
	if err != nil {
		return nil, err
	}
	vo.Receipt = receiptVO
	vo.IsToContract = vo.Receipt.ContractAddress != "" && vo.To != ""
	return &vo, nil
}

func (tx *TX) ToContract() *Contract {
	contract := &Contract{
		ID:        primitive.NewObjectID(),
		ChainID:   tx.ChainID,
		Address:   tx.Receipt.ContractAddress,
		Creator:   tx.From,
		TxHash:    tx.Hash,
		Content:   tx.Input,
		Timestamp: tx.Timestamp,
	}
	return contract
}

func GetRpcResult(endpoint string, method string, params []string) interface{} {
	client := &http.Client{}
	request := NewJsonRpcStruct(method, params)
	jsonstr, _ := json.Marshal(request)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonstr))
	if err != nil {
		logrus.Errorln("request error")
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		logrus.Warnf("node is disable %s", endpoint)
		return nil
	}
	defer response.Body.Close()

	result, _ := ioutil.ReadAll(response.Body)
	res := new(Response)
	json.Unmarshal(result, res)
	return res.Result
}

func NewJsonRpcStruct(method string, params []string) *JsonRpcStruct {
	return &JsonRpcStruct{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      1,
	}
}

func (receipt *Receipt) ToVO() (*ReceiptVO, error) {
	var vo ReceiptVO
	if err := util.SimpleCopyProperties(&vo, receipt); err != nil {
		logrus.Errorln(err)
		return nil, exterr.ErrConvert
	}
	// todo parse status
	vo.StatusName = ""
	return &vo, nil
}
