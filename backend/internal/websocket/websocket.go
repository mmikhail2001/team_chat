package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mmikhail2001/team_chat/internal/database"

	"github.com/gorilla/websocket"
)

type Ws struct {
	// для отправки самому пользователю
	Conn *websocket.Conn
	// точно ля для Handler нужен отдельная структура ???
	Handler *EventHandler
	Db      *database.Database
	User    *database.User
	// все коннекты со всеми пользователями
	// чтобы в случае события у данного пользователя знать, кому отправлять уведомления
	Conns *Connections
}

func (ws *Ws) Write(data []byte) {
	err := ws.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Println(err)
	}
}

func (ws *Ws) Read() ([]byte, error) {
	_, data, err := ws.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (ws *Ws) ReadLoop() {
	for {
		data, err := ws.Read()
		if err != nil {
			ws.Disconnect()
			return
		}
		ws.HandleWSMessage(data)
	}
}

func (ws *Ws) HandleWSMessage(data []byte) {
	var ws_message WS_Message
	err := json.Unmarshal(data, &ws_message)
	if err != nil {
		fmt.Println(err)
	}

	data_json, err := json.Marshal(ws_message.Data)
	if err != nil {
		fmt.Println(err)
	}

	ctx := Context{
		Ws:    ws,
		Event: strings.ToUpper(ws_message.Event),
		Data:  data_json,
		// TODO: зачем db, если в WS есть DB
		// почему DB есть во всех  структурах..........
		// дичайшее зацепление
		Db: ws.Db,
	}

	ws.Handler.Handle(ctx)
}

func (ws *Ws) Close() {
	err := ws.Conn.Close()
	if err != nil {
		log.Println(err)
	}
}
