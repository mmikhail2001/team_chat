package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/websocket"
)

type Context struct {
	Res  http.ResponseWriter
	Req  *http.Request
	Db   *database.Database
	User database.User
	Conn *websocket.Connections
}

func (ctx *Context) WriteJSON(res_object interface{}) {
	res, err := json.Marshal(res_object)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx.Res.Header().Set("Content-Type", "application/json")
	ctx.Res.Write(res)
}
