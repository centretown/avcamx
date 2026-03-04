package avcamx

import (
	"testing"
	"time"
)

func TestUdp(t *testing.T) {
	done := make(chan int)
	update := make(chan string)
	host := NewAvHost("", "all", []string{}, 0, nil)
	go host.PollUDP(done, update)
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

func TestUDPAddr(t *testing.T) {
	t.Log(UDPAddress())
}
