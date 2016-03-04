package doccache

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	l sync.Mutex

	docs *cache
	once sync.Once
)

type data struct {
	lastUsed time.Time
	buf      []byte
}

type cache struct {
	url *url.URL
	sync.RWMutex
	m       map[string]*data
	size    int64
	maxSize int64
}

//GetInstance returns a singelton instance of the cache
func GetInstance() *cache {
	once.Do(func() {
		docs = &cache{
			m: make(map[string]*data),
		}
	})
	return docs
}

func (d *data) getData() []byte {
	log.Debug("DocMap: get data")
	return d.buf
}

func (d *data) setLastUsed(time time.Time) {
	log.Debug("DocMap: set last used")
	d.lastUsed = time
}

func (d *data) getLastUsed() time.Time {
	log.Debug("DocMap: LastUsed get last used")
	return d.lastUsed
}

// SetURL sets the url which is used to query the backing data store
func (docs *cache) SetURL(url *url.URL) {
	docs.Lock()
	defer docs.Unlock()
	docs.url = url
	log.Debug("DocMap: Setting document cache url ", url.String())
}

//SetCacheSize sets the max size of the cache in byte
func (docs *cache) SetCacheSize(size int) {
	docs.Lock()
	defer docs.Unlock()
	docs.maxSize = int64(size)
	log.Debug("DocMap: Set document cache size to ", size, "byte")
}

//GetDoc returns the requested document
func (docs *cache) GetDoc(uuid string) ([]byte, error) {
	l.Lock()
	defer l.Unlock()
	log.Debug("DocMap: Get document with uuid ", uuid)
	docs.RLock()
	if val, ok := docs.m[uuid]; ok {
		log.Debug("DocMap: Document already in cache")
		val.setLastUsed(time.Now().UTC())
		buf := val.getData()
		docs.RUnlock()
		return buf, nil
	}
	url := *docs.url
	docs.RUnlock()
	buf, err := fetchDocViaHTTP(url, uuid)
	if err != nil {
		return nil, err
	}
	docs.persistDocInCache(uuid, buf)
	return buf, nil
}

func (docs *cache) persistDocInCache(uuid string, buf []byte) {
	log.Debug("DocMap: persist document in cache")
	docs.Lock()
	defer docs.Unlock()
	d := &data{
		lastUsed: time.Now().UTC(),
		buf:      buf}
	dataSize := int64(len(d.getData()))
	if docs.maxSize < (docs.size + dataSize) {
		log.Info("DocMap: cache reached max size, can NOT persist document in cache")
		log.Debug(fmt.Sprintf("DocMap: cachesize is %d / %d", docs.size, docs.maxSize))
		return
	}
	docs.size += dataSize
	log.Debug(fmt.Sprintf("DocMap: cachesize is %d / %d", docs.size, docs.maxSize))
	docs.m[uuid] = d
}

//DeleteDoc deletes the specified document
func (docs *cache) DeleteDoc(uuid string) {
	log.Debug("DocMap: Remove document with uuid ", uuid)
	docs.Lock()
	defer docs.Unlock()
	val, ok := docs.m[uuid]
	if ok {
		log.Debug(fmt.Sprintf("DocMap: cleanup Remove document %s from map", uuid))
		docs.size -= int64(len(val.getData()))
		delete(docs.m, uuid)
		log.Debug(fmt.Sprintf("DocMap: cachesize is %d / %d", docs.size, docs.maxSize))
	}
}

//CleanUp is used to delete documents which are older than the specified
//retention period
func (docs *cache) CleanUp(retention time.Duration) {
	docs.Lock()
	defer docs.Unlock()
	for uuid, val := range docs.m {
		now := time.Now().UTC()
		if now.Sub(val.getLastUsed()) > retention {
			log.Debug(fmt.Sprintf("DocMap: cleanup Remove document %s from map", uuid))
			docs.size -= int64(len(val.getData()))
			delete(docs.m, uuid)
			log.Debug(fmt.Sprintf("DocMap: cachesize is %d / %d", docs.size, docs.maxSize))
		}
	}
}
