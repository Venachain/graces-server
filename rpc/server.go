package rpc

import (
	"PlatONE-Graces/model"
	"PlatONE-Graces/util"
	"PlatONE-Graces/web/dao"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PlatONEnetwork/PlatONE-Go/rpc"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/client/precompiled"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//// 	默认值设置在 core/vm/sc_param_manager.go:51
//// 为了不改动PlatONE，故在此写死
//var TxGasLimitConst = "1500000000"

const (
	DefaultContractInterpreter = "wasm"
)

// Ping ping 指定 url 看是否能 ping 通
// 支持的 Scheme 有：http、https、ws、wss、stdio、stdio、ipc
func Ping(url string) (bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	_, err := rpc.DialContext(ctx, url)
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetLatestBlockFromChain 获取链上的最新区块
func GetLatestBlockFromChain(chainID string) (*types.Block, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	return getLatestBlock(ctx, cli)
}

// GetLatestBlockFromDB 获取数据库中的最新区块
func GetLatestBlockFromDB(chainID string) (*model.Block, error) {
	id, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, err
	}
	return dao.DefaultBlockDao.LatestBlock(id)
}

// GetBlockByHash 通过 hash 从链上获取区块，并组装为数据库 model
func GetBlockByHash(chainID string, hash string) (*model.Block, error) {
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return getBlockByHash(ctx, cli, chainID, hash)
}

// GetBlockHeadByHash 通过 hash 从链上获取区块头，并组装为数据库 model
func GetBlockHeadByHash(chainID string, hash string) (*model.BLockHead, error) {
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return getBlockHeadByHash(ctx, cli, hash)
}

// GetBlockByNumber 通过 number（高度） 从链上获取区块，并组装为数据库 model
func GetBlockByNumber(chainID string, number int64) (*model.Block, error) {
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return getBlockByNumber(ctx, cli, chainID, number)
}

// GetBlockHeadByNumber 通过 number（高度） 从链上获取区块头，并组装为数据库 model
func GetBlockHeadByNumber(chainID string, number int64) (*model.BLockHead, error) {
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return getBlockHeadByNumber(ctx, cli, number)
}

// GetTXDataByBlockHash 通过 区块hash 从链上获取区块内的交易数据，并组装为数据库 model
func GetTXDataByBlockHash(chainID string, blockID string, blockHash string) ([]*model.TX, error) {
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	block, err := cli.EthClient().BlockByHash(ctx, common.HexToHash(blockHash))
	if err != nil {
		return nil, err
	}
	return getDBTXDataByBlock(chainID, blockID, block)
}

// GetTXDataByBlockNumber 通过 区块高度 从链上获取区块内的交易数据，并组装为数据库 model
func GetTXDataByBlockNumber(chainID string, blockID string, number int64) ([]*model.TX, error) {
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	block, err := cli.EthClient().BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, err
	}
	return getDBTXDataByBlock(chainID, blockID, block)
}

// GetTXReceiptByTXHash 通过 交易hash 从链上获取该交易的收据数据，并组装为数据库 model
func GetTXReceiptByTXHash(chainID string, txHash string) (*model.Receipt, error) {
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return getTXReceiptByTXHash(ctx, cli, txHash)
}

// GetAllCNS 从链上获取所有的 CNS 信息
func GetAllCNS(chainID string) ([]*model.CNS, error) {
	// 1、通过 chainID 获取其对应的链 rpc 连接客户端
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	// 2、通过客户端获取消息调用器
	caller := NewMsgCaller(cli)

	// 3、通过消息调用器执行 getRegisteredContracts 合约调用，获取所有已注册进 CNS 管理器的合约cns映射信息
	txParams := &TxParams{}
	contractParams := buildGetRegisteredContractsParams(DefaultContractInterpreter)
	res, err := caller.Call(txParams, contractParams)
	if err != nil {
		return nil, err
	}
	// 4、解析所有已注册的合约cns映射信息，获取到去重后的cns数据
	cnsMap, err2 := parseRegisteredContracts(caller, chainID, res)
	if err2 != nil {
		return nil, err2
	}
	// 5、组装cns数据进行返回
	cnsData := make([]*model.CNS, 0)
	for _, v := range cnsMap {
		cnsData = append(cnsData, v)
	}
	return cnsData, nil
}

// GetRPCClientByChainID 通过 链id 信息获取其对应的 rpc 客户端连接
func GetRPCClientByChainID(chainID string) (*Client, error) {
	chain, err := getChainByID(chainID)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return getRPCClientByChain(ctx, *chain)
}

