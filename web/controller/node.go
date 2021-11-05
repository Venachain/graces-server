package controller

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"PlatONE-Graces/web/service"
	"PlatONE-Graces/web/util/response"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

var (
	DefaultNodeController *NodeController
)

func init() {
	DefaultNodeController = newNodeController()
}

func newNodeController() *NodeController {
	return &NodeController{
		service: service.DefaultNodeService,
	}
}

//NodeSync go doc
//@Summary 节点监控模块
//@Description 通过 节点的ip和port查询节点状态
//@Tags 节点信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param  condition body model.NodeSyncReq true "节点的ip和port"
//@Success 200 {object} model.SyncNodeResult 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/nodes/sync [POST]
func (c *NodeController) NodeSync(ctx *gin.Context) {
	result := model.Result{}
	nodeReq := model.NodeSyncReq{}
	if err := ctx.ShouldBind(&nodeReq); nil != err {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.NodeSyncServer(&nodeReq)
	if err != nil {
		logrus.Errorln("node sync error")
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//NodeByID go doc
//@Summary 查询节点信息
//@Description 通过 id 查询节点信息
//@Tags 节点信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param id path string true "id" "节点信息id"
//@Success 200 {object} model.Result{data=model.NodeVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/node/id/{id} [GET]
func (c *NodeController) NodeByID(ctx *gin.Context) {
	result := model.Result{}
	id := ctx.Param("id")
	if len(id) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.NodeByID(id)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//Nodes go doc
//@Summary 查询节点信息
//@Description 按条件查询节点信息
//@Tags 节点信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.NodeQueryCondition true "节点信息查询条件"
//@Success 200 {object} model.Result{data=model.PageInfo{items=[]model.NodeVO}} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/nodes [post]
func (c *NodeController) Nodes(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.NodeQueryCondition{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	items, e := c.service.Nodes(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	total, e := c.service.Count(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	pageInfo := &model.PageInfo{}
	pageData, e := pageInfo.Build(dto.PageDTO, items, total)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = pageData
	response.Success(ctx, result)
	return
}

func (c *NodeController) CreateNode(nodeInfo []string, ch chan []byte) {
	//chainInfo, err := service.DefaultChainService.ChainByName(nodeInfo[1])
	//if err != nil {
	//	logrus.Errorln(err)
	//	return
	//}

	//ws.DefaultDeploy.DeployNewNode(chainInfo, nodeInfo, ch)
}
