package internal

type Consumer interface {
	Start(msg string)
}
type consumer struct {
	callbackFunc func(msg string)
}

func NewConsumer(cb func(string)) Consumer {
	return &consumer{callbackFunc: cb}
}
func (c *consumer) Start(msg string) {
	c.callbackFunc(msg)
}
