package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jstauff/feats-api/internal/middleware"
	"github.com/jstauff/feats-api/internal/models"
	"github.com/jstauff/feats-api/internal/services"
)

type GoalHandler struct {
	goalService *services.GoalService
}

func NewGoalHandler(goalService *services.GoalService) *GoalHandler {
	return &GoalHandler{
		goalService: goalService,
	}
}

func (h *GoalHandler) GetUserGoals(c *gin.Context) {
	userID := c.Param("id")

	goals, err := h.goalService.GetUserGoals(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(goals))
}

func (h *GoalHandler) CreateGoal(c *gin.Context) {
	var input services.CreateGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	goal, err := h.goalService.CreateGoal(input, userID)
	if err != nil {
		if err == services.ErrInvalidPeriod {
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invalid period. Must be 'daily', 'weekly', or 'monthly'",
			))
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(goal))
}

func (h *GoalHandler) UpdateGoal(c *gin.Context) {
	goalID := c.Param("id")

	var input services.UpdateGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.ErrCodeValidation,
			"Invalid request body",
		))
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)

	goal, err := h.goalService.UpdateGoal(goalID, userID, input)
	if err != nil {
		switch err {
		case services.ErrGoalNotFound:
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Goal not found",
			))
		case services.ErrInvalidPeriod:
			c.JSON(http.StatusBadRequest, models.ErrorResponse(
				models.ErrCodeValidation,
				"Invalid period",
			))
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				models.ErrCodeInternalError,
				"An error occurred",
			))
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(goal))
}

func (h *GoalHandler) DeleteGoal(c *gin.Context) {
	goalID := c.Param("id")
	userID, _ := middleware.GetCurrentUserID(c)

	err := h.goalService.DeleteGoal(goalID, userID)
	if err != nil {
		if err == services.ErrGoalNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				models.ErrCodeNotFound,
				"Goal not found",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.ErrCodeInternalError,
			"An error occurred",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(gin.H{
		"message": "Goal deleted",
	}))
}
