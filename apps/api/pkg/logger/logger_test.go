package logger_test

import (
	"testing"

	"github.com/financeos/api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_DevelopmentMode(t *testing.T) {
	log, err := logger.New("development", "debug")
	require.NoError(t, err)
	assert.NotNil(t, log)
}

func TestNew_ProductionMode(t *testing.T) {
	log, err := logger.New("production", "info")
	require.NoError(t, err)
	assert.NotNil(t, log)
}

func TestNew_InvalidLevel(t *testing.T) {
	_, err := logger.New("development", "invalid-level")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing log level")
}

func TestNew_AllLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			log, err := logger.New("development", level)
			require.NoError(t, err)
			assert.NotNil(t, log)
		})
	}
}
