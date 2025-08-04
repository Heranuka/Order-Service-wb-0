package tests

import (
	"encoding/json"
	"log"
	"wb-l0/internal/domain"
)

var (
	InstanceStruct = domain.Order{
		OrderUID:    "1b563feb7b2b84b6testCase",
		TrackNumber: "1WBILMtestCaseTRACK",
		Entry:       "WBIL",
		Delivery: domain.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: domain.Payment{
			Transaction:  "b563feb7b2b84b6testCase",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
		},
		Items: []domain.Items{
			{
				ChrtID:     9934930,
				Price:      453,
				Rid:        "ab4219087a764ae0btestCase",
				Name:       "Mascaras",
				Sale:       30,
				Size:       "0",
				TotalPrice: 317,
				NmID:       2389212,
				Brand:      "Vivienne Sabo",
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "testCase",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
	}
	InstanceString string

	InstanceKafka = `{"order_uid":"b563feb7b2b84b6test","track_number":"WBILMTESTTRACK","entry":"WBIL","delivery":{"name":"Test Testov","phone":"+9720000000","zip":"2639809","city":"Kiryat Mozkin","address":"Ploshad Mira 15","region":"Kraiot","email":"test@gmail.com"},"payment":{"transaction":"b563feb7b2b84b6test","request_id":"","currency":"USD","provider":"wbpay","amount":1817,"payment_dt":1637907727,"bank":"alpha","delivery_cost":1500,"goods_total":317,"custom_fee":0},"items":[{"chrt_id":9934930,"track_number":"WBILMTESTTRACK","price":453,"rid":"ab4219087a764ae0btest","name":"Mascaras","sale":30,"size":"0","total_price":317,"nm_id":2389212,"brand":"Vivienne Sabo","status":202}],"locale":"en","internal_signature":"","customer_id":"test","delivery_service":"meest","shardkey":"9","sm_id":99,"date_created":"2021-11-26T06:22:19Z","oof_shard":"1"}`
)

func init() {
	orderJSON, err := json.Marshal(InstanceStruct)
	if err != nil {
		log.Fatalf("failed to marshal InstanceStruct: %s", err)
	}
	InstanceString = string(orderJSON)
}
