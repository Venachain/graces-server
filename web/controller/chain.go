package controller

import (
	"encoding/json"
	"fmt"

	"graces/exterr"
	"graces/model"
	"graces/syncer"
	"graces/web/service"
	"graces/web/util/response"
	"graces/ws"

	"github.com/Venachain/Venachain/cmd/vcl/client/packet"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	DefaultChainController *ChainController
)

func init() {
	DefaultChainController = newChainController()
}

func newChainController() *ChainController {
	return &ChainController{
		service: service.DefaultChainService,
	}
}

//InsertChain godoc
//@Summary 添加链信息
//@Description 添加链信息进 graces 平台
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param chainDTO body model.ChainDTO true "链信息"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain [post]
func (c *ChainController) InsertChain(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.ChainDTO{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}

	if e := c.service.InsertChain(dto); e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	response.Success(ctx, result)
	return
}

//ChainById go doc
//@Summary 查询链信息
//@Description 通过 id 查询链信息
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param id path string true "id" "链信息id"
//@Success 200 {object} model.Result{data=model.ChainVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/id/{id} [GET]
func (c *ChainController) ChainById(ctx *gin.Context) {
	result := model.Result{}
	id := ctx.Param("id")
	if len(id) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, err := c.service.ChainByID(id)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//ChainByName go doc
//@Summary 查询链信息
//@Description 通过 name 查询链信息
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param name path string true "name" "链名称"
//@Success 200 {object} model.Result{data=model.ChainVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/name/{name} [GET]
func (c *ChainController) ChainByName(ctx *gin.Context) {
	result := model.Result{}
	name := ctx.Param("name")
	if len(name) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	data, e := c.service.ChainByName(name)
	if e != nil {
		response.ErrorHandler(ctx, e)
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//Chains godoc
//@Summary 查询链信息
//@Description 按条件查询链信息
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.ChainQueryCondition true "链信息查询条件"
//@Success 200 {object} model.Result{data=model.PageInfo{items=[]model.ChainVO}} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chains [post]
func (c *ChainController) Chains(ctx *gin.Context) {
	result := model.Result{}
	// 数据绑定
	dto := model.ChainQueryCondition{}
	if e := ctx.BindJSON(&dto); e != nil {
		response.ErrorHandler(ctx, exterr.NewError(exterr.ErrCodeParameterInvalid, e.Error()))
		return
	}
	items, e := c.service.Chains(dto)
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

//IncrSyncStart go doc
//@Summary 链数据增量同步：开始
//@Description 开始增量同步链数据
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param chainid path string true "chainid" "链ID"
//@Success 200 {object} model.Result{} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/incrsync/start/{chainid} [GET]
func (c *ChainController) IncrSyncStart(ctx *gin.Context) {
	result := model.Result{}
	chainID := ctx.Param("chainid")
	if len(chainID) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}
	if chain, err := c.service.ChainByID(chainID); chain == nil || err != nil {
		msg := fmt.Sprintf("chain[%s] does not exist", chainID)
		result.Msg = msg
		result.Code = exterr.ErrChainDataSync.Code
		response.Fail(ctx, result)
		return
	}
	syncer.DefaultChainDataSyncManager.IncrSyncStart(chainID, true)
	result.Data = chainID
	response.Success(ctx, result)
	return
}

//FullSyncStart go doc
//@Summary 链数据全量同步：开始
//@Description 开始全量同步链数据
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param chainid path string true "chainid" "链ID"
//@Success 200 {object} model.Result{} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/fullsync/start/{chainid} [GET]
func (c *ChainController) FullSyncStart(ctx *gin.Context) {
	result := model.Result{}
	chainID := ctx.Param("chainid")
	if len(chainID) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}
	if chain, err := c.service.ChainByID(chainID); chain == nil || err != nil {
		msg := fmt.Sprintf("chain[%s] does not exist", chainID)
		result.Msg = msg
		result.Code = exterr.ErrChainDataSync.Code
		response.Fail(ctx, result)
		return
	}
	syncer.DefaultChainDataSyncManager.FullSyncStart(chainID, true)
	result.Data = chainID
	response.Success(ctx, result)
	return
}

//ChainDataSyncInfo go doc
//@Summary 链数据同步信息
//@Description 查询链数据同步信息
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param chainid path string true "chainid" "链ID"
//@Success 200 {object} model.Result{data=model.ChainDataSyncInfoVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/sync/info/{chainid} [GET]
func (c *ChainController) ChainDataSyncInfo(ctx *gin.Context) {
	result := model.Result{}
	chainID := ctx.Param("chainid")
	if len(chainID) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}
	info, ok := syncer.DefaultChainDataSyncManager.GetChainDataSyncInfo(chainID)
	if !ok {
		result.Data = nil
		response.Success(ctx, result)
		return
	}
	vo, err := info.ToVO()
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}
	result.Data = *vo
	response.Success(ctx, result)
	return
}

