package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mmikhail2001/team_chat/internal/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: почему нельзя было передать в параметре пользователя
func (ws *Ws) ConnectUser() {
	// почему response ?
	log.Println("ws ConnectUser")

	user := ws.User
	res_user := response.NewUser(user, 1)

	log.Printf("%s Connected\n", user.Username)
	ws.Conns.AddUser(user.ID, ws)
	fmt.Println(1)
	res_channels := []response.Channel{}
	channels := ws.Db.GetChannels(user)
	for _, channel := range channels {
		recipients := []response.User{}
		for _, recipient := range channel.Recipients {
			// почему самому себе не отправляем только в личных чатах?
			if channel.Type == 1 && recipient == user.ID {
				continue
			}
			recipient, _ := ws.Db.GetUser(recipient)
			recipients = append(recipients, response.NewUser(recipient, ws.Conns.GetUserStatus(recipient.ID)))
		}
		res_channels = append(res_channels, response.NewChannel(&channel, recipients))

		status := response.Status{
			UserID:    user.ID,
			Status:    1,
			Type:      1,
			ChannelID: channel.ID.Hex(),
		}
		// наверное, по поводу онлайна можно отправлять STATUS_UPDATE по каждому сотруднику всем сотрудникам
		ws.Conns.BroadcastToChannel(channel.ID.Hex(), "STATUS_UPDATE", status)
		ws.Conns.AddUserToChannel(user.ID, channel.ID.Hex())
	}
	fmt.Println(2)

	relationships := ws.Db.GetRelationships(user.ID)
	for _, relationship := range relationships {
		// TODO: где enum ??? relationship.Type
		if relationship.Type != 1 {
			continue
		}
		status := response.Status{
			UserID: user.ID,
			Status: 1,
			Type:   0,
		}
		ws.Conns.SendToUser(relationship.ToUserID, "STATUS_UPDATE", status)
	}
	var personalChatIDs []primitive.ObjectID
	for _, chat := range res_channels {
		chatID, _ := primitive.ObjectIDFromHex(chat.ID)
		personalChatIDs = append(personalChatIDs, chatID)
	}

	messagesCollection := ws.Db.Mongo.Collection("messages")
	filter := bson.M{"channel_id": bson.M{"$in": personalChatIDs}}
	cursor, _ := messagesCollection.Find(context.TODO(), filter)
	defer cursor.Close(context.Background())

	messagesMap := make(map[string]bool)
	for cursor.Next(context.Background()) {
		var message bson.M
		cursor.Decode(&message)
		channelID := message["channel_id"].(primitive.ObjectID)
		messagesMap[channelID.Hex()] = true
	}

	var filteredPersonalChats []response.Channel
	for _, chat := range res_channels {
		if messagesMap[chat.ID] || chat.Type == 2 {
			filteredPersonalChats = append(filteredPersonalChats, chat)
		}
	}

	ws_msg := WS_Message{
		Event: "READY",
		Data: Ready{
			User:     res_user,
			Channels: filteredPersonalChats,
		},
	}

	fmt.Println(4)
	ws_res, _ := json.Marshal(ws_msg)
	ws.Write(ws_res)
}

// func (db *Database) SearchChannels(query string, currentUserID string) ([]Channel, error) {
// 	var channels []Channel

// 	// Фильтрация групповых чатов по запросу
// 	groupChatsFilter := bson.M{"type": 2, "name": bson.M{"$regex": query, "$options": "i"}}
// 	cursor, err := db.Mongo.Collection("channels").Find(context.TODO(), groupChatsFilter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = cursor.All(context.TODO(), &channels)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Фильтрация персональных чатов
// 	var personalChats []Channel
// 	personalChatsFilter := bson.M{"type": 1, "recipients": currentUserID}
// 	cursor, err = db.Mongo.Collection("channels").Find(context.TODO(), personalChatsFilter)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = cursor.All(context.TODO(), &personalChats)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Получение списка идентификаторов персональных чатов с соответствующими сообщениями
// 	var personalChatIDs []primitive.ObjectID
// 	for _, chat := range personalChats {
// 		personalChatIDs = append(personalChatIDs, chat.ID)
// 	}

// 	// Фильтрация персональных чатов по наличию сообщений
// 	messagesCollection := db.Mongo.Collection("messages")
// 	filter := bson.M{"channel_id": bson.M{"$in": personalChatIDs}}
// 	cursor, err = messagesCollection.Find(context.TODO(), filter)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Создание карты для отслеживания наличия сообщений в персональных чатах
// 	messagesMap := make(map[primitive.ObjectID]bool)
// 	for cursor.Next(context.Background()) {
// 		var message bson.M
// 		if err := cursor.Decode(&message); err != nil {
// 			return nil, err
// 		}
// 		channelID := message["channel_id"].(primitive.ObjectID)
// 		messagesMap[channelID] = true
// 	}

// 	// Фильтрация персональных чатов по наличию сообщений и добавление в итоговый список
// 	var filteredPersonalChats []Channel
// 	for _, chat := range personalChats {
// 		if messagesMap[chat.ID] {
// 			filteredPersonalChats = append(filteredPersonalChats, chat)
// 		}
// 	}

// 	// Добавление отфильтрованных персональных чатов к общему списку каналов
// 	channels = append(channels, filteredPersonalChats...)

// 	return channels, nil
// }
