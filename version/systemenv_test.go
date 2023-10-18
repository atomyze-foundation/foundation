package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/atomyze-foundation/foundation/version"
)

func TestSystemEnv(t *testing.T) {
	s := version.SystemEnv()
	_, ok := s["/etc/issue"]
	assert.True(t, ok)
}
