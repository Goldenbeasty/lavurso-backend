package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	mux := httprouter.New()
	mux.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowed)

	// administrator
	mux.HandlerFunc(http.MethodGet, "/users", app.listAllUsers)
	mux.HandlerFunc(http.MethodPost, "/users/create", app.createUser)

	return mux
}
