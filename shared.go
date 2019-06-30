package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func retornaClientMongoDB() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://mongo:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// Response is a struct that represents a http response body
type Response map[string]interface{}

// JSON is method to send a JSON http response
func JSON(w http.ResponseWriter, reposta Response, codigo int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	r, _ := json.Marshal(reposta)
	w.Write(r)
}
