package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewOfflineService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewOfflineService(nil, nil, logger)
	require.NotNil(t, svc)
}

func TestNewOfflineService_NilDeps(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := NewOfflineService(nil, nil, logger)
	assert.NotNil(t, svc)
	assert.Nil(t, svc.routeService)
	assert.Nil(t, svc.locationService)
}
