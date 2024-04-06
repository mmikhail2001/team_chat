package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/mail"
	"github.com/mmikhail2001/team_chat/internal/restapi"
	"github.com/mmikhail2001/team_chat/internal/websocket"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// TODO:
// первое сообщение в личное беседе дублируется

var (
	HOST           = os.Getenv("SERVER_HOST")
	PORT           = os.Getenv("SERVER_PORT")
	MONGO_URI      = os.Getenv("MONGO_URI")
	MONGO_DATABASE = os.Getenv("MONGO_DATABASE")
	SMTP_SERVER    = os.Getenv("SMTP_SERVER")
	SMTP_USERNAME  = os.Getenv("SMTP_USERNAME")
	SMTP_PASSWORD  = os.Getenv("SMTP_PASSWORD")
)

var (
	// TODO: с бд нужно работать не так!
	// test
	db      = database.NewDatabase(MONGO_URI, MONGO_DATABASE)
	handler = &websocket.EventHandler{}
	conns   = websocket.NewConnections()
)

var keycloackHost string = "http://localhost:8080"
var clientHost string = "http://localhost:3000"
var realm string = "myrealm"
var clientID string = "myclient"
var clientSecret string = "hYApd1fWgPbABkbn6zTY5r66DrKLlWn4"

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	mail.NewMailSystem(SMTP_SERVER, SMTP_USERNAME, SMTP_PASSWORD)

	router := mux.NewRouter()
	headers := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"})
	// TODO: ИБ
	origins := handlers.AllowedOrigins([]string{"*"})
	router.Use(handlers.CORS(headers, methods, origins))
	router.Use(handlers.RecoveryHandler())

	handler.Add("PING", func(ctx *websocket.Context) {
		ws_msg, _ := json.Marshal(websocket.WS_Message{Event: "PONG", Data: ""})
		ctx.Send(ws_msg)
	})

	oauthCfg := restapi.OAuthConfig{
		KeycloackHost: keycloackHost,
		ClientHost:    clientHost,
		Realm:         realm,
		ClientID:      clientID,
		ClientSecret:  clientSecret,
	}

	oauthHandler, err := restapi.NewHandler(oauthCfg, db, conns)
	if err != nil {
		fmt.Println(err)
		return
	}

	api := router.PathPrefix("/api").Subrouter()

	// Auth
	api.HandleFunc("/login", oauthHandler.HandleLogin).Methods("POST")
	api.HandleFunc("/logout", oauthHandler.AuthMiddleware(oauthHandler.Logout)).Methods("POST")
	api.HandleFunc("/loginCallback", oauthHandler.HandleLoginCallback).Methods("GET")
	api.HandleFunc("/checkLogin", oauthHandler.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).Methods("POST")

	api.HandleFunc("/search", oauthHandler.Authenticated(restapi.Search)).Methods("GET")

	api.HandleFunc("/channels/{id}", oauthHandler.Authenticated(restapi.GetChannel)).Methods("GET")
	api.HandleFunc("/channels/{id}", oauthHandler.Authenticated(restapi.EditChannel)).Methods("PATCH")
	api.HandleFunc("/channels/{id}", oauthHandler.Authenticated(restapi.DeleteChannel)).Methods("DELETE")
	// Recipients
	api.HandleFunc("/channels/{id}/recipients/{uid}", oauthHandler.Authenticated(restapi.AddRecipient)).Methods("PUT")
	api.HandleFunc("/channels/{id}/recipients/{uid}", oauthHandler.Authenticated(restapi.RemoveRecipient)).Methods("DELETE")
	// Messages
	api.HandleFunc("/channels/{id}/messages", oauthHandler.Authenticated(restapi.GetMessages)).Methods("GET")
	api.HandleFunc("/channels/{id}/messages", oauthHandler.Authenticated(restapi.CreateMessage)).Methods("POST")
	api.HandleFunc("/channels/{id}/messages/{mid}", oauthHandler.Authenticated(restapi.GetMessage)).Methods("GET")
	api.HandleFunc("/channels/{id}/messages/{mid}", oauthHandler.Authenticated(restapi.EditMessage)).Methods("PATCH")
	api.HandleFunc("/channels/{id}/messages/{mid}", oauthHandler.Authenticated(restapi.DeleteMessage)).Methods("DELETE")

	// Threads
	api.HandleFunc("/channels/{id}/messages/{mid}", oauthHandler.Authenticated(restapi.CreateThread)).Methods("POST")
	// Reactions
	api.HandleFunc("/messages/{mid}/react", oauthHandler.Authenticated(restapi.CreateReaction)).Methods("POST")

	// Pin Messages
	api.HandleFunc("/channels/{id}/pins", oauthHandler.Authenticated(restapi.GetPins)).Methods("GET")
	api.HandleFunc("/channels/{id}/pins/{mid}", oauthHandler.Authenticated(restapi.PinMsg)).Methods("PUT")
	api.HandleFunc("/channels/{id}/pins/{mid}", oauthHandler.Authenticated(restapi.UnpinMsg)).Methods("DELETE")
	// Invites
	api.HandleFunc("/invites/{id}", oauthHandler.Authenticated(restapi.JoinInvite)).Methods("GET")
	api.HandleFunc("/channels/{id}/invites", oauthHandler.Authenticated(restapi.GetInvites)).Methods("GET")
	api.HandleFunc("/channels/{id}/invites", oauthHandler.Authenticated(restapi.CreateInvite)).Methods("POST")
	api.HandleFunc("/channels/{id}/invites/{iid}", oauthHandler.Authenticated(restapi.DeleteInvite)).Methods("DELETE")
	// Bans
	api.HandleFunc("/channels/{id}/bans", oauthHandler.Authenticated(restapi.GetAllBans)).Methods("GET")
	api.HandleFunc("/channels/{id}/bans/{bid}", oauthHandler.Authenticated(restapi.GetBan)).Methods("GET")
	api.HandleFunc("/channels/{id}/bans/{bid}", oauthHandler.Authenticated(restapi.DeleteBan)).Methods("DELETE")
	// Users
	api.HandleFunc("/users/@me", oauthHandler.Authenticated(restapi.GetUser)).Methods("GET")
	api.HandleFunc("/users/@me", oauthHandler.Authenticated(restapi.EditUser)).Methods("PATCH")
	api.HandleFunc("/users/@me/channels", oauthHandler.Authenticated(restapi.GetChannels)).Methods("GET")
	api.HandleFunc("/users/@me/channels", oauthHandler.Authenticated(restapi.CreateChannel)).Methods("POST")
	// Relationship
	api.HandleFunc("/users/@me/relationships", oauthHandler.Authenticated(restapi.GetRelationships)).Methods("GET")
	api.HandleFunc("/users/@me/relationships/{rid}", oauthHandler.Authenticated(restapi.GetRelationship)).Methods("GET")
	// TODO: это параметр, а не путь
	// внутри хэндлеров возможно копипаста
	// все равно надо убирать этих друзей...
	api.HandleFunc("/users/@me/relationships/{rid}/default", oauthHandler.Authenticated(restapi.ChangeRelationshipToDefault)).Methods("PUT")
	api.HandleFunc("/users/@me/relationships/{rid}/friend", oauthHandler.Authenticated(restapi.ChangeRelationshipToFriend)).Methods("PUT")
	api.HandleFunc("/users/@me/relationships/{rid}/block", oauthHandler.Authenticated(restapi.ChangeRelationshipToBlock)).Methods("PUT")
	// Files
	// а чтобы запросить аватарки не нужно быть авторизированным?...
	api.HandleFunc("/avatars/{user_id}/{avatar_id}/{filename}", IncludeDB(restapi.GetAvatars)).Methods("GET")
	api.HandleFunc("/icons/{channel_id}/{icon_id}/{filename}", IncludeDB(restapi.GetIcons)).Methods("GET")
	api.HandleFunc("/attachments/{channel_id}/{message_id}/{attachment_id}/{filename}", IncludeDB(restapi.GetAttachments)).Methods("GET")
	// Gateway
	api.HandleFunc("/ws", oauthHandler.Authenticated(Gateway))

	fileServer := http.FileServer(http.Dir("./web/dist/"))
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("./web/dist" + r.URL.Path); err != nil {
			http.ServeFile(w, r, "./web/dist/index.html")
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})

	server_uri := fmt.Sprintf("%s:%s", HOST, PORT)
	log.Println("Listening on ", server_uri)

	server := http.Server{
		Addr:         server_uri,
		Handler:      router,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
		IdleTimeout:  time.Second * 3,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err.Error())
	}
}
