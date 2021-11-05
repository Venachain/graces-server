package rpc

import (
	"context"
	"testing"
	"time"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/client/precompiled"

	"github.com/stretchr/testify/assert"
)

func buildMsgCallerParams() (*TxParams, *ContractParams) {
	contractAddr := precompile.CnsManagementAddress
	funcName := "getRegisteredContracts"
	defaultInter := "wasm"

	funcParams := &struct {
		Name    string
		Address string
		Origin  string
		Range   string
	}{}
	funcParams.Range = "(0,0)"

	contract := &ContractParams{
		ContractAddr: contractAddr,
		Method:       funcName,
		Interpreter:  defaultInter,
		AbiMethods:   nil,
		Data:         funcParams,
	}
	return &TxParams{}, contract
}

func TestNewMsgCaller(t *testing.T) {
	url := "http://127.0.0.1:6791"
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	client, err := NewClient(ctx, url, "", "")
	assert.True(t, err == nil && client != nil)
	caller := NewMsgCaller(client)
	txParams, contractParams := buildMsgCallerParams()
	res, err := caller.Call(txParams, contractParams)
	assert.True(t, err == nil)
	t.Log(res)
}
