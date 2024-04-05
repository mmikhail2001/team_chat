package database

import (
	"context"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func (db *Database) GetRelationship(from string, to string) (*Relationship, int) {
	relationshipsCollection := db.Mongo.Collection("relationships")

	var relationship Relationship
	err := relationshipsCollection.FindOne(context.TODO(), bson.M{"from_user_id": from, "to_user_id": to}).Decode(&relationship)
	if err != nil {
		log.Println("GetRelationship: FindOne: ", err)
		return nil, http.StatusNotFound
	}

	return &relationship, http.StatusOK
}

func (db *Database) GetRelationships(user_id string) []Relationship {
	relationshipsCollection := db.Mongo.Collection("relationships")

	cursor, err := relationshipsCollection.Find(context.TODO(), bson.M{"from_user_id": user_id})
	if err != nil {
		log.Println("GetRelationships: Find: ", err)
		return []Relationship{}
	}

	var relationships []Relationship

	for cursor.Next(context.TODO()) {
		var relationship Relationship
		cursor.Decode(&relationship)

		relationships = append(relationships, relationship)
	}

	return relationships
}
