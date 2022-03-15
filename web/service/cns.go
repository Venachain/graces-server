package service

import (
	"fmt"
	"reflect"
	"strings"

	"graces/exterr"
	"graces/model"
	"graces/rpc"
	"graces/syncer"
	"graces/util"
	"graces/web/dao"

	precompile "github.com/Venachain/Venachain/cmd/vcl/client/precompiled"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DefaultCNSService ICNSService
)

func init() {
	DefaultCNSService = newCNSService()
}

func newCNSService() ICNSService {
	return &cnsService{dao: dao.DefaultCNSDao}
}

type cnsService struct {
	dao dao.ICNSDao
}

func (s *cnsService) CNSByID(id string) (*model.CNSVO, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"_id": objectId,
	}
	cns, err := s.dao.CNS(filter)
	if err != nil {
		return nil, err
	}
	return cns.ToVO()
}

func (s *cnsService) CNS(chainID string, name, address, version string) (*model.CNSVO, error) {
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"chain_id": cid,
		"name":     bson.M{"$regex": fmt.Sprintf("^(?i)%s$", name)},
		"address":  bson.M{"$regex": fmt.Sprintf("^(?i)%s$", address)},
		"version":  version,
	}
	cns, err := s.dao.CNS(filter)
	if err != nil {
		return nil, err
	}
	return cns.ToVO()
}

func (s *cnsService) CNSs(condition model.CNSQueryCondition) ([]*model.CNSVO, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return nil, err
	}
	findOps := util.BuildOptionsByQuery(condition.PageIndex, condition.PageSize)
	if !reflect.ValueOf(condition.Sort).IsZero() {
		sort := bson.D{}
		for k, v := range condition.Sort {
			sort = append(sort, bson.E{k, v})
		}
		findOps.Sort = sort
	} else {
		sort := bson.D{{"_id", -1}}
		findOps.Sort = sort
	}
	var vos []*model.CNSVO
	cnss, err := s.dao.CNSs(filter, findOps)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	for _, cns := range cnss {
		vo, err := cns.ToVO()
		if err != nil {
			return nil, err
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *cnsService) Count(condition model.CNSQueryCondition) (int64, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return 0, err
	}
	findOps := options.Count()
	return s.dao.Count(filter, findOps)
}

func (s *cnsService) Register(dto model.CNSRegisterDTO) (*model.ContractCallResult, error) {
	account, err := DefaultAccountService.FirstAccount(dto.ChainID)
	if err != nil {
		return nil, err
	}
	accountDTO := model.UnlockAccountDTO{
		LockAccountDTO: model.LockAccountDTO{
			AccountDTO: model.AccountDTO{ChainID: dto.ChainID, NodeID: ""},
			Account:    account,
		},
		Password: "0",
		Duration: 0,
	}
	unlock, err := DefaultAccountService.UnlockAccount(accountDTO)
	if err != nil || !unlock {
		return nil, err
	}

	client, err := rpc.GetRPCClientByChainID(dto.ChainID)
	if err != nil {
		return nil, err
	}

	caller := rpc.NewMsgCaller(client)
	txParams := &rpc.TxParams{From: account}
	contractParams := s.buildCnsRegisterParam(dto)
	res, err := caller.Call(txParams, contractParams)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("cnsRegister result：%+v", res)
	results, err := DefaultContractService.ParseContractCallResult(dto.ChainID, res)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		logrus.Errorf("cns register error: no result")
		return nil, exterr.NewError(exterr.ErrorContractWithCNS, "no result")
	}
	result := results[0]
	if !strings.Contains(result.Logs[0], "register succeed") {
		return nil, exterr.NewError(exterr.ErrorContractWithCNS, result.Logs[0])
	}

	// 触发 CNS 数据同步
	s.fireCNSSync(dto.ChainID)
	logrus.Infof("CNS register success: %+v", results)
	return result, nil
}

