package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"graces/exterr"
	"graces/model"
	"graces/rpc"
	"graces/util"
	"graces/web/dao"

	precompile "github.com/Venachain/Venachain/cmd/vcl/client/precompiled"
	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/rlp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var DefaultContractService IContractService

func init() {
	DefaultContractService = newContractService()
}

func newContractService() IContractService {
	return &contractService{
		dao: dao.DefaultContractDao,
	}
}

type contractService struct {
	dao dao.IContractDao
}

func (s *contractService) OpenFireWall(fireWallParam model.FireWall) (string, error) {
	var result string
	var err error
	var contractAddr = precompile.FirewallManagementAddress
	defaultInter := "wasm"
	chainid, _ := rpc.GetRPCClientByChainID(fireWallParam.Chainid)
	caller := rpc.NewMsgCaller(chainid)
	funcName := "__sys_FwOpen"
	funcParams := &struct {
		address string
	}{address: fireWallParam.ContractAddress}

	data := rpc.NewContractParams(contractAddr, funcName, defaultInter, nil, funcParams)
	txParams := &rpc.TxParams{}
	txParams.From, err = DefaultAccountService.FirstAccount(fireWallParam.Chainid)
	if err != nil {
		logrus.Errorln("get first account is error")
		return "", err
	}

	//txParams.From = GetListAccount(fireWallParam.Chainid)
	res, err := caller.Call(txParams, data)
	if err != nil {
		logrus.Errorln("call open firewall error")
		return "", exterr.NewError(exterr.ErrCodeUpdate, err.Error())
	}
	result = fmt.Sprintf("%v", res[0])
	return result, nil
}

func (s *contractService) CloseFireWall(fireWallParam model.FireWall) (string, error) {
	var result string
	var err error
	var contractAddr = precompile.FirewallManagementAddress
	defaultInter := "wasm"
	chainid, _ := rpc.GetRPCClientByChainID(fireWallParam.Chainid)
	caller := rpc.NewMsgCaller(chainid)
	funcName := "__sys_FwClose"
	funcParams := &struct {
		address string
	}{address: fireWallParam.ContractAddress}

	data := rpc.NewContractParams(contractAddr, funcName, defaultInter, nil, funcParams)
	txParams := &rpc.TxParams{}
	txParams.From, err = DefaultAccountService.FirstAccount(fireWallParam.Chainid)
	if err != nil {
		logrus.Errorln("get first account is error")
		return "", err
	}

	//txParams.From = GetListAccount(fireWallParam.Chainid)
	res, err := caller.Call(txParams, data)
	if err != nil {
		logrus.Errorln("call close firewall error")
		return "", exterr.NewError(exterr.ErrCodeUpdate, err.Error())
	}
	result = fmt.Sprintf("%v", res[0])
	return result, nil
}

