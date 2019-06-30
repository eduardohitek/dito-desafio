package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func createRqRs(t *testing.T, body interface{}, method string, url string) (*http.Request, *httptest.ResponseRecorder) {
	var req *http.Request
	var err error
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader := bytes.NewReader(bodyBytes)
		req, _ = http.NewRequest(method, url, bodyReader)
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	return req, rr
}

func checkStatus(t *testing.T, rr *httptest.ResponseRecorder, expectedStatusCode int) {
	if status := rr.Code; status != expectedStatusCode {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, expectedStatusCode)
	}
}

func checkResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedResponse string) {
	if rr.Body.String() != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}
}

func Test_handleEventOK(t *testing.T) {
	event := Event{Event: "test-1", Timestamp: time.Now()}

	rr := httptest.NewRecorder()
	req, rr := createRqRs(t, event, "POST", "/event")

	handler := http.HandlerFunc(handleEvent)
	handler.ServeHTTP(rr, req)

	checkStatus(t, rr, http.StatusOK)

	expected := `{"msg":"Event saved successfully!"}`
	checkResponse(t, rr, expected)

}

func Test_handleEventNameEmpty(t *testing.T) {
	event := Event{Timestamp: time.Now()}
	req, rr := createRqRs(t, event, "POST", "/event")

	handler := http.HandlerFunc(handleEvent)
	handler.ServeHTTP(rr, req)

	checkStatus(t, rr, http.StatusBadRequest)

	expected := `{"err":"Field event should no be empty","msg":"Error on validating the Event from request"}`
	checkResponse(t, rr, expected)
}

func Test_handleEventTSEmpty(t *testing.T) {
	event := Event{Event: "test-1"}
	req, rr := createRqRs(t, event, "POST", "/event")

	handler := http.HandlerFunc(handleEvent)
	handler.ServeHTTP(rr, req)

	checkStatus(t, rr, http.StatusBadRequest)

	expectedResponse := `{"err":"Field timestamp should no be empty","msg":"Error on validating the Event from request"}`
	checkResponse(t, rr, expectedResponse)

}

func Test_handleEventAC(t *testing.T) {

	req, rr := createRqRs(t, nil, "GET", "/event?event=re")
	handler := http.HandlerFunc(handleEvent)

	handler.ServeHTTP(rr, req)

	checkStatus(t, rr, http.StatusOK)

	// Check the response body is what we expect.
	expected := `{"events":["resell"]}`
	checkResponse(t, rr, expected)
}

func Test_handleEventEmpty(t *testing.T) {

	req, rr := createRqRs(t, nil, "GET", "/event")
	handler := http.HandlerFunc(handleEvent)
	handler.ServeHTTP(rr, req)

	checkStatus(t, rr, http.StatusOK)
	expected := ``
	checkResponse(t, rr, expected)
}

func Test_categorizeEvents(t *testing.T) {
	timeStampEvent1 := time.Now()
	event1 := Event{Event: "comprou-produto",
		Timestamp: timeStampEvent1,
		CustomData: []CustomData{
			CustomData{Key: "product_name", Value: "Camisa Azul"},
			CustomData{Key: "transaction_id", Value: "3029384"},
			CustomData{Key: "product_price", Value: 100},
		},
		customDataMap: CustomDataMap{"product_name": "Camisa Azul", "transaction_id": "3029384", "product_price": 100},
	}
	timeStampEvent2 := time.Now()
	event2 := Event{Event: "comprou",
		Timestamp: timeStampEvent2,
		CustomData: []CustomData{
			CustomData{Key: "store_name", Value: "Patio Savassi"},
			CustomData{Key: "transaction_id", Value: "3029384"},
		},
		customDataMap: CustomDataMap{"store_name": "Patio Savassi", "transaction_id": "3029384"},
	}
	type args struct {
		events Events
	}
	tests := []struct {
		name               string
		args               args
		wantComprou        []Event
		wantComprouProduto []Event
	}{
		{name: "Test1",
			args:               args{events: Events{Events: []Event{event1, event2}}},
			wantComprou:        []Event{event2},
			wantComprouProduto: []Event{event1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotComprou, gotComprouProduto := categorizeEvents(tt.args.events)
			if !reflect.DeepEqual(gotComprou, tt.wantComprou) {
				t.Errorf("categorizeEvents() gotComprou = %v, want %v", gotComprou, tt.wantComprou)
			}
			if !reflect.DeepEqual(gotComprouProduto, tt.wantComprouProduto) {
				t.Errorf("categorizeEvents() gotComprouProduto = %v, want %v", gotComprouProduto, tt.wantComprouProduto)
			}
		})
	}
}

