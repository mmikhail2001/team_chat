package restapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/request"
	"github.com/mmikhail2001/team_chat/internal/response"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// func GetUser(ctx *Context) {
// 	user_res := response.NewUser(&ctx.User, ctx.Conn.GetUserStatus(ctx.User.ID))
// 	ctx.WriteJSON(user_res)
// }

func GetUser(ctx *Context) {
	filter := bson.M{
		"reactions": bson.M{
			"$elemMatch": bson.M{"user_id": ctx.User.ID},
		},
	}

	messagesCollection := ctx.Db.Mongo.Collection("messages")

	cur, err := messagesCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.TODO())

	var messages []database.Message
	for cur.Next(context.TODO()) {
		var msg database.Message
		if err := cur.Decode(&msg); err != nil {
			log.Fatal(err)
		}
		messages = append(messages, msg)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	var userReactions []response.ReactionMessage
	for _, msg := range messages {
		for _, reaction := range msg.Reactions {
			if reaction.UserID == ctx.User.ID {
				userReactions = append(userReactions, response.ReactionMessage{
					MessageID: msg.ID.Hex(),
					Reaction:  reaction.Reaction,
				})
			}
		}
	}

	user_res := response.NewUser(&ctx.User, ctx.Conn.GetUserStatus(ctx.User.ID))
	user_res.Reactions = userReactions
	ctx.WriteJSON(user_res)
}

func EditUser(ctx *Context) {
	log.Println("handler EditUser")
	var request request.User
	err := json.NewDecoder(ctx.Req.Body).Decode(&request)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	avatar := request.Avatar
	if avatar == "" {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	file_type_regx := regexp.MustCompile("image/(png|jpeg|gif)")
	file_ext_regx := regexp.MustCompile("png|jpeg|gif")

	file_type := file_type_regx.FindString(avatar)
	if file_type == "" {
		ctx.Res.WriteHeader(http.StatusBadRequest)
		return
	}

	file_ext := file_ext_regx.FindString(file_type)

	avatarB64 := avatar[strings.Index(avatar, ",")+1:]

	avatar_db := database.Avatar{
		ID:     primitive.NewObjectID(),
		Ext:    file_ext,
		Type:   file_type,
		Avatar: avatarB64,
	}

	users := ctx.Db.Mongo.Collection("users")
	_, err = users.UpdateOne(context.TODO(), bson.M{"_id": ctx.User.ID}, bson.M{"$set": bson.M{"avatar": avatar_db}})
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx.User.Avatar = avatar_db
	user_res := response.NewUser(&ctx.User, ctx.Conn.GetUserStatus(ctx.User.ID))
	ctx.WriteJSON(user_res)
}