func (s *contractService) FireWallStatus(request model.FireWall) (string, error) {
	var result string
	var err error
	var contractAddr = precompile.FirewallManagementAddress
	defaultInter := "wasm"
	chainid, _ := rpc.GetRPCClientByChainID(request.Chainid)
	caller := rpc.NewMsgCaller(chainid)
	funcName := "__sys_FwStatus"
	funcParams := &struct {
		address string
	}{address: request.ContractAddress}

	data := rpc.NewContractParams(contractAddr, funcName, defaultInter, nil, funcParams)
	txParams := &rpc.TxParams{}
	txParams.From, err = DefaultAccountService.FirstAccount(request.Chainid)
	if err != nil {
		logrus.Errorln("get first account is error")
		return "", err
	}

	//txParams.From = GetListAccount(fireWallParam.Chainid)
	res, err := caller.Call(txParams, data)
	if err != nil {
		logrus.Errorln("call close firewall error")
		return "", exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	result = fmt.Sprintf("%v", res[0])
	return result, nil
}

func (s *contractService) Contracts(condition model.ContractQueryCondition) ([]*model.ContractVO, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return nil, err
	}
	findOps := util.BuildOptionsByQuery(condition.PageIndex, condition.PageSize)
	if !reflect.ValueOf(condition.Sort).IsZero() {
		sort := bson.D{}
		for k, v := range condition.Sort {
			if k == "id" {
				k = "_id"
			}
			sort = append(sort, bson.E{Key: k, Value: v})
		}
		findOps.Sort = sort
	} else {
		sort := bson.D{{"timestamp", -1}}
		findOps.Sort = sort
	}
	contracts, err := s.dao.Contracts(filter, findOps)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	vos := make([]*model.ContractVO, 0)
	for _, contract := range contracts {
		vo, err := contract.ToVO()
		if err != nil {
			return nil, err
		}

		c := model.CNSQueryCondition{
			ChainID: contract.ChainID.Hex(),
			Address: contract.Address,
		}
		// 按 CNS 名称过滤
		if condition.CNSName != "" {
			c.Name = condition.CNSName
			cnss, err := DefaultCNSService.CNSs(c)
			if err == nil && len(cnss) > 0 {
				vo.CNS = cnss
				vos = append(vos, vo)
			}
			continue
		}
		cnss, err := DefaultCNSService.CNSs(c)
		if err == nil && len(cnss) > 0 {
			vo.CNS = cnss
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *contractService) Count(condition model.ContractQueryCondition) (int64, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return 0, err
	}
	contracts, err := s.dao.Contracts(filter, nil)
	if err != nil {
		return 0, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	var cnt int64
	for _, contract := range contracts {
		// 按 CNS 名称过滤
		if condition.CNSName != "" {
			c := model.CNSQueryCondition{
				ChainID: contract.ChainID.Hex(),
				Address: strings.ToLower(contract.Address),
				Name:    condition.CNSName,
			}
			cnss, err := DefaultCNSService.CNSs(c)
			if err == nil && len(cnss) > 0 {
				cnt++
			}
			continue
		}
		cnt++
	}
	return cnt, nil
}

func (s *contractService) ContractByAddress(chainID string, address string) (*model.ContractVO, error) {
	objectId, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{}
	filter["chain_id"] = objectId
	filter["address"] = bson.M{"$regex": fmt.Sprintf("^(?i)%s$", address)}

	contract, err := s.dao.Contract(filter)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	c := model.CNSQueryCondition{
		ChainID: contract.ChainID.Hex(),
		Address: strings.ToLower(contract.Address),
	}
	vo, err := contract.ToVO()
	if err != nil {
		return nil, err
	}
	content, _, err := s.ShowContract(contract.Content)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	vo.Content = content
	cnss, err := DefaultCNSService.CNSs(c)
	if err == nil && len(cnss) > 0 {
		vo.CNS = cnss
	}
	return vo, nil
}

// ShowContract 解析合约内容
func (s *contractService) ShowContract(input string) ([]map[string]interface{}, []byte, error) {
	data := "0x" + input
	res, _ := hexutil.Decode(data)

	ptr := new(interface{})
	err := rlp.Decode(bytes.NewReader(res), &ptr)
	if err != nil {
		return nil, nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	deref := reflect.ValueOf(ptr).Elem().Interface()
	wasm := deref.([]interface{})[1].([]byte)

	obj := deref.([]interface{})[2].([]byte)
	var content []map[string]interface{}
	err = json.Unmarshal(obj, &content)
	if err != nil {
		return nil, nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return content, wasm, nil
}

// ParseContractCallResult 解析合约调用返回值
func (s *contractService) ParseContractCallResult(chainID string, callResults []interface{}) ([]*model.ContractCallResult, error) {
	if callResults == nil {
		return nil, nil
	}
	result := make([]*model.ContractCallResult, len(callResults))
	for i, res := range callResults {
		resultMap := make(map[string]interface{}, 0)
		resStr, ok := res.(string)
		if !ok {
			return nil, errors.New("callResult isn't a string")
		}
		err := json.Unmarshal([]byte(resStr), &resultMap)
		if err != nil {
			return nil, err
		}
		var ccr model.ContractCallResult
		ccr.ChainID = chainID
		ccr.Status = resultMap["status"].(string)
		ccr.BlockNumber = uint64(resultMap["blockNumber"].(float64))
		ccr.GasUsed = uint64(resultMap["GasUsed"].(float64))
		ccr.From = resultMap["From"].(string)
		ccr.To = resultMap["To"].(string)
		ccr.TxHash = resultMap["TxHash"].(string)
		ilogs, ok := resultMap["logs"].([]interface{})
		if ok {
			ccr.Logs = make([]string, len(ilogs))
			for j, v := range ilogs {
				ccr.Logs[j] = v.(string)
			}
		}
		errMsg, ok := resultMap["err"].(string)
		if ok {
			ccr.ErrMsg = errMsg
		}

		result[i] = &ccr
	}
	return result, nil
}

// 构建查询条件过滤器
func (s *contractService) buildFilterByCondition(condition model.ContractQueryCondition) (interface{}, error) {
	filter := bson.M{}
	if !reflect.ValueOf(condition.ID).IsZero() {
		id, err := primitive.ObjectIDFromHex(condition.ID)
		if err != nil {
			return nil, exterr.ErrObjectIDInvalid
		}
		filter["_id"] = id
	}
	if !reflect.ValueOf(condition.ChainID).IsZero() {
		chainID, err := primitive.ObjectIDFromHex(condition.ChainID)
		if err != nil {
			return nil, exterr.ErrObjectIDInvalid
		}
		filter["chain_id"] = chainID
	}
	if !reflect.ValueOf(condition.TxHash).IsZero() {
		filter["tx_hash"] = bson.M{"$regex": fmt.Sprintf("^(?i)%s$", condition.TxHash)}
	}
	if !reflect.ValueOf(condition.Address).IsZero() {
		filter["address"] = bson.M{"$regex": fmt.Sprintf("^(?i)%s$", condition.Address)}
	}
	if !reflect.ValueOf(condition.Creator).IsZero() {
		filter["creator"] = condition.Creator
	}
	if !reflect.ValueOf(condition.TimeStart).IsZero() || !reflect.ValueOf(condition.TimeEnd).IsZero() {
		if !reflect.ValueOf(condition.TimeStart).IsZero() && !reflect.ValueOf(condition.TimeEnd).IsZero() {
			filter["timestamp"] = bson.D{
				{"$gte", condition.TimeStart},
				{"$lte", condition.TimeEnd},
			}
		} else if !reflect.ValueOf(condition.TimeStart).IsZero() {
			filter["timestamp"] = bson.D{
				{"$gte", condition.TimeStart},
				{"$lte", time.Now().UnixNano() / 1e6},
			}
		} else {
			filter["timestamp"] = bson.D{
				{"$gte", 0},
				{"$lte", condition.TimeEnd},
			}
		}
	}
	return filter, nil
}
