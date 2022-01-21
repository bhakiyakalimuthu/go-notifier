package internal

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

func Test_notifier_Process(t *testing.T) {

	tests := map[string]struct {
		httpClientFunc func(*httpClientMock) HttpClient
		interval       time.Duration
		consumerChan   chan string
		data           []string
		want           int
	}{
		"Should successfully message should be notified(notify method called) when single message passed": {
			httpClientFunc: func(client *httpClientMock) HttpClient {
				client.On("Notify", mock.Anything).Return()
				return client
			},
			interval:     time.Nanosecond,
			consumerChan: make(chan string),
			data:         []string{"msg1"},
			want:         1,
		},
		"Should successfully message should be notified(notify method called) when multiple message passed": {
			httpClientFunc: func(client *httpClientMock) HttpClient {
				client.On("Notify", mock.Anything).Return()
				return client
			},
			interval:     time.Nanosecond,
			consumerChan: make(chan string),
			data:         []string{"msg1", "msg2", "msg3", "msg4", "msg5"},
			want:         5,
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			client := &httpClientMock{}
			n := &notifier{
				logger:       zap.NewNop(),
				interval:     testCase.interval,
				consumerChan: testCase.consumerChan,
			}
			if testCase.httpClientFunc != nil {
				n.httpClient = testCase.httpClientFunc(client)
			}
			wg := new(sync.WaitGroup)
			wg.Add(1)
			go func() {
				for _, i := range testCase.data {
					testCase.consumerChan <- i
				}
				close(testCase.consumerChan)
			}()
			go func() {
				n.Process(wg, 10)
			}()
			wg.Wait()
			client.AssertNumberOfCalls(t, "Notify", testCase.want)
		})
	}
}

func Test_notifier_Start(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	ctx3, cancel3 := context.WithCancel(context.Background())
	tests := map[string]struct {
		ctx          context.Context
		cancelFunc   context.CancelFunc
		producerChan chan string
		consumerChan chan string
		data         []string
		want         int
	}{
		"Should successfully msg received on consumer channel when single msg passed to producer channel": {
			ctx:          ctx1,
			cancelFunc:   cancel1,
			producerChan: make(chan string),
			consumerChan: make(chan string, 1),
			data:         []string{"msg1"},
			want:         1,
		},
		"Should successfully msg received on consumer channel when multiple msg passed to producer channel": {
			ctx:          ctx2,
			cancelFunc:   cancel2,
			producerChan: make(chan string, 1),
			consumerChan: make(chan string, 3),
			data:         []string{"msg1", "msg2", "msg3"},
			want:         3,
		},
		"Should not fail  when no msg passed to producer channel": {
			ctx:          ctx3,
			cancelFunc:   cancel3,
			producerChan: make(chan string, 1),
			consumerChan: make(chan string, 1),
			data:         []string{},
			want:         0,
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			n := &notifier{
				logger:       zap.NewNop(),
				interval:     1 * time.Nanosecond,
				producerChan: testCase.producerChan,
				consumerChan: testCase.consumerChan,
			}
			wg := new(sync.WaitGroup)
			wg.Add(1)
			go func() {
				defer wg.Done()
				n.Start(testCase.ctx)
			}()

			for _, i := range testCase.data {
				testCase.producerChan <- i
			}
			testCase.cancelFunc()
			wg.Wait()
			assert.Equal(t, testCase.want, len(testCase.consumerChan))
		})

	}
}
