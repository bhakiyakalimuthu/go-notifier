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
	CallbackFunc(msg string)
	Process(wg *sync.WaitGroup, workerID int)
	Start(ctx context.Context)
}

type notifier struct {
	logger       *zap.Logger
	url          string
	consumerChan chan string
	jobsChan     chan string
	httpClient   *http.Client
}

func NewNotifier(logger *zap.Logger, url string, consumerChan, jobsChan chan string) Notifier {
	client := &http.Client{Timeout: time.Second * 10}
	return &notifier{
		logger:       logger,
		url:          url,
		consumerChan: consumerChan,
		jobsChan:     jobsChan,
		httpClient:   client,
	}
}

func (n *notifier) CallbackFunc(msg string) {
	n.consumerChan <- msg
}
func (n *notifier) Process(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	for job := range n.jobsChan {
		n.makeRequest(job)
	}
}

func (n *notifier) Start(ctx context.Context) {
	for {
		select {
		case job := <-n.consumerChan:
			n.jobsChan <- job
		case <-ctx.Done():
			n.logger.Info("Received cancellation")
			close(n.jobsChan)
			return
		}
	}
}

func (n *notifier) makeRequest(msg string) {

	req, err := http.NewRequest("POST", n.url, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
		return
	}
	_, err = n.httpClient.Do(req)
	if err != nil {
		n.logger.Error("failed to make new request", zap.Error(err))
	} else {
		n.logger.Info("successfully notified the message", zap.String("msg", msg))
	}
}
