package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
)

func handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		event, err := retrieveEvent(r)
		if err != nil {
			JSON(w, Response{"msg": "Error on retrieving the Event from request", "err": err.Error()}, http.StatusBadRequest)
			return
		}
		err = event.validate()
		if err != nil {
			JSON(w, Response{"msg": "Error on validating the Event from request", "err": err.Error()}, http.StatusBadRequest)
			return
		}
		err = saveEventDB(event)
		if err != nil {
			JSON(w, Response{"msg": "Error on saving the Event from request", "err": err.Error()}, http.StatusInternalServerError)
			return
		}
		JSON(w, Response{"msg": "Event saved successfully!"}, http.StatusOK)
		return
	} else if r.Method == "GET" {
		query := r.URL.Query()
		if query != nil {
			eventQuery := query.Get("event")
			if eventQuery != "" && len(eventQuery) > 1 {
				events, err := autoCompleteEvent(eventQuery)
				if err != nil {
					JSON(w, Response{"msg": "Error on searching the Event autocomplete", "err": err.Error()}, http.StatusInternalServerError)
					return
				}
				JSON(w, Response{"events": events}, http.StatusOK)
			}
			return
		}
		return
	}
}

func categorizeEvents(events Events) (comprou []Event, comprouProduto []Event) {
	var c []Event
	var cp []Event
	for _, event := range events.Events {
		event.customDataMap = make(map[string]interface{})
		for _, produto := range event.CustomData {
			event.customDataMap[produto.Key] = produto.Value
		}
		if event.Event == "comprou" {
			c = append(c, event)
		} else {
			cp = append(cp, event)
		}
	}
	return c, cp
}

func groupProducts(comprouProduto []Event) ProductGroup {
	productsGrouped := ProductGroup{}
	for _, event := range comprouProduto {
		p := Product{Name: event.customDataMap["product_name"].(string),
			Price: event.customDataMap["product_price"].(float64)}
		productsGrouped[event.customDataMap["transaction_id"].(string)] = append(productsGrouped[event.customDataMap["transaction_id"].(string)], p)
	}
	return productsGrouped
}

func groupTimeline(comprou []Event, productsGrouped ProductGroup) Timeline {
	var timeline Timeline
	for _, event := range comprou {
		e := EventGrouped{
			Timestamp:     event.Timestamp,
			Revenue:       event.Revenue,
			TransactionID: event.customDataMap["transaction_id"].(string),
			StoreName:     event.customDataMap["store_name"].(string),
			Products:      productsGrouped[event.customDataMap["transaction_id"].(string)],
		}
		timeline.Timeline = append(timeline.Timeline, e)
	}
	return timeline
}

func sortTimeline(timeline Timeline) Timeline {
	sort.Sort(timeline)
	return timeline
}

func groupEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		req, _ := http.NewRequest("GET", "https://storage.googleapis.com/dito-questions/events.json", nil)
		client := &http.Client{}
		response, _ := client.Do(req)
		body, err := ioutil.ReadAll(io.LimitReader(response.Body, 1048576))
		if err != nil {
			JSON(w, Response{"msg": "Error on retrieving the Events from request", "err": err.Error()}, http.StatusInternalServerError)
			return
		}

		var events Events
		err = json.Unmarshal(body, &events)
		if err != nil {
			JSON(w, Response{"msg": "Error on parsing the Events from request", "err": err.Error()}, http.StatusInternalServerError)
			return
		}
		comprou, comprouProduto := categorizeEvents(events)
		productsGrouped := groupProducts(comprouProduto)

		timeline := groupTimeline(comprou, productsGrouped)

		timeline = sortTimeline(timeline)
		JSON(w, Response{"timeline": timeline}, http.StatusOK)
		return
	}

}

func main() {
	http.HandleFunc("/event", handleEvent)
	http.HandleFunc("/groupEvents", groupEvents)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error on starting the server", err.Error())
	}
}

func retrieveEvent(r *http.Request) (Event, error) {
	event := Event{}
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Println("Error on retrieving the data from the Request", err)
		return event, err
	}
	err = json.Unmarshal(body, &event)
	if err != nil {
		log.Println("Error on transforming the data from the Request", err)
		return event, err
	}
	return event, err
}
