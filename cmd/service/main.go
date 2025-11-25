package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"wb-l0/internal/cache"
	"wb-l0/internal/db"
	"wb-l0/internal/model"

	"github.com/nats-io/stan.go"
)

const (
	clusterID = "test-cluster"
	clientID  = "order-service"
	dbConn    = "postgres://user:password@localhost:5432/orders_db?sslmode=disable"
)

func main() {

	storage, err := db.New(dbConn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	localCache := cache.New()
	orders, err := storage.LoadOrders()
	if err != nil {
		log.Fatalf("Failed to restore cache: %v", err)
	}
	for _, order := range orders {
		localCache.Set(order)
	}
	log.Printf("Restored %d orders from DB to Cache", len(orders))

	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL("nats://localhost:4223"))
	if err != nil {
		log.Fatalf("NATS connection error: %v", err)
	}
	defer sc.Close()

	_, err = sc.Subscribe("orders", func(m *stan.Msg) {
		var order model.Order
		if err := json.Unmarshal(m.Data, &order); err != nil {
			log.Printf("Invalid data received: %v", err)
			m.Ack()
			return
		}

		if order.OrderUID == "" {
			log.Println("Received order without UID")
			m.Ack()
			return
		}

		if err := storage.SaveOrder(order); err != nil {
			log.Printf("Failed to save order to DB: %v", err)
			return
		}

		localCache.Set(order)

		m.Ack()
		log.Printf("Order %s processed", order.OrderUID)
	}, stan.SetManualAckMode(), stan.DurableName("my-durable-queue"))

	if err != nil {
		log.Fatalf("Subscription error: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/index.html")
		if err != nil {
			log.Printf("Template error: %v", err)
			http.Error(w, "Template not found", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		order, err := localCache.Get(id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	
}
