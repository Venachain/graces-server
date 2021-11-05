package main

import (
	"PlatONE-Graces/config"
	"PlatONE-Graces/syncer"
	"PlatONE-Graces/web/router"
	"PlatONE-Graces/ws"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)
	if err := config.MakeLogConfig(); err != nil {
		log.Fatalf("%v", err)
	}
	gracesRouter := router.InitRouter()
	ws.DefaultWSSubscriber.ChainWSTopicAutoSubDelayStart(config.Config.Syncer.Delay)
	syncer.DefaultChainDataSyncManager.ChainDataIncrSyncDelayStart(config.Config.Syncer.Delay)
	err := gracesRouter.Run(config.Config.HttpConf.Addr())
	if err != nil {
		logrus.Errorf("Graces start err: %v", err)
		return
	}
}