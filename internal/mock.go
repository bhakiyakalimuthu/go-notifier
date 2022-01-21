/*
Copyright © 2022
Author Bhakiyaraj Kalimuthu
Email bhakiya.kalimuthu@gmail.com
*/

package internal

import (
	"github.com/stretchr/testify/mock"
)

//type notifierMock struct {
//	mock.Mock
//}
//
//func (n *notifierMock) Process(wg *sync.WaitGroup, workerID int) { n.Called() }
//func (n *notifierMock) Start(ctx context.Context)                { n.Called() }

type httpClientMock struct {
	mock.Mock
}

func (n *httpClientMock) Notify(msg string) { n.Called() }
