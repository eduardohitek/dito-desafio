package main

import (
	"errors"
	"time"
)

// Event represent an object with a name and a timestamp
type Event struct {
	Event         string       `json:"event"`
	Timestamp     time.Time    `json:"timestamp"`
	Revenue       float64      `json:"revenue"`
	CustomData    []CustomData `json:"custom_data,omitempty"`
	customDataMap CustomDataMap
}


type CustomDataMap map[string]interface{}

type ProductGroup map[string][]Product

type Events struct {
	Events []Event `json:"events"`
}

type CustomData struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (e Event) validate() error {
	if e.Event == "" {
		return errors.New("Field event should no be empty")
	}
	if e.Timestamp.IsZero() {
		return errors.New("Field timestamp should no be empty")
	}
	return nil
}

type Timeline struct {
	Timeline []EventGrouped `json:"timeline"`
}

func (t Timeline) Len() int {
	return len(t.Timeline)
}

func (t Timeline) Less(i, j int) bool {
	return t.Timeline[i].Timestamp.After(t.Timeline[j].Timestamp)
}

func (t Timeline) Swap(i, j int) {
	t.Timeline[i], t.Timeline[j] = t.Timeline[j], t.Timeline[i]
}

type EventGrouped struct {
	Timestamp     time.Time `json:"timestamp"`
	Revenue       float64   `json:"revenue"`
	TransactionID string    `json:"transaction_id"`
	StoreName     string    `json:"store_name"`
	Products      []Product `json:"products"`
}

type Product struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
