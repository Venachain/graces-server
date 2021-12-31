package service

import (
	"PlatONE-Graces/db"
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"PlatONE-Graces/util"
	"PlatONE-Graces/web/dao"
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DefaultBlockService IBlockService
)

func init() {
	DefaultBlockService = newBlockService()
}

func newBlockService() IBlockService {
	return &blockService{
		dao: dao.DefaultBlockDao,
	}
}

type blockService struct {
	dao dao.IBlockDao
}

func (s *blockService) BlockByID(id string) (*model.BlockVO, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"_id": objectId,
	}
	block, err := s.dao.Block(filter)
	if err != nil {
		return nil, err
	}
	return block.ToVO()
}

func (s *blockService) BlockByHash(chainID string, hash string) (*model.BlockVO, error) {
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"chain_id": cid,
		"hash":     bson.M{"$regex": fmt.Sprintf("^(?i)%s$", hash)},
	}
	block, err := s.dao.Block(filter)
	if err != nil {
		return nil, err
	}
	return block.ToVO()
}

func (s *blockService) Blocks(condition model.BlockQueryCondition) ([]*model.BlockVO, error) {
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
		sort := bson.D{{"height", -1}}
		findOps.Sort = sort
	}
	var vos []*model.BlockVO
	blocks, err := s.dao.Blocks(filter, findOps)
	if err != nil {
		return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	for _, block := range blocks {
		vo, err := block.ToVO()
		if err != nil {
			return nil, err
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func (s *blockService) Count(condition model.BlockQueryCondition) (int64, error) {
	filter, err := s.buildFilterByCondition(condition)
	if err != nil {
		return 0, err
	}
	findOps := options.Count()
	return s.dao.Count(filter, findOps)
}

func (s *blockService) ChainStats(chainID string) (model.StatsVO, error) {
	var result model.StatsVO
	result.TotalTx = getTotalTx(chainID)
	result.TotalContract = getTotalContract(chainID)
	result.TotalNode = getTotalNode(chainID)
	result.LatestBlock = getLatestBlock(chainID, s)
	return result, nil
}

// 构建查询条件过滤器
func (s *blockService) buildFilterByCondition(condition model.BlockQueryCondition) (interface{}, error) {
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
	if !reflect.ValueOf(condition.Proposer).IsZero() {
		filter["proposer"] = condition.Proposer
	}
	if !reflect.ValueOf(condition.Hash).IsZero() {
		filter["hash"] = condition.Hash
	}
	if !reflect.ValueOf(condition.Height).IsZero() {
		filter["height"] = condition.Height
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

func getTotalTx(chainID string) int64 {
	collection := db.DefaultDB.Collection(collectionNameTX)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)

	objectId, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return 0
	}
	filter := bson.M{"chain_id": objectId}
	count, err := collection.CountDocuments(ctx, filter)
	if nil != err {
		logrus.Errorln("get tx count error")
		return 0
	}
	return count
}

func getTotalContract(chainID string) int64 {
	//collection := db.DefaultDB.Collection(collectionNameTX)
	//ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	objectId, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return 0
	}
	filter := bson.M{"chain_id": objectId}

	filter["receipt.contract_address"] = bson.M{"$ne": ""}

	//count, err := collection.CountDocuments(ctx, filter)
	contractCondition := model.ContractQueryCondition{
		ChainID: chainID,
	}
	count, err := DefaultContractService.Count(contractCondition)
	if nil != err {
		logrus.Errorln("get contract count error")
		return 0
	}
	return count
}

func getTotalNode(chainID string) int64 {
	collection := db.DefaultDB.Collection(collectionNameNode)
	ctx, _ := context.WithTimeout(context.Background(), defaultTimeout)
	objectId, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return 0
	}
	filter := bson.M{"chain_id": objectId}
	count, err := collection.CountDocuments(ctx, filter)
	if nil != err {
		logrus.Errorln("get node count error")
		return 0
	}
	return count
}

func getLatestBlock(chainID string, s *blockService) uint64 {
	objectId, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return 0
	}
	//obj := model.BlockQueryCondition{
	//	ChainID:objectId.Hex(),
	//}
	//res,err := DefaultBlockService.Count(obj)
	//if err != nil{
	//	logrus.Errorln("get block count is error")
	//	return 0
	//}
	//return res
	block, err := dao.DefaultBlockDao.LatestBlock(objectId)
	if err != nil {
		return 0
	}
	return block.Height

}
