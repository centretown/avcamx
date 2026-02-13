package ffav

import (
	"testing"
)

func TestStep(t *testing.T) {
	ch := make(chan int)
	ch2 := make(chan int)
	go EncodeStep("test_data/teststep.mp4", ch)
	go EncodeStep("test_data/teststep2.mp4", ch2)
	<-ch
	<-ch2
}

func TestDraw(t *testing.T) {
	ch := make(chan int)
	ch2 := make(chan int)
	go EncodeDraw("test_data/testdraw.mp4", ch)
	go EncodeDraw("test_data/testdraw2.mp4", ch2)
	<-ch
	<-ch2
}
