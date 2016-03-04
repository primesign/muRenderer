package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"renderservice/doccache"
	"renderservice/renderer"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	minZoom = 0.05
	maxZoom = 4.0
)

// Default handels requests to /
func Default(w http.ResponseWriter, r *http.Request) {
	log.Debug("defaultHandler")
	w.Write([]byte("RenderService"))
}

// Health shows the service is ok
func Health(w http.ResponseWriter, r *http.Request) {
	log.Debug("healthHandler")
	w.Write([]byte("OK"))
}

// CloseDocument frees the resources held by the given document
func CloseDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	log.Debug("closeDocumentHandler: Working on close request for uuid ", uuid)
	cache := doccache.GetInstance()
	cache.DeleteDoc(uuid)
	w.WriteHeader(http.StatusNoContent)
}

func parseInt(w http.ResponseWriter, v string) (int, error) {
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Debug("Could not parse path variable", err)
		log.Warn("Could not parse int from path variable")
		http.Error(w, fmt.Sprintf("The request parameter %s can not be processed.", v), http.StatusBadRequest)
		return 0, errors.New("could not convert int")
	}
	return i, nil
}

func getRenderer(w http.ResponseWriter, uuid string) (renderer.Renderer, error) {
	cache := doccache.GetInstance()
	buf, err := cache.GetDoc(uuid)
	if err != nil {
		log.Debug("Could not fetch document ", err)
		log.Warn("Could not fetch document")
		http.Error(w, "Could not fetch the requested document.", http.StatusInternalServerError)
		return nil, err
	}
	return createRenderer(w, buf)
}

func createRenderer(w http.ResponseWriter, buf []byte) (renderer.Renderer, error) {

	// TODO: try to get error from NewRenderer if there are problem creating the renderer
	r := renderer.NewRenderer(buf)
	//	if err != nil {
	//		log.Debug("Could not create renderer", err)
	//		log.Warn("Could not create renderer for a document")
	//		http.Error(w, "Could not create renderer for the requested document.", http.StatusInternalServerError)
	//		return nil, err
	//	}
	return r, nil
}

// RenderPNG renders a document
func RenderPNG(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pageNr, err := parseInt(w, vars["pagenr"])
	if err != nil {
		return
	}

	zoom := r.FormValue("z")
	width := r.FormValue("w")
	height := r.FormValue("h")

	if zoom == "" && (width == "" || height == "") {
		log.Debug("No handler for request available.")
		http.Error(w, "No handler for request available.", http.StatusBadRequest)
		return
	}

	var ren renderer.Renderer
	if r.Method == "GET" {
		uuid := vars["uuid"]
		ren, err = getRenderer(w, uuid)
	} else if r.Method == "POST" {
		buf, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			log.Debug("Error reading POST request body", err)
			log.Warn("Error reading POST request body")
			http.Error(w, "Could not render the requested document.", http.StatusBadRequest)
			return
		}
		ren, err = createRenderer(w, buf)

	}
	if err != nil {
		return
	}
	defer ren.CloseRenderer()
	maxPage := ren.NumPages()
	if pageNr > maxPage {
		log.Debug(fmt.Sprintf("The requested page %d is bigger than max: %d", pageNr, maxPage))
		http.Error(w, "The requested page is not contained in this document.", http.StatusNotFound)
		return
	}

	var pngReader io.Reader
	if zoom != "" {
		pngReader = renderPNGZoom(w, ren, pageNr, zoom)
	} else if width != "" && height != "" {
		pngReader = renderPNGRectangle(w, ren, pageNr, width, height)
	}

	if pngReader == nil {
		return
	}
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, pngReader)
}

// renderPNGRectangle renders a document to fit in a given rectangle
func renderPNGRectangle(w http.ResponseWriter, ren renderer.Renderer, pageNr int, width string, height string) io.Reader {

	maxWidth, err := parseInt(w, width)
	if err != nil {
		return nil
	}

	maxHeight, err := parseInt(w, height)
	if err != nil {
		return nil
	}

	pngReader, err := ren.RenderPageForRectangle(pageNr, maxWidth, maxHeight)

	if err != nil {
		log.Debug("Error rendering image", err)
		log.Warn("Error rendering image")
		http.Error(w, "Could not render the requested document.", http.StatusBadRequest)
		return nil
	}

	return pngReader
}

// renderPNGZoom renders a document with a given zoom factor
func renderPNGZoom(w http.ResponseWriter, ren renderer.Renderer, pageNr int, z string) io.Reader {

	zoom, err := strconv.ParseFloat(z, 64)
	if err != nil {
		log.Debug("renderPngHandler: error converting zoomlevel for document ", err)
		http.Error(w, "The requested zoom level can not be processed.", http.StatusBadRequest)
		return nil
	}
	if zoom < minZoom || zoom > maxZoom {
		log.Debug("The requested zoom is not in the appropriate range.")
		http.Error(w, "The requested zoom is not in the appropriate range.", http.StatusBadRequest)
		return nil
	}

	pngReader, err := ren.RenderPageForZoom(pageNr, zoom)

	if err != nil {
		log.Debug("Error rendering image", err)
		log.Warn("Error rendering image")
		http.Error(w, "Could not render the requested document.", http.StatusBadRequest)
		return nil
	}

	return pngReader
}

// PageNumber states how many pages the given document has
func PageNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	log.Debug("pageNumberHandler: Working on page number request for uuid ", uuid)
	ren, err := getRenderer(w, uuid)
	if err != nil {
		return
	}
	defer ren.CloseRenderer()
	pageNrs := ren.NumPages()
	b, err := json.Marshal(pageNrs)
	if err != nil {
		log.Debug("pageNumberHandler: could not marshal number of pages ", err)
		log.Warn("Could not marshal response")
		http.Error(w, "Could not marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

// PageInfo returns information about a page
func PageInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	log.Debug("pageInfoHandler: Working on page info request for uuid ", uuid)

	pageNr, err := parseInt(w, vars["pagenr"])
	if err != nil {
		return
	}

	ren, err := getRenderer(w, uuid)
	if err != nil {
		return
	}
	defer ren.CloseRenderer()
	pi := ren.PageInfo(pageNr)
	b, err := json.Marshal(pi)
	if err != nil {
		log.Debug("pageInfoHandler: could not marshal page info ", err)
		log.Warn("Could not marshal response")
		http.Error(w, "Could not marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
