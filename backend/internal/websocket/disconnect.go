package websocket

import (
	"log"

	"github.com/mmikhail2001/team_chat/internal/response"
)

func (ws *Ws) Disconnect() {
	if ws.User == nil {
		ws.Close()
		return
	}

	channels := ws.Db.GetChannels(ws.User)
	for _, channel := range channels {
		status := response.Status{
			UserID:    ws.User.ID,
			Status:    0,
			Type:      1,
			ChannelID: channel.ID.Hex(),
		}
		ws.Conns.RemoveUserFromChannel(ws.User.ID, channel.ID.Hex())
		ws.Conns.BroadcastToChannel(channel.ID.Hex(), "STATUS_UPDATE", status)
	}

	relationships := ws.Db.GetRelationships(ws.User.ID)
	for _, relationship := range relationships {
		if relationship.Type != 1 {
			continue
		}

		status := response.Status{
			UserID: ws.User.ID,
			Status: 0,
			Type:   0,
		}
		ws.Conns.SendToUser(relationship.ToUserID, "STATUS_UPDATE", status)
	}

	ws.Conns.RemoveUser(ws.User.ID)
	ws.Close()
	log.Printf("%s Disconnected\n", ws.User.Username)
}
