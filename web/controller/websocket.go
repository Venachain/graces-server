package controller

import (
	"graces/exterr"
	"graces/model"
	"graces/web/service"
	"graces/web/util/response"

	"github.com/gin-gonic/gin"
)

var (
	DefaultWebsocketController *WebsocketController
)

func init() {
	DefaultWebsocketController = newWebSocketController()
}

func newWebSocketController() *WebsocketController {
	return &WebsocketController{
		service: service.DefaultWebsocketService,
	}
}

//Manager go doc
//@Summary WebSocket 管理器信息
//@Description 查询 WebSocket 管理器当前状态信息
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Success 200 {object} model.Result{data=model.WSManagerVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/manager [GET]
func (c *WebsocketController) Manager(ctx *gin.Context) {
	result := model.Result{
		Data: c.service.Manager(),
	}
	response.Success(ctx, result)
	return
}

//Groups go doc
//@Summary 所有 WebSocket 组信息
//@Description 查询 WebSocket 中所有组的详细信息
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Success 200 {object} model.Result{data=[]model.WSGroupVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/groups [GET]
func (c *WebsocketController) Groups(ctx *gin.Context) {
	groups, err := c.service.Groups()
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result := model.Result{
		Data: groups,
	}
	response.Success(ctx, result)
	return
}

//GroupByName go doc
//@Summary 单个 WebSocket 组信息
//@Description 通过 组名称 查询 WebSocket 中指定组的详细信息
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param group path string true "group" "组名称"
//@Success 200 {object} model.Result{data=model.WSGroupVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/group/{group} [GET]
func (c *WebsocketController) GroupByName(ctx *gin.Context) {
	result := model.Result{}
	name := ctx.Param("group")
	if len(name) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, e := c.service.Group(name)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//Send go doc
//@Summary 向单个 WebSocket 客户端发送信息
//@Description 向指定 id 和 group 的 WebSocket 客户端发送信息
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param wsMessageDto body model.WSMessageDTO true "数据信息"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/send [POST]
func (c *WebsocketController) Send(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.WSMessageDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	e := c.service.Send(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	response.Success(ctx, result)
	return
}

//SendGroup go doc
//@Summary 向一个组中的所有 WebSocket 客户端广播信息
//@Description 向指定 group 中的所有 WebSocket 客户端广播信息
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param wsGroupMessageDTO body model.WSGroupMessageDTO true "数据信息"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/sendgroup [POST]
func (c *WebsocketController) SendGroup(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.WSGroupMessageDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	e := c.service.SendGroup(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	response.Success(ctx, result)
	return
}

//SendAll go doc
//@Summary 向所有 WebSocket 客户端广播信息
//@Description 向所有 WebSocket 客户端广播信息
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param wsBroadCastMessageDTO body model.WSBroadCastMessageDTO true "数据信息"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/sendall [POST]
func (c *WebsocketController) SendAll(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.WSBroadCastMessageDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	c.service.SendAll(dto)
	response.Success(ctx, result)
	return
}

//Dial go doc
//@Summary 拨号连接
//@Description 作为 websocket 客户端向其他 websocket 服务端拨号建立连接
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param wsDialDTO body model.WSDialDTO true "拨号信息"
//@Success 200 {object} model.Result{data=model.WSGroupVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/dial [POST]
func (c *WebsocketController) Dial(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.WSDialDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, e := c.service.Dial(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//ClientSend go doc
//@Summary websocket 客户端向 websocket 服务端发送消息
//@Description 让指定 id 和 group 的 WebSocket 客户端向它所连接的服务端发送信息
//@Tags WebSocket 管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param wsMessageDto body model.WSMessageDTO true "数据信息"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/ws/clientsend [POST]
func (c *WebsocketController) ClientSend(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.WSMessageDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	e := c.service.ClientSend(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	response.Success(ctx, result)
	return
}
