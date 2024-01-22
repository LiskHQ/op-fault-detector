package api

import (
	"testing"

	"github.com/LiskHQ/op-fault-detector/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties/assert"
)

func TestGetGinModeFromSysLogLevel(t *testing.T) {
	testCases := []struct {
		name        string
		logLevel    string
		wantGinMode string
	}{
		{
			name:        "should return debug gin_mode with trace log_level",
			logLevel:    log.LevelTrace,
			wantGinMode: gin.DebugMode,
		},
		{
			name:        "should return debug gin_mode with debug log_level",
			logLevel:    log.LevelDebug,
			wantGinMode: gin.DebugMode,
		},
		{
			name:        "should return release gin_mode with info log_level",
			logLevel:    log.LevelInfo,
			wantGinMode: gin.ReleaseMode,
		},
		{
			name:        "should return release gin_mode with warn log_level",
			logLevel:    log.LevelWarn,
			wantGinMode: gin.ReleaseMode,
		},
		{
			name:        "should return release gin_mode with error log_level",
			logLevel:    log.LevelError,
			wantGinMode: gin.ReleaseMode,
		},
		{
			name:        "should return release gin_mode with fatal log_level",
			logLevel:    log.LevelFatal,
			wantGinMode: gin.ReleaseMode,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotGinMode := getGinModeFromSysLogLevel(tc.logLevel)
			assert.Equal(t, gotGinMode, tc.wantGinMode)
		})
	}
}
