package service

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"graces/config"
	"graces/exterr"
	"graces/model"
	"graces/rpc"
	"graces/util"
	"graces/web/dao"
	"graces/ws"

	precompile "github.com/Venachain/Venachain/cmd/vcl/client/precompiled"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionNameTX     = "txs"
	collectionNameNode   = "nodes"
	collectionNameBlocks = "blocks"
	defaultTimeout       = 120 * time.Second
)

var (
	DefaultChainService IChainService
)

func init() {
	DefaultChainService = newChainService(dao.DefaultChainDao)
}

func newChainService(dao dao.IChainDao) *chainService {
	return &chainService{
		dao: dao,
	}
}

type chainService struct {
	dao dao.IChainDao
}

//IsExist
// 判断一条链是否存在，当前的判断方法是判断 name 相同，或 IP 相同且 prc_port、p2p_port、ws_port 至少有一个相同
func (s chainService) IsExist(chainDTO model.ChainDTO) bool {
	filter := bson.M{}
	filter["$or"] = []bson.D{
		{{"name", chainDTO.Name}},
		{{"$and", []bson.D{
			{{"ip", chainDTO.IP}},
			{{"$or", []bson.D{
				{{"rpc_port", chainDTO.RPCPort}},
				{{"p2p_port", chainDTO.P2PPort}},
				{{"ws_port", chainDTO.WSPort}},
			}},
			}},
		}},
	}
	// 同SQL：select * from chains where name = #{name} or (ip = #{ip} and (rpc_port = #{rpc_port} or p2p_port = #{p2p_port} or ws_port = #{ws_port}))
	chain, err := s.dao.Chain(filter)
	if err != nil || chain.ID.IsZero() {
		return false
	}
	return true
}

func (s *chainService) InsertChain(chainDTO model.ChainDTO) error {
	// 1、验重
	if s.IsExist(chainDTO) {
		msg := fmt.Sprintf("chain[%s] is existed or it's port is conflicted in ip[%v]", chainDTO.Name, chainDTO.IP)
		return exterr.NewError(exterr.ErrCodeParameterInvalid, msg)
	}

	// 2、构建链数据
	chain := &model.Chain{}
	err := chain.ValueOfDTO(chainDTO)
	if err != nil {
		return err
	}
	if chain.ChainConfig == nil {
		chain.ChainConfig = config.Config.ChainConfig
	}

	// 3、ping
	_, err = s.ping(*chain)
	if err != nil {
		return exterr.NewError(exterr.ErrCodeInsert, err.Error())
	}

	// 4、保存链数据
	err = s.dao.InsertChain(*chain)
	if err != nil {
		return exterr.NewError(exterr.ErrCodeInsert, err.Error())
	}

	// 5、给新增的链订阅事件
	err = ws.DefaultWSSubscriber.SubTopicsForChain(chain)
	if err != nil {
		return exterr.NewError(exterr.ErrCodeInsert, err.Error())
	}
	return nil
}

func (s *chainService) ChainByID(id string) (*model.ChainVO, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{"_id": objectId}
	chain, err := s.dao.Chain(filter)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return chain.ToVO()
}

func (s *chainService) ChainByName(name string) (*model.ChainVO, error) {
	filter := bson.M{"name": name}
	chain, err := s.dao.Chain(filter)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return chain.ToVO()
}

func (s *chainService) ChainByAddress(ip string, port int64) (*model.ChainVO, error) {
	filter := bson.M{"ip": ip, "port": port}
	chain, err := s.dao.Chain(filter)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return chain.ToVO()
}

