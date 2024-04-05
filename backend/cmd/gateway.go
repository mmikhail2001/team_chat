package main

import (
	"log"
	"net/http"

	"github.com/mmikhail2001/team_chat/internal/restapi"

	ws "github.com/mmikhail2001/team_chat/internal/websocket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Gateway(ctx *restapi.Context) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(ctx.Res, ctx.Req, nil)

	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	ws := &ws.Ws{
		Conn:    conn,
		Handler: handler,
		Db:      db,
		User:    &ctx.User,
		Conns:   conns,
	}

	// TODO: а где массив всех ws соединений?
	ws.ConnectUser()
	// вместо чтения всех поступающих сообщений в цикле нужно обработчики навешивать на события
	// это только для пинг понга ???
	// т.е. клиент сам ничего не пишет в веб сокет, кроме пинга
	// клиент общается по http, а далее сервер уже рассылает всем ws уведомления
	// стоило ли ради одного пинг понга все эти функции писать с обработчиками? ... и EventHandler-ами
	ws.ReadLoop()
}
