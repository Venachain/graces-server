package syncer

import (
	"errors"
	"fmt"
	"graces/config"
	"graces/exterr"
	"graces/model"
	"graces/rpc"
	"graces/web/dao"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// GC 频率，次/30分钟
	gcInterval = 30 * 60
	// 同步已完成的数据至少存留的时间：2分钟
	keepTime = 2 * 60 * 1000

	// StatusPrepare 同步状态：准备同步
	StatusPrepare = "prepare"
	// StatusSyncing 同步状态：同步中
	StatusSyncing = "syncing"
	// StatusError 同步状态：同步出错
	StatusError = "error"
	// StatusSuccess 同步状态：同步成功
	StatusSuccess = "success"

	// ErrTypeNodeSync 节点数据同步错误类型
	ErrTypeNodeSync = 1
	// ErrTypeCNSSync CNS数据同步错误类型
	ErrTypeCNSSync = 2
	// ErrTypeBlockOrTXSync 区块数据或交易数据同步错误类型
	ErrTypeBlockOrTXSync = 3
)

var (
	DefaultChainDataSyncManager *chainDataSyncManager
)

func init() {
	DefaultChainDataSyncManager = newChainSyncManager()
	go DefaultChainDataSyncManager.errProcessAndGC()
}

func newChainSyncManager() *chainDataSyncManager {
	return &chainDataSyncManager{
		syncInfoContainer: make(map[string]*model.ChainDataSyncInfo),
		ErrChan:           make(chan *model.SyncErrMsg),
	}
}

type chainDataSyncManager struct {
	syncInfoContainer map[string]*model.ChainDataSyncInfo
	lock              sync.Mutex
	ErrChan           chan *model.SyncErrMsg
}

// ChainDataIncrSyncDelayStart 链数据循环增量同步延迟启动
func (manager *chainDataSyncManager) ChainDataIncrSyncDelayStart(delay time.Duration) {
	if delay < 0 {
		delay = 0
	}
	target := time.Until(time.Now().Add(delay * time.Second))
	timer := time.NewTimer(target)
	select {
	case <-timer.C:
		go manager.loopIncrSync()
		break
	}
}

// ChainDataIncrSyncStart 链数据循环增量同步立即启动
func (manager *chainDataSyncManager) ChainDataIncrSyncStart() {
	go manager.loopIncrSync()
}

// 增量循环同步
func (manager *chainDataSyncManager) loopIncrSync() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("unknown panic，loopIncrSync：%+v", err)
		}
	}()

	interval := config.Config.Syncer.IncrInterval * time.Second
	logrus.Infof("chain data increment synchronize [start], sync interval: [%v/once]", interval)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			filter := bson.M{}
			findOps := options.Find().SetProjection(bson.D{{"_id", 1}, {"name", 1}})
			chains, err := dao.DefaultChainDao.Chains(filter, findOps)
			if err != nil || len(chains) == 0 {
				logrus.Infof("no chains need to increment synchronize")
				continue
			}
			for _, chain := range chains {
				manager.IncrSyncStart(chain.ID.Hex(), true)
			}
		}
	}
}

// GetChainDataSyncInfo 获取同步信息
func (manager *chainDataSyncManager) GetChainDataSyncInfo(chainID string) (*model.ChainDataSyncInfo, bool) {
	info, ok := manager.syncInfoContainer[chainID]
	return info, ok
}

// PutChainDataSyncInfo 往同步管理器的容器中添加链的同步信息
func (manager *chainDataSyncManager) PutChainDataSyncInfo(info *model.ChainDataSyncInfo) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.syncInfoContainer[info.ChainID] = info
}

// IncrSyncStart 开始增量同步
func (manager *chainDataSyncManager) IncrSyncStart(chainID string, isAsync bool) {
	if isAsync {
		go manager.syncStart(chainID, false)
		return
	}
	manager.syncStart(chainID, false)
}

// FullSyncStart 开始全量同步
func (manager *chainDataSyncManager) FullSyncStart(chainID string, isAsync bool) {
	if isAsync {
		go manager.syncStart(chainID, true)
		return
	}
	manager.syncStart(chainID, true)
}

