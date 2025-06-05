package ffav

import (
	"os"
	"os/signal"
	"testing"

	"github.com/centretown/avcamx"
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

func TestProxy(t *testing.T) {
	remoteAddr := "http://192.168.10.197:8080"
	host := avcamx.NewAvHost("", "9900")
	remote, err := host.FetchRemote(remoteAddr)
	if err != nil {
		t.Fatal(err)
	}
	host.MakeProxy(remote, nil)

	httpErr := make(chan error, 1)
	go func() {
		httpErr <- host.ListenAndServe()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-httpErr:
		t.Logf("failed to serve http: %v", err)
	case sig := <-sigs:
		t.Logf("terminating: %v", sig)
	}

	host.Quit()
}