func (s *chainService) Chains(condition model.ChainQueryCondition) ([]*model.ChainVO, error) {
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
	var vos []*model.ChainVO
	chains, err := s.dao.Chains(filter, findOps)
	if err != nil {
		return vos, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	for _, chain := range chains {
		vo, err := chain.ToVO()
		if err != nil {
			return nil, err
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *chainService) Count(condition model.ChainQueryCondition) (int64, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return 0, err
	}
	findOps := options.Count()
	return s.dao.Count(filter, findOps)
}

func (s *chainService) GetSysConfigString(id string, funcName string) (string, error) {
	var result string
	contractAddr := precompile.ParameterManagementAddress
	defaultInter := "wasm"
	chain, _ := rpc.GetRPCClientByChainID(id)
	caller := rpc.NewMsgCaller(chain)

	data := rpc.NewContractParams(contractAddr, funcName, defaultInter, nil, nil)
	txParams := &rpc.TxParams{}
	res, err := caller.Call(txParams, data)
	if err != nil {
		logrus.Errorln("call blockGasLimit contract error")
		return "", err
	}
	result = fmt.Sprintf("%v", res[0])
	if result == "" {
		logrus.Errorln("GetSysConfig result get nothing")
		return "", errors.New("GetSysConfig result get nothing")
	}
	return result, nil
}

func (s *chainService) SetSysConfigString(id string, funcName string, funcParams interface{}) (string, error) {
	var result string
	var err error
	contractAddr := precompile.ParameterManagementAddress
	defaultInter := "wasm"
	chainid, _ := rpc.GetRPCClientByChainID(id)
	caller := rpc.NewMsgCaller(chainid)

	data := rpc.NewContractParams(contractAddr, funcName, defaultInter, nil, funcParams)
	txParams := &rpc.TxParams{}
	txParams.From, err = DefaultAccountService.FirstAccount(id)
	if err != nil {
		logrus.Errorln("get first account is error")
		return "", err
	}
	res, err := caller.Call(txParams, data)
	if err != nil {
		logrus.Errorln("call blockGasLimit contract error")
		return "", err
	}

	result = fmt.Sprintf("%v", res[0])
	return result, nil
}

// 构建查询条件过滤器
func (s *chainService) buildFilterByCondition(condition model.ChainQueryCondition) (interface{}, error) {
	filter := bson.M{}
	if !reflect.ValueOf(condition.ID).IsZero() {
		id, err := primitive.ObjectIDFromHex(condition.ID)
		if err != nil {
			return nil, exterr.ErrObjectIDInvalid
		}
		filter["_id"] = id
	}
	if !reflect.ValueOf(condition.Name).IsZero() {
		filter["name"] = condition.Name
	}
	if !reflect.ValueOf(condition.IP).IsZero() {
		filter["ip"] = condition.IP
	}
	if !reflect.ValueOf(condition.RPCPort).IsZero() {
		filter["rpc_port"] = condition.RPCPort
	}
	if !reflect.ValueOf(condition.P2PPort).IsZero() {
		filter["p2p_port"] = condition.P2PPort
	}
	if !reflect.ValueOf(condition.WSPort).IsZero() {
		filter["ws_port"] = condition.WSPort
	}
	return filter, nil
}

// 对链的 RPC、P2P、WS 端口进行 ping，看是否都能 ping 通
func (s *chainService) ping(chain model.Chain) (bool, error) {
	rpcHost := fmt.Sprintf("%s:%v", chain.IP, chain.RPCPort)
	rpcUri := url.URL{
		Scheme: "http",
		Host:   rpcHost,
	}
	rpcPing, err := rpc.Ping(rpcUri.String())
	if err != nil {
		return false, err
	}

	p2pHost := fmt.Sprintf("%s:%v", chain.IP, chain.P2PPort)
	p2pUri := url.URL{
		Scheme: "http",
		Host:   p2pHost,
	}
	p2pPing, err := rpc.Ping(p2pUri.String())
	if err != nil {
		return false, err
	}

	wsHost := fmt.Sprintf("%s:%v", chain.IP, chain.WSPort)
	wsUri := url.URL{
		Scheme: "ws",
		Host:   wsHost,
	}
	wsPing, err := rpc.Ping(wsUri.String())
	if err != nil {
		return false, err
	}
	return rpcPing && p2pPing && wsPing, nil
}