func (manager *chainDataSyncManager) syncStart(chainID string, isFullSync bool) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("unknown panic，syncStart：%+v", err)
		}
	}()
	syncType := "increment sync"
	if isFullSync {
		syncType = "full sync"
	}

	logrus.Debugf("chain[%s] data %s [start]", chainID, syncType)
	defer logrus.Debugf("chain[%s] data %s [end]", chainID, syncType)
	manager.syncProcess(chainID, isFullSync)
}

// 数据同步处理
func (manager *chainDataSyncManager) syncProcess(chainID string, isFullSync bool) {
	chainSyncInfo := manager.BuildChainSyncInfo(chainID)
	if chainSyncInfo.Status == StatusSyncing {
		logrus.Infof("this chain[%s] is syncing, don't repeat sync for it", chainID)
		return
	}

	chainSyncInfo.Status = StatusSyncing
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(3)

	// 同步节点
	go func(waitGroup *sync.WaitGroup) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("unknown panic，SyncNode：%+v", err)
			}
			waitGroup.Done()
		}()
		err := manager.SyncNode(chainID, isFullSync)
		if err != nil {
			chainSyncInfo.NodeDataSyncInfo.ErrMsg = err.Error()
			chainSyncInfo.NodeDataSyncInfo.Status = StatusError
			DefaultChainDataSyncManager.ErrChan <- &model.SyncErrMsg{
				ChainID: chainID,
				ErrType: ErrTypeNodeSync,
				Err:     err,
			}
			return
		}
	}(waitGroup)

	// 同步 CNS
	go func(waitGroup *sync.WaitGroup) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("unknown panic，SyncCNS：%+v", err)
			}
			waitGroup.Done()
		}()
		err := manager.SyncCNS(chainID, isFullSync)
		if err != nil {
			chainSyncInfo.CNSDataSyncInfo.ErrMsg = err.Error()
			chainSyncInfo.CNSDataSyncInfo.Status = StatusError
			DefaultChainDataSyncManager.ErrChan <- &model.SyncErrMsg{
				ChainID: chainID,
				ErrType: ErrTypeCNSSync,
				Err:     err,
			}
			return
		}
	}(waitGroup)

	// 同步区块和区块内的交易
	go func(waitGroup *sync.WaitGroup) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("unknown panic，SyncBlockAndTX：%+v", err)
			}
			waitGroup.Done()
		}()
		err := manager.SyncBlockAndTX(chainID, isFullSync)
		if err != nil {
			chainSyncInfo.BlockDataSyncInfo.ErrMsg = err.Error()
			chainSyncInfo.BlockDataSyncInfo.Status = StatusError
			DefaultChainDataSyncManager.ErrChan <- &model.SyncErrMsg{
				ChainID: chainID,
				ErrType: ErrTypeBlockOrTXSync,
				Err:     err,
			}
			return
		}
	}(waitGroup)

	waitGroup.Wait()

	if chainSyncInfo.NodeDataSyncInfo.Status == StatusSuccess && chainSyncInfo.CNSDataSyncInfo.Status == StatusSuccess &&
		chainSyncInfo.BlockDataSyncInfo.Status == StatusSuccess {
		chainSyncInfo.Status = StatusSuccess
		logrus.Infof("chain[%s] data sync success", chainID)
	}
	return
}

// BuildChainSyncInfo 构建链数据同步信息
func (manager *chainDataSyncManager) BuildChainSyncInfo(chainID string) *model.ChainDataSyncInfo {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	chainSyncInfo, ok := manager.GetChainDataSyncInfo(chainID)
	if !ok {
		chainSyncInfo = &model.ChainDataSyncInfo{
			ChainID:   chainID,
			Status:    StatusPrepare,
			StartTime: time.Now().Unix(),
		}
		manager.syncInfoContainer[chainID] = chainSyncInfo
	}
	return chainSyncInfo
}

