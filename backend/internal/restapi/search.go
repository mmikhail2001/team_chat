package restapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mmikhail2001/team_chat/internal/response"
)

func Search(ctx *Context) {
	query := ctx.Req.URL.Query().Get("query")
	searchEmployees := ctx.Req.URL.Query().Get("employees") == "true"
	searchChats := ctx.Req.URL.Query().Get("chats") == "true"

	var res struct {
		Users    []response.Channel
		Channels []response.Channel
	}

	if searchEmployees {
		dbChannelsUsers, err := ctx.Db.SearchUsers(query, ctx.User.ID)
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
					recipients = append(recipients, userResponse)
				}
			}
			channel := response.NewChannel(&dbChannel, recipients)
			res.Users = append(res.Users, channel)
		}
	}

	if searchChats {
		dbChannels, err := ctx.Db.SearchChannels(query, ctx.User.ID)
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

	if len(res.Users) == 0 && len(res.Channels) == 0 {
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
