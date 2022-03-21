package service

import (
	"context"
	"time"

	"graces/exterr"
	"graces/model"
	"graces/rpc"
	"graces/util"
	"graces/web/dao"

	"github.com/Venachain/Venachain/common"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DefaultAccountService IAccountService
)

func init() {
	DefaultAccountService = newAccountService()
}

func newAccountService() IAccountService {
	return &accountService{}
}

type accountService struct {
}

func (s *accountService) LockAccount(dto model.LockAccountDTO) (bool, error) {
	client, err := rpc.GetRPCClientByChainID(dto.ChainID)
	if err != nil {
		return false, exterr.NewError(exterr.ErrCodeUpdate, err.Error())
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var res bool
	err = client.RpcClient().CallContext(ctx, &res, "personal_lockAccount", common.HexToAddress(dto.Account))
	if err != nil {
		return false, exterr.NewError(exterr.ErrCodeUpdate, err.Error())
	}
	logrus.Debugf("account[%v] lock：%v", dto.Account, res)
	return res, nil
}

func (s *accountService) UnlockAccount(dto model.UnlockAccountDTO) (bool, error) {
	client, err := rpc.GetRPCClientByChainID(dto.ChainID)
	if err != nil {
		return false, exterr.NewError(exterr.ErrCodeUpdate, err.Error())
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var res bool
	err = client.RpcClient().CallContext(ctx, &res, "personal_unlockAccount", common.HexToAddress(dto.Account), dto.Password, dto.Duration)
	if err != nil {
		return false, exterr.NewError(exterr.ErrCodeUpdate, err.Error())
	}
	logrus.Debugf("account[%v] unlock：%v", dto.Account, res)
	return res, nil
}

func (s *accountService) FirstAccount(chainID string) (string, error) {
	client, err := rpc.GetRPCClientByChainID(chainID)
	if err != nil {
		return "", exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var addresses []string
	err = client.RpcClient().CallContext(ctx, &addresses, "personal_listAccounts")
	if err != nil {
		return "", exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	return addresses[0], nil
}

func (s *accountService) ListAccounts(dto model.AccountDTO) ([]*model.AccountVO, error) {
	if dto.ChainID == "" {
		logrus.Errorln("Missing Chain ID!")
		return nil, exterr.NewError(exterr.ErrCodeFind, "chain ID must not be null")
	}

	var accounts []*model.AccountVO
	// 获取指定节点的账户信息
	if dto.NodeID != "" {
		client, err := rpc.GetRPCClientByChainIDAndNodeID(dto.ChainID, dto.NodeID)
		if err != nil {
			return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
		}
		var addresses []string
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		err = client.RpcClient().CallContext(ctx, &addresses, "personal_listAccounts")
		if err != nil {
			return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
		}
		for _, address := range addresses {
			account := &model.AccountVO{
				ChainID: dto.ChainID,
				NodeID:  dto.NodeID,
				Account: address,
			}
			accounts = append(accounts, account)
		}
		return accounts, nil
	}

	// 获取链上所有节点的账户信息
	filter := bson.M{}
	filter["chain_id"], _ = primitive.ObjectIDFromHex(dto.ChainID)
	condition := model.NodeQueryCondition{}
	findOps := util.BuildOptionsByQuery(condition.PageIndex, condition.PageSize)
	nodes, err := dao.DefaultNodeDao.Nodes(filter, findOps)
	if err != nil {
		logrus.Errorln(err)
	}

	for _, node := range nodes {
		client, err := rpc.GetRPCClientByChainIDAndNodeID(dto.ChainID, node.ID.Hex())
		if err != nil {
			return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
		}

		var addresses []string
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		err = client.RpcClient().CallContext(ctx, &addresses, "personal_listAccounts")
		if err != nil {
			return nil, exterr.NewError(exterr.ErrCodeFind, err.Error())
		}
		for _, address := range addresses {
			account := &model.AccountVO{
				ChainID: dto.ChainID,
				NodeID:  node.ID.Hex(),
				Account: address,
			}
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}
