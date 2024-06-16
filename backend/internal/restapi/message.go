package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/request"
	"github.com/mmikhail2001/team_chat/internal/response"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetMessages(ctx *Context) {
	log.Println("handler GetMessages")
	vars := mux.Vars(ctx.Req)
	channel_id := vars["id"]
	querys := ctx.Req.URL.Query()

	limit, _ := strconv.ParseInt(querys.Get("limit"), 10, 64)
	offset, _ := strconv.ParseInt(querys.Get("offset"), 10, 64)
	if limit == 0 {
		limit = 100
	} else if limit > 100 {
		limit = 100
	}

	messages, statusCode := ctx.Db.GetMessages(channel_id, limit, offset, &ctx.User)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	// channel, statusCode := ctx.Db.GetChannel(channel_id, &ctx.User)
	// if statusCode != http.StatusOK {
	// 	ctx.Res.WriteHeader(statusCode)
	// 	return
	// }

	messages_res := []response.Message{}
	for _, message := range messages {
		if message.SystemMessage {
			messages_res = append(messages_res, response.NewMessage(&message, response.User{}))
			continue
		}

		user, statusCode := ctx.Db.GetUser(message.AccountID)
		if statusCode != http.StatusOK {
			fmt.Println("err: get user === ", message.AccountID)
			continue
		}
		messages_res = append(messages_res, response.NewMessage(&message, response.NewUser(user, 0)))
		// if channel.Type != 4 {
		// } else {
		// 	user := database.User{Username: "gitlab-user"}
		// 	messages_res = append(messages_res, response.NewMessage(&message, response.NewUser(&user, 0)))
		// }
	}

	ctx.WriteJSON(messages_res)
}

func GetMessage(ctx *Context) {
	log.Println("handler GetMessage")
	vars := mux.Vars(ctx.Req)
	channel_id := vars["id"]
	message_id := vars["mid"]

	message, _, statusCode := ctx.Db.GetMessage(message_id, channel_id, &ctx.User)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	var message_res response.Message

	if message.SystemMessage {
		var user response.User
		message_res = response.NewMessage(message, user)
	} else {
		user, statusCode := ctx.Db.GetUser(message.AccountID)
		if statusCode != http.StatusOK {
			ctx.Res.WriteHeader(statusCode)
			return
		}

		message_res = response.NewMessage(message, response.NewUser(user, 0))
	}

	ctx.WriteJSON(message_res)
}

