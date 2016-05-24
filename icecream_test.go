package icecream

import (
	"testing"
	"time"
)

func TestIceCream(t *testing.T) {
	icecream, _ := New()
	err := icecream.Start(":3333")
	if err != nil {
		println(err)
	}

	client, _ := New()
	client.Connect("test", "127.0.0.1:3333")

	for {
		select {
		case <-time.After(1000):
		}
	}
}
