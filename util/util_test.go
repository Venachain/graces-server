package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Addr struct {
	Name string
}

func TestSimpleCopyProperties(t *testing.T) {
	src := struct {
		Name  string
		Age   int64
		Books []string
		Addr  *Addr
	}{}
	src.Name = "Tom"
	src.Age = 18
	src.Books = []string{"1", "2", "3"}
	addr := &Addr{Name: "sh"}
	src.Addr = addr

	dst := struct {
		Name  string
		Age   int64
		Books []string
		Addr  *Addr
	}{}
	err := SimpleCopyProperties(&dst, src)
	t.Log(err)
	assert.True(t, err == nil)
}
