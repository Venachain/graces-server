package controller

import (
	"graces/exterr"
	"graces/model"
	"graces/web/service"
	"graces/web/util/response"

	"github.com/gin-gonic/gin"
)

var (
	DefaultCNSController *CNSController
)

func init() {
	DefaultCNSController = newCNSController()
}

func newCNSController() *CNSController {
	return &CNSController{service: service.DefaultCNSService}
}

//CNSByID go doc
//@Summary 查询CNS映射信息
//@Description 通过 id 查询CNS映射信息
//@Tags CNS映射信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param id path string true "id" "CNS映射信息id"
//@Success 200 {object} model.Result{data=model.CNSVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/cns/{id} [GET]
func (c *CNSController) CNSByID(ctx *gin.Context) {
	result := model.Result{}
	id := ctx.Param("id")
	if len(id) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.CNSByID(id)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//CNSs godoc
//@Summary 查询CNS映射信息
//@Description 按条件查询CNS映射信息
//@Tags CNS映射信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.CNSQueryCondition true "CNS映射信息查询条件"
//@Success 200 {object} model.Result{data=model.PageInfo{items=[]model.CNSVO}} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/cnss [post]
func (c *CNSController) CNSs(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.CNSQueryCondition{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	items, e := c.service.CNSs(dto)
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

//Register godoc
//@Summary 注册CNS映射信息
//@Description 把合约注册进合约命名系统（CNS）中
//@Tags CNS映射信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param dto body model.CNSRegisterDTO true "CNS映射信息注册DTO"
//@Success 200 {object} model.Result{data=model.ContractCallResult} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/cns/register [post]
func (c *CNSController) Register(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.CNSRegisterDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	data, e := c.service.Register(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//Redirect godoc
//@Summary CNS映射信息版本重定向
//@Description 重定向该CNS映射信息的版本，默认情况下使用最新版本
//@Tags CNS映射信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param dto body model.CNSRedirectDTO true "CNS重定向信息DTO"
//@Success 200 {object} model.Result{data=model.ContractCallResult} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/cns/redirect [post]
func (c *CNSController) Redirect(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.CNSRedirectDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	data, e := c.service.Redirect(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}
