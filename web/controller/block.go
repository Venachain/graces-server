package controller

import (
	"PlatONE-Graces/exterr"
	"PlatONE-Graces/model"
	"PlatONE-Graces/web/service"
	"PlatONE-Graces/web/util/response"
	"github.com/gin-gonic/gin"
)

var (
	DefaultBlockController *BlockController
)

func init() {
	DefaultBlockController = newBlockController()
}

func newBlockController() *BlockController {
	return &BlockController{
		service: service.DefaultBlockService,
	}
}

//Blocks go doc
//@Summary 查询区块信息
//@Description 按条件查询区块信息
//@Tags 区块信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.BlockQueryCondition true "区块信息查询条件"
//@Success 200 {object} model.Result{data=model.PageInfo{items=[]model.BlockVO}} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/blocks [post]
func (c *BlockController) Blocks(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.BlockQueryCondition{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	items, e := c.service.Blocks(dto)
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

//BlockByID go doc
//@Summary 查询区块信息
//@Description 通过 id 查询区块信息
//@Tags 区块信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param id path string true "id" "区块信息id"
//@Success 200 {object} model.Result{data=model.BlockVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/block/id/{id} [GET]
func (c *BlockController) BlockByID(ctx *gin.Context) {
	result := model.Result{}
	id := ctx.Param("id")
	if len(id) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.BlockByID(id)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//BlockByHash go doc
//@Summary 查询区块信息
//@Description 通过 hash 查询区块信息
//@Tags 区块信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.BlockByHashDTO true "区块信息查询条件"
//@Success 200 {object} model.Result{data=[]model.BlockVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/block/hash [post]
func (c *BlockController) BlockByHash(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.BlockByHashDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}

	data, err := c.service.BlockByHash(dto.ChainID, dto.Hash)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//Stats godoc
//@Summary 获取当前链上交易、区块、合约的总量
//@Description 链数据查询
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param chainid path string true "chainid" "链ID"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/stats/{chainid} [GET]
func (c *BlockController) Stats(ctx *gin.Context) {
	result := model.Result{}
	chainID := ctx.Param("chainid")
	if len(chainID) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}
	res, err := c.service.ChainStats(chainID)
	if err != nil {
		response.ErrorHandler(ctx, exterr.ErrorGetStats)
		return
	}
	result.Data = res
	response.Success(ctx, result)
	return
}
