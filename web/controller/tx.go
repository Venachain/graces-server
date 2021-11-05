package controller

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"PlatONE-Graces/web/service"
	"PlatONE-Graces/web/util/response"
	"reflect"

	"github.com/gin-gonic/gin"
)

var (
	DefaultTXController *TXController
)

func init() {
	DefaultTXController = newTXController()
}

func newTXController() *TXController {
	return &TXController{
		service: service.DefaultTXService,
	}
}

//TXs godoc
//@Summary 查询交易信息
//@Description 按条件查询交易信息
//@Tags 交易信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.TXQueryCondition true "交易信息查询条件"
//@Success 200 {object} model.Result{data=model.PageInfo{items=[]model.TXVO}} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/txs [post]
func (c *TXController) TXs(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.TXQueryCondition{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	items, e := c.service.TXs(dto)
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

//TXByID go doc
//@Summary 查询交易信息
//@Description 通过 id 查询交易信息
//@Tags 交易信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param id path string true "id" "交易信息id"
//@Success 200 {object} model.Result{data=model.TXVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/tx/id/{id} [GET]
func (c *TXController) TXByID(ctx *gin.Context) {
	result := model.Result{}
	id := ctx.Param("id")
	if len(id) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.TXByID(id)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//TXByHash go doc
//@Summary 查询交易信息
//@Description 通过 hash 查询交易信息
//@Tags 交易信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.TXByHashDTO true "交易信息查询条件"
//@Success 200 {object} model.Result{data=model.TXVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/tx/hash [POST]
func (c *TXController) TXByHash(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.TXByHashDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}

	data, err := c.service.TXByHash(dto.ChainID, dto.Hash)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//TXsForContractCall godoc
//@Summary 查询属于合约调用的交易信息
//@Description 按条件查询属于合约调用的交易信息
//@Tags 交易信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.TXQueryCondition true "交易信息查询条件"
//@Success 200 {object} model.Result{data=model.PageInfo{items=[]model.TXVO}} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/txs/contractcall [post]
func (c *TXController) TXsForContractCall(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.TXQueryCondition{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	if reflect.ValueOf(dto.ContractAddress).IsZero() {
		// - 是专门用于查询合约的特定标识
		dto.ContractAddress = "-"
	}
	items, e := c.service.TXsForContractCall(dto)
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

//TxAmountStats godoc
//@Summary 获取某个时间段内的交易总数
//@Description 链数据查询
//@Tags 交易信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param chainid path string true "chainid"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/stats/tx/count/{chainid} [GET]
func (c *TXController) TxAmountStats(ctx *gin.Context) {
	result := model.Result{}
	chainid := ctx.Param("chainid")
	if len(chainid) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	res, err := c.service.History(chainid)
	if err != nil {
		response.ErrorHandler(ctx, exterr.ErrorGetStats)
		return
	}
	result.Data = res
	response.Success(ctx, result)
	return
}
