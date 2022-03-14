package service

import (
	"fmt"
	"testing"

	"graces/model"

	"github.com/stretchr/testify/assert"
)

func TestGetRpcResult(t *testing.T) {
	id := "6128b643192c48ceac3986a1"
	chain, err := DefaultChainService.ChainByID(id)
	if err != nil {
		fmt.Errorf("get chain by chainid is error")
	}
	endpoint := fmt.Sprintf("http://%v:%v", chain.IP, chain.P2PPort)
	//endpoint := syncer.DefaultSyncer.GetEndPointByChainID(id)
	method := "personal_listAccounts"
	res := model.GetRpcResult(endpoint, method, nil)
	assert.True(t, res != nil)
}
