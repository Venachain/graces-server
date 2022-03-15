package ws

type MsgProcessorContext struct {
	client *Client
}

type MsgProcessorStrategy interface {
	Process(ctx *MsgProcessorContext, msg interface{}) error
}

func NewMsgProcessor(client *Client, strategy MsgProcessorStrategy) *MsgProcessor {
	return &MsgProcessor{
		context:  &MsgProcessorContext{client: client},
		strategy: strategy,
	}
}

// MsgProcessor websocket 消息处理策略模式
type MsgProcessor struct {
	context  *MsgProcessorContext
	strategy MsgProcessorStrategy
}

func (p *MsgProcessor) Process(msg interface{}) error {
	return p.strategy.Process(p.context, msg)
}
