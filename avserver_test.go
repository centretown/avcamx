package avcamx

import (
	"fmt"
	"testing"
	"time"

	"github.com/korandiz/v4l"
)

type testListener struct {
}

func (t *testListener) StreamOn(id int) {
	fmt.Println(id, " is on")
}

func (t *testListener) StreamOff(id int) {
	fmt.Println(id, " is off")
}

func TestServerFind(t *testing.T) {
	webcams := FindLocalCams()
	for k, v := range webcams {
		t.Log(k, v)
	}
}

func TestServer(t *testing.T) {
	var info v4l.DeviceInfo

	webcam := NewLocalCam(&info)
	config := &VideoConfig{
		Codec:  "MJPG",
		Width:  1920,
		Height: 1080,
		FPS:    30,
	}

	server := NewAvServer(0, webcam, config, nil, nil)

	err := webcam.Open(config)
	if err != nil {
		t.Fatal(err)
	}

	if !webcam.isOpened {
		t.Fatal(fmt.Errorf("Not isOpen"))
	}

	go server.Serve()

	time.Sleep(1 * time.Second)
	server.quit <- 1

	time.Sleep(100 * time.Millisecond)
	if server.Source.IsOpened() {
		server.Source.Close()
		t.Fatal("source still open")
	}
}
