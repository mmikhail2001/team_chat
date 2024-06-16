package database

import (
	"context"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func (db *Database) SearchGroupChannels(query string, currentUserID string) ([]Channel, error) {
	var channels []Channel

	groupChatsFilter := bson.M{
		"type":       2,
		"name":       bson.M{"$regex": query, "$options": "i"},
		"recipients": bson.M{"$elemMatch": bson.M{"$eq": currentUserID}},
	}
	cursor, err := db.Mongo.Collection("channels").Find(context.TODO(), groupChatsFilter)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.TODO(), &channels)
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func (db *Database) SearchPersonalChannels(query string, currentUserID string) ([]Channel, error) {
	var channels []Channel
	var personalChats []Channel
	personalChatsFilter := bson.M{"type": 1, "recipients": currentUserID}
	cursor, err := db.Mongo.Collection("channels").Find(context.TODO(), personalChatsFilter)
	if err != nil {
		return []Channel{}, err
	}

	err = cursor.All(context.TODO(), &personalChats)
	if err != nil {
		return nil, err
	}

	for _, chat := range personalChats {
		for _, recipientID := range chat.Recipients {
			if recipientID != currentUserID {
				user, status := db.GetUser(recipientID)
				if status == http.StatusOK && strings.Contains(strings.ToLower(user.Username), strings.ToLower(query)) {
					channels = append(channels, chat)
					break
				}
			}
		}
	}
	return channels, nil
}