// SyncBlockAndTX 同步区块和交易
func (manager *chainDataSyncManager) SyncBlockAndTX(chainID string, isFullSync bool) error {
	logrus.Debugf("chain[%v] block data and tx data sync [start]", chainID)
	defer logrus.Debugf("chain[%v] block data and tx data sync [end]", chainID)
	chainSyncInfo, ok := manager.GetChainDataSyncInfo(chainID)
	if !ok {
		return fmt.Errorf("chain[%v] block data and tx data sync fail：chainSyncInfo is nil", chainID)
	}
	blockSyncInfo := chainSyncInfo.BlockDataSyncInfo
	if blockSyncInfo == nil {
		blockSyncInfo = &model.BlockDataSyncInfo{
			LatestHeight:         0,
			CurrentHeight:        0,
			Status:               StatusPrepare,
			StartTime:            time.Now().Unix(),
			BlockSyncTimeAvg:     0,
			EstimateCompleteTime: time.Now().Unix(),
			ErrMsg:               "",
		}
		chainSyncInfo.BlockDataSyncInfo = blockSyncInfo
	}
	blockSyncInfo.Status = StatusSyncing
	// 全量同步从块高为 0 开始进行同步
	if !isFullSync {
		blockSyncInfo.CurrentHeight = 0
		// 非全量同步，即增量同步才数据库中查询块高最高的区块
		dbBLock, _ := rpc.GetLatestBlockFromDB(chainID)
		if dbBLock != nil {
			blockSyncInfo.CurrentHeight = dbBLock.Height
		}
	}
	latestBlock, err := rpc.GetLatestBlockFromChain(chainID)
	if err != nil {
		return exterr.NewError(exterr.ErrCodeChainDataSync, err)
	}
	blockSyncInfo.LatestHeight = latestBlock.NumberU64()
	// 如果数据库的区块高度和链上的高度相同，则无需同步
	if blockSyncInfo.CurrentHeight == blockSyncInfo.LatestHeight {
		blockSyncInfo.Status = StatusSuccess
		return nil
	}
	err = manager.syncBlockBySyncInfo(chainID, blockSyncInfo, isFullSync)
	if err != nil {
		return err
	}
	blockSyncInfo.Status = StatusSuccess
	logrus.Infof("chain[%v] block data and tx data sync success", chainID)
	return nil
}

// SyncCNS 同步 cns
func (manager *chainDataSyncManager) SyncCNS(chainID string, isFullSync bool) error {
	logrus.Debugf("chain[%v] cns data sync [start]", chainID)
	defer logrus.Debugf("chain[%v] cns data sync [end]", chainID)
	chainSyncInfo, ok := manager.GetChainDataSyncInfo(chainID)
	if !ok {
		return fmt.Errorf("chain[%v] cns data sync fail：chainSyncInfo is nil", chainID)
	}
	sncDataSyncInfo := chainSyncInfo.CNSDataSyncInfo
	if sncDataSyncInfo == nil {
		sncDataSyncInfo = &model.CNSDataSyncInfo{
			Size:                 0,
			Index:                0,
			Status:               StatusPrepare,
			StartTime:            time.Now().Unix(),
			SyncTimeAvg:          0,
			EstimateCompleteTime: time.Now().Unix(),
			ErrMsg:               "",
		}
		chainSyncInfo.CNSDataSyncInfo = sncDataSyncInfo
	}
	sncDataSyncInfo.Status = StatusSyncing
	allCNS, err := rpc.GetAllCNS(chainID)
	if err != nil {
		sncDataSyncInfo.ErrMsg = err.Error()
		return err
	}
	sncDataSyncInfo.Size = len(allCNS)
	for i, cns := range allCNS {
		sncDataSyncInfo.Index = i + 1
		err = DefaultSyncer.saveCNS(*cns, isFullSync)
		if err != nil {
			return err
		}
		// 计算已经消耗的时间
		timeConsume := time.Now().Unix() - sncDataSyncInfo.StartTime
		// 计算同步每个区块需要的平均时间
		t := int64(i)
		if t == 0 {
			sncDataSyncInfo.SyncTimeAvg = timeConsume
		} else {
			sncDataSyncInfo.SyncTimeAvg = timeConsume / t
		}
		// 更新预计完成时间
		sncDataSyncInfo.EstimateCompleteTime = time.Now().Unix() + int64(sncDataSyncInfo.Size-sncDataSyncInfo.Index)*sncDataSyncInfo.SyncTimeAvg
		manager.setEstimateCompleteTime(chainID, sncDataSyncInfo.EstimateCompleteTime)
	}
	sncDataSyncInfo.Status = StatusSuccess
	logrus.Infof("chain[%v] cns data sync success", chainID)
	return nil
}

