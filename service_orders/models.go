package main

import (
	"database/sql"
	"encoding/json"
	"time"
)

type OrderStatus string

const (
	StatusCreated    OrderStatus = "created"
	StatusInProgress OrderStatus = "in_progress"
	StatusDone       OrderStatus = "done"
	StatusCancelled  OrderStatus = "cancelled"
)

type OrderItem struct {
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}

type Order struct {
	ID          string      `json:"id"`
	UserID      string      `json:"userId"`
	Items       []OrderItem `json:"items"`
	Status      OrderStatus `json:"status"`
	TotalAmount float64     `json:"totalAmount"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func insertOrder(o *Order) error {
	now := time.Now()
	o.CreatedAt = now
	o.UpdatedAt = now

	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT INTO orders (id, user_id, items_json, status, total_amount, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		o.ID, o.UserID, string(itemsJSON), string(o.Status), o.TotalAmount, o.CreatedAt, o.UpdatedAt,
	)
	return err
}

func getOrderByID(id string) (*Order, error) {
	row := db.QueryRow(
		`SELECT id, user_id, items_json, status, total_amount, created_at, updated_at
		 FROM orders WHERE id = ?`,
		id,
	)

	var o Order
	var itemsJSON string
	var statusStr string

	if err := row.Scan(&o.ID, &o.UserID, &itemsJSON, &statusStr, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(itemsJSON), &o.Items); err != nil {
		return nil, err
	}
	o.Status = OrderStatus(statusStr)

	return &o, nil
}

func getOrdersCountForUser(userID string) (int, error) {
	row := db.QueryRow(`SELECT COUNT(*) FROM orders WHERE user_id = ?`, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func listOrdersForUser(userID string, limit, offset int, sortDesc bool) ([]*Order, error) {
	orderDir := "ASC"
	if sortDesc {
		orderDir = "DESC"
	}

	rows, err := db.Query(
		`SELECT id, user_id, items_json, status, total_amount, created_at, updated_at
		 FROM orders
		 WHERE user_id = ?
		 ORDER BY created_at `+orderDir+`
		 LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		var itemsJSON, statusStr string
		if err := rows.Scan(&o.ID, &o.UserID, &itemsJSON, &statusStr, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(itemsJSON), &o.Items); err != nil {
			return nil, err
		}
		o.Status = OrderStatus(statusStr)
		orders = append(orders, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// проверить, допустим ли переход статуса
func canTransitionStatus(from, to OrderStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case StatusCreated:
		return to == StatusInProgress || to == StatusCancelled
	case StatusInProgress:
		return to == StatusDone || to == StatusCancelled
	case StatusDone, StatusCancelled:
		// финальные – дальше нельзя
		return false
	default:
		return false
	}
}

// обновление статуса в БД (и в объекте)
func updateOrderStatus(o *Order, newStatus OrderStatus) error {
	if !canTransitionStatus(o.Status, newStatus) {
		// используем ErrNoRows как маркер "нельзя перейти"
		return sql.ErrNoRows
	}

	now := time.Now()
	o.Status = newStatus
	o.UpdatedAt = now

	_, err := db.Exec(
		`UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`,
		string(o.Status), o.UpdatedAt, o.ID,
	)
	return err
}

// удаление заказа
func deleteOrder(o *Order) error {
	_, err := db.Exec(`DELETE FROM orders WHERE id = ?`, o.ID)
	return err
}
