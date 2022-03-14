package ws

import (
	"errors"
	"fmt"
	"graces/util"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewStringMsgProcessor() *StringMsgProcessor {
	return &StringMsgProcessor{}
}

// StringMsgProcessor 字符串消息处理器
type StringMsgProcessor struct{}

func (s *StringMsgProcessor) Process(ctx *MsgProcessorContext, msg interface{}) error {
	logrus.Debugf("string message process [start]:\n%+v", msg)
	defer logrus.Debugln("string message process [end]")

	data, ok := msg.(string)
	if !ok {
		errStr := fmt.Sprintf("can't response the msg:%+v, it's not [ping]", msg)
		return errors.New(errStr)
	}
	fields := strings.Fields(data)
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

func (s *StringMsgProcessor) Ping(ctx *MsgProcessorContext, msg string) error {
	if msg != "ping" {
		errStr := fmt.Sprintf("can't response the msg: %v, it's not [ping]", msg)
		return errors.New(errStr)
	}
	ctx.client.Message <- []byte("pong")
	return nil
}
