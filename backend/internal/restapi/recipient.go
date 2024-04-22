package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/request"
	"github.com/mmikhail2001/team_chat/internal/response"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	INVITE_JOIN      = "%s has joined using the invite."
	ADD_RECIPIENT    = "%s added %s to the channel."
	REMOVE_RECIPIENT = "%s %s %s from the channel."
	REASON           = "Reason: %s."
	RECIPIENT_LEAVE  = "%s has left the channel."
)

func AddRecipient(ctx *Context) {
	log.Println("handler AddRecipient")
	url_vars := mux.Vars(ctx.Req)
	channel_id := url_vars["id"]
	user_id := url_vars["uid"]

	channelCollection := ctx.Db.Mongo.Collection("channels")

	// channel, statusCode := ctx.Db.GetChannel(channel_id, &ctx.User)
	channel, statusCode := ctx.Db.GetChannelWithoutUser(channel_id)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	if !(channel.Type == 4 || channel.Type == 5) {
		if channel.Type == 1 || channel.OwnerID != ctx.User.ID {
			ctx.Res.WriteHeader(http.StatusForbidden)
			return
		}
	}

	user, statusCode := ctx.Db.GetUser(user_id)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	for _, rid := range channel.Recipients {
		if rid == user.ID {
			ctx.Res.WriteHeader(http.StatusNotAcceptable)
			return
		}
	}

	rd := options.After
	result := channelCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": channel.ID}, bson.M{"$push": bson.M{"recipients": user.ID}}, &options.FindOneAndUpdateOptions{ReturnDocument: &rd})
	if result.Err() != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result.Decode(&channel)

	recipients := []response.User{}
	for _, recipient := range channel.Recipients {
		recipient, _ := ctx.Db.GetUser(recipient)
		recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
	}

	res_channel := response.NewChannel(channel, recipients)
	ctx.WriteJSON(res_channel)
	ctx.Conn.AddUserToChannel(user_id, channel_id)
	ctx.Conn.BroadcastToChannel(res_channel.ID, "CHANNEL_MODIFY", res_channel)

	add := fmt.Sprintf(ADD_RECIPIENT, ctx.User.Username, user.Username)
	message, statusCode := ctx.Db.CreateMessage(add, channel_id, true, nil)

	if statusCode != http.StatusOK {
		return
	}

	res_message := response.NewMessage(message, response.User{})
	ctx.Conn.BroadcastToChannel(channel_id, "MESSAGE_CREATE", res_message)
}

func AddRoleToChannel(ctx *Context) {
	log.Println("handler AddRecipient")
	url_vars := mux.Vars(ctx.Req)
	channel_id := url_vars["id"]
	role := url_vars["role"]

	channelCollection := ctx.Db.Mongo.Collection("channels")

	// channel, statusCode := ctx.Db.GetChannel(channel_id, &ctx.User)
	channel, statusCode := ctx.Db.GetChannelWithoutUser(channel_id)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	if !(channel.Type == 4 || channel.Type == 5) {
		if channel.Type == 1 || channel.OwnerID != ctx.User.ID {
			ctx.Res.WriteHeader(http.StatusForbidden)
			return
		}
	}

	rolesCollection := ctx.Db.Mongo.Collection("roles")
	cursor, err := rolesCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while fetching roles:", err)
		return
	}

	// Находим ID роли по названию роли запроса
	var roles []database.Role
	err = cursor.All(context.TODO(), &roles)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while decoding roles:", err)
		return
	}

	var roleID primitive.ObjectID
	for _, r := range roles {
		if r.Name == role {
			roleID = r.ID
			break
		}
	}

	usersCollection := ctx.Db.Mongo.Collection("users")
	cursor, err = usersCollection.Find(context.TODO(), bson.M{"roles": roleID})
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while fetching users with role:", err)
		return
	}

	var users []database.User
	err = cursor.All(context.TODO(), &users)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Error while decoding users with role:", err)
		return
	}

	// Добавляем пользователей в канал, если они еще не добавлены
	for _, u := range users {
		var isRecipientExist bool
		for _, rid := range channel.Recipients {
			if rid == u.ID {
				isRecipientExist = true
				break
			}
		}
		if !isRecipientExist {
			rd := options.After
			result := channelCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": channel.ID}, bson.M{"$push": bson.M{"recipients": u.ID}}, &options.FindOneAndUpdateOptions{ReturnDocument: &rd})
			if result.Err() != nil {
				ctx.Res.WriteHeader(http.StatusInternalServerError)
				log.Println("Error while adding user to channel:", result.Err())
				return
			}
		}
	}

	channel, statusCode = ctx.Db.GetChannelWithoutUser(channel_id)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	recipients := []response.User{}
	for _, rid := range channel.Recipients {
		// Получаем данные добавленного пользователя
		recipient, statusCode := ctx.Db.GetUser(rid)
		if statusCode != http.StatusOK {
			ctx.Res.WriteHeader(http.StatusInternalServerError)
			log.Println("Error while getting user:", err)
			return
		}

		// Создаем объект ответа для пользователя
		recipientResponse := response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID))

		// Добавляем пользователя к списку получателей
		recipients = append(recipients, recipientResponse)
	}

	// Создаем объект ответа для канала
	resChannel := response.NewChannel(channel, recipients)

	// Отправляем обновление канала всем участникам
	for _, recipient := range recipients {
		if recipient.ID != ctx.User.ID {
			ctx.Conn.AddUserToChannel(recipient.ID, channel_id)
			addMessage := fmt.Sprintf(ADD_RECIPIENT, ctx.User.Username, recipient.Username)
			message, statusCode := ctx.Db.CreateMessage(addMessage, channel_id, true, nil)
			if statusCode != http.StatusOK {
				ctx.Res.WriteHeader(http.StatusInternalServerError)
				log.Println("Error while creating message:", statusCode)
				return
			}
			// Создаем объект ответа для сообщения
			resMessage := response.NewMessage(message, response.User{})
			// Отправляем сообщение о добавлении пользователя в канал
			ctx.Conn.BroadcastToChannel(channel_id, "MESSAGE_CREATE", resMessage)
		}
	}
	// Отправляем сообщение о добавлении пользователя в канал
	ctx.Conn.BroadcastToChannel(resChannel.ID, "CHANNEL_MODIFY", resChannel)

}

