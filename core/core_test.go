package core

import (
	"testing"
)

func TestInvokePanic(t *testing.T) {
	cc := ChainCode{}
	cc.Invoke(nil)
}
