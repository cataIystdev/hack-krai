package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"kudytudy-api/internal/models"
	"kudytudy-api/internal/services"
)

type ReviewHandler struct {
	reviewService *services.ReviewService
	logger        *zap.Logger
}

func NewReviewHandler(reviewService *services.ReviewService, logger *zap.Logger) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
		logger:        logger.Named("review_handler"),
	}
}

func (h *ReviewHandler) Create(c fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	var req models.CreateReviewRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "некорректный формат запроса"})
	}

	review, err := h.reviewService.CreateReview(c.Context(), req, userID.String())
	if err != nil {
		if errors.Is(err, services.ErrReviewValidation) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
		}
		h.logger.Error("failed to create review", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "внутренняя ошибка сервера"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    review,
	})
}

func (h *ReviewHandler) ListLocationReviews(c fiber.Ctx) error {
	locID := c.Params("id")
	
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	reviews, err := h.reviewService.GetLocationReviews(c.Context(), locID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list location reviews", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "внутренняя ошибка сервера"})
	}
	
	if reviews == nil {
		reviews = make([]models.ReviewResponse, 0)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    reviews,
	})
}

func (h *ReviewHandler) ListMyReviews(c fiber.Ctx) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	reviews, err := h.reviewService.GetUserReviews(c.Context(), userID.String(), limit, offset)
	if err != nil {
		h.logger.Error("failed to list my reviews", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "внутренняя ошибка сервера"})
	}

	if reviews == nil {
		reviews = make([]models.ReviewResponse, 0)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    reviews,
	})
}
