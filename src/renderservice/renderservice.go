package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"renderservice/doccache"
	"runtime"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	log "github.com/Sirupsen/logrus"
)

// used in cache buffer calculation
const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
)

func setLogLvl(lvl string) {
	lrLvl, err := log.ParseLevel(lvl)
	if err != nil {
		fmt.Println("Unable to parse logging level:", lvl)
		os.Exit(1)
	}
	log.SetLevel(lrLvl)
}

func main() {

	var (
		_         = runtime.GOMAXPROCS(4)
		ip        = flag.String("ip", "localhost", "the ip the server should listen on")
		port      = flag.String("port", "8080", "the port the server should listen on")
		delay     = flag.Duration("shutdowndelay", 5*time.Second, "the delay for the graceful shutdown")
		cleanup   = flag.Duration("cleanupinterval", 3*time.Minute, "how often stored documents should be garbage collected")
		retention = flag.Duration("retention", 10*time.Minute, "how long documents should be cached in memory")
		loglvl    = flag.String("loglevel", "info", "supported levels: (debug|info|warn|error|fatal|panic)")
		docURL    = flag.String("docurl", "http://localhost:9000/", "the url to retrieve documents from")
		cacheSize = flag.Int("cacheSize", 500, "the cache size in megabyte")
	)

	flag.Parse()

	url, err := url.Parse(*docURL)
	if err != nil {
		fmt.Println("Unable to parse the document url:", docURL, "Error:", err)
		os.Exit(2)
	}

	cache := doccache.GetInstance()
	cache.SetURL(url)
	cache.SetCacheSize((*cacheSize) * MEGABYTE)

	setLogLvl(*loglvl)
	log.SetFormatter(&log.JSONFormatter{})

	log.Debug("schedule cleanup task")
	t := time.NewTicker(*cleanup)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-t.C:
				cache := doccache.GetInstance()
				cache.CleanUp(*retention)
			case <-quit:
				t.Stop()
				return
			}
		}
	}()

	router := NewRouter()
	fmt.Println("server listens on ", *ip, ":", *port, "Close with Ctrl-C and wait for graceful shutdown")

	srv := &graceful.Server{
		Timeout: *delay,
		Server: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", *ip, *port),
			Handler: router,
		},
	}

	// start http server
	srv.ListenAndServe()

	log.Debug("closing cleanup task")
	close(quit)
}
