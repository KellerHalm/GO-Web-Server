package main

import (
	"encoding/json"
	"log"
	"time"
	"wb-l0/internal/model"

	"github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("test-cluster", "publisher-client", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	order := model.Order{
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: model.Delivery{
			Name:  "Test User",
			Phone: "+9720000000",
			Zip:   "2639809",
			City:  "Kiryat Mozkin",
		},
		Payment: model.Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:  9934930,
				TrackID: "WBILMTESTTRACK",
				Price:   453,
				Rid:     "ab4219087a7642122000061238",
				Name:    "Mascaras",
				Sale:    30,
				Size:    "0",
			},
		},
		Locale:      "en",
		InternalSig: "",
		CustomerID:  "test",
		DeliveryService: "meest",
		Shardkey:    "9",
		SmID:        99,
		DateCreated: "2021-11-26T06:22:19Z",
		OofShard:    "1",
	}

	data, _ := json.Marshal(order)

	err = sc.Publish("orders", data)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Published valid order")

	sc.Publish("orders", []byte("{invalid-json}"))
	log.Println("Published garbage")
	
	time.Sleep(1 * time.Second)
}
