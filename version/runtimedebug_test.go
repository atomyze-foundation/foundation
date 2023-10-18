package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/atomyze-foundation/foundation/version"
)

func TestBuildInfo(t *testing.T) {
	bi, err := version.BuildInfo()
	assert.NoError(t, err)
	assert.NotNil(t, bi)
}
