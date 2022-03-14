package controller

import (
	"graces/exterr"
	"graces/model"
	"graces/web/service"
	"graces/web/util/response"

	"github.com/gin-gonic/gin"
)

var (
	DefaultAccountController *AccountController
)

func init() {
	DefaultAccountController = newAccountController()
}

func newAccountController() *AccountController {
	return &AccountController{service: service.DefaultAccountService}
}

//LockAccount go doc
//@Summary 锁定账户
//@Description 锁定指定链的指定节点上的指定账户
//@Tags 链账户管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param dto body model.LockAccountDTO true "账户DTO"
//@Success 200 {object} model.Result{data=bool} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/account/lock [post]
func (c *AccountController) LockAccount(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.LockAccountDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	data, e := c.service.LockAccount(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//UnlockAccount go doc
//@Summary 解锁账户
//@Description 解锁指定链的指定节点上的指定账户
//@Tags 链账户管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param dto body model.UnlockAccountDTO true "解锁账户DTO"
//@Success 200 {object} model.Result{data=bool} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/account/unlock [post]
func (c *AccountController) UnlockAccount(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.UnlockAccountDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	data, e := c.service.UnlockAccount(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//ListAccount go doc
//@Summary 展示节点的账户列表
//@Description 展示指定链、指定节点的账户列表
//@Tags 链账户管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param dto body model.AccountDTO true "账户信息查询DTO"
//@Success 200 {object} model.Result{data=[]model.AccountVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/account/list [post]
func (c *AccountController) ListAccount(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.AccountDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}

	data, e := c.service.ListAccounts(dto)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}
