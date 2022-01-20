package internal

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Notifier is the interface that groups the Start and Process methods
type Notifier interface {
	Process(wg *sync.WaitGroup, workerID int)
	Start(ctx context.Context)
}

// notifier type
type notifier struct {
	logger       *zap.Logger   // logger
	url          string        // url where notification to be sent
	interval     time.Duration // interval in which notification to be sent
	producerChan chan string   // channel to receive from stdio
	consumerChan chan string   // chanel to consume the data
	httpClient   *http.Client  // http client for sending notification
}

// NewNotifier constructor
func NewNotifier(logger *zap.Logger, url string, interval time.Duration, producerChan, consumerChan chan string) Notifier {
	client := &http.Client{Timeout: time.Second * 5} // default timeout set to 5s
	return &notifier{
		logger:       logger,
		url:          url,
		interval:     interval,
		producerChan: producerChan,
		consumerChan: consumerChan,
		httpClient:   client,
	}
}

// Process starts the worker process based on the number items in the consumer channel until it closes
func (n *notifier) Process(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	for job := range n.consumerChan {
		<-time.After(n.interval) // wait for the provided interval
		n.logger.Debug("starting job", zap.Int("workerID", workerID))
		n.notify(job) // call http client to make notification
	}
	n.logger.Warn("gracefully finishing job", zap.Int("workerID", workerID))
}

// Start acts as a proxy between producer and consumer channel,also supports the gracefull cancellation
func (n *notifier) Start(ctx context.Context) {
	for {
		select {
		case job := <-n.producerChan: // fetch job from producer
			n.logger.Debug("received msg from consumerChan")
			n.consumerChan <- job // pass job to consumer
		case <-ctx.Done():
			n.logger.Warn("received context cancellation......")
			close(n.consumerChan) // when context is done, close the consumer channel
			return
		}
	}
}

// notify using http client to make http post request
func (n *notifier) notify(msg string) {
	n.logger.Debug("making http request", zap.String("msg", msg))
	// create http request
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