func GetRPCClientByChainIDAndNodeID(chainID string, nodeID string) (*Client, error) {
	chain, err := getChainByID(chainID)
	node, err := getNodeByID(nodeID)
	if err != nil {
		return nil, err
	}
	return getRPCClientByChainAndNodeID(*chain, *node)
}

//todo
func getRPCClientByChainAndNodeID(chain model.Chain, node model.Node) (*Client, error) {
	host := fmt.Sprintf("%v:%v", node.ExternalIP, node.RPCPort)
	uri := url.URL{
		Scheme: "http",
		Host:   host,
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	nodeConfig, ok := chain.ChainConfig["node"].(map[string]interface{})
	if !ok {
		return nil, errors.New("chain [%v] without node config")
	}
	keyfilePath := nodeConfig["keyfile_path"].(string)
	passphrase := nodeConfig["passphrase"].(string)
	client, err := NewClient(ctx, uri.String(), passphrase, keyfilePath)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// GetAllNodes 获取指定链上的所有节点数据
func GetAllNodes(chainID string) ([]*model.Node, error) {
	// 1、通过 chainID 获取其对应的链 rpc 连接客户端
	cli, err := GetRPCClientByChainID(chainID)
	if err != nil {
		return nil, err
	}
	// 2、通过客户端获取消息调用器
	caller := NewMsgCaller(cli)
	// 3、通过消息调用器执行 getRegisteredContracts 合约调用，获取所有已注册进 CNS 管理器的合约cns映射信息
	txParams := &TxParams{}
	contractParams := buildGetAllNodesParams(DefaultContractInterpreter)
	res, err := caller.Call(txParams, contractParams)

	logrus.Debugf("res:%+v", res)

	nodes, err := parseNodes(chainID, res)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// 通过 id 获取链信息
func getChainByID(chainID string) (*model.Chain, error) {
	id, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{
		"_id": id,
	}
	chain, err := dao.DefaultChainDao.Chain(filter)
	if err != nil {
		return nil, err
	}
	return chain, nil
}

// 通过 id 获取节点信息
func getNodeByID(nodeID string) (*model.Node, error) {
	id, err := primitive.ObjectIDFromHex(nodeID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{
		"_id": id,
	}
	node, err := dao.DefaultNodeDao.Node(filter)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// 通过链配置信息获取其对应的 rpc 客户端连接
func getRPCClientByChain(ctx context.Context, chain model.Chain) (*Client, error) {
	host := fmt.Sprintf("%v:%v", chain.IP, chain.RPCPort)
	uri := url.URL{
		Scheme: "http",
		Host:   host,
	}
	nodeConfig, ok := chain.ChainConfig["node"].(map[string]interface{})
	if !ok {
		return nil, errors.New("chain [%v] without node config")
	}
	keyfilePath := nodeConfig["keyfile_path"].(string)
	passphrase := nodeConfig["passphrase"].(string)
	client, err := NewClient(ctx, uri.String(), passphrase, keyfilePath)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// GetBlockNumber 获取指定节点的最新区块的高度
func GetBlockNumber(endpoint string) (uint64, error) {
	var err error
	height := model.GetRpcResult(endpoint, "eth_blockNumber", []string{})
	if height == nil {
		return 0, err
	}
	if blockHeight, ok := height.(string); ok {
		blockNumber, err := util.Hex2Number(blockHeight)
		if err != nil {
			return 0, err
		}
		return blockNumber, nil
	}
	return 0, errors.New("height not be a string")
}

func getBlockByNumber(ctx context.Context, client *Client, chainID string, number int64) (*model.Block, error) {
	block, err := client.EthClient().BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, err
	}
	dbBlock := buildDBBlock(*block)
	head, err := getBlockHeadByNumber(ctx, client, number)
	if err != nil {
		return nil, err
	}
	dbBlock.Head = head

	objectID, _ := primitive.ObjectIDFromHex(chainID)
	dbBlock.ChainID = objectID
	dbBlock.ID = primitive.NewObjectID()
	return dbBlock, nil
}

func getBlockHeadByNumber(ctx context.Context, client *Client, number int64) (*model.BLockHead, error) {
	head, err := client.EthClient().HeaderByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, err
	}
	return buildDBBlockHead(head), nil
}

func getLatestBlock(ctx context.Context, client *Client) (*types.Block, error) {
	return client.EthClient().BlockByNumber(ctx, nil)
}

func getBlockHeadByHash(ctx context.Context, client *Client, hash string) (*model.BLockHead, error) {
	head, err := client.EthClient().HeaderByHash(ctx, common.HexToHash(hash))
	if err != nil {
		return nil, err
	}
	return buildDBBlockHead(head), nil
}

// 通过 *types.Block 获取 []model.TX 交易数据
func getDBTXDataByBlock(chainID string, blockID string, block *types.Block) ([]*model.TX, error) {
	if block == nil {
		return nil, errors.New("block is nil")
	}
	txs := make([]*model.TX, block.Transactions().Len())
	for i, tx := range block.Transactions() {

		dbTX, err := buildDBTX(tx)
		if err != nil {
			logrus.Errorln(err)
			continue
		}
		dbTX.Height = block.NumberU64()
		dbTX.Timestamp = block.Time().Int64()
		receipt, err := GetTXReceiptByTXHash(chainID, tx.Hash().Hex())
		if nil != err {
			logrus.Errorln("fail to get transaction receipt.err:", err)
			return nil, err
		}
		dbTX.Receipt = receipt
		dbTX.ID = primitive.NewObjectID()
		dbTX.BlockID, _ = primitive.ObjectIDFromHex(blockID)
		dbTX.ChainID, _ = primitive.ObjectIDFromHex(chainID)

		// 转换gaslimit
		intNum, _ := strconv.Atoi(model.TxGasLimitConst)
		gaslimit := uint64(intNum)
		dbTX.GasLimit = gaslimit

		txs[i] = dbTX
	}
	return txs, nil
}

func getBlockByHash(ctx context.Context, client *Client, chainID string, hash string) (*model.Block, error) {
	block, err := client.EthClient().BlockByHash(ctx, common.HexToHash(hash))
	if err != nil {
		return nil, err
	}
	dbBlock := buildDBBlock(*block)
	head, err := getBlockHeadByHash(ctx, client, hash)
	if err != nil {
		return nil, err
	}
	dbBlock.Head = head

	objectID, _ := primitive.ObjectIDFromHex(chainID)
	dbBlock.ChainID = objectID
	dbBlock.ID = primitive.NewObjectID()
	return dbBlock, nil
}

func getTXReceiptByTXHash(ctx context.Context, client *Client, txHash string) (*model.Receipt, error) {
	receipt, err := client.EthClient().TransactionReceipt(ctx, common.HexToHash(txHash))
	if nil != err {
		logrus.Errorln("fail to get transaction receipt.err:", err)
		return nil, err
	}
	return buildDBReceipt(*receipt)
}

func buildDBReceipt(receipt types.Receipt) (*model.Receipt, error) {
	dbReceipt := &model.Receipt{}
	dbReceipt.ContractAddress = receipt.ContractAddress.Hex()
	dbReceipt.Status = receipt.Status
	eventBytes, err := json.Marshal(receipt.Logs)
	if err != nil {
		logrus.Errorln("fail to Marshal receipt.Logs:", err)
		return nil, err
	}
	// todo parse event
	dbReceipt.Event = string(eventBytes)
	dbReceipt.GasUsed = receipt.GasUsed
	return dbReceipt, nil
}

func buildDBBlockHead(head *types.Header) *model.BLockHead {
	if head == nil {
		logrus.Warningf("head is nil")
		return nil
	}
	var blockHead model.BLockHead
	blockHead.ParentHash = head.ParentHash.Hex()
	blockHead.Miner = head.Coinbase.Hex()
	blockHead.StateRoot = head.Root.Hex()
	blockHead.TransactionsRoot = head.TxHash.Hex()
	blockHead.ReceiptsRoot = head.ReceiptHash.Hex()
	blockHead.LogsBloom = hex.EncodeToString(head.Bloom.Bytes())
	blockHead.Height = head.Number.Uint64()
	blockHead.GasLimit = head.GasLimit
	blockHead.GasUsed = head.GasUsed
	blockHead.Timestamp = int64(head.Time.Uint64())
	blockHead.ExtraData = hex.EncodeToString(head.Extra)
	blockHead.MixHash = head.MixDigest.Hex()
	blockHead.Nonce = head.Nonce.Uint64()
	blockHead.Hash = head.Hash().Hex()
	return &blockHead
}

func sender(tx *types.Transaction) (common.Address, error) {
	//first try Frontier
	signer := types.FrontierSigner{}
	addr, err := signer.Sender(tx)
	if nil == err {
		return addr, nil
	}

	addr, err = types.NewEIP155Signer(tx.ChainId()).Sender(tx)
	if nil == err {
		return addr, nil
	}
	return types.HomesteadSigner{}.Sender(tx)
}

func buildDBBlock(block types.Block) *model.Block {
	var dbBlock model.Block
	dbBlock.Height = block.NumberU64()
	dbBlock.ExtraData = hex.EncodeToString(block.Extra())
	dbBlock.GasLimit = block.GasLimit()
	dbBlock.GasUsed = block.GasUsed()
	dbBlock.Hash = block.Hash().Hex()
	dbBlock.ParentHash = block.ParentHash().Hex()
	dbBlock.Proposer = block.Coinbase().Hex()
	dbBlock.Timestamp = block.Time().Int64()
	dbBlock.TxAmount = uint64(block.Transactions().Len())
	dbBlock.Size = block.Size().String()
	return &dbBlock
}

func buildDBTX(tx *types.Transaction) (*model.TX, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	var dbTX model.TX
	dbTX.Hash = tx.Hash().Hex()
	from, err := sender(tx)
	if err != nil {
		logrus.Errorln("fail to get sender of tx.err:", err)
		return nil, err
	}
	dbTX.From = from.Hex()
	if tx.To() != nil {
		dbTX.To = tx.To().Hex()
	}
	dbTX.GasLimit = tx.Gas()
	dbTX.GasPrice = tx.GasPrice().Uint64()
	dbTX.Nonce = fmt.Sprintf("%d", tx.Nonce())
	dbTX.Input = hex.EncodeToString(tx.Data())
	dbTX.Value = tx.Value().Uint64()

	return &dbTX, nil
}

func buildGetRegisteredContractsParams(interpreter string) *ContractParams {
	contractAddr := precompile.CnsManagementAddress
	funcName := "getRegisteredContracts"
	if interpreter == "" {
		interpreter = DefaultContractInterpreter
	}

	// 构造 getRegisteredContracts() 函数的查询条件参数
	funcParams := &struct {
		Name    string
		Address string
		Origin  string
		Range   string
	}{}
	funcParams.Range = "(0,0)"

	// 构造合约参数
	contract := &ContractParams{
		ContractAddr: contractAddr,
		Method:       funcName,
		Interpreter:  interpreter,
		AbiMethods:   nil,
		Data:         funcParams,
	}
	return contract
}

// 解析已注册进CNS的合约信息
func parseRegisteredContracts(caller *MsgCaller, chainID string, msgCallResults []interface{}) (map[string]*model.CNS, error) {
	// 1、数据校验
	if caller == nil {
		return nil, errors.New("caller must not be nil")
	}
	if msgCallResults == nil {
		logrus.Infof("no msgCallResults need to parse")
		return nil, nil
	}
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, err
	}

	// 2、遍历所有合约cns映射信息一一处理
	// cnsMap 用于去重
	cnsMap := make(map[string]*model.CNS, 0)
	for _, res := range msgCallResults {
		// 2.1、解析返回 json 字符串
		jsonContract, err := util.JsonStringPatch(res)
		if err != nil {
			return nil, err
		}
		contractMap, ok := jsonContract.(map[string]interface{})
		if !ok {
			errMsg := fmt.Sprintf("jsonContract is not a map struct: %v", contractMap)
			logrus.Errorln(errMsg)
			return nil, errors.New(errMsg)
		}
		// 2.2、拿到 json 中的 data 数据，这里面的数据才是合约cns映射信息
		contractData, ok := contractMap["data"].([]interface{})
		if !ok {
			errMsg := fmt.Sprintf("contractMap.data is not a slice struct: %v", contractMap)
			logrus.Errorln(errMsg)
			return nil, errors.New(errMsg)
		}
		// 2.3、处理每一条合约cns映射信息
		for _, contractInfo := range contractData {
			contract, ok := contractInfo.(map[string]interface{})
			if !ok {
				errMsg := fmt.Sprintf("this data not a map struct: %v", contract)
				logrus.Errorln(errMsg)
				return nil, errors.New(errMsg)
			}

			// 2.3.1、提取映射信息
			name := contract["name"].(string)
			version := contract["version"].(string)
			address := contract["address"].(string)
			// 2.3.2、校验参数
			if !cmd_common.ParamValidWrap(name, "name") || !cmd_common.ParamValidWrap(version, "version") {
				errMsg := "[name] or [version] param valid fail"
				logrus.Errorln(errMsg)
				return nil, errors.New(errMsg)
			}

			// 2.3.3、使用消息调用器执行 getContractAddress 合约调用，通过 name 和 version 查询对应的合约地址
			txParams := &TxParams{}
			contractParams := buildGetContractAddressParams(DefaultContractInterpreter, name, version)
			contractAddresses, err := caller.Call(txParams, contractParams)
			if err != nil {
				return nil, err
			}
			// 2.3.4、解析合约地址，去除尾部的 0 ASCII码
			addr, ok := contractAddresses[0].(string)
			addr = strings.TrimFunc(addr, func(r rune) bool {
				// 对应的字符是：' '
				return r == 32 || r == 0
			})
			// 2.3.5、如果地址不同，则跳过
			if strings.ToLower(address) != strings.ToLower(addr) {
				continue
			}
			// 2.3.6、cns信息的初始化和去重
			cns, ok := cnsMap[name]
			if !ok {
				cns = &model.CNS{
					ID:      primitive.NewObjectID(),
					ChainID: cid,
				}
				cnsMap[name] = cns
			}
			cns.Address = address
			cns.Name = name
			cns.Version = version
		}
	}
	return cnsMap, nil
}

func buildGetContractAddressParams(interpreter string, name string, version string) *ContractParams {
	mappingContractAddr := precompile.CnsManagementAddress
	funcName := "getContractAddress"
	if interpreter == "" {
		interpreter = DefaultContractInterpreter
	}

	funcParams := &struct {
		Name    string
		Version string
	}{name, version}
	// 构造合约参数
	contract := &ContractParams{
		ContractAddr: mappingContractAddr,
		Method:       funcName,
		Interpreter:  interpreter,
		AbiMethods:   nil,
		Data:         funcParams,
	}
	return contract
}

func buildGetAllNodesParams(interpreter string) *ContractParams {
	contractAddr := precompile.NodeManagementAddress
	funcName := "getAllNodes"
	if interpreter == "" {
		interpreter = DefaultContractInterpreter
	}

	// 构造 getAllNodes() 函数的查询条件参数
	var funcParams interface{}

	// 构造合约参数
	contract := &ContractParams{
		ContractAddr: contractAddr,
		Method:       funcName,
		Interpreter:  interpreter,
		AbiMethods:   nil,
		Data:         funcParams,
	}
	return contract
}

// 解析节点数据
func parseNodes(chainID string, msgCallResults []interface{}) ([]*model.Node, error) {
	if msgCallResults == nil {
		logrus.Infof("no msgCallResults need to parse")
		return nil, nil
	}
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, err
	}
	nodes := make([]*model.Node, 0)
	for _, res := range msgCallResults {
		jsonNodes, err := util.JsonStringPatch(res)
		if err != nil {
			return nil, err
		}
		nodesMap, ok := jsonNodes.(map[string]interface{})
		if !ok {
			return nil, errors.New("jsonNodes isn't a map[string]interface{}")
		}
		data, ok := nodesMap["data"]
		if !ok {
			return nil, errors.New("msgCallResult haven't data property")
		}
		nodesData, ok := data.([]interface{})
		if !ok {
			return nil, errors.New("msgCallResult isn't a slice")
		}
		for _, nodeData := range nodesData {
			nodeMap, ok := nodeData.(map[string]interface{})
			if !ok {
				return nil, errors.New("msgCallResult isn't a map slice")
			}
			var node model.Node
			node.ID = primitive.NewObjectID()
			node.ChainID = cid
			node.Name = nodeMap["name"].(string)
			node.PublicKey = nodeMap["publicKey"].(string)
			node.Desc = nodeMap["desc"].(string)
			node.InternalIP = nodeMap["internalIP"].(string)
			node.ExternalIP = nodeMap["externalIP"].(string)
			node.RPCPort = int(nodeMap["rpcPort"].(float64))
			node.P2PPort = int(nodeMap["p2pPort"].(float64))
			node.Type = int(nodeMap["type"].(float64))
			node.Status = int(nodeMap["status"].(float64))
			owner := nodeMap["owner"].(string)
			if owner == "0x" {
				node.Owner = "no owner"
			} else {
				node.Owner = owner
			}

			nodes = append(nodes, &node)
		}
	}
	return nodes, err
}
