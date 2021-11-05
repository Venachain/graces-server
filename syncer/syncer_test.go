package syncer

import (
	"PlatONE-Graces/rpc"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	chainID = "6143ee44b29c57f7aa740d49"
)

func Test_Init(t *testing.T) {
	log.Printf("DefaultSyncer: %+v", DefaultSyncer)
	assert.True(t, DefaultSyncer != nil)
	time.Sleep(5 * time.Second)
}

func TestSyncer_BlockSyncIncr(t *testing.T) {
	block, err := rpc.GetLatestBlockFromChain(chainID)
	assert.True(t, err == nil)
	if block == nil {
		return
	}
	err = DefaultSyncer.BlockIncrSync(chainID, 0, block.NumberU64())
	assert.True(t, err == nil)
}

func TestSyncer_BlockSyncFull(t *testing.T) {
	err := DefaultSyncer.BlockFullSync(chainID)
	assert.True(t, err == nil)
}

func TestSyncer_SyncNode(t *testing.T) {
	err := DefaultSyncer.SyncNode(chainID, true)
	assert.True(t, err == nil)
}

func TestSyncer_SyncCNS(t *testing.T) {
	err := DefaultSyncer.SyncCNS(chainID, true)
	assert.True(t, err == nil)
}
