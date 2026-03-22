package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestReviewHandler_Create_Unauthorized(t *testing.T) {
	app := fiber.New()
	logger := zap.NewNop()
	
	// Initialize with nil service, as we only test the early 401 return
	handler := NewReviewHandler(nil, logger)
	
	app.Post("/reviews", handler.Create)

	req := httptest.NewRequest("POST", "/reviews", nil)
	resp, err := app.Test(req)
	
	assert.NoError(t, err)
	// We expect 401 because the JWT middleware is not attached and extractUserID will fail
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
