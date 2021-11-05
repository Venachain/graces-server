package ws

import (
	"PlatONE-Graces/util"
	"PlatONE-Graces/web/dao"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewSendReplyMsgProcessor() *SendReplyMsgProcessor {
	return &SendReplyMsgProcessor{}
}

type SendReplyMsgProcessor struct{}

func (s *SendReplyMsgProcessor) Process(ctx *MsgProcessorContext, msg interface{}) error {
	logrus.Debugf("websocket send reply message process [start]:\n%+v", msg)
	defer logrus.Debugln("websocket send reply message process [end]")

	data, ok := msg.(map[string]interface{})
	if !ok {
		errStr := fmt.Sprintf("can't process unknown sendReplyMsg:\n %+v", msg)
		return errors.New(errStr)
	}
	// 主要是从发送消息的 id 中提取标识数据
	sendId := data["id"].(string)
	fields := strings.Fields(sendId)
	msgType := fields[0]
	methodName := util.MethodCapitalized(msgType)

	// 通过反射进行方法调用
	reType := reflect.TypeOf(s)
	method, ok := reType.MethodByName(methodName)
	if !ok {
		return errors.New(fmt.Sprintf("no process method for msgType[%v]", msgType))
	}
	methodParams := make([]reflect.Value, 3)
	// 第一个参数为方法的持有者
	methodParams[0] = reflect.ValueOf(s)
	methodParams[1] = reflect.ValueOf(ctx)
	methodParams[2] = reflect.ValueOf(data)
	resValues := method.Func.Call(methodParams)
	if len(resValues) > 0 {
		if err, ok := resValues[len(resValues)-1].Interface().(error); ok {
			return err
		}
	}
	return nil
}

// Subscription 处理请求订阅成功后响应的消息数据
func (s *SendReplyMsgProcessor) Subscription(ctx *MsgProcessorContext, msg interface{}) error {
	data, ok := msg.(map[string]interface{})
	if !ok {
		errStr := fmt.Sprintf("can't process unknown sendReplyMsg:\n %+v", msg)
		return errors.New(errStr)
	}
	// 主要是从发送消息的 id 中提取标识数据
	sendId := data["id"].(string)
	fields := strings.Fields(sendId)
	topic := fields[1]
	id := fields[2]
	hash := data["result"].(string)
	// 将订阅成功后返回的订阅哈希值设置到消息记录中
	if err := dao.DefaultWSMsgDao.UpdateWSMsgHash(id, topic, hash); err != nil {
		return err
	}
	logrus.Infof("topic[%v] subscribe success, update subMsg[%v] hash to [%v]", topic, id, hash)
	return nil
}
