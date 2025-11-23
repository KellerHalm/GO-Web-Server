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
	// 1. Инициализация БД
	storage, err := db.New(dbConn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	
	// 2. Инициализация Кэша и восстановление данных
	localCache := cache.New()
	orders, err := storage.LoadOrders()
	if err != nil {
		log.Fatalf("Failed to restore cache: %v", err)
	}
	for _, order := range orders {
		localCache.Set(order)
	}
	log.Printf("Restored %d orders from DB to Cache", len(orders))

	// 3. Подключение к NATS Streaming
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatalf("NATS connection error: %v", err)
	}
	defer sc.Close()

	// 4. Подписка на канал
	_, err = sc.Subscribe("orders", func(m *stan.Msg) {
		var order model.Order
		// Валидация: пробуем распарсить JSON
		if err := json.Unmarshal(m.Data, &order); err != nil {
			log.Printf("Invalid data received: %v", err)
			// Можно подтвердить сообщение (Ack), чтобы не блокировать очередь,
			// так как этот JSON мы всё равно не обработаем.
			m.Ack() 
			return
		}

		// Дополнительная валидация на пустоту ID
		if order.OrderUID == "" {
			log.Println("Received order without UID")
			m.Ack()
			return
		}

		// Сохраняем в БД
		if err := storage.SaveOrder(order); err != nil {
			log.Printf("Failed to save order to DB: %v", err)
			// Если БД упала, мы НЕ делаем Ack, чтобы NATS переслал сообщение позже
			return 
		}

		// Обновляем кэш
		localCache.Set(order)
		
		// Подтверждаем получение
		m.Ack()
		log.Printf("Order %s processed", order.OrderUID)
	}, stan.SetManualAckMode(), stan.DurableName("my-durable-queue")) 
	// DurableName гарантирует, что NATS запомнит позицию, если сервис упадет

	if err != nil {
		log.Fatalf("Subscription error: %v", err)
	}

	// 5. HTTP Сервер
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("../../web/index.html")
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
