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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/mux"
)

func CreateChannel(ctx *Context) {
	log.Println("handler CreateChannel")
	var request request.Channel
	_ = json.NewDecoder(ctx.Req.Body).Decode(&request)

	name := strings.TrimSpace(request.Name)
	icon := strings.TrimSpace(request.Icon)
	recipientID := strings.TrimSpace(request.RecipientID)

	if name == "" && recipientID == "" {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	channel, statusCode := ctx.Db.CreateChannel(name, icon, recipientID, &ctx.User, request.IsNews)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	recipients := []response.User{}
	if channel.Type == 1 {
		recipient, _ := ctx.Db.GetUser(recipientID)
		recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
	} else {
		recipients = append(recipients, response.NewUser(&ctx.User, ctx.Conn.GetUserStatus(ctx.User.ID)))
	}

	res_channel := response.NewChannel(channel, recipients)
	ctx.WriteJSON(res_channel)
	ctx.Conn.AddUserToChannel(ctx.User.ID, channel.ID.Hex())
	if channel.Type == 1 {
		recipient := res_channel.Recipients[0]
		ctx.Conn.AddUserToChannel(recipient.ID, res_channel.ID)
		res_user := response.NewUser(&ctx.User, ctx.Conn.GetUserStatus(ctx.User.ID))
		ctx.Conn.SendToUser(recipient.ID, "CHANNEL_CREATE", response.NewChannel(channel, []response.User{res_user}))
	}
}

func CreateThread(ctx *Context) {
	url_vars := mux.Vars(ctx.Req)
	channelID := url_vars["id"]
	messageID := url_vars["mid"]

	channelObjectID, err := primitive.ObjectIDFromHex(channelID)
	if err != nil {
		log.Println("CreateThread: primitive.ObjectIDFromHex", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		log.Println("CreateThread: primitive.ObjectIDFromHex", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message database.Message
	err = ctx.Db.Mongo.Collection("messages").FindOne(context.Background(), bson.M{"_id": messageObjectID}).Decode(&message)
	if err != nil {
		log.Println("CreateThread: ctx.Db.Mongo.Collection.FindOne", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var channel database.Channel
	err = ctx.Db.Mongo.Collection("channels").FindOne(context.Background(), bson.M{"_id": channelObjectID}).Decode(&channel)
	if err != nil {
		log.Println("CreateThread: ctx.Db.Mongo.Collection.FindOne", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if message.HasThread {
		log.Println("thread already exists")
		var recipients []response.User
		for _, recipientID := range channel.Recipients {
			ctx.Conn.AddUserToChannel(recipientID, message.ThreadID.Hex())
			user, status := ctx.Db.GetUser(recipientID)
			if status == http.StatusOK {
				recipients = append(recipients, response.NewUser(user, status))
			} else {
				log.Println("CreateThread: ctx.Db.GetUser failed for recipient", recipientID)
				ctx.Res.WriteHeader(status)
				return
			}
		}
		channel, status := ctx.Db.GetChannel(message.ThreadID.Hex(), &ctx.User)
		if status != http.StatusOK {
			ctx.Res.WriteHeader(status)
			return
		}
		ctx.Res.WriteHeader(http.StatusOK)
		responseChannel := response.NewChannel(channel, recipients)
		json.NewEncoder(ctx.Res).Encode(responseChannel)
		return
	}

	newChannel := database.Channel{
		Type:       3,
		Recipients: channel.Recipients,
		Name:       message.Content,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	result, err := ctx.Db.Mongo.Collection("channels").InsertOne(context.Background(), newChannel)
	if err != nil {
		log.Println("CreateThread: ctx.Db.Mongo.Collection.InsertOne", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	newChannel.ID = result.InsertedID.(primitive.ObjectID)

	_, err = ctx.Db.Mongo.Collection("messages").UpdateOne(
		context.Background(),
		bson.M{"_id": messageObjectID},
		bson.M{"$set": bson.M{"thread_id": result.InsertedID, "has_thread": true}},
	)
	if err != nil {
		log.Println("CreateThread: ctx.Db.Mongo.Collection.UpdateOne", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var recipients []response.User
	for _, recipientID := range channel.Recipients {
		user, status := ctx.Db.GetUser(recipientID)
		if status == http.StatusOK {
			recipients = append(recipients, response.NewUser(user, status))
		} else {
			log.Println("CreateThread: ctx.Db.GetUser failed for recipient", recipientID)
			ctx.Res.WriteHeader(status)
			return
		}
	}

	for _, recipient := range recipients {
		ctx.Conn.AddUserToChannel(recipient.ID, newChannel.ID.Hex())
	}

	responseChannel := response.NewChannel(&newChannel, recipients)
	ctx.Res.WriteHeader(http.StatusOK)
	json.NewEncoder(ctx.Res).Encode(responseChannel)
}

func GetChannels(ctx *Context) {
	log.Println("handler GetChannels")
	res_channels := []response.Channel{}
	channels := ctx.Db.GetChannels(&ctx.User)
	for _, channel := range channels {
		recipients := []response.User{}
		for _, recipient := range channel.Recipients {
			if channel.Type == 1 && recipient == ctx.User.ID {
				continue
			}
			recipient, _ := ctx.Db.GetUser(recipient)
			recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
		}

		res_channels = append(res_channels, response.NewChannel(&channel, recipients))
	}

	ctx.WriteJSON(res_channels)
}

func GetChannelsMailings(ctx *Context) {
	log.Println("handler GetChannelsMailings")
	res_channels := []response.Channel{}
	channels := ctx.Db.GetChannelsMailings()
	for _, channel := range channels {
		recipients := []response.User{}
		for _, recipient := range channel.Recipients {
			recipient, _ := ctx.Db.GetUser(recipient)
			recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
		}
		res_channels = append(res_channels, response.NewChannel(&channel, recipients))
	}

	ctx.WriteJSON(res_channels)
}

func GetChannel(ctx *Context) {
	log.Println("handler GetChannel")
	url_vars := mux.Vars(ctx.Req)
	channel_id := url_vars["id"]

	channel, statusCode := ctx.Db.GetChannel(channel_id, &ctx.User)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	recipients := []response.User{}
	for _, recipient := range channel.Recipients {
		if channel.Type == 1 && recipient == ctx.User.ID {
			continue
		}
		recipient, _ := ctx.Db.GetUser(recipient)
		recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
	}

	res_channel := response.NewChannel(channel, recipients)
	ctx.WriteJSON(res_channel)
}

func EditChannel(ctx *Context) {
	log.Println("handler EditChannel")
	url_vars := mux.Vars(ctx.Req)
	channel_id := url_vars["id"]
	var request request.EditChannel
	_ = json.NewDecoder(ctx.Req.Body).Decode(&request)

	name := strings.TrimSpace(request.Name)
	icon := strings.TrimSpace(request.Icon)

	if name == "" && icon == "" {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	channel, statusCode := ctx.Db.ModifyChannel(channel_id, name, icon, &ctx.User)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	recipients := []response.User{}
	for _, recipient := range channel.Recipients {
		recipient, _ := ctx.Db.GetUser(recipient)
		recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
	}

	res_channel := response.NewChannel(channel, recipients)

	ctx.WriteJSON(res_channel)
	ctx.Conn.BroadcastToChannel(res_channel.ID, "CHANNEL_MODIFY", res_channel)
}

func DeleteChannel(ctx *Context) {
	log.Println("handler DeleteChannel")
	url_vars := mux.Vars(ctx.Req)
	channel_id := url_vars["id"]

	channel, statusCode := ctx.Db.DeleteChannel(channel_id, &ctx.User)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	recipients := []response.User{}
	for _, recipient := range channel.Recipients {
		recipient, _ := ctx.Db.GetUser(recipient)
		recipients = append(recipients, response.NewUser(recipient, ctx.Conn.GetUserStatus(recipient.ID)))
	}

	res_channel := response.NewChannel(channel, recipients)

	ctx.WriteJSON(res_channel)
	ctx.Conn.RemoveUserFromChannel(ctx.User.ID, channel_id)
	ctx.Conn.BroadcastToChannel(channel.ID.Hex(), "CHANNEL_MODIFY", res_channel)

	recipient_leave := fmt.Sprintf(RECIPIENT_LEAVE, ctx.User.Username)
	message, statusCode := ctx.Db.CreateMessage(recipient_leave, res_channel.ID, true, nil)

	if statusCode != http.StatusOK {
		return
	}

	res_message := response.NewMessage(message, response.User{})
	ctx.Conn.BroadcastToChannel(res_channel.ID, "MESSAGE_CREATE", res_message)
}
