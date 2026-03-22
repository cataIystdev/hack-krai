package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kudytudy-api/internal/services"
)

func TestNewWSHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	hub := services.NewWSHub(nil, logger)
	handler := NewWSHandler(hub, logger)
	require.NotNil(t, handler)
	assert.NotNil(t, handler.hub)
}