//GetSystemConfig godoc
//@Summary 获取当前账户的系统参数
//@Description 获取系统参数
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param id path string true "chain ID"
//@Success 200 {object} model.Result{data=model.SystemConfigVO} 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/getsystemconfig/{id} [GET]
func (c *ChainController) GetSystemConfig(ctx *gin.Context) {
	result := model.Result{}
	id := ctx.Param("id")
	if len(id) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}
	if chain, err := c.service.ChainByID(id); chain == nil || err != nil {
		msg := fmt.Sprintf("chain[%s] is not existed", id)
		result.Msg = msg
		result.Code = exterr.ErrorGetSysconfig.Code
		response.Fail(ctx, result)
		return
	}
	data_blockGasLimit, err := c.service.GetSysConfigString(id, "getBlockGasLimit")
	data_TxGasLimit, err := c.service.GetSysConfigString(id, "getTxGasLimit")
	model.TxGasLimitConst = data_TxGasLimit

	data_GasContractName, err := c.service.GetSysConfigString(id, "getGasContractName")
	data_IsProduceEmptyBlock, err := c.service.GetSysConfigString(id, "getIsProduceEmptyBlock")
	data_CheckContractDeployPermission, err := c.service.GetSysConfigString(id, "getCheckContractDeployPermission")
	data_IsApproveDeployedContract, err := c.service.GetSysConfigString(id, "getIsApproveDeployedContract")
	data_IsTxUseGas, err := c.service.GetSysConfigString(id, "getIsTxUseGas")
	data := model.SystemConfigVO{
		ChainID:                   id,
		BlockGasLimit:             data_blockGasLimit,
		TxGasLimit:                data_TxGasLimit,
		IsUseGas:                  data_IsTxUseGas,
		IsApproveDeployedContract: data_IsApproveDeployedContract,
		IsCheckDeployPermission:   data_CheckContractDeployPermission,
		IsProduceEmptyBlock:       data_IsProduceEmptyBlock,
		GasContractName:           data_GasContractName,
	}
	if err != nil {
		response.ErrorHandler(ctx, exterr.ErrorGetSysconfig)
		logrus.Errorln("getSysConfig error")
		return
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//SetSystemConfig godoc
//@Summary 设置当前账户的系统参数
//@Description 设置系统参数
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param condition body model.SystemConfigVO true "系统参数设置信息"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/setsystemconfig [post]
func (c *ChainController) SetSystemConfig(ctx *gin.Context) {
	result := model.Result{}
	var err error
	var systemConfig model.SystemConfigVO
	var data model.SystemConfigVO
	if err = ctx.ShouldBind(&systemConfig); nil != err {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	if systemConfig.BlockGasLimit != "" {
		funcParams := &struct {
			BlockGasLimit string
		}{BlockGasLimit: systemConfig.BlockGasLimit}
		data_blockGasLimit, err := c.service.SetSysConfigString(systemConfig.ChainID, "setBlockGasLimit", funcParams)
		if err != nil {
			response.ErrorHandler(ctx, exterr.ErrorSetSysconfig)
			return
		}
		data.BlockGasLimit = data_blockGasLimit
	}
	if systemConfig.TxGasLimit != "" {
		funcParams := &struct {
			TxGasLimit string
		}{TxGasLimit: systemConfig.TxGasLimit}
		data_TxGasLimit, err := c.service.SetSysConfigString(systemConfig.ChainID, "setTxGasLimit", funcParams)
		if err == nil {
			model.TxGasLimitConst = systemConfig.TxGasLimit
		}
		if err != nil {
			response.ErrorHandler(ctx, exterr.ErrorSetSysconfig)
			return
		}
		data.TxGasLimit = data_TxGasLimit
	}
	if systemConfig.IsUseGas != "" {
		funcParams := &struct {
			IsTxUseGas string
		}{IsTxUseGas: systemConfig.IsUseGas}
		data_IsTxUseGas, err := c.service.SetSysConfigString(systemConfig.ChainID, "setIsTxUseGas", funcParams)
		if err != nil {
			response.ErrorHandler(ctx, exterr.ErrorSetSysconfig)
			return
		}
		data.IsUseGas = data_IsTxUseGas
	}
	if systemConfig.IsApproveDeployedContract != "" {
		funcParams := &struct {
			IsApproveDeployedContract string
		}{IsApproveDeployedContract: systemConfig.IsApproveDeployedContract}

		data_IsApproveDeployedContract, err := c.service.SetSysConfigString(systemConfig.ChainID, "setIsApproveDeployedContract", funcParams)
		if err != nil {
			response.ErrorHandler(ctx, exterr.ErrorSetSysconfig)
			return
		}
		data.IsApproveDeployedContract = data_IsApproveDeployedContract
	}
	if systemConfig.IsCheckDeployPermission != "" {
		funcParams := &struct {
			CheckContractDeployPermission string
		}{CheckContractDeployPermission: systemConfig.IsCheckDeployPermission}
		data_CheckContractDeployPermission, err := c.service.SetSysConfigString(systemConfig.ChainID, "setCheckContractDeployPermission", funcParams)
		if err != nil {
			response.ErrorHandler(ctx, exterr.ErrorSetSysconfig)
			return
		}
		data.IsCheckDeployPermission = data_CheckContractDeployPermission
	}
	if systemConfig.IsProduceEmptyBlock != "" {
		funcParams := &struct {
			IsProduceEmptyBlock string
		}{IsProduceEmptyBlock: systemConfig.IsProduceEmptyBlock}
		data_IsProduceEmptyBlock, err := c.service.SetSysConfigString(systemConfig.ChainID, "setIsProduceEmptyBlock", funcParams)
		if err != nil {
			response.ErrorHandler(ctx, exterr.ErrorSetSysconfig)
			return
		}
		data.IsProduceEmptyBlock = data_IsProduceEmptyBlock
	}
	if systemConfig.GasContractName != "" {
		funcParams := &struct {
			GasContractName string
		}{GasContractName: systemConfig.GasContractName}
		data_GasContractName, err := c.service.SetSysConfigString(systemConfig.ChainID, "setGasContractName", funcParams)
		if err != nil {
			response.ErrorHandler(ctx, exterr.ErrorSetSysconfig)
			return
		}
		data.GasContractName = data_GasContractName
	}
	result.Data = data
	response.Success(ctx, result)
	return
}

//DeployContract godoc
//@Summary 为指定链部署智能合约
//@Description 部署智能合约
//@Tags 链信息管理
//@version 1.0
//@Accept json
//@Produce  json
//@Param data formData file true "file"
//@Success 200 {object} model.Result 成功后返回值
//@Failure 400 {object} model.Result 请求参数有误
//@Router /api/chain/deploy/contract/:chainid [GET]
func (c *ChainController) DeployContract(ctx *gin.Context) {
	result := model.Result{}

	chainID := ctx.Param("chainid")
	if len(chainID) == 0 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}
	if chain, err := c.service.ChainByID(chainID); chain == nil || err != nil {
		msg := fmt.Sprintf("chain[%s] is not existed", chainID)
		result.Msg = msg
		result.Code = exterr.ErrContractDeploy.Code
		response.Fail(ctx, result)
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	files := form.File["file"]
	if len(files) == 1 {
		response.ErrorHandler(ctx, exterr.ErrParameterInvalid)
		return
	}

	//unlock account
	address, err := service.DefaultAccountService.FirstAccount(chainID)
	if err != nil {
		response.ErrorHandler(ctx, err)
		return
	}

	unlockAccountDTO := model.UnlockAccountDTO{
		LockAccountDTO: model.LockAccountDTO{
			AccountDTO: model.AccountDTO{ChainID: chainID, NodeID: ""},
			Account:    address,
		},
		Password: "0",
		Duration: 0,
	}

	unlock, err := service.DefaultAccountService.UnlockAccount(unlockAccountDTO)
	if err != nil || !unlock {
		response.ErrorHandler(ctx, err)
		return
	}

	//todo 待优化：可根据特定account来部署合约
	account, _ := service.DefaultAccountService.FirstAccount(chainID)
	res, err := ws.DefaultDeploy.DeployContract(chainID, account, files)
	if err != nil {
		response.ErrorHandler(ctx, err)
	} else {
		result.Data = res

		var receipt packet.ReceiptParsingReturn
		resString := res.([]interface{})[0].(string)

		if err := json.Unmarshal([]byte(resString), &receipt); err != nil {
			response.ErrorHandler(ctx, err)
			return
		}
		result.Data = &receipt
		response.Success(ctx, result)
	}
	return
}
