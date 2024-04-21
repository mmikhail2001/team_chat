package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mmikhail2001/team_chat/internal/response"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (ws *Ws) ConnectUser() {
	user := ws.User
	res_user := response.NewUser(user, 1)

	log.Printf("%s Connected\n", user.Username)
	ws.Conns.AddUser(user.ID, ws)
	res_channels := []response.Channel{}
	channels := ws.Db.GetChannels(user)
	for _, channel := range channels {
		recipients := []response.User{}
		for _, recipient := range channel.Recipients {
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

	relationships := ws.Db.GetRelationships(user.ID)
	for _, relationship := range relationships {
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
	// нужно вернуть только те личные каналы, в которых есть хотя бы одно сообщение, и все групповые каналы
	var personalChatIDs []primitive.ObjectID
	for _, chat := range res_channels {
		chatID, _ := primitive.ObjectIDFromHex(chat.ID)
		personalChatIDs = append(personalChatIDs, chatID)
	}

	messagesMap := make(map[string]bool)
	if len(personalChatIDs) != 0 {
		messagesCollection := ws.Db.Mongo.Collection("messages")
		filter := bson.M{"channel_id": bson.M{"$in": personalChatIDs}}
		cursor, err := messagesCollection.Find(context.TODO(), filter)
		if err != nil {
			log.Println("find personalChatIDs err:", err)
			return
		}
		defer cursor.Close(context.Background())

		// map[channel_id]true - если у канала channel_id есть сообщения
		for cursor.Next(context.Background()) {
			var message bson.M
			cursor.Decode(&message)
			channelID := message["channel_id"].(primitive.ObjectID)
			messagesMap[channelID.Hex()] = true
		}
	}

	var filteredChannels []response.Channel
	for _, chat := range res_channels {
		if messagesMap[chat.ID] || chat.Type == 2 || chat.Type == 3 || chat.Type == 4 || chat.Type == 5 {
			filteredChannels = append(filteredChannels, chat)
		}
	}

	ws_msg := WS_Message{
		Event: "READY",
		Data: Ready{
			User:     res_user,
			Channels: filteredChannels,
		},
	}

	ws_res, _ := json.Marshal(ws_msg)
	ws.Write(ws_res)
}
