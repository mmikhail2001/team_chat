package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/websocket"

	oidc "github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/oauth2"
)

// userInfo, err := h.Provider.UserInfo(context.Background(), oauth2.StaticTokenSource(&token))

type Handler struct {
	KeycloackHost string
	ClientHost    string
	Realm         string
	ClientID      string
	ClientSecret  string
	OAuth2Config  oauth2.Config
	Provider      *oidc.Provider
	Verifier      *oidc.IDTokenVerifier

	mu          sync.Mutex
	oauthStates map[string]bool

	Db   *database.Database
	Conn *websocket.Connections
}

type Mytoken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenID      string `json:"token_id"`
}

var ctxToken string = "token"

type OAuthConfig struct {
	KeycloackHost string
	ClientHost    string
	Realm         string
	ClientID      string
	ClientSecret  string
}

// Db   *database.Database
// Conn *websocket.Connections

// пока так, но нужно создавать клиента
func NewHandler(cfg OAuthConfig, Db *database.Database, Conn *websocket.Connections) (Handler, error) {
	provider, err := oidc.NewProvider(context.Background(), cfg.KeycloackHost+"/realms/"+cfg.Realm)
	if err != nil {
		fmt.Printf("This is an error with regard to the context: %v", err)
		return Handler{}, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	OAuth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.ClientHost + "/api/loginCallback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return Handler{
		KeycloackHost: cfg.KeycloackHost,
		ClientHost:    cfg.ClientHost,
		Realm:         cfg.Realm,
		ClientID:      cfg.ClientID,
		ClientSecret:  cfg.ClientSecret,
		OAuth2Config:  OAuth2Config,
		Provider:      provider,
		Verifier:      verifier,
		mu:            sync.Mutex{},
		oauthStates:   make(map[string]bool),
		Db:            Db,
		Conn:          Conn,
	}, nil
}

type Response struct {
	Redirect string `json:"redirect,omitempty"`
}

type AuthFunction func(ctx *Context)

func (h *Handler) Authenticated(function AuthFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenIDRaw, err := r.Cookie("token_id")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_ = tokenIDRaw
		accessTokenRaw, err := r.Cookie("access_token")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		refreshTokenRaw, err := r.Cookie("refresh_token")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// mytoken := Mytoken{
		// 	AccessToken:  accessTokenRaw.Value,
		// 	RefreshToken: refreshTokenRaw.Value,
		// 	TokenID:      tokenIDRaw.Value,
		// }

		token := oauth2.Token{
			AccessToken:  accessTokenRaw.Value,
			RefreshToken: refreshTokenRaw.Value,
		}
		userInfo, err := h.Provider.UserInfo(context.Background(), oauth2.StaticTokenSource(&token))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		user, status := h.Db.GetUser(userInfo.Subject)
		if status != http.StatusOK {
			log.Println("Authenticated: GetUser: status: ", status)
			w.WriteHeader(status)
		}

		ctx := Context{
			Res:  w,
			Req:  r,
			Db:   h.Db,
			User: *user,
			Conn: h.Conn,
		}

		function(&ctx)
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleLogin\n\n")
	oauthStateString := uuid.New().String()
	url := h.OAuth2Config.AuthCodeURL(oauthStateString)
	h.mu.Lock()
	h.oauthStates[oauthStateString] = true
	h.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{Redirect: url})
}