func Test_groupProducts(t *testing.T) {
	timeStampEvent1 := time.Now()
	event1 := Event{Event: "comprou-produto",
		Timestamp: timeStampEvent1,
		CustomData: []CustomData{
			CustomData{Key: "product_name", Value: "Camisa Azul"},
			CustomData{Key: "transaction_id", Value: "3029384"},
			CustomData{Key: "product_price", Value: float64(100)},
		},
		customDataMap: CustomDataMap{"product_name": "Camisa Azul", "transaction_id": "3029384", "product_price": float64(100)},
	}
	type args struct {
		comprouProduto []Event
	}
	tests := []struct {
		name string
		args args
		want ProductGroup
	}{
		{name: "Test1",
			args: args{comprouProduto: []Event{event1}},
			want: ProductGroup{"3029384": []Product{Product{Name: "Camisa Azul", Price: float64(100)}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := groupProducts(tt.args.comprouProduto); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_groupTimeline(t *testing.T) {
	timeStampEvent2 := time.Now()
	event2 := Event{Event: "comprou",
		Timestamp: timeStampEvent2,
		Revenue:   float64(100),
		CustomData: []CustomData{
			CustomData{Key: "store_name", Value: "Patio Savassi"},
			CustomData{Key: "transaction_id", Value: "3029384"},
		},
		customDataMap: CustomDataMap{"store_name": "Patio Savassi", "transaction_id": "3029384"},
	}
	productGroup := ProductGroup{"3029384": []Product{
		Product{
			Name:  "Camisa Azul",
			Price: float64(100)}},
	}
	type args struct {
		comprou         []Event
		productsGrouped ProductGroup
	}
	tests := []struct {
		name string
		args args
		want Timeline
	}{
		{name: "Test1", args: args{
			comprou:         []Event{event2},
			productsGrouped: productGroup},
			want: Timeline{
				Timeline: []EventGrouped{
					EventGrouped{
						Timestamp:     timeStampEvent2,
						Revenue:       float64(100),
						TransactionID: "3029384",
						StoreName:     "Patio Savassi",
						Products:      []Product{Product{Name: "Camisa Azul", Price: float64(100)}}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := groupTimeline(tt.args.comprou, tt.args.productsGrouped); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupTimeline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sortTimeline(t *testing.T) {
	event1 := EventGrouped{Timestamp: time.Now()}
	time.Sleep(200 * time.Millisecond)
	event2 := EventGrouped{Timestamp: time.Now()}

	type args struct {
		timeline Timeline
	}
	tests := []struct {
		name string
		args args
		want Timeline
	}{
		{name: "Test1",
			args: args{timeline: Timeline{Timeline: []EventGrouped{event1, event2}}},
			want: Timeline{Timeline: []EventGrouped{event2, event1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortTimeline(tt.args.timeline); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortTimeline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_groupEvents(t *testing.T) {

	req, rr := createRqRs(t, nil, "GET", "/groupEvents")
	handler := http.HandlerFunc(groupEvents)
	handler.ServeHTTP(rr, req)

	checkStatus(t, rr, http.StatusOK)
	expected := `{"timeline":{"timeline":[{"timestamp":"2016-10-02T11:37:31.2300892-03:00","revenue":120,"transaction_id":"3409340","store_name":"BH Shopping","products":[{"name":"Tenis Preto","price":120}]},{"timestamp":"2016-09-22T13:57:31.2311892-03:00","revenue":250,"transaction_id":"3029384","store_name":"Patio Savassi","products":[{"name":"Camisa Azul","price":100},{"name":"Calça Rosa","price":150}]}]}}`
	checkResponse(t, rr, expected)
}
