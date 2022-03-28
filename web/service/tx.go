package service

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"

	"graces/exterr"
	"graces/model"
	"graces/util"
	"graces/web/dao"

	"github.com/Venachain/Venachain/cmd/vcl/client/packet"
	cmd_common "github.com/Venachain/Venachain/cmd/vcl/common"
	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/rlp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DefaultTXService     ITXService
	InvokeContractAction = "InvokeContract"
	DeployContractAction = "DeployContract"
	TranslateAction      = "TransferFunds"
	ZeroAddress          = "0x0000000000000000000000000000000000000000"
)

var SysContractList = map[string]string{
	"0x0000000000000000000000000000000000000000": "sysContract",
	"0x1000000000000000000000000000000000000001": "userManagerContract",
	"0x1000000000000000000000000000000000000002": "nodeManagerContract",
	"0x0000000000000000000000000000000000000011": "cnsManagerContract",
	"0x1000000000000000000000000000000000000004": "paramManagerContract",
	"0x1000000000000000000000000000000000000005": "fireWallContract",
	"0x1000000000000000000000000000000000000006": "groupManagerContract",
	"0x1000000000000000000000000000000000000007": "contractDataContract",
}

func init() {
	DefaultTXService = newTXService()
}

func newTXService() ITXService {
	return &txService{
		dao: dao.DefaultTXDao,
	}
}

type txService struct {
	dao dao.ITXDao
}

func (s *txService) TXByID(id string) (*model.TXVO, error) {
	var result *model.TXVO
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"_id": objectId,
	}
	tx, err := s.dao.TX(filter)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	result, err = s.TXShow(tx)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return result, nil
}

func (s *txService) TXByHash(chainID string, hash string) (*model.TXVO, error) {
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"chain_id": cid,
		"hash":     bson.M{"$regex": fmt.Sprintf("^(?i)%s$", hash)},
	}
	tx, err := s.dao.TX(filter)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	result, err := s.TXShow(tx)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return result, nil
}

