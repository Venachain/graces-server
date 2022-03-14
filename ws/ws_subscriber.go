package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"graces/config"
	"graces/model"
	"graces/util"
	"graces/web/dao"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// DefaultWSSubscriber 默认的 websocket 订阅器
	DefaultWSSubscriber *wsSubscriber
)

func init() {
	logrus.Debugf("DefaultWSSubscriber init [start]")
	DefaultWSSubscriber = newWSSubscriber()
	logrus.Debugf("DefaultWSSubscriber init [end]")
}

type wsSubscriber struct {
	chainDao   dao.IChainDao
	wsDao      dao.IWSMsgDao
	wsManager  *Manager
	chains     map[string]*model.Chain
	maxChainID string
}

func newWSSubscriber() *wsSubscriber {
	return &wsSubscriber{
		chainDao:  dao.DefaultChainDao,
		wsDao:     dao.DefaultWSMsgDao,
		wsManager: DefaultWebsocketManager,
	}
}

// ChainWSTopicAutoSubDelayStart 延迟启动
func (s *wsSubscriber) ChainWSTopicAutoSubDelayStart(delay time.Duration) {
	if delay < 0 {
		delay = 0
	}
	target := time.Until(time.Now().Add(delay * time.Second))
	timer := time.NewTimer(target)
	select {
	case <-timer.C:
		s.ChainWSTopicAutoSubStart()
		break
	}
}

// ChainWSTopicAutoSubStart 立即启动
func (s *wsSubscriber) ChainWSTopicAutoSubStart() {
	logrus.Debugf("websocket subscribe [strat]")
	defer logrus.Debugf("websocket subscribe [end]")
	s.initChains()
	err := s.subTopicsForEveryChain()
	if err != nil {
		logrus.Errorf("sycer subNewHeads error: %+v", err)
	}
	logrus.Info("chain websocket topic auto subscribe success")
}

func (s *wsSubscriber) loadChainsFromDB() ([]*model.Chain, error) {
	filter := bson.M{}
	findOps := options.Find()
	// 按 _id 逆序
	findOps.SetSort(bson.D{{"_id", -1}})
	chains, err := s.chainDao.Chains(filter, findOps)

	if err != nil {
		return nil, err
	}
	logrus.Debugf("load chains from DB: %+v", chains)
	return chains, nil
}

func (s *wsSubscriber) initChains() {
	logrus.Debugf("load chains [start]")
	defer logrus.Debugf("load chains [end]")
	chains, err := s.loadChainsFromDB()
	if err != nil {
		logrus.Errorf("syncer init error: %+v", err)
		return
	}
	chainMap := make(map[string]*model.Chain, len(chains))
	for _, chain := range chains {
		chainMap[chain.ID.Hex()] = chain
	}
	s.chains = chainMap
	if len(chains) > 0 {
		s.maxChainID = chains[0].ID.Hex()
	}
	logrus.Infof("load chains success")
}

func (s *wsSubscriber) subTopicsForEveryChain() error {
	logrus.Debugf("subscribe topics from websocket for every chain [start]")
	defer logrus.Debugf("subscribe topics from websocket for every chain [end]")
	if len(s.chains) == 0 {
		logrus.Infof("no chains need to subscribe topics from websocket")
	}
	for _, chain := range s.chains {
		err := s.SubTopicsForChain(chain)
		if err != nil {
			logrus.Errorln(err)
		}
	}
	logrus.Info("subscribe topics from websocket for every chain success")
	return nil
}

// SubTopicsForChain 为指定的链订阅它所配置的所有 websocket topics
func (s *wsSubscriber) SubTopicsForChain(chain *model.Chain) error {
	if chain == nil {
		return errors.New("can't subscribe topics for nil chain")
	}

	// 1、获取 websocket 客户端连接
	client, err := s.getWSClientByChain(*chain)
	if err != nil {
		return err
	}

	// 2、提取当前链订阅的 topics 配置信息
	topics, ok := chain.ChainConfig["ws"].(map[string]interface{})["topics"].(map[string]interface{})
	logrus.Debug(topics)
	if !ok {
		msg := fmt.Sprintf("chain[%s][%s:%v] lost websocket subscription config, can't subscribe websocket topics for it",
			chain.Name, chain.IP, chain.RPCPort)
		return errors.New(msg)
	}

	// 3 处理 topics 订阅
	for k, topic := range topics {
		if !ok {
			logrus.Warningf("topic[%v] lost config, can't subscribe it from websocket", k)
			continue
		}
		topicName := topic.(map[string]interface{})["name"]
		topicParams := topic.(map[string]interface{})["params"]
		err := s.wsSubTopicProcessor(*chain, client, topicName.(string), topicParams.(string))
		if err != nil {
			logrus.Warningln(err)
		}
	}
	return nil
}

