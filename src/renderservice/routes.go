package main

import (
	"net/http"
	"renderservice/handler"
)

// Route respresents a http route
type Route struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

//Routes contains all routes for our services
type Routes []Route

var routes = Routes{
	Route{
		"GET",
		"/",
		handler.Default,
	},
	Route{
		"GET",
		"/health",
		handler.Health,
	},
	Route{
		"GET",
		"/renderservice/{uuid}/{pagenr:[1-9]+[0-9]*}/pageinfo",
		handler.PageInfo,
	},
	Route{
		"GET",
		"/renderservice/{uuid}/numpages",
		handler.PageNumber,
	},
	Route{
		"DELETE",
		"/renderservice/{uuid}",
		handler.CloseDocument,
	},
	Route{
		"GET",
		"/renderservice/{uuid}/{pagenr:[1-9]+[0-9]*}",
		handler.RenderPNG,
	},
	Route{
		"POST",
		"/renderservice/{pagenr:[1-9]+[0-9]*}",
		handler.RenderPNG,
	},
}
