package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// NewRouter returns a new http router
func NewRouter() *mux.Router {

	log.Debug("creating new router")
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {

		handler := route.HandlerFunc
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Handler(Log(handler))
	}
	return router
}
