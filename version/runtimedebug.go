package version

import (
	"fmt"
	"runtime/debug"
)

func BuildInfo() (*debug.BuildInfo, error) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, fmt.Errorf("fetching build info failed")
	}

	if bi == nil {
		return nil, fmt.Errorf("build information is empty")
	}

	return bi, nil
}
