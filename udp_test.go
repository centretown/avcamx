package avcamx

import (
	"testing"
	"time"
)

func TestUdp(t *testing.T) {
	go PollUDP()

	time.Sleep(time.Second)
	err := DialUDP("hello world")

	if err != nil {
		t.Fatal(err)
	}

	err = DialUDP("hello universe")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 30)
}
