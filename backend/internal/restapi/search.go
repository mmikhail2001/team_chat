package restapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/response"
	"go.mongodb.org/mongo-driver/bson"
)

func Search(ctx *Context) {
	query := ctx.Req.URL.Query().Get("query")
	searchEmployees := ctx.Req.URL.Query().Get("employees") == "true"
	searchChats := ctx.Req.URL.Query().Get("chats") == "true"
	searchRoles := ctx.Req.URL.Query().Get("roles") == "true"

	var res struct {
		Users    []response.Channel
		Channels []response.Channel
		Roles    []string
	}

	if searchRoles {
		collection := ctx.Db.Mongo.Collection("roles")

		filter := bson.M{"name": bson.M{"$regex": query, "$options": "i"}}

		cursor, err := collection.Find(context.Background(), filter)
		if err != nil {
			log.Println("Find roles: err", err)
			return
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var role database.Role
			if err := cursor.Decode(&role); err != nil {
				log.Println("Decode roles: err", err)
				continue
			}
			res.Roles = append(res.Roles, role.Name)
		}
		if err := cursor.Err(); err != nil {
			log.Println("cursor roles: err", err)
			return
		}
	}

	if searchEmployees {
		dbChannelsUsers, err := ctx.Db.SearchPersonalChannels(query, ctx.User.ID)
		if err != nil {
			log.Println("Search: err", err)
			ctx.Res.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, dbChannel := range dbChannelsUsers {
			var recipients []response.User
			for _, recipientID := range dbChannel.Recipients {
				if recipientID != ctx.User.ID {
					user, status := ctx.Db.GetUser(recipientID)
					if status != http.StatusOK {
						continue
					}
					userResponse := response.NewUser(user, ctx.Conn.GetUserStatus(user.ID))
					// массив из 1 элемента...
					recipients = append(recipients, userResponse)
				}
			}
			channel := response.NewChannel(&dbChannel, recipients)
			res.Users = append(res.Users, channel)
		}
	}

	if searchChats {
		dbChannels, err := ctx.Db.SearchGroupChannels(query, ctx.User.ID)
		if err != nil {
			return
		}
		for _, dbChannel := range dbChannels {
			var recipients []response.User
			for _, recipientID := range dbChannel.Recipients {
				user, status := ctx.Db.GetUser(recipientID)
				if status != http.StatusOK {
					continue
				}
				userResponse := response.NewUser(user, ctx.Conn.GetUserStatus(user.ID))
				recipients = append(recipients, userResponse)
			}
			channel := response.NewChannel(&dbChannel, recipients)
			res.Channels = append(res.Channels, channel)
		}
	}

	if len(res.Users) == 0 && len(res.Channels) == 0 && len(res.Roles) == 0 {
		ctx.Res.WriteHeader(http.StatusNotFound)
		return
	}

	jsonResponse, err := json.Marshal(res)
	if err != nil {
		log.Println("Marshal: err", err)
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx.Res.Header().Set("Content-Type", "application/json")
	ctx.Res.WriteHeader(http.StatusOK)
	ctx.Res.Write(jsonResponse)
}
