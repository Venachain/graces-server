package ws

import (
	"PlatONE-Graces/config"
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"PlatONE-Graces/rpc"
	"PlatONE-Graces/util"
	"PlatONE-Graces/web/dao"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func NewSubMsgProcessor() *SubMsgProcessor {
	return &SubMsgProcessor{}
}

// SubMsgProcessor 订阅消息处理
type SubMsgProcessor struct{}

func (s *SubMsgProcessor) Process(ctx *MsgProcessorContext, msg interface{}) error {
	logrus.Debugf("websocket subscription message process [start]:\n%+v", msg)
	defer logrus.Debugln("websocket subscription message process [end]")

	data, ok := msg.(map[string]interface{})
	if !ok {
		errStr := fmt.Sprintf("can't process unknown subMsg:\n %+v", msg)
		return errors.New(errStr)
	}
	params, ok := data["params"].(map[string]interface{})
	if !ok {
		errStr := fmt.Sprintf("subMsg without [params] property cannot be processed:\n [%+v]", data)
		return errors.New(errStr)
	}
	subHash, ok := params["subscription"].(string)
	if !ok {
		errStr := fmt.Sprintf("subMsg without [subscription] property cannot be processed:\n [%+v]", data)
		return errors.New(errStr)
	}
	// 通过哈希拿到存储的消息数据
	wsMsg, err := dao.DefaultWSMsgDao.WSMsg(bson.M{"hash": subHash})
	if err != nil {
		return err
	}

	topic, ok := wsMsg.Extra["topic"].(map[string]interface{})
	if !ok {
		errStr := fmt.Sprintf("subMsg without [topic] property cannot be processed:\n [%+v]", data)
		return errors.New(errStr)
	}

	topicName, ok := topic["name"].(string)
	if !ok {
		errStr := fmt.Sprintf("subMsg without [topic.name] property cannot be processed:\n [%+v]", data)
		return errors.New(errStr)
	}
	chainID := wsMsg.ChainID.Hex()
	err = s.process(ctx, topicName, chainID, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *SubMsgProcessor) process(ctx *MsgProcessorContext, topicName string, chainID string, params map[string]interface{}) error {
	methodName := util.MethodCapitalized(topicName)
	reType := reflect.TypeOf(s)
	method, ok := reType.MethodByName(methodName)
	if !ok {
		return errors.New(fmt.Sprintf("no process method for topic[%v]", topicName))
	}
	methodParams := make([]reflect.Value, 4)
	// 第一个参数为方法的持有者
	methodParams[0] = reflect.ValueOf(s)
	methodParams[1] = reflect.ValueOf(ctx.client)
	methodParams[2] = reflect.ValueOf(chainID)
	methodParams[3] = reflect.ValueOf(params)
	resValues := method.Func.Call(methodParams)
	if len(resValues) > 0 {
		if err, ok := resValues[len(resValues)-1].Interface().(error); ok {
			return err
		}
	}
	return nil
}

// NewHeads newHeads topic 的订阅数据处理方法，方法名称需要与 topic 保持一致，但首字母需要是大写的
func (s *SubMsgProcessor) NewHeads(client *Client, chainID string, params map[string]interface{}) error {
	// 拿到区块头数据
	result, ok := params["result"].(map[string]interface{})
	if !ok {
		errStr := fmt.Sprintf("subMsg params without result property cannot be processed:\n [%+v]", params)
		return exterr.NewError(exterr.ErrCodeWebsocketSubMsgProcess, errStr)
	}
	blockHash, ok := result["hash"].(string)
	if !ok {
		errStr := fmt.Sprintf("subMsg params without result.hash property cannot be processed:\n [%+v]", params)
		return exterr.NewError(exterr.ErrCodeWebsocketSubMsgProcess, errStr)
	}

	logrus.Debugf("block[%s] sync start...", blockHash)
	defer logrus.Debugf("block[%s] sync completed", blockHash)

	head, err := rpc.GetBlockHeadByHash(chainID, blockHash)
	if err != nil {
		return err
	}
	block, err := rpc.GetBlockByHash(chainID, blockHash)
	if err != nil {
		return err
	}
	block.Head = head
	err = dao.DefaultBlockDao.InsertBlock(*block)
	if err != nil {
		return err
	}
	txs, err := rpc.GetTXDataByBlockHash(chainID, block.ID.Hex(), blockHash)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		// 保存合约
		if tx.To == "" {
			contract := tx.ToContract()
			err = dao.DefaultContractDao.InsertContract(*contract)
			if err != nil {
				return err
			}
		}

		err = dao.DefaultTXDao.InsertTX(*tx)
		if err != nil {
			logrus.Errorln(err)
		}
	}
	logrus.Infof("sync success block[%v][%v]", block.Height, block.Hash)

	// 把从链上接收到的新数据转发到订阅该事件的 ws 前端客户端
	err = s.forwardBlock(chainID, block)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		err = s.forwardTX(chainID, tx)
		if err != nil {
			return err
		}
	}
	stats := s.getStat(chainID)
	err = s.forwardStats(chainID, stats)
	if err != nil {
		return err
	}

	nodes, err := s.getNodes(chainID)
	if err != nil {
		return err
	}
	err = s.forwardNodes(chainID, nodes)
	if err != nil {
		return err
	}

	logrus.Infof("forward success block[%v][%v]", block.Height, block.Hash)
	return nil
}

