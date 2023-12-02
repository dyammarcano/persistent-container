package routes

import "github.com/gorilla/mux"

type MuxRoute interface {
	AddRoute() *mux.Route
}
