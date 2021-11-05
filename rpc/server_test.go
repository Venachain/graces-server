package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	chainID = "614afb76cd5466d1c64c4cc4"
)

func TestRPC_GetTXReceiptByTXHash(t *testing.T) {
	hash := "0x9a49ef3f9a32eb27ae73ee9a713490c543512645a2a544de9c1bfae32dbe80dc"
	receipt, err := GetTXReceiptByTXHash(chainID, hash)
	assert.True(t, err == nil)
	t.Logf("receipt:\n%+v", receipt)
}

func TestRPC_GetTXDataByBlockHash(t *testing.T) {
	blockHash := "0xe0cc8dac28903efab33d5f494131bc246ec4198efc9791c40b858e6f784ec609"
	blockID := "61245d5770fa43c7684bb666"
	txs, err := GetTXDataByBlockHash(chainID, blockID, blockHash)
	assert.True(t, err == nil)
	t.Logf("txs:\n%+v", txs)
}

func TestRPC_GetTXDataByBlockNumber(t *testing.T) {
	var blockNumber int64 = 33
	blockID := "61245d5770fa43c7684bb666"
	txs, err := GetTXDataByBlockNumber(chainID, blockID, blockNumber)
	assert.True(t, err == nil)
	t.Logf("txs:\n%+v", txs)
}

func TestRPC_GetBlockHeadByNumber(t *testing.T) {
	var blockNumber int64 = 33
	head, err := GetBlockHeadByNumber(chainID, blockNumber)
	assert.True(t, err == nil)
	t.Logf("head:\n%+v", head)
}

func TestRPC_GetBlockByNumber(t *testing.T) {
	var blockNumber int64 = 33
	block, err := GetBlockByNumber(chainID, blockNumber)
	assert.True(t, err == nil)
	t.Logf("block:\n%+v", block)
}

func TestRPC_GetBlockHeadByHash(t *testing.T) {
	blockHash := "0xe0cc8dac28903efab33d5f494131bc246ec4198efc9791c40b858e6f784ec609"
	head, err := GetBlockHeadByHash(chainID, blockHash)
	assert.True(t, err == nil)
	t.Logf("head:\n%+v", head)
}

func TestRPC_GetBlockByHash(t *testing.T) {
	blockHash := "0xe0cc8dac28903efab33d5f494131bc246ec4198efc9791c40b858e6f784ec609"
	block, err := GetBlockByHash(chainID, blockHash)
	assert.True(t, err == nil)
	t.Logf("block:\n%+v", block)
}

func TestRPC_GetLatestBlockFromChain(t *testing.T) {
	block, err := GetLatestBlockFromChain(chainID)
	assert.True(t, err == nil)
	t.Logf("block:\n%+v", block)
}

func TestRPC_GetLatestBlockFromDB(t *testing.T) {
	block, err := GetLatestBlockFromDB(chainID)
	assert.True(t, err == nil)
	t.Logf("block:\n%+v", block)
}

func TestRPC_GetAllCNS(t *testing.T) {
	cns, err := GetAllCNS(chainID)
	assert.True(t, err == nil)
	for _, v := range cns {
		t.Logf("cns: %+v\n", v)
	}
}

func TestRPC_GetAllNodes(t *testing.T) {
	nodes, err := GetAllNodes(chainID)
	assert.True(t, err == nil)
	for _, v := range nodes {
		t.Logf("nodes: %+v\n", v)
	}
}
