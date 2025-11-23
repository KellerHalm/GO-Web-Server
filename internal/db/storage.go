package db

import (
	"database/sql"
	"encoding/json"
	"wb-l0/internal/model"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func New(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{conn: db}, nil
}

func (d *DB) SaveOrder(order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	_, err = d.conn.Exec("INSERT INTO orders (order_uid, data) VALUES ($1, $2) ON CONFLICT (order_uid) DO UPDATE SET data = $2", order.OrderUID, data)
	return err
}

func (d *DB) LoadOrders() (map[string]model.Order, error) {
	rows, err := d.conn.Query("SELECT data FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]model.Order)
	for rows.Next() {
		var data []byte
		var order model.Order
		if err := rows.Scan(&data); err != nil {
			continue
		}
		if err := json.Unmarshal(data, &order); err != nil {
			continue
		}
		result[order.OrderUID] = order
	}
	return result, nil
}