func (s *SubMsgProcessor) forwardBlock(group string, block *model.Block) error {
	if block == nil {
		errStr := "can not to forward nil block"
		return exterr.NewError(exterr.ErrCodeWebsocketSubMsgProcess, errStr)
	}
	vo, err := block.ToVO()
	if err != nil {
		return err
	}
	dto := model.WSSubMsgDTO{
		ID:      group,
		Type:    config.Config.WSConf.WsMsgTypesConf.Pub.BlockType,
		Content: vo,
	}
	// 对该组的客户端进行广播
	err = s.Forward("", group, dto)
	if err != nil {
		return err
	}
	return nil
}

func (s *SubMsgProcessor) forwardTX(group string, tx *model.TX) error {
	if tx == nil {
		errStr := "can not to forward nil tx"
		return exterr.NewError(exterr.ErrCodeWebsocketSubMsgProcess, errStr)
	}
	vo, err := tx.ToVO()
	if err != nil {
		return err
	}
	dto := model.WSSubMsgDTO{
		ID:      group,
		Type:    config.Config.WSConf.WsMsgTypesConf.Pub.TXType,
		Content: vo,
	}
	// 对该组的客户端进行广播
	err = s.Forward("", group, dto)
	if err != nil {
		return err
	}
	return nil
}

func (s *SubMsgProcessor) forwardStats(group string, stats *model.StatsVO) error {
	if stats == nil {
		errStr := "can not to forward nil stats"
		return exterr.NewError(exterr.ErrCodeWebsocketSubMsgProcess, errStr)
	}
	dto := model.WSSubMsgDTO{
		ID:      group,
		Type:    config.Config.WSConf.WsMsgTypesConf.Pub.StatsType,
		Content: stats,
	}
	// 对该组的客户端进行广播
	err := s.Forward("", group, dto)
	if err != nil {
		return err
	}
	return nil
}

func (s *SubMsgProcessor) forwardNodes(group string, nodes []*model.NodeVO) error {
	if nodes == nil {
		errStr := "can not to forward nil nodes"
		return exterr.NewError(exterr.ErrCodeWebsocketSubMsgProcess, errStr)
	}
	dto := model.WSSubMsgDTO{
		ID:      group,
		Type:    config.Config.WSConf.WsMsgTypesConf.Pub.NodeInfoType,
		Content: nodes,
	}
	// 对该组的客户端进行广播
	err := s.Forward("", group, dto)
	if err != nil {
		return err
	}
	return nil
}

// Forward 把从链上接收到的事件数据转发到前端
func (s *SubMsgProcessor) Forward(clientID string, group string, dto model.WSSubMsgDTO) error {
	if group == "" && clientID == "" {
		errStr := "client is nil, can not to forward"
		return exterr.NewError(exterr.ErrCodeWebsocketSubMsgProcess, errStr)
	}
	jsonMsg, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	if clientID == "" {
		DefaultWebsocketManager.SendGroup(group, jsonMsg)
		return nil
	}
	DefaultWebsocketManager.Send(clientID, group, jsonMsg)
	return nil
}

// 获取统计数据
func (s *SubMsgProcessor) getStat(chainID string) *model.StatsVO {
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil
	}

	var result model.StatsVO
	filter := bson.M{"chain_id": cid}
	findOps := options.Count()
	txCnt, _ := dao.DefaultTXDao.Count(filter, findOps)
	contractCnt, _ := dao.DefaultContractDao.Count(filter, findOps)
	nodeCnt, _ := dao.DefaultNodeDao.Count(filter, findOps)
	latestBlock, _ := dao.DefaultBlockDao.LatestBlock(cid)

	result.TotalTx = txCnt
	result.TotalContract = contractCnt
	result.TotalNode = nodeCnt
	result.LatestBlock = latestBlock.Height

	return &result
}

func (s *SubMsgProcessor) getNodes(chainID string) ([]*model.NodeVO, error) {
	cid, err := primitive.ObjectIDFromHex(chainID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"chain_id": cid}
	findOps := util.BuildOptionsByQuery(1, 10)
	nodes, err := dao.DefaultNodeDao.Nodes(filter, findOps)
	if err != nil {
		return nil, err
	}
	var vos []*model.NodeVO
	for _, node := range nodes {
		vo, err := node.ToVO()
		if err != nil {
			return nil, err
		}
		endpoint := fmt.Sprintf("http://%v:%v", node.ExternalIP, node.RPCPort)
		blockNumber, err := rpc.GetBlockNumber(endpoint)
		if err != nil {
			return nil, err
		}
		vo.Blocknumber = uint32(blockNumber)
		vo.IsAlive = true
		vos = append(vos, vo)
	}
	return vos, nil
}