// SyncNode 节点同步
func (manager *chainDataSyncManager) SyncNode(chainID string, isFullSync bool) error {
	logrus.Debugf("chain[%v] node data sync [start]", chainID)
	defer logrus.Debugf("chain[%v] node data sync [end]", chainID)
	chainSyncInfo, ok := manager.GetChainDataSyncInfo(chainID)
	if !ok {
		return fmt.Errorf("chain[%v] node data sync fail：chainSyncInfo is nil", chainID)
	}
	nodeDataSyncInfo := chainSyncInfo.NodeDataSyncInfo
	if nodeDataSyncInfo == nil {
		nodeDataSyncInfo = &model.NodeDataSyncInfo{
			Size:                 0,
			Index:                0,
			Status:               StatusPrepare,
			StartTime:            time.Now().Unix(),
			SyncTimeAvg:          0,
			EstimateCompleteTime: time.Now().Unix(),
			ErrMsg:               "",
		}
		chainSyncInfo.NodeDataSyncInfo = nodeDataSyncInfo
	}
	chainSyncInfo.Status = StatusSyncing
	allNodes, err := rpc.GetAllNodes(chainID)
	if err != nil {
		return err
	}
	nodeDataSyncInfo.Size = len(allNodes)
	for i, node := range allNodes {
		nodeDataSyncInfo.Index = i + 1
		err = DefaultSyncer.saveNode(*node, isFullSync)
		if err != nil {
			return err
		}
		// 计算已经消耗的时间
		timeConsume := time.Now().Unix() - nodeDataSyncInfo.StartTime
		// 计算同步每个节点需要的平均时间
		t := int64(i)
		if t == 0 {
			nodeDataSyncInfo.SyncTimeAvg = timeConsume
		} else {
			nodeDataSyncInfo.SyncTimeAvg = timeConsume / t
		}
		// 更新预计完成时间
		nodeDataSyncInfo.EstimateCompleteTime = time.Now().Unix() + int64(nodeDataSyncInfo.Size-nodeDataSyncInfo.Index)*nodeDataSyncInfo.SyncTimeAvg
		manager.setEstimateCompleteTime(chainID, nodeDataSyncInfo.EstimateCompleteTime)
	}
	nodeDataSyncInfo.Status = StatusSuccess
	logrus.Infof("chain[%v] node data sync success", chainID)
	return nil
}

// 同步区块和交易
func (manager *chainDataSyncManager) syncBlockBySyncInfo(chainID string, blockSyncInfo *model.BlockDataSyncInfo, isFullSync bool) error {
	if blockSyncInfo == nil {
		return errors.New("blockSyncInfo must not be nil")
	}
	// TODO：需要处理可能因为中间数据同步出错而导致的区块数据不全问题
	startHeight := blockSyncInfo.CurrentHeight
	for ; blockSyncInfo.CurrentHeight <= blockSyncInfo.LatestHeight; blockSyncInfo.CurrentHeight++ {
		err := DefaultSyncer.syncBlockByNumber(chainID, int64(blockSyncInfo.CurrentHeight), isFullSync)
		if err != nil {
			blockSyncInfo.ErrMsg = err.Error()
			return err
		}
		// 计算已经消耗的时间
		timeConsume := time.Now().Unix() - blockSyncInfo.StartTime
		// 计算同步每个区块需要的平均时间
		t := int64(blockSyncInfo.CurrentHeight - startHeight)
		if t == 0 {
			blockSyncInfo.BlockSyncTimeAvg = timeConsume
		} else {
			blockSyncInfo.BlockSyncTimeAvg = timeConsume / t
		}
		// 更新预计完成时间
		blockSyncInfo.EstimateCompleteTime = time.Now().Unix() + int64(blockSyncInfo.LatestHeight-blockSyncInfo.CurrentHeight)*blockSyncInfo.BlockSyncTimeAvg
		manager.setEstimateCompleteTime(chainID, blockSyncInfo.EstimateCompleteTime)
	}
	// 正常执行完成时，当前高度会比最新区块的高度大 1
	blockSyncInfo.CurrentHeight = blockSyncInfo.LatestHeight
	return nil
}

