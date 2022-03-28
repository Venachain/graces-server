package syncer

import (
	"graces/model"
	"graces/rpc"
	"graces/web/dao"

	"github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DefaultSyncer *syncer
)

func init() {
	DefaultSyncer = newSyncer()
}

func newSyncer() *syncer {
	return &syncer{
		chainDao: dao.DefaultChainDao,
		nodeDao:  dao.DefaultNodeDao,
	}
}

type syncer struct {
	chainDao dao.IChainDao
	nodeDao  dao.INodeDao
}

// BlockFullSync 区块全量同步
func (s *syncer) BlockFullSync(chainID string) error {
	latestBlock, err := rpc.GetLatestBlockFromChain(chainID)
	if err != nil {
		return err
	}
	for i := 0; i <= int(latestBlock.NumberU64()); i++ {
		err := s.syncBlockByNumber(chainID, int64(i), true)
		if err != nil {
			return err
		}
	}
	return nil
}

// BlockIncrSync 区块增量同步
func (s *syncer) BlockIncrSync(chainID string, curHeight uint64, targetHeight uint64) error {
	for i := curHeight + 1; i <= targetHeight; i++ {
		err := s.syncBlockByNumber(chainID, int64(i), false)
		if err != nil {
			return err
		}
	}
	return nil
}

// 通过块高同步单个区块
func (s *syncer) syncBlockByNumber(chainID string, number int64, isFullSync bool) error {
	block, err := rpc.GetBlockByNumber(chainID, number)
	if err != nil {
		return err
	}
	err = s.saveBlockData(*block, isFullSync)
	if err != nil {
		return err
	}
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return err
	}
	filter := bson.M{
		"chain_id": cid,
		"hash":     block.Hash,
	}
	dbBlock, err := dao.DefaultBlockDao.Block(filter)
	if err != nil {
		return err
	}
	return s.syncTXDataByBlockNumber(dbBlock.ChainID.Hex(), dbBlock.ID.Hex(), int64(dbBlock.Height), isFullSync)
}

// 保存区块数据入库
// 增量同步，对于不存在的数据则插入，对于已存在的数据则不做任何处理，因为链上的区块数据不会被修改
// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
func (s *syncer) saveBlockData(block model.Block, isFullSync bool) error {
	filter := bson.M{
		"chain_id": block.ChainID,
		"hash":     block.Hash,
	}
	// 增量同步，对于不存在的数据则插入，对于已存在的数据则不做任何处理，因为链上的区块数据不会被修改
	if !isFullSync {
		dbBlock, err := dao.DefaultBlockDao.Block(filter)
		if err != nil || dbBlock.ID.IsZero() {
			err = dao.DefaultBlockDao.InsertBlock(block)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
	update := bson.M{
		"$set": bson.M{
			"height":      block.Height,
			"timestamp":   block.Timestamp,
			"tx_amount":   block.TxAmount,
			"proposer":    block.Proposer,
			"gas_used":    block.GasUsed,
			"gas_limit":   block.GasLimit,
			"parent_hash": block.ParentHash,
			"extra_data":  block.ExtraData,
			"size":        block.Size,
			"head":        block.Head,
		},
	}
	updateOptions := options.Update()
	upsert := true
	updateOptions.Upsert = &upsert
	err := dao.DefaultBlockDao.Update(filter, update, updateOptions)
	if err != nil {
		return err
	}
	return nil
}

// SyncCNS 同步 CNS 数据
func (s *syncer) SyncCNS(chainID string, isFullSync bool) error {
	allCNS, err := rpc.GetAllCNS(chainID)
	if err != nil {
		return err
	}
	for _, cns := range allCNS {
		err = s.saveCNS(*cns, isFullSync)
		if err != nil {
			return err
		}
	}
	return err
}

// 保存 CNS 数据入库
// 增量同步，对于不存在的数据则插入，对于已存在的数据则不做任何处理，因为链上的区块数据不会被修改
// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
func (s *syncer) saveCNS(cns model.CNS, isFullSync bool) error {
	filter := bson.M{
		"chain_id": cns.ChainID,
		"name":     cns.Name,
		"address":  cns.Address,
		"version":  cns.Version,
	}
	// 增量同步，对于不存在的数据则插入，对于已存在的数据则不做任何处理，因为链上的区块数据不会被修改
	if !isFullSync {
		dbCNS, err := dao.DefaultCNSDao.CNS(filter)
		if err != nil || dbCNS.ID.IsZero() {
			err = dao.DefaultCNSDao.InsertCNS(cns)
			if err != nil {
				return err
			}
		}
		return nil
	}
	// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
	update := bson.M{
		"$set": bson.M{
			"chain_id": cns.ChainID,
			"name":     cns.Name,
			"address":  cns.Address,
			"version":  cns.Version,
		},
	}
	updateOptions := options.Update()
	upsert := true
	updateOptions.Upsert = &upsert
	err := dao.DefaultCNSDao.Update(filter, update, updateOptions)
	if err != nil {
		return err
	}
	return nil
}

// SyncNode 同步节点数据
func (s *syncer) SyncNode(chainID string, isFullSync bool) error {
	nodes, err := rpc.GetAllNodes(chainID)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		err = s.saveNode(*node, isFullSync)
		if err != nil {
			return err
		}
	}
	return nil
}

// 保存 node 数据入库
// 节点状态是会变化的，所以无论是增量还是全量更新都要更新节点数据
func (s *syncer) saveNode(node model.Node, isFullSync bool) error {
	chain, err := dao.DefaultChainDao.Chain(bson.M{"_id": node.ChainID})
	if err != nil {
		return err
	}
	// 如果节点的公网IP使用的是本地IP，则使用链的IP来作为它的公网IP
	if node.ExternalIP == "127.0.0.1" || node.ExternalIP == "localhost" {
		node.ExternalIP = chain.IP
	}
	find := bson.M{
		"chain_id": node.ChainID,
		"name":     node.Name,
	}
	update := bson.M{"$set": bson.M{
		"public_key":  node.PublicKey,
		"desc":        node.Desc,
		"internal_ip": node.InternalIP,
		"external_ip": node.ExternalIP,
		"rpc_port":    node.RPCPort,
		"p2p_port":    node.P2PPort,
		"type":        node.Type,
		"status":      node.Status,
		"owner":       node.Owner,
	}}
	upsert := true
	updateOptions := options.Update()
	updateOptions.Upsert = &upsert
	// 数据存在则更新，不存在则插入
	err = dao.DefaultNodeDao.Update(find, update, updateOptions)
	if err != nil {
		return err
	}
	return nil
}

// 通过块高同步区块内的交易数据
func (s *syncer) syncTXDataByBlockNumber(chainID string, blockID string, blockNumber int64, isFullSync bool) error {
	txs, err := rpc.GetTXDataByBlockNumber(chainID, blockID, blockNumber)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		err = s.saveTXData(*tx, isFullSync)
		if err != nil {
			logrus.Errorf("save tx data error：%v", err)
		}
	}
	return nil
}

