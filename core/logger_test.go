package core

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerWrongEnv(t *testing.T) {
	err := os.Setenv("CORE_CHAINCODE_LOGGING_FORMAT", "wrongFormat")
	assert.NoError(t, err)
	logger := Logger()
	assert.NotNil(t, logger)
}

func TestLoggerCorrectEnv(t *testing.T) {
	formatStr := "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
	err := os.Setenv("CORE_CHAINCODE_LOGGING_FORMAT", formatStr)
	assert.NoError(t, err)
	logger := Logger()
	assert.NotNil(t, logger)
}

func TestLoggerEmptyEnv(t *testing.T) {
	err := os.Setenv("CORE_CHAINCODE_LOGGING_FORMAT", "")
	assert.NoError(t, err)
	logger := Logger()
	assert.NotNil(t, logger)
}

func TestLoggerNotSetEnv(t *testing.T) {
	logger := Logger()
	assert.NotNil(t, logger)
}
