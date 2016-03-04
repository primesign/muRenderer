package main

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

//Log ads basic logging to the http.HandlerFunc
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		log.WithFields(log.Fields{
			"client":   r.RemoteAddr,
			"method":   r.Method,
			"url":      r.RequestURI,
			"duration": time.Since(start),
		}).Info("Handled Request")
	})
}
