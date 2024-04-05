package database

import (
	"context"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func (db *Database) GetUser(id string) (*User, int) {
	var user User
	users := db.Mongo.Collection("users")

	err := users.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		log.Println("GetUser: ", err)
		return nil, http.StatusNotFound
	}

	return &user, http.StatusOK
}

func (db *Database) GetUserByEmail(mail string) (*User, int) {
	var user User
	users := db.Mongo.Collection("users")

	err := users.FindOne(context.TODO(), bson.M{"email": mail}).Decode(&user)
	if err != nil {
		log.Println("GetUserByEmail: ", err)
		return nil, http.StatusNotFound
	}

	return &user, http.StatusOK
}