// фронт не наш тянет эту ручку? значит, http.Redirect ???
func (h *Handler) HandleLoginCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleLoginCallback\n\n")
	state := r.FormValue("state")
	if _, ok := h.oauthStates[state]; !ok {
		fmt.Printf("invalid oauth state, expected '%s'\n", state)
		http.Redirect(w, r, "/api/login", http.StatusTemporaryRedirect)
		return
	}
	delete(h.oauthStates, state)
	code := r.FormValue("code")
	token, err := h.OAuth2Config.Exchange(r.Context(), code)
	if err != nil {
		fmt.Printf("Code exchange failed with '%v'\n", err)
		http.Redirect(w, r, "/api/login", http.StatusTemporaryRedirect)
		return
	}

	userInfo, err := h.Provider.UserInfo(context.Background(), oauth2.StaticTokenSource(token))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
	}
	subject := userInfo.Subject

	collectionUser := h.Db.Mongo.Collection("users")

	_, status := h.Db.GetUser(subject)
	if status == http.StatusNotFound {
		newUser := bson.M{"_id": subject, "email": userInfo.Email, "username": strings.Split(userInfo.Email, "@")[0]}
		_, err := collectionUser.InsertOne(context.Background(), newUser)
		if err != nil {
			log.Println("InsertOne new user")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	userFromDB, status := h.Db.GetUser(subject)
	if status == http.StatusNotFound {
		log.Println("GetUser new user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	existingChannels := make(map[string]database.Channel)
	channels := h.Db.GetChannels(userFromDB)

	for _, channel := range channels {
		if channel.Type == 1 {
			if channel.Recipients[0] == subject {
				existingChannels[channel.Recipients[1]] = channel
			} else {
				existingChannels[channel.Recipients[0]] = channel
			}
		}
	}

	cursor, err := collectionUser.Find(context.Background(), bson.M{"_id": bson.M{"$ne": subject}})
	if err != nil {
		log.Println("Find user except new")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user database.User
		if err := cursor.Decode(&user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		recipientID := user.ID
		if _, ok := existingChannels[recipientID]; !ok {
			log.Println("create new channel:", userFromDB.Username, "with:", user.Username)
			newChannel, statusCode := h.Db.CreateChannel("", "", recipientID, userFromDB)
			if statusCode != http.StatusOK {
				w.WriteHeader(statusCode)
				return
			}
			existingChannels[recipientID] = *newChannel
		}
	}
	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenID := token.Extra("id_token").(string)

	http.SetCookie(w, &http.Cookie{
		Name:     "token_id",
		Value:    tokenID,
		Path:     "/",
		Expires:  token.Expiry,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		Path:     "/",
		Expires:  token.Expiry,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token.RefreshToken,
		Path:     "/",
		Expires:  token.Expiry,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/channels", http.StatusTemporaryRedirect)
}

func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	log.Println("AuthMiddleware\n\n")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenIDRaw, err := r.Cookie("token_id")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		accessTokenRaw, err := r.Cookie("access_token")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		refreshTokenRaw, err := r.Cookie("refresh_token")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		mytoken := Mytoken{
			AccessToken:  accessTokenRaw.Value,
			RefreshToken: refreshTokenRaw.Value,
			TokenID:      tokenIDRaw.Value,
		}

		token := oauth2.Token{
			AccessToken:  accessTokenRaw.Value,
			RefreshToken: refreshTokenRaw.Value,
		}
		userInfo, err := h.Provider.UserInfo(context.Background(), oauth2.StaticTokenSource(&token))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		}

		log.Println("userInfo.Subject = ", userInfo.Subject)

		// log.Println("middleware with", userInfo.Subject, userInfo.Profile, userInfo.Email)

		// это как будто должно работать, но не работает...
		// мне нужно обновлять просроченный токен...

		// tokenSource := h.OAuth2Config.TokenSource(context.Background(), &token)
		// newToken, err := tokenSource.Token()
		// if err != nil {
		// 	log.Fatalln(err)
		// }

		// if newToken.AccessToken != token.AccessToken {
		// 	log.Println("newToken.AccessToken = ", newToken.AccessToken)
		// 	log.Println("token.AccessToken = ", token.AccessToken)
		// 	mytoken.AccessToken = newToken.AccessToken
		// }

		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxToken, mytoken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("Handle Logout")

	http.SetCookie(w, &http.Cookie{
		Name:     "token_id",
		Value:    "", // Пустое значение
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour), // Установка срока действия в прошлое время
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	mytokenRaw := r.Context().Value(ctxToken)
	mytoken, ok := mytokenRaw.(Mytoken)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res := Response{
		Redirect: h.KeycloackHost + "/realms/" + h.Realm + "/protocol/openid-connect/logout?post_logout_redirect_uri=" + h.ClientHost + "/" + "&id_token_hint=" + mytoken.TokenID,
	}
	json.NewEncoder(w).Encode(res)
}
