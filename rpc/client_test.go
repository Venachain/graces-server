package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	url := "http://127.0.0.1:6791"
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	client, err := NewClient(ctx, url, "0", "./keystore")
	assert.True(t, client != nil && err == nil)
}
