/*
Copyright © 2022
Author Bhakiyaraj Kalimuthu
Email bhakiya.kalimuthu@gmail.com
*/

package internal

import (
	"context"
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
	httpClient   HttpClient    // http client for sending notification
	interval     time.Duration // interval in which notification to be sent
	producerChan chan string   // channel to receive from stdio
	consumerChan chan string   // chanel to consume the data

}

// NewNotifier constructor
func NewNotifier(logger *zap.Logger, httpClient HttpClient, interval time.Duration, producerChan, consumerChan chan string) Notifier {
	return &notifier{
		logger:       logger,
		interval:     interval,
		producerChan: producerChan,
		consumerChan: consumerChan,
		httpClient:   httpClient,
	}
}

// Process starts the worker process based on the number items in the consumer channel until it closes
func (n *notifier) Process(wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	for job := range n.consumerChan {
		<-time.After(n.interval) // wait for the provided interval
		n.logger.Debug("starting job", zap.Int("workerID", workerID))
		n.httpClient.Notify(job) // call http client to make notification
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
