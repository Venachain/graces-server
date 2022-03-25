package service

import (
	"fmt"
	"reflect"

	"graces/exterr"
	"graces/model"
	"graces/rpc"
	"graces/util"
	"graces/web/dao"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DefaultNodeService INodeService

func init() {
	DefaultNodeService = newNodeService()
}

func newNodeService() INodeService {
	return &nodeService{
		dao: dao.DefaultNodeDao,
	}
}

type nodeService struct {
	dao dao.INodeDao
}

func (s *nodeService) NodeSyncServer(node *model.NodeSyncReq) (*model.SyncNodeResult, error) {
	res := &model.SyncNodeResult{}
	var err error
	endpoint := fmt.Sprintf("http://%v:%v", node.Ip, node.Port)
	ping, err := rpc.Ping(endpoint)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, fmt.Sprintf("failed to connect: %s", err.Error()))
	}
	if !ping {
		return nil, exterr.NewError(exterr.ErrCodeFind, "failed to connect: connect timeout")
	}
	blockNumber, err := rpc.GetBlockNumber(endpoint)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	res.BlockNumber = uint32(blockNumber)

	isMining, err := model.GetRpcResult(endpoint, "eth_mining", []string{})
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	res.IsMining = isMining.(bool)
	price, err := model.GetRpcResult(endpoint, "eth_gasPrice", []string{})
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	gasprice := price.(string)
	gasPrice, err := util.Hex2Number(gasprice)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeParameterInvalid, err.Error())
	}
	res.GasPrice = uint32(gasPrice)
	pendingtx, err := model.GetRpcResult(endpoint, "eth_pendingTransactions", []string{})
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}

	pendingtxs, ok := pendingtx.([]string)
	if ok {
		res.PendingTx = pendingtxs
	} else {
		res.PendingTx = nil
	}

	res.PendingNumber = len(res.PendingTx)

	return res, nil
}

func (s *nodeService) NodeByID(id string) (*model.NodeVO, error) {
	nodeId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"_id": nodeId,
	}
	node, err := s.dao.Node(filter)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	vo, err := node.ToVO()
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	endpoint := fmt.Sprintf("http://%v:%v", node.ExternalIP, node.RPCPort)
	blockNumber, err := rpc.GetBlockNumber(endpoint)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	vo.Blocknumber = uint32(blockNumber)
	vo.IsAlive = true
	return vo, nil
}

func (s *nodeService) Nodes(condition model.NodeQueryCondition) ([]*model.NodeVO, error) {
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
			sort = append(sort, bson.E{k, v})
		}
		findOps.Sort = sort
	} else {
		sort := bson.D{{"name", 1}}
		findOps.Sort = sort
	}
	var vos []*model.NodeVO
	nodes, err := s.dao.Nodes(filter, findOps)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	for _, node := range nodes {
		vo, err := node.ToVO()
		if err != nil {
			return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
		}
		endpoint := fmt.Sprintf("http://%v:%v", node.ExternalIP, node.RPCPort)
		blockNumber, err := rpc.GetBlockNumber(endpoint)
		if err != nil {
			return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
		}
		vo.Blocknumber = uint32(blockNumber)
		vo.IsAlive = true
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *nodeService) Count(condition model.NodeQueryCondition) (int64, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return 0, err
	}
	findOps := options.Count()
	var cnt int64
	if cnt, err = s.dao.Count(filter, findOps); err != nil {
		return 0, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return cnt, nil
}

// 构建查询条件过滤器
func (s *nodeService) buildFilterByCondition(condition model.NodeQueryCondition) (interface{}, error) {
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
		filter["name"] = condition.Name
	}
	if !reflect.ValueOf(condition.InternalIP).IsZero() {
		filter["internal_ip"] = condition.InternalIP
	}
	if !reflect.ValueOf(condition.ExternalIP).IsZero() {
		filter["external_ip"] = condition.ExternalIP
	}
	return filter, nil
}
