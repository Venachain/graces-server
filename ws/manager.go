package ws

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"graces/config"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

var (
	DefaultWebsocketManager *Manager
)

func init() {
	DefaultWebsocketManager = newManager()
	go DefaultWebsocketManager.Start()
	go DefaultWebsocketManager.SendService()
	go DefaultWebsocketManager.SendGroupService()
	go DefaultWebsocketManager.SendAllService()
}

func newManager() *Manager {
	return &Manager{
		Group:            make(map[string]map[string]*Client),
		Register:         make(chan *Client, config.Config.WSConf.BuffSize),
		UnRegister:       make(chan *Client, config.Config.WSConf.BuffSize),
		GroupMessage:     make(chan *GroupMessageData, config.Config.WSConf.BuffSize),
		Message:          make(chan *MessageData, config.Config.WSConf.BuffSize),
		BroadCastMessage: make(chan *BroadCastMessageData, config.Config.WSConf.BuffSize),
		groupCount:       0,
		clientCount:      0,
	}
}

// Start 启动 websocket 管理器
func (manager *Manager) Start() {
	logrus.Infof("websocket manage start")
	for {
		select {
		// 注册
		case client := <-manager.Register:
			logrus.Infof("client [%s] connect", client.Id)
			logrus.Infof("register client [%s] to group [%s]", client.Id, client.Group)

			manager.Lock.Lock()
			if manager.Group[client.Group] == nil {
				manager.Group[client.Group] = make(map[string]*Client)
				manager.groupCount += 1
			}
			manager.Group[client.Group][client.Id] = client
			manager.clientCount += 1
			manager.Lock.Unlock()

		// 注销
		case client := <-manager.UnRegister:
			logrus.Infof("unregister client [%s] from group [%s]", client.Id, client.Group)
			manager.Lock.Lock()
			if _, ok := manager.Group[client.Group]; ok {
				if _, ok := manager.Group[client.Group][client.Id]; ok {
					close(client.Message)
					delete(manager.Group[client.Group], client.Id)
					manager.clientCount -= 1
					if len(manager.Group[client.Group]) == 0 {
						logrus.Infof("delete empty group [%s]", client.Group)
						delete(manager.Group, client.Group)
						manager.groupCount -= 1
					}
				}
			}
			manager.Lock.Unlock()
		}
	}
}

// SendService 处理单个 client 发送数据
func (manager *Manager) SendService() {
	for {
		select {
		case data := <-manager.Message:
			if groupMap, ok := manager.Group[data.Group]; ok {
				if conn, ok := groupMap[data.Id]; ok {
					conn.Message <- data.Message
				}
			}
		}
	}
}

// SendGroupService 处理 group 广播数据
func (manager *Manager) SendGroupService() {
	for {
		select {
		// 发送广播数据到某个组的 channel 变量 Send 中
		case data := <-manager.GroupMessage:
			if groupMap, ok := manager.Group[data.Group]; ok {
				for _, conn := range groupMap {
					conn.Message <- data.Message
				}
			}
		}
	}
}

// SendAllService 处理广播数据【数据广播到所有分组的 websocket 中】
func (manager *Manager) SendAllService() {
	for {
		select {
		case data := <-manager.BroadCastMessage:
			for _, v := range manager.Group {
				for _, conn := range v {
					conn.Message <- data.Message
				}
			}
		}
	}
}

// Send 向指定的 client 发送数据
func (manager *Manager) Send(id string, group string, message []byte) {
	data := &MessageData{
		Id:      id,
		Group:   group,
		Message: message,
	}
	manager.Message <- data
}

// SendGroup 向指定的 Group 广播
func (manager *Manager) SendGroup(group string, message []byte) {
	data := &GroupMessageData{
		Group:   group,
		Message: message,
	}
	manager.GroupMessage <- data
}

// SendAll 向所有分组广播
func (manager *Manager) SendAll(message []byte) {
	data := &BroadCastMessageData{
		Message: message,
	}
	manager.BroadCastMessage <- data
}

