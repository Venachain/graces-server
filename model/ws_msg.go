package model

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/util"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WSMsg struct {
	ID         primitive.ObjectID     `json:"id" bson:"_id"`
	ChainID    primitive.ObjectID     `json:"chain_id" bson:"chain_id"`
	Type       string                 `json:"type" bson:"type"`
	Message    string                 `json:"message" bson:"message"`
	Hash       string                 `json:"hash" bson:"hash"`
	Extra      map[string]interface{} `json:"extra" bson:"extra"`
	CreateTime int64                  `json:"create_time" bson:"create_time"`
	UpdateTime int64                  `json:"update_time" bson:"update_time"`
	DeleteTime int64                  `json:"delete_time" bson:"delete_time"`
}

// WSMessageDTO websocket 单个客户端发送数据信息
type WSMessageDTO struct {
	// websocket 客户端连接 ID
	ID string `json:"id" binding:"required"`
	// websocket 客户端连接所在的分组
	Group string `json:"group" binding:"required"`
	// 要发送的消息内容
	Message string `json:"message" binding:"required"`
}

// WSGroupMessageDTO websocket 组客户端广播数据信息
type WSGroupMessageDTO struct {
	// websocket 客户端连接所在的分组
	Group string `json:"group" binding:"required"`
	// 要发送的消息内容
	Message string `json:"message" binding:"required"`
}

// WSBroadCastMessageDTO websocket 所有客户端广播数据信息
type WSBroadCastMessageDTO struct {
	// 要发送的消息内容
	Message string `json:"message" binding:"required"`
}

type WSDialDTO struct {
	// websocket 服务端 IP 地址
	IP string `json:"ip" binding:"required,min=7,max=15"`
	// websocket 服务端端口号
	Port int64 `json:"port" binding:"required,min=0,max=65535"`
	// websocket 服务端请求路径
	Path string `json:"path"`
	// 当前 websocket 连接被分配到的分组
	Group string `json:"group" binding:"required,min=1,max=50"`
}

type WSMsgDTO struct {
	// 主键ID
	ID string `json:"id"`
	// 所属链ID
	ChainID string `json:"chain_id"`
	// 消息类型
	Type string `json:"type"`
	// 消息内容
	Message string `json:"message"`
	// 消息哈希值
	Hash string `json:"hash"`
	// 额外的消息属性
	Extra map[string]interface{} `json:"extra" bson:"extra"`
}

// WSSubMsgDTO 与前端 ws 连接交互所需的消息 dto
type WSSubMsgDTO struct {
	// 消息id
	ID string `json:"id"`
	// 消息类型
	Type string `json:"type"`
	// 消息内容
	Content interface{} `json:"content"`
}

type WSManagerVO struct {
	// 当前 websocket 客户端分组数量
	GroupLen int64 `json:"group_len"`
	// 当前 websocket 客户端分组数量
	ClientLen int64 `json:"client_len"`
	// websocket 注册缓冲队列长度
	ChanRegisterLen int64 `json:"chan_register_len"`
	// websocket 注销缓冲队列长度
	ChanUnregisterLen int64 `json:"chan_unregister_len"`
	// websocket 向单客户端发送消息时，消息缓冲队列的长度
	ChanMessageLen int64 `json:"chan_message_len"`
	// websocket 向组客户端广播消息时，消息缓冲队列的长度
	ChanGroupMessageLen int64 `json:"chan_group_message_len"`
	// websocket 向所有客户端广播消息时，消息缓冲队列的长度
	ChanBroadCastMessageLen int64 `json:"chan_broad_cast_message_len"`
}

type WSGroupVO struct {
	// websocket 组名称
	Name string `json:"name"`
	// 该组所包含的客户端信息
	Clients []WSClientVO `json:"clients"`
}

type WSClientVO struct {
	// websocket 客户端连接 id
	ID string `json:"id"`
	// websocket 该客户端所在的组
	Group string `json:"group"`
	// websocket 连接本地地址
	LocalAddr string `json:"local_addr"`
	// websocket 连接远程地址
	RemoteAddr string `json:"remote_addr"`
	// websocket 请求服务端连接的路径
	Path string `json:"path"`
	// 连接是否存活
	IsAlive bool `json:"is_alive"`
	// 是否是 graces 的主动拨号的连接
	IsDial bool `json:"is_dial"`
	// 连接断线后已重试连接的次数
	RetryCnt int64 `json:"retry_cnt"`
}

func (w *WSMsg) ValueOfDTO(dto WSMsgDTO) error {
	if err := util.SimpleCopyProperties(w, dto); err != nil {
		logrus.Errorln(err)
		return exterr.ErrConvert
	}
	id, err := primitive.ObjectIDFromHex(dto.ID)
	if err != nil {
		logrus.Errorln(err)
		return exterr.ErrObjectIDInvalid
	}
	chainID, err := primitive.ObjectIDFromHex(dto.ChainID)
	if err != nil {
		logrus.Errorln(err)
		return exterr.ErrObjectIDInvalid
	}
	w.ID = id
	w.ChainID = chainID
	w.CreateTime = time.Now().Unix()
	return nil
}
