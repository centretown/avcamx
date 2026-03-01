package avcamx

import (
	"testing"
	"time"
)

func TestUdp(t *testing.T) {
	done := make(chan int)
	update := make(chan string)
	go PollUDP(done, update)
	go func() {
		for {
			select {
			case remoteAddr := <-update:
				t.Log((remoteAddr))
				continue
			default:
				time.Sleep(time.Second)
			}
		}
	}()
	time.Sleep(time.Second * 30)
	done <- 1
	time.Sleep(time.Second)

}