func RemoveRecipient(ctx *Context) {
	log.Println("handler RemoveRecipient")
	url_vars := mux.Vars(ctx.Req)
	channel_id := url_vars["id"]
	user_id := url_vars["uid"]

	var request_ request.RemoveRecipient
	_ = json.NewDecoder(ctx.Req.Body).Decode(&request_)

	isBan := request_.IsBan
	reason := strings.TrimSpace(request_.Reason)

	channelCollection := ctx.Db.Mongo.Collection("channels")
	bansCollection := ctx.Db.Mongo.Collection("bans")

	// channel, statusCode := ctx.Db.GetChannel(channel_id, &ctx.User)
	channel, statusCode := ctx.Db.GetChannelWithoutUser(channel_id)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	if !(channel.Type == 4 || channel.Type == 5) {
		if channel.Type == 1 || channel.OwnerID != ctx.User.ID {
			ctx.Res.WriteHeader(http.StatusForbidden)
			return
		}
	}

	user, statusCode := ctx.Db.GetUser(user_id)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	if isBan {
		ban := database.Ban{
			ID:         primitive.NewObjectID(),
			BannedUser: user.ID,
			ChannelID:  channel.ID,
			BannedBy:   ctx.User.ID,
			Reason:     reason,
			CreatedAt:  time.Now().Unix(),
		}

		_, err := bansCollection.InsertOne(context.TODO(), ban)
		if err != nil {
			ctx.Res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	rd := options.After
	result := channelCollection.FindOneAndUpdate(context.TODO(), bson.M{"_id": channel.ID}, bson.M{"$pull": bson.M{"recipients": user.ID}}, &options.FindOneAndUpdateOptions{ReturnDocument: &rd})
	if result.Err() != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result.Decode(&channel)
	recipients := []response.User{}
	for _, recipient := range channel.Recipients {
		recipient, _ := ctx.Db.GetUser(recipient)
		recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
	}

	res_channel := response.NewChannel(channel, recipients)
	ctx.WriteJSON(res_channel)

	ctx.Conn.RemoveUserFromChannel(user.ID, channel.ID.Hex())
	ctx.Conn.BroadcastToChannel(channel.ID.Hex(), "CHANNEL_MODIFY", res_channel)
	ctx.Conn.SendToUser(user.ID, "CHANNEL_DELETE", res_channel)

	kickorban := "kicked"
	if isBan {
		kickorban = "banned"
	}

	remove := fmt.Sprintf(REMOVE_RECIPIENT, ctx.User.Username, kickorban, user.Username)

	if reason != "" {
		remove = fmt.Sprint(remove, " ", fmt.Sprintf(REASON, reason))
	}

	message, statusCode := ctx.Db.CreateMessage(remove, channel_id, true, nil)

	if statusCode != http.StatusOK {
		return
	}

	res_message := response.NewMessage(message, response.User{})
	ctx.Conn.BroadcastToChannel(channel_id, "MESSAGE_CREATE", res_message)
}
