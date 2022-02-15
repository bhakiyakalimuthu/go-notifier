package internal

import (
	"go.uber.org/zap"
	"testing"
)

func BenchmarkHttpClient_Notify(b *testing.B) {
	b.Run("benchHttpClient", func(b *testing.B) {
		n := NewHttpClient(zap.NewNop(), "https://httpbin.org/")
		benchtHttpClient(n)
	})
	b.Run("benchHttpClient1", func(b *testing.B) {
		n1 := NewHttpClient1(zap.NewNop(), "https://httpbin.org/")
		benchtHttpClient(n1)
	})
}

func benchtHttpClient(client HttpClient) {
	s := []string{"hello", "hi", "how are you", "beautiful"}
	for _, i := range s {
		client.Notify(i)
	}

}
