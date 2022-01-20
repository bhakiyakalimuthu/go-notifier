package internal

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Notifier interface {
	Process(wg *sync.WaitGroup, workerID int)
	Start(ctx context.Context)
}

type notifier struct {
	logger       *zap.Logger
	url          string
	interval     time.Duration
	producerChan chan string
	consumerChan chan string
	httpClient   *http.Client
}

func NewNotifier(logger *zap.Logger, url string, interval time.Duration, producerChan, consumerChan chan string) Notifier {
	client := &http.Client{Timeout: time.Second * 10}
	return &notifier{
		logger:       logger,
		url:          url,
		interval:     interval,
		producerChan: producerChan,
		consumerChan: consumerChan,
		httpClient:   client,
	}
}

func (n *notifier) Process(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	for job := range n.consumerChan {
		<-time.After(n.interval)
		n.logger.Debug("starting job", zap.Int("workerID", workerID))
		n.makeRequest(job)
	}
	n.logger.Warn("gracefully finishing job", zap.Int("workerID", workerID))
}

func (n *notifier) Start(ctx context.Context) {
	for {
		select {
		case job := <-n.producerChan:
			n.logger.Debug("received msg from consumerChan")
			n.consumerChan <- job
		case <-ctx.Done():
			n.logger.Warn("received context cancellation......")
			close(n.consumerChan)
			return
		}
	}
}

func (n *notifier) makeRequest(msg string) {
	n.logger.Debug("making http request", zap.String("msg", msg))
	req, err := http.NewRequest(http.MethodPost, n.url, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
		return
	}
	_, err = n.httpClient.Do(req)
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
		return
	}
	n.logger.Info("successfully notified the message", zap.String("msg", msg))

}