func (s *cnsService) fireCNSSync(chainID string) {
	syncInfo := syncer.DefaultChainDataSyncManager.BuildChainSyncInfo(chainID)
	if syncInfo.CNSDataSyncInfo != nil && syncInfo.CNSDataSyncInfo.Status == syncer.StatusSyncing {
		logrus.Infof("this chain[%s] CNS data is syncing, don't repeat sync for it", chainID)
		return
	}
	err := syncer.DefaultChainDataSyncManager.SyncCNS(chainID, false)
	if err != nil {
		syncInfo.CNSDataSyncInfo.ErrMsg = err.Error()
		syncInfo.CNSDataSyncInfo.Status = syncer.StatusError
		syncer.DefaultChainDataSyncManager.ErrChan <- &model.SyncErrMsg{
			ChainID: chainID,
			ErrType: syncer.ErrTypeBlockOrTXSync,
			Err:     err,
		}
		return
	}
}

func (s *cnsService) buildCnsRegisterParam(dto model.CNSRegisterDTO) *rpc.ContractParams {
	contractAddr := precompile.CnsManagementAddress
	funcName := "cnsRegister"
	interpreter := "wasm"

	// 构造 cnsRegister() 函数的查询条件参数
	funcParams := &struct {
		Name    string
		Version string
		Address string
	}{dto.Name, dto.Version, dto.Address}

	// 构造合约参数
	contract := &rpc.ContractParams{
		ContractAddr: contractAddr,
		Method:       funcName,
		Interpreter:  interpreter,
		AbiMethods:   nil,
		Data:         funcParams,
	}
	return contract
}

func (s *cnsService) Redirect(dto model.CNSRedirectDTO) (*model.ContractCallResult, error) {
	account, err := DefaultAccountService.FirstAccount(dto.ChainID)
	if err != nil {
		return nil, err
	}
	accountDTO := model.UnlockAccountDTO{
		LockAccountDTO: model.LockAccountDTO{
			AccountDTO: model.AccountDTO{ChainID: dto.ChainID, NodeID: ""},
			Account:    account,
		},
		Password: "0",
		Duration: 0,
	}
	unlock, err := DefaultAccountService.UnlockAccount(accountDTO)
	if err != nil || !unlock {
		return nil, err
	}

	client, err := rpc.GetRPCClientByChainID(dto.ChainID)
	if err != nil {
		return nil, err
	}

	caller := rpc.NewMsgCaller(client)
	txParams := &rpc.TxParams{From: account}
	contractParams := s.buildCnsRedirectParam(dto)
	res, err := caller.Call(txParams, contractParams)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("cnsRedirect result：%+v", res)

	results, err := DefaultContractService.ParseContractCallResult(dto.ChainID, res)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		logrus.Errorf("cns redirect error: no result")
		return nil, exterr.NewError(exterr.ErrorContractWithCNS, "no result")
	}
	result := results[0]
	if !strings.Contains(result.Logs[0], "redirect succeed") {
		return nil, exterr.NewError(exterr.ErrorContractWithCNS, result.Logs[0])
	}

	// 触发 CNS 数据同步
	s.fireCNSSync(dto.ChainID)
	logrus.Infof("CNS redirect success: %+v", result)
	return result, nil
}

func (s *cnsService) buildCnsRedirectParam(dto model.CNSRedirectDTO) *rpc.ContractParams {
	contractAddr := precompile.CnsManagementAddress
	funcName := "cnsRedirect"
	interpreter := "wasm"

	// 构造 cnsRegister() 函数的查询条件参数
	funcParams := &struct {
		Name    string
		Version string
	}{dto.Name, dto.Version}

	// 构造合约参数
	contract := &rpc.ContractParams{
		ContractAddr: contractAddr,
		Method:       funcName,
		Interpreter:  interpreter,
		AbiMethods:   nil,
		Data:         funcParams,
	}
	return contract
}

// 构建查询条件过滤器
func (s *cnsService) buildFilterByCondition(condition model.CNSQueryCondition) (interface{}, error) {
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
	if !reflect.ValueOf(condition.Name).IsZero() {
		filter["name"] = bson.M{"$regex": fmt.Sprintf("^(?i)%s$", condition.Name)}
	}
	if !reflect.ValueOf(condition.Address).IsZero() {
		filter["address"] = bson.M{"$regex": fmt.Sprintf("^(?i)%s$", condition.Address)}
	}
	if !reflect.ValueOf(condition.Version).IsZero() {
		filter["version"] = condition.Version
	}
	return filter, nil
}
