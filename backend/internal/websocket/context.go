package websocket

import (
	"github.com/mmikhail2001/team_chat/internal/database"
)

type Context struct {
	Ws    *Ws
	Event string
	Data  []byte
	// TODO: точно нужно таким образом передавать бд?
	Db *database.Database
}

func (ctx *Context) Send(data []byte) {
	ctx.Ws.Write(data)
}
