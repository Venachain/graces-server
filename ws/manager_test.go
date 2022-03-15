package ws

import (
	"log"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestManager_RegisterClient(t *testing.T) {
	go DefaultWebsocketManager.Start()
	client := &Client{
		Id:      uuid.NewV4().String(),
		Group:   "test",
		Socket:  nil,
		Message: nil,
	}
	time.Sleep(2 * time.Second)
	DefaultWebsocketManager.RegisterClient(client)
	log.Println(DefaultWebsocketManager.Info())
	assert.True(t, DefaultWebsocketManager.LenClient() > 0)
}

func TestManager_UnRegisterClient(t *testing.T) {
	go DefaultWebsocketManager.Start()
	client := &Client{
		Id:      uuid.NewV4().String(),
		Group:   "test",
		Socket:  nil,
		Message: nil,
	}
	time.Sleep(2 * time.Second)
	DefaultWebsocketManager.RegisterClient(client)
	log.Println(DefaultWebsocketManager.Info())
	assert.True(t, DefaultWebsocketManager.LenClient() > 0)
	time.Sleep(2 * time.Second)
	DefaultWebsocketManager.UnRegisterClient(client)
	log.Println(DefaultWebsocketManager.Info())
	assert.True(t, DefaultWebsocketManager.LenClient() == 0)
}

func TestManager_Dial(t *testing.T) {
	ip := "127.0.0.1"
	port := 26791
	group := "Venachain"
	path := ""
	go DefaultWebsocketManager.Start()
	client, err := DefaultWebsocketManager.Dial(ip, int64(port), path, group)
	assert.True(t, client != nil && err == nil)
}
