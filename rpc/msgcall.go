package rpc

import (
	"encoding/json"
	"errors"
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/Venachain/Venachain/accounts/abi"
	"github.com/Venachain/Venachain/cmd/vcl/client/packet"
	precompile "github.com/Venachain/Venachain/cmd/vcl/client/precompiled"
	"github.com/Venachain/Venachain/cmd/vcl/client/utils"
	cmd_common "github.com/Venachain/Venachain/cmd/vcl/common"
	"github.com/Venachain/Venachain/common"
)

// NewMsgCaller 创建一个消息调用器
// rpcClient 链 rpc 连接客户端
func NewMsgCaller(rpcClient *Client) *MsgCaller {
	return &MsgCaller{
		rpcClient,
	}
}

// Call RPC 消息调用
func (caller *MsgCaller) Call(txParams *TxParams, contractParams *ContractParams) ([]interface{}, error) {
	// 解析合约 abi
	var funcAbi []byte
	if p := precompile.List[contractParams.ContractAddr]; p != "" {
		funcAbi, _ = precompile.Asset(p)
	} else {
		funcAbi = contractParams.AbiMethods
	}
	contractAbi, _ := packet.ParseAbiFromJson(funcAbi)
	methodAbi, err := contractAbi.GetFuncFromAbi(contractParams.Method)
	if err != nil {
		return nil, err
	}

	// 解析合约函数参数
	funcParams, _ := caller.getDataParams(contractParams.Data)
	funcArgs, _ := methodAbi.StringToArgs(funcParams)

	// 解析 CNS
	cns, to, err := cmd_common.CnsParse(contractParams.ContractAddr)
	if err != nil {
		return nil, err
	}

	// 生成合约数据
	data := packet.NewData(funcArgs, methodAbi)
	dataGenerator := packet.NewContractDataGen(data, contractAbi, cns.TxType)
	// 设置执行合约的虚拟机（VM）
	dataGenerator.SetInterpreter(contractParams.Interpreter, cns.Name, cns.TxType)

	// 构造交易数据
	from := common.HexToAddress(txParams.From)
	tx := packet.NewTxParams(from, &to, "", txParams.Gas, "", "")

	// 解析 keyfile 数据
	//keyfile, err := caller.parseKeyfile(txParams.From)
	//if err == nil {
	//	keyfile.Passphrase = caller.passphrase
	//
	//	err := keyfile.ParsePrivateKey()
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	keyfile := &utils.Keyfile{
		Address:    txParams.From,
		Json:       nil,
		Passphrase: "",
	}
	return caller.MessageCallV2(dataGenerator, tx, keyfile, true)
}

// DeployContract RPC 合约部署
func (caller *MsgCaller) DeployContract(txParams *TxParams, contractParams *ContractParams) ([]interface{}, error) {
	var consArgs = make([]interface{}, 0)
	var constructor *packet.FuncDesc

	vm := contractParams.Interpreter
	data, _ := caller.getDataParams(contractParams.Data)
	codeBytes := []byte(data[0])
	abiBytes := []byte(data[1])
	consParams := data[2:]

	conAbi, _ := packet.ParseAbiFromJson(abiBytes)
	if constructor = conAbi.GetConstructor(); constructor != nil {
		consArgs, _ = constructor.StringToArgs(consParams)
	}

	dataGenerator := packet.NewDeployDataGen(conAbi)
	dataGenerator.SetInterpreter(vm, abiBytes, codeBytes, consArgs, constructor)

	from := common.HexToAddress(txParams.From)
	tx := packet.NewTxParams(from, nil, "", "", "", "")
	keyfile := utils.Keyfile{}
	//keyfile, err := caller.parseKeyfile(txParams.From)
	//if err == nil {
	//	keyfile.Passphrase = "0" //todo just for test, should be real passphrase
	//
	//	err := keyfile.ParsePrivateKey()
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	return caller.MessageCallV2(dataGenerator, tx, &keyfile, true)
}

// 解析合约函数参数
func (caller *MsgCaller) getDataParams(i interface{}) ([]string, error) {
	var funcParams []string
	if i == nil {
		return nil, nil
	}

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, errors.New("data is not struct type")
	}

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)

		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() == reflect.Struct || value.Kind() == reflect.Interface {
			marshalBytes, _ := json.Marshal(value.Interface())
			funcParams = append(funcParams, string(marshalBytes))
			continue
		}
		if value.Type().Kind() != reflect.String {
			return nil, errors.New("data type not support")
		}
		temp := value.String()
		temp = strings.TrimSpace(temp)
		if temp != "" {
			if strings.Index(temp, "(") == 0 && strings.LastIndex(temp, ")") == len(temp)-1 {
				/// temp = abi.TrimSpace(temp)
				funcParams = append(funcParams, abi.GetFuncParams(temp[1:len(temp)-1])...)
			} else {
				funcParams = append(funcParams, temp)
			}
		}
	}

	return funcParams, nil
}

// 解析 keyfile 数据
func (caller *MsgCaller) parseKeyfile(from string) (*utils.Keyfile, error) {
	if strings.HasPrefix(from, "0x") {
		from = from[2:]
	}

	keyfileDirt := caller.getKeyFilePath()
	fileName, err := utils.GetFileByKey(keyfileDirt, from)
	if err != nil {
		return &utils.Keyfile{}, err
	}

	keyfilePath := keyfileDirt + "/" + fileName
	keyfile, _ := utils.NewKeyfile(keyfilePath)
	return keyfile, nil
}

func (caller *MsgCaller) getKeyFilePath() string {
	_, filename, _, ok := runtime.Caller(1)
	var cwdPath string
	if ok {
		cwdPath = path.Join(path.Dir(filename), "")
	} else {
		cwdPath = "./"
	}
	cwdPath = cwdPath + "/../../../release/linux/data/node-0/"

	keyfileDirt := cwdPath + caller.keyfilePath
	return keyfileDirt
}
