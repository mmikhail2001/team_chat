package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/mmikhail2001/team_chat/internal/request"
	"github.com/mmikhail2001/team_chat/internal/response"

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

	channel, statusCode := ctx.Db.CreateChannel(name, icon, recipientID, &ctx.User)
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
