package service

import (
	"graces/exterr"
	"graces/model"
	"graces/util"
	"graces/web/dao"
	"graces/ws"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	DefaultWebsocketService IWebsocketService
)

func init() {
	DefaultWebsocketService = newWebsocketService()
}

func newWebsocketService() IWebsocketService {
	return &websocketService{
		dao:     dao.DefaultWSMsgDao,
		manager: ws.DefaultWebsocketManager,
	}
}

type websocketService struct {
	dao     dao.IWSMsgDao
	manager *ws.Manager
}

func (s *websocketService) Manager() model.WSManagerVO {
	vo := model.WSManagerVO{
		GroupLen:                int64(s.manager.LenGroup()),
		ClientLen:               int64(s.manager.LenClient()),
		ChanRegisterLen:         int64(len(s.manager.Register)),
		ChanUnregisterLen:       int64(len(s.manager.UnRegister)),
		ChanMessageLen:          int64(len(s.manager.Message)),
		ChanGroupMessageLen:     int64(len(s.manager.GroupMessage)),
		ChanBroadCastMessageLen: int64(len(s.manager.BroadCastMessage)),
	}
	return vo
}

func (s *websocketService) Group(name string) (model.WSGroupVO, error) {
	vo := model.WSGroupVO{}
	group, ok := s.manager.Group[name]
	if !ok {
		return vo, exterr.ErrWebsocketGroupInvalid
	}
	clients := make([]model.WSClientVO, 0)
	for k, v := range group {
		client := model.WSClientVO{}
		err := util.SimpleCopyProperties(&client, v)
		if err != nil {
			logrus.Errorln(err)
			return vo, exterr.ErrConvert
		}
		client.ID = k
		clients = append(clients, client)
	}
	vo.Name = name
	vo.Clients = clients

	return vo, nil
}

func (s *websocketService) Groups() ([]model.WSGroupVO, error) {
	vos := make([]model.WSGroupVO, 0)
	groups := s.manager.Group
	for gk, gv := range groups {
		group := model.WSGroupVO{}
		clients := make([]model.WSClientVO, 0)
		for k, v := range gv {
			client := model.WSClientVO{}
			err := util.SimpleCopyProperties(&client, v)
			if err != nil {
				logrus.Errorln(err)
				return vos, exterr.ErrConvert
			}
			client.ID = k
			clients = append(clients, client)
		}
		group.Name = gk
		group.Clients = clients
		vos = append(vos, group)
	}
	return vos, nil
}

func (s *websocketService) Send(dto model.WSMessageDTO) error {
	if !s.isGroupExist(dto.Group) {
		return exterr.ErrWebsocketGroupNotExist
	}
	if !s.isClientInGroup(dto.ID, dto.Group) {
		return exterr.ErrWebsocketClientNotInGroup
	}
	s.manager.Send(dto.ID, dto.Group, []byte(dto.Message))
	return nil
}

func (s *websocketService) SendGroup(dto model.WSGroupMessageDTO) error {
	if !s.isGroupExist(dto.Group) {
		return exterr.ErrWebsocketGroupNotExist
	}
	s.manager.SendGroup(dto.Group, []byte(dto.Message))
	return nil
}

func (s *websocketService) SendAll(dto model.WSBroadCastMessageDTO) {
	s.manager.SendAll([]byte(dto.Message))
}

func (s *websocketService) Dial(dto model.WSDialDTO) (model.WSClientVO, error) {
	vo := model.WSClientVO{}
	client, err := s.manager.Dial(dto.IP, dto.Port, dto.Path, dto.Group)
	if err != nil {
		if _, ok := err.(exterr.ExtError); ok {
			return vo, err
		}
		if _, ok := err.(*exterr.ExtError); ok {
			return vo, err
		}
		return vo, exterr.ErrWebsocketDial
	}
	err = util.SimpleCopyProperties(&vo, client)
	vo.ID = client.Id
	if err != nil {
		logrus.Errorln(err)
		return vo, exterr.ErrConvert
	}
	return vo, nil
}

func (s *websocketService) ClientSend(dto model.WSMessageDTO) error {
	if !s.isGroupExist(dto.Group) {
		return exterr.ErrWebsocketGroupNotExist
	}
	if !s.isClientInGroup(dto.ID, dto.Group) {
		return exterr.ErrWebsocketClientNotInGroup
	}
	client := s.manager.GetClient(dto.Group, dto.ID)
	if !client.IsDial {
		return exterr.ErrWebsocketClientSend
	}
	client.Message <- []byte(dto.Message)
	return nil
}

func (s *websocketService) SaveMessage(dto model.WSMsgDTO) error {
	msg := &model.WSMsg{}
	err := msg.ValueOfDTO(dto)
	if err != nil {
		return err
	}

	err = s.dao.InsertWSMsg(*msg)
	if err != nil {
		err = exterr.NewError(exterr.ErrCodeInsert, err.Error())
		return err
	}
	return nil
}

func (s *websocketService) UpdateMessageHash(msgID string, msgHash string) error {
	id, err := primitive.ObjectIDFromHex(msgID)
	if err != nil {
		return exterr.ErrObjectIDInvalid
	}
	filter := bson.M{
		"_id": id,
	}
	msg, err := s.dao.WSMsg(filter)
	if err != nil {
		return exterr.NewError(exterr.ErrCodeFind, err.Error())
	}
	extra := msg.Extra
	extra["hash"] = msgHash
	update := bson.M{
		"$set": bson.M{
			"extra": extra,
		},
	}
	return s.dao.UpdateWSMsg(filter, update)
}

// 判断组是否存在
func (s *websocketService) isGroupExist(group string) bool {
	_, ok := s.manager.Group[group]
	return ok
}

// 判断客户端是否存在
func (s *websocketService) isClientExist(clientId string) bool {
	groups := s.manager.Group
	for _, group := range groups {
		_, ok := group[clientId]
		if ok {
			return true
		}
	}
	return false
}

// 判断指定客户端是否在指定的组中
func (s *websocketService) isClientInGroup(clientId string, group string) bool {
	if !s.isGroupExist(group) {
		return false
	}
	g := s.manager.Group[group]
	_, ok := g[clientId]
	return ok
}