func (s *txService) TXs(condition model.TXQueryCondition) ([]*model.TXVO, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
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
	var vos []*model.TXVO
	txs, err := s.dao.TXs(filter, findOps)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	for _, tx := range txs {
		vo, err := tx.ToVO()
		if err != nil {
			return nil, err
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *txService) TXsForContractCall(condition model.TXQueryCondition) ([]*model.TXVO, error) {
	if reflect.ValueOf(condition.ContractAddress).IsZero() {
		// - 是专门用于查询合约的特定标识
		condition.ContractAddress = "-"
	}
	return s.TXs(condition)
}

func (s *txService) Count(condition model.TXQueryCondition) (int64, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return 0, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	findOps := options.Count()
	var cnt int64
	if cnt, err = s.dao.Count(filter, findOps); err != nil {
		return 0, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return cnt, nil
}

func (s *txService) TXShow(txdata *model.TX) (*model.TXVO, error) {
	res, err := txdata.ToVO()
	if err != nil {
		return nil, err
	}
	res.Detail = &model.TxDetail{}

	if txdata.To == "" {
		// 合约部署
		res.Action = DeployContractAction
		content, code, err := DefaultContractService.ShowContract(txdata.Input)
		if err != nil {
			return nil, err
		}
		res.Detail.Extra = content
		res.Wasm = code
	} else if txdata.Input == "" {
		// 处理转账
		res.Action = TranslateAction
		res.Detail.Params = append(res.Detail.Params, txdata.Value)
	} else if txdata.To == ZeroAddress && txdata.Input != "" {
		// cns 调用
		return ParseCnsInvoke(res, txdata)
	} else {
		// 调用合约
		contract, ok := SysContractList[txdata.To]
		temp, err := ParseData(txdata.ChainID.Hex(), txdata.To, txdata.Input)
		if temp == nil || err != nil {
			res.Detail.Txtype = 0
			res.Detail.Method = ""
			res.Detail.Params = nil
			res.Detail.Extra = ""
		} else {
			res.Detail.Txtype = temp.Detail.Txtype
			res.Detail.Method = temp.Detail.Method
			res.Detail.Params = temp.Detail.Params
			res.Detail.Extra = temp.Detail.Extra
		}
		res.Action = InvokeContractAction
		if ok {
			res.Detail.Contract = contract
		} else {
			res.Detail.Contract = txdata.To
		}
	}
	return res, nil
}

func ParseCnsInvoke(res *model.TXVO, txdata *model.TX) (*model.TXVO, error) {
	var functype []string
	var cns *cmd_common.Cns
	res.Action = "CnsInvoke"
	data := "0x" + txdata.Input
	resbyte, err := hexutil.Decode(data)
	if err != nil {
		logrus.Errorln("decode is error")
		return nil, err
	}
	ptr := new(interface{})
	err = rlp.Decode(bytes.NewReader(resbyte), &ptr)
	if err != nil {
		logrus.Errorln("decode trasaction is error", err)
		return nil, err
	}
	deref := reflect.ValueOf(ptr).Elem().Interface()
	if deref == nil {
		return nil, nil
	}
	for i, v := range deref.([]interface{}) {
		// cns 调用第一个解析出来的参数为txtype
		if i == 0 {
			res.Detail.Txtype = common.BytesToInt64(v.([]byte))
			//cns 调用第二个解析出来的参数为cns name
		} else if i == 1 {
			cns, _, err = cmd_common.CnsParse(string(v.([]byte)))
			if err != nil {
				logrus.Fatalf(err.Error())
			}
			// 第三个解析出来的参数为cns方法
		} else if i == 2 {
			res.Detail.Method = string(v.([]byte))

			condition := model.CNSQueryCondition{
				ChainID: txdata.ChainID.Hex(),
				Name:    cns.Name,
			}
			cnsVo, _ := DefaultCNSService.CNSs(condition)
			if cnsVo == nil {
				return res, errors.New("[CNS] name and version is not registered in CNS")
			}
			cnsAddress := cnsVo[0].Address
			if cns != nil {
				res.Detail.Contract = "(" + cns.Name + ")" + cnsAddress
			}
			functype = getfunctype(txdata.ChainID.Hex(), cnsAddress, res.Detail.Method)
			if functype == nil {
				// 处理调用系统合约的交易
				temp, err := ParseData(txdata.ChainID.Hex(), txdata.To, txdata.Input)
				if temp == nil || err != nil {
					res.Detail.Txtype = 0
					res.Detail.Method = ""
					res.Detail.Params = nil
					res.Detail.Extra = "the address of the contract is incorrect"
				} else {
					res.Detail.Txtype = temp.Detail.Txtype
					res.Detail.Method = temp.Detail.Method
					res.Detail.Params = temp.Detail.Params
					res.Detail.Extra = temp.Detail.Extra
				}
				res.Action = InvokeContractAction

				res.Detail.Contract = txdata.To
				return res, nil
			}
			// 剩下的为调用的参数
		} else {
			j := functype[i-3]
			tmp := getParams(j, v)
			res.Detail.Params = append(res.Detail.Params, tmp)
		}
	}
	return res, nil
}

func ParseData(chainid string, to string, input string) (*model.TXVO, error) {
	result := model.TXVO{}
	result.Detail = &model.TxDetail{}
	var functype []string
	data := "0x" + input
	res, err := hexutil.Decode(data)
	if err != nil {
		logrus.Errorln("decode is error")
		return nil, err
	}
	ptr := new(interface{})
	err = rlp.Decode(bytes.NewReader(res), &ptr)
	if err != nil {
		logrus.Errorln("decode trasaction is error", err)
		return nil, err
	}
	deref := reflect.ValueOf(ptr).Elem().Interface()
	if deref == nil {
		return nil, nil
	}
	for i, v := range deref.([]interface{}) {
		if i == 0 {
			result.Detail.Txtype = common.BytesToInt64(v.([]byte))
		} else if i == 1 {
			result.Detail.Method = string(v.([]byte))
			functype = getfunctype(chainid, to, result.Detail.Method)
			if functype == nil {
				result.Detail.Extra = "the address of the contract is incorrect"
				return &result, nil
			}
		} else {
			j := functype[i-2]
			tmp := getParams(j, v)
			result.Detail.Params = append(result.Detail.Params, tmp)
		}
	}
	return &result, nil
}

func getParams(functype string, v interface{}) interface{} {
	var Params interface{}
	if functype == "string" {
		Params = string(v.([]byte))
	} else if functype == "int32" {
		Params = common.BytesToInt32(v.([]byte))
	} else if functype == "int64" {
		Params = common.BytesToInt64(v.([]byte))
	} else if functype == "float32" {
		Params = common.BytesToFloat32(v.([]byte))
	} else if functype == "float64" {
		Params = common.BytesToFloat64(v.([]byte))
	} else if functype == "uint64" {
		Params = binary.BigEndian.Uint64(v.([]byte))
	} else if functype == "uint32" {
		Params = binary.BigEndian.Uint32(v.([]byte))
	} else if functype == "Uint16" {
		Params = binary.BigEndian.Uint16(v.([]byte))
	} else {
		Params = "no suitable type"
	}
	return Params
}

func getfunctype(chainid string, to string, funcName string) []string {
	var functype []string
	var funcAbi []byte
	_, ok := SysContractList[to]
	if ok {
		funcAbi = cmd_common.AbiParse("", to)
	} else {
		funcAbi = getFuncAbi(chainid, to)
		//tmp := getFuncAbi(chainid, to)
		//funcAbi = tmp
		if funcAbi == nil {
			return nil
		}
	}

	contractAbi, err := packet.ParseAbiFromJson(funcAbi)
	if err != nil {
		logrus.Errorf("ParseAbiFromJson is wrong", err.Error())
		return functype
	}
	methodAbi, err := contractAbi.GetFuncFromAbi(funcName)
	if err != nil {
		logrus.Errorf("GetFuncFromAbi is wrong", err.Error())
		return functype
	}
	for _, j := range methodAbi.Inputs {
		functype = append(functype, j.Type)
	}
	return functype
}

func getFuncAbi(id string, to string) []byte {
	chain, err := DefaultChainService.ChainByID(id)
	if err != nil {
		logrus.Errorf("failed to ChainByID: %s", err.Error())
		return nil
	}
	endpoint := fmt.Sprintf("http://%v:%v", chain.IP, chain.RPCPort)
	var param []string
	param = append(param, to)
	param = append(param, "latest")
	code, err := model.GetRpcResult(endpoint, "eth_getCode", param)
	if err != nil {
		logrus.Errorf("failed to GetRpcResult: %s", err.Error())
		return nil
	}
	if code == "0x" {
		return nil
	}
	res, _ := hexutil.Decode(code.(string))

	ptr := new(interface{})
	err = rlp.Decode(bytes.NewReader(res), &ptr)
	if err != nil {
		return nil
	}
	deref := reflect.ValueOf(ptr).Elem().Interface()
	funcAbi := deref.([]interface{})[2].([]byte)
	return funcAbi
}

// History 统计一周内的交易总量
func (s *txService) History(chainID string) (*[]model.TxStatsVO, error) {
	result := make([]model.TxStatsVO, 0)
	// 统计7天的交易数量
	num := 7
	for i := -1; i >= (-num); i-- {
		temp := model.TxStatsVO{}
		day := time.Now().AddDate(0, 0, i)

		temp.Date = day.Format("2006-01-02")
		y, m, d := day.Date()
		start := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
		end := time.Date(y, m, d, 23, 59, 59, 0, time.Local)

		condition := model.TXQueryCondition{
			ChainID:   chainID,
			TimeStart: start.UnixNano() / int64(time.Millisecond),
			TimeEnd:   end.UnixNano() / int64(time.Millisecond),
		}
		cnt, _ := s.Count(condition)
		temp.TxAmount = int(cnt)
		result = append(result, temp)
	}

	return &result, nil
}

// 构建查询条件过滤器
func (s *txService) buildFilterByCondition(condition model.TXQueryCondition) (interface{}, error) {
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
	if !reflect.ValueOf(condition.BlockID).IsZero() {
		blockID, err := primitive.ObjectIDFromHex(condition.BlockID)
		if err != nil {
			return nil, exterr.ErrObjectIDInvalid
		}
		filter["block_id"] = blockID
	}
	if !reflect.ValueOf(condition.Hash).IsZero() {
		filter["hash"] = bson.M{"$regex": fmt.Sprintf("^(?i)%s$", condition.Hash)}
	}
	if !reflect.ValueOf(condition.Height).IsZero() {
		filter["height"] = condition.Height
	}
	if !reflect.ValueOf(condition.Status).IsZero() {
		filter["receipt.status"] = condition.Status
	}
	if !reflect.ValueOf(condition.ContractAddress).IsZero() {
		// - 是专门用于查询合约的特定标识
		if condition.ContractAddress == "-" {
			filter["receipt.contract_address"] = bson.M{"$ne": ""}
			filter["to"] = bson.M{"$ne": ""}
		} else {
			filter["receipt.contract_address"] = bson.M{"$regex": fmt.Sprintf("^(?i)%s$", condition.ContractAddress)}
		}
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
	if !reflect.ValueOf(condition.ParticipantHash).IsZero() {
		filter["$or"] = []bson.D{
			{{"from", condition.ParticipantHash}},
			{{"to", condition.ParticipantHash}},
		}
	}
	return filter, nil
}