// 保存单个交易数据入库
// 增量同步，对于不存在的数据则插入，对于已存在的数据则不做任何处理，因为链上的区块数据不会被修改
// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
func (s *syncer) saveTXData(tx model.TX, isFullSync bool) error {
	filter := bson.M{
		"chain_id": tx.ChainID,
		"hash":     tx.Hash,
	}
	// 增量同步，对于不存在的数据则插入，对于已存在的数据则不做任何处理，因为链上的区块数据不会被修改
	if !isFullSync {
		dbTX, err := dao.DefaultTXDao.TX(filter)
		// 如果数据不存在则插入新的数据
		if err != nil || dbTX.ID.IsZero() {
			err = dao.DefaultTXDao.InsertTX(tx)
			if err != nil {
				return err
			}
			// 保存合约
			if tx.To == "" {
				err = s.saveContractData(tx, isFullSync)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
	update := bson.M{
		"$set": bson.M{
			"block_id":  tx.BlockID,
			"height":    tx.Height,
			"timestamp": tx.Timestamp,
			"from":      tx.From,
			"to":        tx.To,
			"gas_limit": tx.GasLimit,
			"gas_price": tx.GasPrice,
			"nonce":     tx.Nonce,
			"input":     tx.Input,
			"value":     tx.Value,
			"receipt":   tx.Receipt,
		},
	}
	updateOptions := options.Update()
	upsert := true
	updateOptions.Upsert = &upsert
	err := dao.DefaultTXDao.Update(filter, update, updateOptions)
	if err != nil {
		return err
	}

	// 保存合约
	if tx.To == "" {
		err = s.saveContractData(tx, isFullSync)
		if err != nil {
			return err
		}
	}
	return nil
}

// 保存 合约 数据
// 增量同步，对于不存在的数据则插入，对于已存在的数据则不做任何处理，因为链上的区块数据不会被修改
// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
func (s *syncer) saveContractData(tx model.TX, isFullSync bool) error {
	contract := (&tx).ToContract()
	if !isFullSync {
		filter := bson.M{
			"chain_id": contract.ChainID,
			"tx_hash":  contract.TxHash,
			"address":  contract.Address,
		}
		dbContract, err := dao.DefaultContractDao.Contract(filter)
		// 如果数据不存在则插入新的数据
		if err != nil || dbContract.ID.IsZero() {
			err = dao.DefaultContractDao.InsertContract(*contract)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// 全量同步，对于不存在的数据则插入，对于已存在的数据则更新
	filter := bson.M{
		"chain_id": contract.ChainID,
		"tx_hash":  contract.TxHash,
		"address":  contract.Address,
	}
	update := bson.M{
		"$set": bson.M{
			"creator":   contract.Creator,
			"content":   contract.Content,
			"timestamp": contract.Timestamp,
		},
	}
	updateOptions := options.Update()
	upsert := true
	updateOptions.Upsert = &upsert
	err := dao.DefaultContractDao.Update(filter, update, updateOptions)
	if err != nil {
		return err
	}
	return nil
}
