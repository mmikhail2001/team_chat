package main

import (
	"net/http"

	"github.com/mmikhail2001/team_chat/internal/database"
	"github.com/mmikhail2001/team_chat/internal/restapi"
)

type IDBFunction func(w http.ResponseWriter, r *http.Request, db *database.Database)
type AuthFunction func(ctx *restapi.Context)

func IncludeDB(function IDBFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		function(w, r, db)
	}
}
