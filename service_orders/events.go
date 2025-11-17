package main

import "log"

// Публикация события "создан заказ"
func publishOrderCreated(o *Order, requestID string) {
	log.Printf(
		`event=order.created requestId=%s orderId=%s userId=%s status=%s total=%.2f`,
		requestID, o.ID, o.UserID, o.Status, o.TotalAmount,
	)
}

// Публикация события "обновлён статус"
func publishOrderStatusUpdated(o *Order, oldStatus, newStatus OrderStatus, requestID string) {
	log.Printf(
		`event=order.status_updated requestId=%s orderId=%s userId=%s from=%s to=%s`,
		requestID, o.ID, o.UserID, oldStatus, newStatus,
	)
}