func CreateMessage(ctx *Context) {
	log.Println("handler CreateMessage")
	vars := mux.Vars(ctx.Req)
	channel_id := vars["id"]

	content_type := ctx.Req.Header.Get("Content-Type")
	if content_type == "application/json" {
		var message_req request.Message
		err := json.NewDecoder(ctx.Req.Body).Decode(&message_req)
		if err != nil {
			ctx.Res.WriteHeader(http.StatusBadRequest)
			return
		}

		content := strings.TrimSpace(message_req.Content)

		if content == "" {
			ctx.Res.WriteHeader(http.StatusBadRequest)
			return
		}

		message, statusCode := ctx.Db.CreateMessage(content, channel_id, false, &ctx.User)
		if statusCode != http.StatusOK {
			ctx.Res.WriteHeader(statusCode)
			return
		}

		message_res := response.NewMessage(message, response.NewUser(&ctx.User, 0))

		ctx.WriteJSON(message_res)
		ctx.Conn.BroadcastToChannel(channel_id, "MESSAGE_CREATE", message_res)
	} else {
		// с вложениями
		content := ctx.Req.FormValue("content")
		file, handler, err := ctx.Req.FormFile("file")
		if err != nil {
			ctx.Res.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		if handler.Size > 8388608 {
			ctx.Res.WriteHeader(http.StatusBadRequest)
			return
		}

		file_data, err := io.ReadAll(file)
		if err != nil {
			ctx.Res.WriteHeader(http.StatusBadRequest)
			return
		}

		filename := strings.ReplaceAll(handler.Filename, " ", "_")

		db_attachment := database.Attachment{
			ID:          primitive.NewObjectID(),
			Filename:    filename,
			Size:        handler.Size,
			ContentType: handler.Header["Content-Type"][0],
			Data:        file_data,
		}

		content = strings.TrimSpace(content)
		message, statusCode := ctx.Db.CreateMessage(content, channel_id, false, &ctx.User)
		if statusCode != http.StatusOK {
			ctx.Res.WriteHeader(statusCode)
			return
		}
		message.Attachments = []database.Attachment{db_attachment}

		messageCollection := ctx.Db.Mongo.Collection("messages")
		messageCollection.UpdateOne(context.TODO(), bson.M{"_id": message.ID}, bson.M{"$set": bson.M{"attachments": message.Attachments}})

		message_res := response.NewMessage(message, response.NewUser(&ctx.User, 0))

		ctx.WriteJSON(message_res)
		ctx.Conn.BroadcastToChannel(channel_id, "MESSAGE_CREATE", message_res)
	}
}

func EditMessage(ctx *Context) {
	log.Println("handler EditMessage")
	vars := mux.Vars(ctx.Req)
	channel_id := vars["id"]
	message_id := vars["mid"]

	var message_req request.Message
	err := json.NewDecoder(ctx.Req.Body).Decode(&message_req)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	content := strings.TrimSpace(message_req.Content)

	if content == "" {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	message, statusCode := ctx.Db.EditMessage(message_id, channel_id, content, &ctx.User)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	message_res := response.NewMessage(message, response.NewUser(&ctx.User, 0))

	ctx.WriteJSON(message_res)
	ctx.Conn.BroadcastToChannel(message_res.ChannelID, "MESSAGE_MODIFY", message_res)
}

func DeleteReaction(ctx *Context) {
	log.Println("handler DeleteReaction")
	vars := mux.Vars(ctx.Req)
	messageID := vars["mid"]

	// Получаем ID пользователя
	userID := ctx.User.ID

	// Находим сообщение по его ID
	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		log.Println("DeleteReaction: primitive.ObjectIDFromHex", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message database.Message
	err = ctx.Db.Mongo.Collection("messages").FindOne(context.TODO(), bson.M{"_id": messageObjectID}).Decode(&message)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusNotFound)
		log.Println("Failed to find message:", err)
		return
	}

	// Удаляем реакцию пользователя из списка реакций сообщения
	var newReactions []database.Reaction
	for _, existingReaction := range message.Reactions {
		if existingReaction.UserID != userID {
			newReactions = append(newReactions, existingReaction)
		}
	}

	message.Reactions = newReactions

	// Обновляем сообщение в базе данных с новым списком реакций
	_, err = ctx.Db.Mongo.Collection("messages").UpdateOne(
		context.TODO(),
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"reactions": newReactions}},
	)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to update message reactions:", err)
		return
	}

	// Возвращаем обновленное сообщение в формате JSON
	user, _ := ctx.Db.GetUser(message.AccountID)
	resMessage := response.NewMessage(&message, response.NewUser(user, ctx.Conn.GetUserStatus(user.ID)))
	response, err := json.Marshal(resMessage)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to marshal response:", err)
		return
	}

	ctx.Res.Header().Set("Content-Type", "application/json")
	ctx.Res.WriteHeader(http.StatusOK)
	ctx.Res.Write(response)
}

func CreateReaction(ctx *Context) {
	log.Println("handler CreateReaction")
	vars := mux.Vars(ctx.Req)
	messageID := vars["mid"]
	reaction := ctx.Req.URL.Query().Get("reaction")
	if reaction == "" {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	messageObjectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		log.Println("CreateReaction: primitive.ObjectIDFromHex", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message database.Message
	err = ctx.Db.Mongo.Collection("messages").FindOne(context.TODO(), bson.M{"_id": messageObjectID}).Decode(&message)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusNotFound)
		log.Println("Failed to find message:", err)
		return
	}

	var newReactions []database.Reaction

	for _, existingReaction := range message.Reactions {
		if existingReaction.UserID != ctx.User.ID {
			newReactions = append(newReactions, existingReaction)
		}
	}

	reactionToAdd := database.Reaction{Reaction: reaction, UserID: ctx.User.ID}
	newReactions = append(newReactions, reactionToAdd)
	message.Reactions = newReactions

	_, err = ctx.Db.Mongo.Collection("messages").UpdateOne(
		context.TODO(),
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"reactions": newReactions}},
	)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to update message reactions:", err)
		return
	}

	// user, _ := ctx.Db.GetUser(ctx.User.ID)
	user, _ := ctx.Db.GetUser(message.AccountID)

	resMessage := response.NewMessage(&message, response.NewUser(user, ctx.Conn.GetUserStatus(user.ID)))

	response, err := json.Marshal(resMessage)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to marshal response:", err)
		return
	}

	ctx.Res.Header().Set("Content-Type", "application/json")
	ctx.Res.WriteHeader(http.StatusOK)
	ctx.Res.Write(response)
}

func DeleteMessage(ctx *Context) {
	log.Println("handler DeleteMessage")
	vars := mux.Vars(ctx.Req)
	channel_id := vars["id"]
	message_id := vars["mid"]

	message, statusCode := ctx.Db.DeleteMessage(message_id, channel_id, &ctx.User)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	user, statusCode := ctx.Db.GetUser(message.AccountID)
	if statusCode != http.StatusOK {
		ctx.Res.WriteHeader(statusCode)
		return
	}

	message_res := response.NewMessage(message, response.NewUser(user, 0))

	ctx.WriteJSON(message_res)
	ctx.Conn.BroadcastToChannel(message_res.ChannelID, "MESSAGE_DELETE", message_res)
}
