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
	DefaultContractController *ContractController
)

func init() {
	DefaultContractController = newContractController()
}

func newContractController() *ContractController {
	return &ContractController{
		service: service.DefaultContractService,
	}
}

//FireWallOpen go doc
//@Summary 开启合约防火墙
//@Description 通过 合约地址 开启合约防火墙
//@Tags 合约信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param  condition body model.FireWall true "链id 和合约地址"
//@Success 200 {object} string 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/contract/openfirewall [POST]
func (c *ContractController) FireWallOpen(ctx *gin.Context) {
	result := model.Result{}
	fireWallParam := model.FireWall{}
	if e := ctx.BindJSON(&fireWallParam); e != nil {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}
	data, err := c.service.OpenFireWall(fireWallParam)
	if err != nil {
		logrus.Errorln("open firewall error")
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//FireWallClose go doc
//@Summary 关闭合约防火墙
//@Description 通过 合约地址 关闭合约防火墙
//@Tags 合约信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param  condition body model.FireWall true "链id 和合约地址"
//@Success 200 {object} string 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/contract/closefirewall [POST]
func (c *ContractController) FireWallClose(ctx *gin.Context) {
	result := model.Result{}
	fireWallParam := model.FireWall{}
	if err := ctx.ShouldBind(&fireWallParam); nil != err {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.CloseFireWall(fireWallParam)
	if err != nil {
		logrus.Errorln("close firewall error")
		response.ErrorHandler(ctx, exterr.ErrorContractFirewall)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//FireWallClose go doc
//@Summary 获取防火墙状态
//@Description 通过合约地址获取合约防火墙状态
//@Tags 合约信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param  condition body model.FireWall true "链id 和合约地址"
//@Success 200 {object} string 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/contract/getfirewallstatus [POST]
func (c *ContractController) GetFirewallStatus(ctx *gin.Context) {
	result := model.Result{}
	fireWallParam := model.FireWall{}
	if err := ctx.ShouldBind(&fireWallParam); nil != err {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.FireWallStatus(fireWallParam)
	if err != nil {
		logrus.Errorln("close firewall error")
		response.ErrorHandler(ctx, exterr.ErrorContractFirewall)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//Contracts godoc
//@Summary 查询合约信息
//@Description 按条件查询合约信息
//@Tags 合约信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.ContractQueryCondition true "合约信息查询条件"
//@Success 200 {object} model.Result{data=model.PageInfo{items=[]model.ContractVO}} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/contracts [post]
func (c *ContractController) Contracts(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.ContractQueryCondition{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	items, e := c.service.Contracts(dto)
	if e != nil {
		response.ErrorHandler(ctx, exterr.ErrorContractParam)
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

//ContractByAddress go doc
//@Summary 查询合约信息
//@Description 通过 合约地址 查询合约信息
//@Tags 合约信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.ContractByAddressDTO true "合约信息查询条件"
//@Success 200 {object} model.Result{data=[]model.TXVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/contract/address [POST]
func (c *ContractController) ContractByAddress(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.ContractByAddressDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}

	data, err := c.service.ContractByAddress(dto.ChainID, dto.ContractAddress)
	if err != nil {
		response.ErrorHandler(ctx, exterr.ErrrorContractByCns)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}