// 设置预计完成时间
func (manager *chainDataSyncManager) setEstimateCompleteTime(chainID string, estimateCompleteTime int64) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.syncInfoContainer[chainID].EstimateCompleteTime = manager.max(manager.syncInfoContainer[chainID].EstimateCompleteTime, estimateCompleteTime)
}

func (manager *chainDataSyncManager) max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// 数据同步错误处理和清理已经完成同步的记录信息
func (manager *chainDataSyncManager) errProcessAndGC() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("unknown panic，errProcessAndGC：%+v", err)
		}
	}()

	ticker := time.NewTicker(gcInterval * time.Second)
	for {
		select {
		case syncErrMsg := <-manager.ErrChan:
			logrus.Errorf("chain[%s] data sync fail: %+v", syncErrMsg.ChainID, syncErrMsg.Err)
			manager.errProcess(syncErrMsg)
		case <-ticker.C:
			manager.gc()
		}
	}
}

// 清理已经完成同步的记录信息，回收内存
func (manager *chainDataSyncManager) gc() {
	deleteKey := make([]string, 0)
	for key, chainSyncInfo := range manager.syncInfoContainer {
		// 数据同步已完成，且超过保存时间才删除
		if chainSyncInfo.Status == StatusSuccess && chainSyncInfo.EstimateCompleteTime+keepTime <= time.Now().Unix() {
			deleteKey = append(deleteKey, key)
		}
	}
	manager.lock.Lock()
	for _, key := range deleteKey {
		delete(manager.syncInfoContainer, key)
		logrus.Infof("ChainInfoSyncManager GC: delete completed syncInfo [%v]", key)
	}
	manager.lock.Unlock()
}

// 错误处理
func (manager *chainDataSyncManager) errProcess(syncErrMsg *model.SyncErrMsg) {
	syncInfo, ok := manager.GetChainDataSyncInfo(syncErrMsg.ChainID)
	if !ok {
		logrus.Infof("no sync error msg need to process for chain[%s]", syncErrMsg.ChainID)
		return
	}
	manager.lock.Lock()
	defer manager.lock.Unlock()
	syncInfo.Status = StatusError
	syncInfo.ErrMsg = syncErrMsg.Err.Error()
	//	TODO sync retries?
	switch syncErrMsg.ErrType {
	case ErrTypeNodeSync:
		logrus.Errorf("chain[%s] node data sync fail: %+v", syncInfo.ChainID, syncErrMsg.Err)
		syncInfo.NodeDataSyncInfo.Status = StatusError
		syncInfo.NodeDataSyncInfo.ErrMsg = syncErrMsg.Err.Error()
	case ErrTypeCNSSync:
		logrus.Errorf("chain[%s] CNS data sync fail: %+v", syncInfo.ChainID, syncErrMsg.Err)
		syncInfo.CNSDataSyncInfo.Status = StatusError
		syncInfo.CNSDataSyncInfo.ErrMsg = syncErrMsg.Err.Error()
	case ErrTypeBlockOrTXSync:
		logrus.Errorf("chain[%s] block data or tx data sync fail: %+v", syncInfo.ChainID, syncErrMsg.Err)
		syncInfo.BlockDataSyncInfo.Status = StatusError
		syncInfo.BlockDataSyncInfo.ErrMsg = syncErrMsg.Err.Error()
	}
}
