package rest

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"wb-l0/internal/domain"
	"wb-l0/pkg/e"

	"github.com/gin-gonic/gin"
)

// @title OrderService App Api
// @version 1
//
//go:generate mockgen -source=handlers.go -destination=mocks/mock.go
type OrderStorage interface {
	GetByUID(ctx context.Context, uid int) (domain.Order, error)
	CreateOrder(ctx context.Context, order domain.Order) (int, error)
	GetAllOrderIDs(ctx context.Context) ([]domain.Order, error)
}
type ServiceRender interface {
	Home(http.ResponseWriter)
}
type Handler struct {
	ordSer OrderStorage
	render ServiceRender
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger, orderService OrderStorage, render ServiceRender) *Handler {
	return &Handler{
		ordSer: orderService,
		logger: logger,
		render: render,
	}
}

// GetOrder godoc
// @Summary Get order by UID
// @Description Get details of an order by its unique ID.
// @ID get-order-by-uid
// @Produce  json
// @Param id path int true "Order ID"
// @Success 200 {object} domain.Order "Successful operation"
// @Failure 400 {object} map[string]string "Invalid ID supplied"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	idst := c.Param("id")
	id, err := strconv.Atoi(idst)
	if err != nil {
		h.logger.Error("invalid id", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	order, err := h.ordSer.GetByUID(c, id)
	if err != nil {
		if errors.Is(err, e.ErrNotFound) {
			h.logger.Error("failed to find order", slog.Int("uid", id), slog.String("error", err.Error()))
			c.JSON(http.StatusNotFound, gin.H{"error": e.ErrNotFound.Error()})
			return
		}
		h.logger.Error("failed to get order", slog.Int("uid", id), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to perform func GetByUID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Response": order})

}

func (h *Handler) Homepage(c *gin.Context) {
	h.render.Home(c.Writer)
}

// PostHandler godoc
// @Summary Create a new order
// @Description Create a new order with the given details.
// @ID create-order
// @Accept  json
// @Produce  json
// @Param order body domain.Order true "Order object to be created"
// @Success 200 {object} map[string]interface{} "Successful operation"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Failed to create order"
// @Router /order [post]
func (h *Handler) PostHandler(c *gin.Context) {
	var o domain.Order
	if err := c.Bind(&o); err != nil {
		h.logger.Error("failed to bind data", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.ordSer.CreateOrder(c, o)
	if err != nil {
		h.logger.Error("failed to create order", slog.String("error", err.Error()), slog.Int("id", id))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"The order successfully created, id": id})

}

// GetAllHandler godoc
// @Summary Return all orders
// @Description Return all orders
// @ID get-all-orders
// @Accept  json
// @Produce  json
// @Param ids path int true "Orders"
// @Success 200 {object} map[string]interface{} "Successful operation"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Failed to create order"
// @Router /orders [get]
func (h *Handler) GetAllHandler(c *gin.Context) {

	orders, err := h.ordSer.GetAllOrderIDs(c)
	if err != nil {
		h.logger.Error("failed to return orders", slog.String("error", err.Error()))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if orders == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пусто :)"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}
