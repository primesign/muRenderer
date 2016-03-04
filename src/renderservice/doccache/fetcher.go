package doccache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

func fetchDocViaHTTP(url url.URL, uuid string) ([]byte, error) {

	url.Path += uuid
	log.Debug("DocFetcher: Fetching document from ", url.String())
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Debug(fmt.Sprintf("DocFetcher: Error while creating request for %s ", url), err)
		return nil, err
	}
	req.Close = true

	resp, err := http.DefaultClient.Do(req)
	if resp != nil {
		// avoid leak in case of redirection error where resp and err are non-nil
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		err := fmt.Errorf("Got http status code %d", resp.StatusCode)
		log.Info(fmt.Sprintf("DocFetcher: Error while downloading %s ", url), err)
		return nil, err
	}

	if err != nil {
		log.Debug(fmt.Sprintf("DocFetcher: Error while downloading %s ", url), err)
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debug("DocFetcher: Error while reading response into byte array", err)
		return nil, err
	}

	return data, nil
}
