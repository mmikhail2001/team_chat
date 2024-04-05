package database

import (
	"context"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func (db *Database) SearchChannels(query string, currentUserID string) ([]Channel, error) {
	var channels []Channel

	groupChatsFilter := bson.M{"type": 2, "name": bson.M{"$regex": query, "$options": "i"}}
	cursor, err := db.Mongo.Collection("channels").Find(context.TODO(), groupChatsFilter)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.TODO(), &channels)
	if err != nil {
		return nil, err
	}

	// var personalChats []Channel
	// personalChatsFilter := bson.M{"type": 1, "recipients": currentUserID}
	// cursor, err = db.Mongo.Collection("channels").Find(context.TODO(), personalChatsFilter)
	// if err != nil {
	// 	return nil, err
	// }

	// err = cursor.All(context.TODO(), &personalChats)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, chat := range personalChats {
	// 	for _, recipientID := range chat.Recipients {
	// 		if recipientID != currentUserID {
	// 			user, status := db.GetUser(recipientID)
	// 			if status == http.StatusOK && strings.Contains(strings.ToLower(user.Username), strings.ToLower(query)) {
	// 				channels = append(channels, chat)
	// 				break
	// 			}
	// 		}
	// 	}
	// }
	return channels, nil
}

func (db *Database) SearchUsers(query string, currentUserID string) ([]Channel, error) {
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
	// for _, channel := range channels {
	// 	recipients := []response.User{}
	// 	for _, recipient := range channel.Recipients {
	// 		// почему самому себе не отправляем только в личных чатах?
	// 		if channel.Type == 1 && recipient == user.ID {
	// 			continue
	// 		}
	// 		recipient, _ := ws.Db.GetUser(recipient)
	// 		recipients = append(recipients, response.NewUser(recipient, ws.Conns.GetUserStatus(recipient.ID)))
	// 	}
	// 	res_channels = append(res_channels, response.NewChannel(&channel, recipients))
	// }
}
