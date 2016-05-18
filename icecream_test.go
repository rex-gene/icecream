package icecream

import (
	"testing"
	_ "time"
)

func TestIceCream(t *testing.T) {
	icecream, _ := New()
	err := icecream.Start(":3333")
	if err != nil {
		println(err)
	}

	/*for {
		select {
		case <-time.After(1000):
		}
	}*/
}
