package main

import (
	"log"

	"graces/config"
	"graces/syncer"
	"graces/web/router"
	"graces/ws"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	gin.SetMode(config.Config.HttpConf.Mode)
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
