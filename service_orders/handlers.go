package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderItemRequest struct {
	Product  string `json:"product" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	Items       []OrderItemRequest `json:"items" binding:"required"`
	TotalAmount float64            `json:"totalAmount" binding:"required,gt=0"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"` // in_progress / done / cancelled
}

func parseStatus(s string) (OrderStatus, bool) {
	switch s {
	case "created":
		return StatusCreated, true
	case "in_progress":
		return StatusInProgress, true
	case "done":
		return StatusDone, true
	case "cancelled":
		return StatusCancelled, true
	default:
		return "", false
	}
}

// POST /v1/orders
func handleCreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if len(req.Items) == 0 {
		fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "At least one item is required")
		return
	}
	if req.TotalAmount <= 0 {
		fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Total amount must be > 0")
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		fail(c, http.StatusInternalServerError, "CONTEXT_ERROR", "User ID missing in context")
		return
	}

	items := make([]OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.Product == "" || it.Quantity <= 0 {
			fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid item product or quantity")
			return
		}
		items = append(items, OrderItem{
			Product:  it.Product,
			Quantity: it.Quantity,
		})
	}

	order := &Order{
		ID:          uuid.NewString(),
		UserID:      userID,
		Items:       items,
		Status:      StatusCreated,
		TotalAmount: req.TotalAmount,
	}

	if err := insertOrder(order); err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to create order")
		return
	}

	// доменное событие "создан заказ"
	publishOrderCreated(order, getRequestID(c))

	success(c, order)
}

// GET /v1/orders/:id
func handleGetOrder(c *gin.Context) {
	orderID := c.Param("id")

	order, err := getOrderByID(orderID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get order")
		return
	}
	if order == nil {
		fail(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		fail(c, http.StatusInternalServerError, "CONTEXT_ERROR", "User ID missing in context")
		return
	}

	// правило: владелец или админ/менеджер/директор/заказчик?
	// по ТЗ достаточно "владелец или админ", но можно дать доступ и менеджеру/директору/заказчику для просмотра
	if order.UserID != userID && !hasAdminRole(c) && !isManager(c) && !isDirector(c) && !isCustomer(c) {
		fail(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to view this order")
		return
	}

	success(c, order)
}

// GET /v1/orders
func handleListMyOrders(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		fail(c, http.StatusInternalServerError, "CONTEXT_ERROR", "User ID missing in context")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	sortStr := c.DefaultQuery("sort", "desc") // desc / asc

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	sortDesc := true
	if sortStr == "asc" {
		sortDesc = false
	}

	total, err := getOrdersCountForUser(userID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to count orders")
		return
	}

	orders, err := listOrdersForUser(userID, limit, offset, sortDesc)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to list orders")
		return
	}

	success(c, gin.H{
		"items": orders,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

// PATCH /v1/orders/:id/status
func handleUpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	newStatus, ok := parseStatus(req.Status)
	if !ok {
		fail(c, http.StatusBadRequest, "INVALID_STATUS",
			"Status must be one of: created, in_progress, done, cancelled")
		return
	}

	order, err := getOrderByID(orderID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get order")
		return
	}
	if order == nil {
		fail(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		fail(c, http.StatusInternalServerError, "CONTEXT_ERROR", "User ID missing in context")
		return
	}

	// права:
	// - admin / manager: могут менять любой заказ
	// - engineer: только свои
	// - director/customer: не могут менять
	if hasAdminRole(c) || isManager(c) {
		// ок
	} else if isEngineer(c) {
		if order.UserID != userID {
			fail(c, http.StatusForbidden, "FORBIDDEN", "Engineer can update only own orders")
			return
		}
	} else {
		fail(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to update orders")
		return
	}

	oldStatus := order.Status

	if err := updateOrderStatus(order, newStatus); err != nil {
		if err == sql.ErrNoRows {
			fail(c, http.StatusBadRequest, "INVALID_TRANSITION", "Status transition is not allowed")
			return
		}
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to update order status")
		return
	}

	publishOrderStatusUpdated(order, oldStatus, newStatus, getRequestID(c))

	success(c, order)
}

// POST /v1/orders/:id/cancel
func handleCancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	order, err := getOrderByID(orderID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get order")
		return
	}
	if order == nil {
		fail(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		fail(c, http.StatusInternalServerError, "CONTEXT_ERROR", "User ID missing in context")
		return
	}

	// владелец, менеджер или админ могут отменять
	if !(order.UserID == userID || hasAdminRole(c) || isManager(c)) {
		fail(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to cancel this order")
		return
	}

	oldStatus := order.Status

	if err := updateOrderStatus(order, StatusCancelled); err != nil {
		if err == sql.ErrNoRows {
			fail(c, http.StatusBadRequest, "INVALID_TRANSITION", "Cannot cancel order in this status")
			return
		}
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to cancel order")
		return
	}

	publishOrderStatusUpdated(order, oldStatus, StatusCancelled, getRequestID(c))

	success(c, order)
}

// DELETE /v1/orders/:id
func handleDeleteOrder(c *gin.Context) {
	orderID := c.Param("id")

	order, err := getOrderByID(orderID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to get order")
		return
	}
	if order == nil {
		fail(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		fail(c, http.StatusInternalServerError, "CONTEXT_ERROR", "User ID missing in context")
		return
	}

	// правило удаления:
	// - владелец может удалять только свои заказы в статусе created или cancelled
	// - admin/manager могут удалять любой заказ
	if hasAdminRole(c) || isManager(c) {
		// ok
	} else {
		if order.UserID != userID {
			fail(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to delete this order")
			return
		}
		if order.Status != StatusCreated && order.Status != StatusCancelled {
			fail(c, http.StatusBadRequest, "INVALID_STATE",
				"Only orders in status 'created' or 'cancelled' can be deleted by owner")
			return
		}
	}

	if err := deleteOrder(order); err != nil {
		fail(c, http.StatusInternalServerError, "DB_ERROR", "Failed to delete order")
		return
	}

	success(c, gin.H{
		"id":      order.ID,
		"deleted": true,
	})
}
