package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func saveEventDB(event Event) error {
	client := retornaClientMongoDB()
	defer client.Disconnect(context.TODO())
	collection := client.Database("dito").Collection("events")
	_, err := collection.InsertOne(context.TODO(), event)
	if err != nil {
		log.Fatalln("Error on inserting new Hero", err)
		return err
	}
	return nil
}

func autoCompleteEvent(event string) ([]interface{}, error) {
	client := retornaClientMongoDB()
	defer client.Disconnect(context.TODO())
	collection := client.Database("dito").Collection("events")
	filtro := bson.M{"event": primitive.Regex{Pattern: "^" + event, Options: "i"}}
	events, err := collection.Distinct(context.TODO(), "event", filtro)
	return events, err
}
