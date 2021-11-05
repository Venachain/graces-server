package service

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"PlatONE-Graces/rpc"
	"PlatONE-Graces/util"
	"PlatONE-Graces/web/dao"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/PlatONEnetwork/PlatONE-Go/common"

	"github.com/sirupsen/logrus"
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
		return false, err
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var res bool
	err = client.RpcClient().CallContext(ctx, &res, "personal_lockAccount", common.HexToAddress(dto.Account))
	if err != nil {
		return false, err
	}
	logrus.Debugf("account[%v] lock：%v", dto.Account, res)
	return res, nil
}

func (s *accountService) UnlockAccount(dto model.UnlockAccountDTO) (bool, error) {
	client, err := rpc.GetRPCClientByChainID(dto.ChainID)
	if err != nil {
		return false, err
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var res bool
	err = client.RpcClient().CallContext(ctx, &res, "personal_unlockAccount", common.HexToAddress(dto.Account), dto.Password, dto.Duration)
	if err != nil {
		return false, err
	}
	logrus.Debugf("account[%v] unlock：%v", dto.Account, res)
	return res, nil
}

func (s *accountService) FirstAccount(chainID string) (string, error) {
	client, err := rpc.GetRPCClientByChainID(chainID)
	if err != nil {
		return "", err
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	var addresses []string
	err = client.RpcClient().CallContext(ctx, &addresses, "personal_listAccounts")
	if err != nil {
		return "", err
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
			return nil, err
		}
		var addresses []string
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		err = client.RpcClient().CallContext(ctx, &addresses, "personal_listAccounts")
		if err != nil {
			return nil, err
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
			return nil, err
		}

		var addresses []string
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		err = client.RpcClient().CallContext(ctx, &addresses, "personal_listAccounts")
		if err != nil {
			return nil, err
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
