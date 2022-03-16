package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Venachain/Venachain/cmd/vcl/client/packet"
	"github.com/Venachain/Venachain/cmd/vcl/client/utils"
	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/rpc"
	"github.com/Venachain/Venachain/venaclient"

	"github.com/sirupsen/logrus"
)

var (
	cliContainer map[string]*Client
	lock         sync.Mutex
)

func init() {
	cliContainer = make(map[string]*Client, 0)
}

// NewClient 创建一个可以操作链的 RPC 客户端
// url 链连接地址
// passphrase 链账户密码
// keyfilePath keyfile 存放的相对地址，默认为 "./keystore"
func NewClient(ctx context.Context, url string, passphrase string, keyfilePath string) (*Client, error) {
	lock.Lock()
	defer lock.Unlock()
	if cli, ok := cliContainer[url]; ok {
		return cli, nil
	}
	venaCli, err := venaclient.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	rpcCli, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	client := &Client{
		venaClient:  venaCli,
		rpcClient:   rpcCli,
		passphrase:  passphrase,
		keyfilePath: keyfilePath,
	}
	cliContainer[url] = client
	return client, nil
}

func (client *Client) EthClient() *venaclient.Client {
	return client.venaClient
}

func (client *Client) RpcClient() *rpc.Client {
	return client.rpcClient
}

// ========================= Msg Call ==============================

func (client *Client) MessageCallV2(dataGen packet.MsgDataGen, tx *packet.TxParams, keyfile *utils.Keyfile, isSync bool) ([]interface{}, error) {
	var result = make([]interface{}, 1)
	var err error

	// combine the data based on the types of the calls (contract call, inner call or deploy call)
	tx.Data, err = dataGen.CombineData()
	if err != nil {
		errStr := fmt.Sprintf(utils.ErrPackDataFormat, err.Error())
		return nil, errors.New(errStr)
	}

	if dataGen.GetIsWrite() {
		res, err := client.Send(tx, keyfile)
		if err != nil {
			return nil, err
		}
		result[0] = res

		if isSync {
			polRes, err := client.GetReceiptByPolling(res)
			if err != nil {
				return result, nil
			}

			receiptBytes, _ := json.MarshalIndent(polRes, "", "\t")
			logrus.Debug(string(receiptBytes))

			recpt := dataGen.ReceiptParsing(polRes)
			// recpt := polRes.Parsing()
			if recpt.Status != packet.TxReceiptSuccessMsg {
				result, _ := client.GetRevertMsg(tx, recpt.BlockNumber)
				if len(result) >= 4 {
					recpt.Err, _ = packet.UnpackError(result)
				}
			}

			result[0] = recpt.String()
		}
	} else {
		result, _ = client.Call(dataGen.GetContractDataDen(), tx)
	}

	return result, nil
}

func (client *Client) Send(tx *packet.TxParams, keyfile *utils.Keyfile) (string, error) {
	params, action, err := tx.SendModeV2(keyfile)
	if err != nil {
		return "", err
	}

	// send the RPC calls
	var resp string
	err = client.rpcClient.Call(&resp, action, params...)
	if err != nil {
		errStr := fmt.Sprintf(utils.ErrSendTransacionFormat, err.Error())
		return "", errors.New(errStr)
	}

	return resp, nil
}

func (client *Client) Call(dataGen *packet.ContractDataGen, tx *packet.TxParams) ([]interface{}, error) {
	var params = make([]interface{}, 0)

	params = append(params, tx)
	params = append(params, "latest")
	action := "eth_call"

	// send the RPC calls
	var resp string
	err := client.rpcClient.Call(&resp, action, params...)
	if err != nil {
		errStr := fmt.Sprintf(utils.ErrSendTransacionFormat, err.Error())
		return nil, errors.New(errStr)
	}

	outputType := dataGen.GetMethodAbi().Outputs
	return dataGen.ParseNonConstantResponse(resp, outputType), nil
}

// ============================ Tx Receipt ===================================

func (client *Client) GetTransactionReceipt(txHash string) (*packet.Receipt, error) {

	var response interface{}
	_ = client.rpcClient.Call(&response, "eth_getTransactionReceipt", txHash)
	if response == nil {
		return nil, nil
	}

	// parse the rpc response
	receipt, err := packet.ParseTxReceipt(response)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (client *Client) GetReceiptByPolling(txHash string) (*packet.Receipt, error) {
	ch := make(chan interface{}, 1)
	go client.getReceiptByPolling(txHash, ch)

	select {
	case receipt := <-ch:
		return receipt.(*packet.Receipt), nil

	case <-time.After(time.Second * 10):
		// temp := fmt.Sprintf("\nget contract receipt timeout...more than %d second.\n", 10)
		// return temp + txHash

		errStr := fmt.Sprintf("get contract receipt timeout...more than %d second.", 10)
		return nil, errors.New(errStr)
	}
}

// todo: end goroutine?
func (client *Client) getReceiptByPolling(txHash string, ch chan interface{}) {

	for {
		receipt, err := client.GetTransactionReceipt(txHash)

		// limit the times of the polling
		if err != nil {
			logrus.Debug(err.Error())
			logrus.Debug("try again 5s later...")
			time.Sleep(5 * time.Second)
			logrus.Debug("try again...\n")
			continue
		}

		if receipt == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		ch <- receipt
	}
}

// ========================== Sol require/ =============================

func (client *Client) GetRevertMsg(msg *packet.TxParams, blockNum uint64) ([]byte, error) {

	var hex = new(hexutil.Bytes)
	err := client.rpcClient.Call(hex, "eth_call", msg, hexutil.EncodeUint64(blockNum))
	if err != nil {
		return nil, err
	}

	return *hex, nil
}

func NewContractParams(defaultAddr, defaultMethod, defaultInter string, abiBytes []byte, dataParams interface{}) *ContractParams {
	return &ContractParams{
		ContractAddr: defaultAddr,
		Method:       defaultMethod,
		Interpreter:  defaultInter,
		AbiMethods:   abiBytes,
		Data:         dataParams,
	}
}
