/*
Copyright © 2022
Author Bhakiyaraj Kalimuthu
Email bhakiya.kalimuthu@gmail.com
*/

package internal

import (
	"bytes"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type HttpClient interface {
	Notify(msg string)
}

type httpClient struct {
	logger     *zap.Logger  // logger
	httpClient *http.Client // http client for sending notification
	url        string       // url where notification to be sent
}
type httpClient1 struct {
	logger     *zap.Logger // logger
	httpClient http.Client // http client for sending notification
	url        string      // url where notification to be sent
}

func NewHttpClient(logger *zap.Logger, url string) HttpClient {
	client := &http.Client{Timeout: time.Second * 5} // default timeout set to 5s
	return &httpClient{
		logger:     logger,
		httpClient: client,
		url:        url,
	}
}

func NewHttpClient1(logger *zap.Logger, url string) HttpClient {
	client := http.Client{Timeout: time.Second * 5} // default timeout set to 5s
	return &httpClient1{
		logger:     logger,
		httpClient: client,
		url:        url,
	}
}

// Notify using http client to make http post request
func (n *httpClient) Notify(msg string) {
	n.logger.Debug("making http request", zap.String("msg", msg))
	// create http request
	req, err := http.NewRequest(http.MethodPost, n.url, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
		return
	}
	// make post request
	_, err = n.httpClient.Do(req)
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
		return
	}
	n.logger.Info("successfully notified the message", zap.String("msg", msg))

}

// Notify using http client to make http post request
func (n *httpClient1) Notify(msg string) {
	n.logger.Debug("making http request", zap.String("msg", msg))
	// create http request
	req, err := http.NewRequest(http.MethodPost, n.url, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
		return
	}
	// make post request
	_, err = n.httpClient.Do(req)
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
		return
	}
	n.logger.Info("successfully notified the message", zap.String("msg", msg))

}
