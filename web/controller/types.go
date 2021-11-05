package controller

import (
	"PlatONE-Graces/web/service"
)

type WebsocketController struct {
	service service.IWebsocketService
}

type ChainController struct {
	service service.IChainService
}

type BlockController struct {
	service service.IBlockService
}

type TXController struct {
	service service.ITXService
}

type NodeController struct {
	service service.INodeService
}

type ContractController struct {
	service service.IContractService
}

type CNSController struct {
	service service.ICNSService
}

type AccountController struct {
	service service.IAccountService
}
