package config

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_ConfigInit(t *testing.T) {
	logrus.Infof("config: %+v", Config)
	assert.True(t, Config != nil)
}
