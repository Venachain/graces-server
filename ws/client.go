package ws

import (
	"encoding/json"
	"strings"

	"graces/config"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// 读信息，从 websocket 连接直接读取数据
func (c *Client) Read() {
	defer func() {
		c.IsAlive = false
		DefaultWebsocketManager.UnRegister <- c
		logrus.Infof("client [%s] disconnect", c.Id)
		if err := c.Socket.Close(); err != nil {
			logrus.Errorf("client [%s] disconnect err: %s", c.Id, err)
		}
	}()

	for {
		messageType, message, err := c.Socket.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			break
		}
		msg := string(message)
		logrus.Debugf("client [%s] receive message: %s", c.Id, msg)
		go c.readMessageProcessor(msg)
	}
}

// 写信息，从 channel 变量 Send 中读取数据写入 websocket 连接
func (c *Client) Write() {
	defer func() {
		c.IsAlive = false
		logrus.Infof("client [%s] disconnect", c.Id)
		if err := c.Socket.Close(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			logrus.Errorf("client [%s] disconnect err: %s", c.Id, err)
		}
	}()

	for {
		select {
		case message, ok := <-c.Message:
			if !ok {
				_ = c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			logrus.Debugf("client [%s] write message: %s", c.Id, string(message))
			err := c.Socket.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				logrus.Errorf("client [%s] writemessage err: %s", c.Id, err)
			}
		}
	}
}

func (c *Client) readMessageProcessor(message string) {
	if message == "ping" {
		msgProcessor := NewMsgProcessor(c, NewStringMsgProcessor())
		err := msgProcessor.Process(message)
		if err != nil {
			logrus.Errorln(err)
		}
		return
	}

	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		logrus.Errorln("[readMessageProcessor] the message [%s] can't unmarshal to map[string]interface{}， err: %v", message, err)
		return
	}
	// 处理作为客户端主动发送消息给服务端后收到回复的消息类型
	_, ok := data["id"].(string)
	if ok {
		err = c.sendTypeMsgProcessor(data)
		if err != nil {
			logrus.Errorln(err)
		}
		return
	}

	// 处理连接建立后，作为服务端被动收到客户端请求，或作为客户端被动收到服务端推送的消息类型
	msgType, ok := data["method"].(string)
	if ok {
		err = c.receiveTypeMsgProcessor(msgType, data)
		if err != nil {
			logrus.Errorln(err)
		}
		return
	}

	// 读取到未知类型的消息
	logrus.Warningf("unknown webosocket msg:\n %+v", data)
}

// 处理作为客户端主动发送消息给服务端后收到回复的消息类型
func (c *Client) sendTypeMsgProcessor(data map[string]interface{}) error {
	msgProcessor := NewMsgProcessor(c, NewSendReplyMsgProcessor())
	return msgProcessor.Process(data)
}

// 处理连接建立后，作为服务端被动收到客户端请求，或作为客户端被动收到服务端推送的消息类型
func (c *Client) receiveTypeMsgProcessor(msgType string, data map[string]interface{}) error {
	switch msgType {
	case config.Config.WSConf.WsMsgTypesConf.Sub.ReceiveType:
		msgProcessor := NewMsgProcessor(c, NewSubMsgProcessor())
		err := msgProcessor.Process(data)
		if err != nil {
			return err
		}
	case "deploy":
		c.Message <- []byte("Start deploy! Please wait...")
		DefaultDeploy.DeployNewChain(data, c.Message)
	case "create":
		c.Message <- []byte("Start create node! Please wait...")
		DefaultDeploy.DeployNewNode(data, c.Message)
	case "startNode":
		c.Message <- []byte("Start node!")
		DefaultDeploy.StartNode(data, c.Message)
	case "stopNode":
		c.Message <- []byte("Stop node!")
		DefaultDeploy.StopNode(data, c.Message)
	case "restartNode":
		c.Message <- []byte("Restart node!")
		DefaultDeploy.RestartNode(data, c.Message)
	default:
		logrus.Errorf("unknown msgType[%v]", msgType)
	}
	return nil
}