// RegisterClient 注册
func (manager *Manager) RegisterClient(client *Client) {
	manager.Register <- client
}

// UnRegisterClient 注销
func (manager *Manager) UnRegisterClient(client *Client) {
	manager.UnRegister <- client
}

// LenGroup 当前组个数
func (manager *Manager) LenGroup() uint {
	return manager.groupCount
}

// LenClient 当前连接个数
func (manager *Manager) LenClient() uint {
	return manager.clientCount
}

// Info 获取 wsManager 管理器信息
func (manager *Manager) Info() map[string]interface{} {
	managerInfo := make(map[string]interface{})
	managerInfo["groupLen"] = manager.LenGroup()
	managerInfo["clientLen"] = manager.LenClient()
	managerInfo["chanRegisterLen"] = len(manager.Register)
	managerInfo["chanUnregisterLen"] = len(manager.UnRegister)
	managerInfo["chanMessageLen"] = len(manager.Message)
	managerInfo["chanGroupMessageLen"] = len(manager.GroupMessage)
	managerInfo["chanBroadCastMessageLen"] = len(manager.BroadCastMessage)
	return managerInfo
}

// WsClient gin 处理 websocket handler
func (manager *Manager) WsClient(ctx *gin.Context) {
	group := ctx.Param("group")

	// 创建 websocket 升级器
	upGrader := websocket.Upgrader{
		// cross origin domain
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		// 处理 Sec-WebSocket-Protocol Header
		Subprotocols: []string{ctx.GetHeader("Sec-WebSocket-Protocol")},
	}
	// 将 http 升级为 websocket
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logrus.Errorf("websocket connect error: %s", group)
		return
	}

	client := &Client{
		Id:         uuid.NewV4().String(),
		Group:      group,
		LocalAddr:  conn.LocalAddr().String(),
		RemoteAddr: conn.RemoteAddr().String(),
		Path:       ctx.Request.URL.String(),
		Socket:     conn,
		IsAlive:    true,
		IsDial:     false,
		RetryCnt:   0,
		Message:    make(chan []byte, config.Config.WSConf.BuffSize),
	}
	manager.RegisterClient(client)
	go client.Read()
	go client.Write()
}

// Dial 作为 websocket 客户端拨号去连接其他 websocket 服务端
func (manager *Manager) Dial(ip string, port int64, path string, group string) (*Client, error) {
	host := fmt.Sprintf("%s:%v", ip, port)
	uri := url.URL{
		Scheme: "ws",
		Host:   host,
		Path:   path,
	}
	ctx, _ := context.WithTimeout(context.Background(), config.Config.WSConf.Timeout*time.Second)
	conn, resp, err := websocket.DefaultDialer.DialContext(ctx, uri.String(), nil)
	if err != nil {
		logrus.Errorf("websocket dial [%s] err: %v", uri.String(), err)
		return nil, err
	}

	logrus.Debugf("websocket dial success, response: %+v", resp)

	client := &Client{
		Id:         uuid.NewV4().String(),
		Group:      group,
		LocalAddr:  conn.LocalAddr().String(),
		RemoteAddr: conn.RemoteAddr().String(),
		Path:       path,
		Socket:     conn,
		IsAlive:    true,
		IsDial:     true,
		RetryCnt:   0,
		Message:    make(chan []byte, config.Config.WSConf.BuffSize),
	}
	manager.RegisterClient(client)
	go client.Read()
	go client.Write()
	return client, nil
}

func (manager *Manager) GetClientById(id string) *Client {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()
	var client *Client
	for _, group := range manager.Group {
		if c, ok := group[id]; ok {
			client = c
			break
		}
	}
	return client
}

func (manager *Manager) GetClient(group string, id string) *Client {
	manager.Lock.Lock()
	defer manager.Lock.Unlock()
	return manager.Group[group][id]
}
