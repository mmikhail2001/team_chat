package websocket

import "github.com/mmikhail2001/team_chat/internal/response"

type Connect struct {
	Token string
}

type WS_Message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Ready struct {
	User         response.User           `json:"user"`
	Channels     []response.Channel      `json:"channels"`
	Relationship []response.Relationship `json:"relationship"`
}