// topic 订阅处理器
func (s *wsSubscriber) wsSubTopicProcessor(chain model.Chain, client *Client, topic string, params string) error {
	// 获取到该链所配置订阅的所有 topic
	topics, err := s.getWSTopicsByChain(chain)
	if err != nil {
		return err
	}
	exist := false
	for _, v := range topics {
		if topic == v {
			exist = true
			break
		}
	}
	if !exist {
		return errors.New(fmt.Sprintf("unknown websocket subscription topic: %v", topic))
	}
	methodName := util.MethodCapitalized(topic)
	// 通过反射进行方法调用
	reType := reflect.TypeOf(s)
	method, ok := reType.MethodByName(methodName)
	if !ok {
		return errors.New(fmt.Sprintf("no process method for topic[%v]", topic))
	}
	methodParams := make([]reflect.Value, 5)
	// 第一个参数为方法的持有者
	methodParams[0] = reflect.ValueOf(s)
	methodParams[1] = reflect.ValueOf(chain)
	methodParams[2] = reflect.ValueOf(client)
	methodParams[3] = reflect.ValueOf(topic)
	methodParams[4] = reflect.ValueOf(params)
	resValues := method.Func.Call(methodParams)
	if len(resValues) > 0 {
		if err, ok := resValues[len(resValues)-1].Interface().(error); ok {
			return err
		}
	}
	return nil
}

// NewHeads newHeads 事件的订阅处理
func (s *wsSubscriber) NewHeads(chain model.Chain, client *Client, topic string, params string) error {
	logrus.Debugf("subscribe topic[newHead] from websocket for chain[%v] [start]", chain.Name)
	defer logrus.Debugf("subscribe topic[newHead] from websocket for chain[%v] [end]", chain.Name)
	// 1、处理参数信息
	msgID, paramsStr, err := s.wsSubParamsProcess(topic, params)
	if err != nil {
		return err
	}

	// 2、将订阅消息保存入库
	extra := make(map[string]interface{})
	topicMap := make(map[string]interface{})
	topicMap["name"] = topic
	topicMap["params"] = paramsStr
	extra["topic"] = topicMap
	msgDTO := model.WSMsgDTO{
		ID:      msgID,
		ChainID: chain.ID.Hex(),
		Type:    config.Config.WSConf.WsMsgTypesConf.Sub.SendType,
		Message: paramsStr,
		Extra:   extra,
	}
	wsMsg := &model.WSMsg{}
	err = wsMsg.ValueOfDTO(msgDTO)
	if err != nil {
		return err
	}
	err = s.wsDao.InsertWSMsg(*wsMsg)
	if err != nil {
		return err
	}

	// 3、给服务端发送订阅消息对指定 topic 进行订阅
	dto := model.WSMessageDTO{
		ID:      client.Id,
		Group:   client.Group,
		Message: paramsStr,
	}
	s.wsManager.Send(dto.ID, dto.Group, []byte(dto.Message))
	logrus.Infof("subscribe topic[newHead] from websocket for chain[%v] [success]", chain.Name)
	return nil
}

// 处理 websocket 事件订阅的参数
func (s *wsSubscriber) wsSubParamsProcess(topic string, params string) (string, string, error) {
	data := make(map[string]interface{})
	id := primitive.NewObjectID().Hex()
	err := json.Unmarshal([]byte(params), &data)
	if err != nil {
		logrus.Errorf("websocket params unmarshal error: %v", err)
		return "", "", err
	}
	msgType := data["id"]
	// type topic id
	data["id"] = fmt.Sprintf("%v %v %v", msgType, topic, id)

	p, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("websocket params marshal error: %v", err)
		return "", "", err
	}
	params = string(p)
	return id, params, nil
}

func (s *wsSubscriber) getWSClientByChain(chain model.Chain) (*Client, error) {
	_, ok := chain.ChainConfig["ws"].(map[string]interface{})
	if !ok {
		msg := fmt.Sprintf("chain[%s][%s:%v] lost websocket config, can't sync data from websocket for it",
			chain.Name, chain.IP, chain.RPCPort)
		return nil, errors.New(msg)
	}
	port := chain.WSPort
	path := ""
	group := chain.Name
	client, err := s.wsManager.Dial(chain.IP, int64(port), path, group)
	url := fmt.Sprintf("ws://%s:%v", chain.IP, port)
	if err != nil {
		msg := fmt.Sprintf("chain[%s][%s:%v] websocket dial [%s] error: %v",
			chain.Name, chain.IP, chain.RPCPort, url, err)
		return nil, errors.New(msg)
	}
	logrus.Debugf("chain[%s][%s:%v] websocket dial [%s] success",
		chain.Name, chain.IP, chain.RPCPort, url)
	return client, nil
}

func (s *wsSubscriber) getWSTopicsByChain(chain model.Chain) ([]string, error) {
	ws, ok := chain.ChainConfig["ws"].(map[string]interface{})
	if !ok {
		msg := fmt.Sprintf("chain[%s][%s:%v] without [ws] config, can't subscribe topics for it",
			chain.Name, chain.IP, chain.RPCPort)
		return nil, errors.New(msg)
	}
	topics, ok := ws["topics"].(map[string]interface{})
	if !ok {
		msg := fmt.Sprintf("chain[%s][%s:%v] without websocket [ws.topics] config, can't subscribe topics for it",
			chain.Name, chain.IP, chain.RPCPort)
		return nil, errors.New(msg)
	}
	topicNames := make([]string, 0)
	for _, topic := range topics {
		topicMap, ok := topic.(map[string]interface{})
		if !ok {
			logrus.Warningf("the topic isn't a map type, can't analysis it:\n%+v", topic)
			continue
		}
		name, ok := topicMap["name"].(string)
		if !ok {
			logrus.Warningf("the topic name[%+v] isn't a string type, can't process it", name)
		}
		topicNames = append(topicNames, name)
	}
	return topicNames, nil
}
