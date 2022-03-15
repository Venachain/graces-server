package syncer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChainInfoSyncManager_IncrSyncStart(t *testing.T) {
	DefaultChainDataSyncManager.IncrSyncStart(chainID, true)
	time.Sleep(2 * time.Second)
	info, _ := DefaultChainDataSyncManager.GetChainDataSyncInfo(chainID)
	t.Logf("%+v", info)
	assert.True(t, info != nil && info.Status == StatusSuccess)
}

func TestChainDataSyncManager_FullSyncStart(t *testing.T) {
	DefaultChainDataSyncManager.FullSyncStart(chainID, true)
	time.Sleep(2 * time.Second)
	info, _ := DefaultChainDataSyncManager.GetChainDataSyncInfo(chainID)
	t.Logf("%+v", info)
	assert.True(t, info != nil && info.Status == StatusSuccess)
}
